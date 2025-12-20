package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/cdn_helper"
	"vibe_ip/internal/helper/dns_helper"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/goarray"
	"golang.org/x/sync/errgroup"
)

// IP更新
type IpUpdateService struct {
	db     *ent.Client      // 数据库客户端
	cron   gocron.Scheduler // 定时任务调度器
	dnsJob gocron.Job       // DNS刷新任务
	cdnJob gocron.Job       // CDN刷新任务
	logger *slog.Logger     // 日志记录器
}

// NewIpUpdateService 创建 IP 更新服务
func NewIpUpdateService(db *ent.Client, cron gocron.Scheduler, logger *slog.Logger) *IpUpdateService {
	return &IpUpdateService{
		db:     db,
		cron:   cron,
		logger: logger,
	}
}

// Close 关闭服务，释放资源
func (s *IpUpdateService) Close() {
	if s.cron != nil && s.dnsJob != nil {
		_ = s.cron.RemoveJob(s.dnsJob.ID())
	}
	if s.cron != nil && s.cdnJob != nil {
		_ = s.cron.RemoveJob(s.cdnJob.ID())
	}
}

// 启动 DNS刷新 任务，延迟 1 分钟添加，确保 CDN 先跑
func (s *IpUpdateService) StartDnsJob(ctx context.Context) (err error) {
	// 判断是否存在DNS任务
	if s.dnsJob != nil {
		return nil
	}
	// 延时1分钟
	time.Sleep(1 * time.Minute)
	// 每10分钟执行一次
	s.dnsJob, err = s.cron.NewJob(
		gocron.DurationJob(10*time.Minute),                   // 每10分钟执行一次
		gocron.NewTask(s.runDnsTask),                         // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("DNS刷新"),                             // 任务名称
		gocron.WithTags("dns_update"),                        // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

// 启动 CDN刷新 任务
func (s *IpUpdateService) StartCdnJob(ctx context.Context) (err error) {
	// 判断是否存在CDN任务
	if s.cdnJob != nil {
		return nil
	}
	// 延时1分钟
	// time.Sleep(1 * time.Minute)
	// 每10分钟执行一次
	s.cdnJob, err = s.cron.NewJob(
		gocron.DurationJob(10*time.Minute),                   // 每10分钟执行一次
		gocron.NewTask(s.runCdnTask),                         // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("CDN刷新"),                             // 任务名称
		gocron.WithTags("cdn_update"),                        // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *IpUpdateService) runDnsTask() {

	ctx := context.Background()
	processingType := "DNS"

	s.logger.InfoContext(ctx, "开始执行",
		slog.String("处理类型", processingType),
	)

	// 查询全部DNS记录
	allDnsRecord, err := global.GvaEntDb.DNSRecord.Query().
		Where(dnsrecord.IsActiveEQ(true)).
		WithConfig().
		All(ctx)
	if err == nil {

		var concurrentRecords, serialRecords []*ent.DNSRecord

		// 按策略分组
		for _, item := range allDnsRecord {
			if s.allowConcurrent(item.Edges.Config) {
				concurrentRecords = append(concurrentRecords, item)
			} else {
				serialRecords = append(serialRecords, item)
			}
		}

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					newLog := s.logger.With(
						slog.String("处理方式", "并发"),
						slog.String("处理类型", processingType),
						slog.Int64("记录", item.ID),
						slog.String("类型", item.Type.String()),
						slog.String("子域名", item.SubDomain),
						slog.String("平台", item.Edges.Config.Platform.String()),
					)

					if err = s.DnsUpdate(egCtx, newLog, item); err != nil {
						newLog.ErrorContext(egCtx, "处理失败",
							slog.Any("error", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				s.logger.ErrorContext(ctx, "并发处理错误",
					slog.String("处理类型", processingType),
					slog.Any("error", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			newLog := s.logger.With(
				slog.String("处理方式", "串行"),
				slog.String("处理类型", processingType),
				slog.Int64("记录", item.ID),
				slog.String("类型", item.Type.String()),
				slog.String("子域名", item.SubDomain),
				slog.String("平台", item.Edges.Config.Platform.String()),
			)

			if err = s.DnsUpdate(ctx, newLog, item); err != nil {
				newLog.ErrorContext(ctx, "处理失败",
					slog.Any("error", err),
				)
			}

		}

	}

	nextRun, err := s.dnsJob.NextRun()
	if err == nil {
		s.logger.InfoContext(ctx, "执行完成",
			slog.String("处理类型", processingType),
			slog.String("下次运行时间", nextRun.Format(time.DateTime)),
		)

		// 判断CDN任务是否存在
		// if s.cdnJob == nil {
		// 	go s.StartCdnJob(context.Background())
		// }
	} else {
		s.logger.InfoContext(ctx, "执行完成",
			slog.String("处理类型", processingType),
		)
	}

}

func (s *IpUpdateService) runCdnTask() {

	ctx := context.Background()
	processingType := "CDN"

	s.logger.InfoContext(ctx, "开始执行",
		slog.String("处理类型", processingType),
	)

	// 查询全部CDN记录
	allCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		Where(cdnrecord.IsActiveEQ(true)).
		WithConfig().
		All(ctx)
	if err == nil {

		var concurrentRecords, serialRecords []*ent.CDNRecord

		// 按策略分组
		for _, item := range allCdnRecord {
			if s.allowConcurrent(item.Edges.Config) {
				concurrentRecords = append(concurrentRecords, item)
			} else {
				serialRecords = append(serialRecords, item)
			}
		}

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					newLog := s.logger.With(
						slog.String("处理方式", "并发"),
						slog.String("处理类型", processingType),
						slog.Int64("记录", item.ID),
						slog.String("类型", item.Type.String()),
						slog.String("子域名", item.SubDomain),
						slog.String("平台", item.Edges.Config.Platform.String()),
					)

					if err = s.CdnUpdate(egCtx, newLog, item); err != nil {
						newLog.ErrorContext(egCtx, "处理失败",
							slog.Any("error", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				s.logger.ErrorContext(ctx, "并发处理错误",
					slog.String("处理类型", processingType),
					slog.Any("error", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			newLog := s.logger.With(
				slog.String("处理方式", "串行"),
				slog.String("处理类型", processingType),
				slog.Int64("记录", item.ID),
				slog.String("类型", item.Type.String()),
				slog.String("子域名", item.SubDomain),
				slog.String("平台", item.Edges.Config.Platform.String()),
			)

			if err = s.CdnUpdate(ctx, newLog, item); err != nil {
				newLog.ErrorContext(ctx, "处理失败",
					slog.Any("error", err),
				)
			}

		}

	}

	nextRun, err := s.cdnJob.NextRun()
	if err == nil {
		s.logger.InfoContext(ctx, "执行完成",
			slog.String("处理类型", processingType),
			slog.String("下次运行时间", nextRun.Format(time.DateTime)),
		)

		// 判断DNS任务是否存在
		if s.dnsJob == nil {
			go s.StartDnsJob(context.Background())
		}
	} else {
		s.logger.InfoContext(ctx, "执行完成",
			slog.String("处理类型", processingType),
		)
	}

}

// 更新CDN记录
// WithIPGroup().
func (s *IpUpdateService) CdnUpdate(ctx context.Context, newLog *slog.Logger, firstRecord *ent.CDNRecord) (err error) {

	// 准备参数
	processParams := cdn_helper.CDNProcessParams{
		CDNRecordID: cast.ToString(firstRecord.ID),

		ConfigJSON:     firstRecord.Edges.Config.ConfigJSON,
		SubDomain:      firstRecord.SubDomain,
		OriginProtocol: strings.ToUpper(firstRecord.OriginProtocol.String()),
		HttpPort:       firstRecord.HTTPPort,
		HttpsPort:      firstRecord.HTTPSPort,
	}

	// 查询IP组
	firstIPGroup, err := firstRecord.QueryIPGroup().
		First(ctx)
	if err != nil {
		return fmt.Errorf("找不到IP组: %w", err)
	}
	processParams.GroupRecordID = cast.ToString(firstIPGroup.ID)

	// 查询设备
	allDevice, err := firstIPGroup.QueryDevices().
		Order(ent.Asc(device.FieldID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("找不到设备: %w", err)
	}
	for _, item := range allDevice {
		if item.IPV4 != "" && firstRecord.Type == cdnrecord.TypeA {
			processParams.OriginValues = append(processParams.OriginValues, item.IPV4)
		}
		if item.IPV6 != "" && firstRecord.Type == cdnrecord.TypeAAAA {
			processParams.OriginValues = append(processParams.OriginValues, item.IPV6)
		}
	}

	// 去重源并检查源是否有效
	processParams.OriginValues = goarray.RemoveDuplicateElement(processParams.OriginValues)
	if len(processParams.OriginValues) == 0 {
		return fmt.Errorf("需要更新的源为空")
	}

	var cdnProcess cdn_helper.CDNProcess

	switch firstRecord.Edges.Config.Platform {
	case platformconfig.PlatformTencentEo:

		cdnProcess = cdn_helper.NewCDNProcessTencentEO(
			global.GvaEnv.DEBUG,
			newLog,
			global.GvaEntDb,
			global.GvaHttpLog,
		)

	case platformconfig.PlatformAliyunEsa:

		cdnProcess = cdn_helper.NewCDNProcessAliyunEsa(
			global.GvaEnv.DEBUG,
			newLog,
			global.GvaEntDb,
			global.GvaHttpLog,
		)

	default:
		err = fmt.Errorf("暂不支持该平台: %s", firstRecord.Edges.Config.Platform)
		_ = s.updateDbOfCdnResponse(ctx, firstRecord.ID, err.Error())
		return err
	}

	// 更新CDN记录
	err = cdnProcess.UpsertRecord(ctx, &processParams)
	if err != nil {
		_ = s.updateDbOfCdnResponse(ctx, firstRecord.ID, err.Error())
		return err
	}
	_ = s.updateDbOfCdnResponse(ctx, firstRecord.ID, "ok")
	return nil
}

// 更新DNS录
// WithIPGroup().
// WithCdnRecord()
func (s *IpUpdateService) DnsUpdate(ctx context.Context, newLog *slog.Logger, firstRecord *ent.DNSRecord) (err error) {

	// 准备参数
	processParams := dns_helper.DNSProcessParams{
		DNSRecordID: cast.ToString(firstRecord.ID),

		ConfigJSON: firstRecord.Edges.Config.ConfigJSON,
		SubDomain:  firstRecord.SubDomain,
	}

	if firstRecord.Type == dnsrecord.TypeCNAME {

		// 查询CDN记录
		firstCdnRecord, err := firstRecord.QueryCdnRecord().
			First(ctx)
		if err != nil {
			return fmt.Errorf("找不到CDN记录: %w", err)
		}

		processParams.OriginValue = firstCdnRecord.CdnDomain

	} else {
		// 查询IP组
		firstIPGroup, err := firstRecord.QueryIPGroup().
			First(ctx)
		if err != nil {
			return fmt.Errorf("找不到IP组: %w", err)
		}

		// 查询设备 (心跳时间在 12 分钟内)
		allDevice, err := firstIPGroup.QueryDevices().
			Where(device.LastHeartbeatAtGTE(time.Now().Add(-12 * time.Minute))).
			Order(ent.Asc(device.FieldID)).
			All(ctx)
		if err != nil {
			return fmt.Errorf("找不到设备: %w", err)
		}
		for _, item := range allDevice {
			if item.IPV4 != "" && firstRecord.Type == dnsrecord.TypeA {
				processParams.OriginValue = item.IPV4
			}
			if item.IPV6 != "" && firstRecord.Type == dnsrecord.TypeAAAA {
				processParams.OriginValue = item.IPV6
			}
		}
	}

	// 检查源是否有效
	if processParams.OriginValue == "" {
		return fmt.Errorf("需要更新的源为空")
	}

	var dnsProcess dns_helper.DNSProcess

	switch firstRecord.Edges.Config.Platform {
	case platformconfig.PlatformCloudflare:

		dnsProcess = dns_helper.NewDNSProcessCloudflare(
			global.GvaEnv.DEBUG,
			newLog,
			global.GvaEntDb,
			global.GvaHttpLog,
		)

	case platformconfig.PlatformTencentDNS:

		dnsProcess = dns_helper.NewDNSProcessTencentDns(
			global.GvaEnv.DEBUG,
			newLog,
			global.GvaEntDb,
			global.GvaHttpLog,
		)

	case platformconfig.PlatformAliyunDNS:

		dnsProcess = dns_helper.NewDNSProcessAliyunDns(
			global.GvaEnv.DEBUG,
			newLog,
			global.GvaEntDb,
			global.GvaHttpLog,
		)

	case platformconfig.PlatformBaiduDNS:

		dnsProcess = dns_helper.NewDNSProcessBaiduDns(
			global.GvaEnv.DEBUG,
			newLog,
			global.GvaEntDb,
			global.GvaHttpLog,
		)

	default:
		err = fmt.Errorf("暂不支持该平台: %s", firstRecord.Edges.Config.Platform)
		_ = s.updateDbOfDnsResponse(ctx, firstRecord.ID, err.Error())
		return err
	}

	// 更新DNS记录
	err = dnsProcess.UpsertRecord(ctx, &processParams)
	if err != nil {
		_ = s.updateDbOfDnsResponse(ctx, firstRecord.ID, err.Error())
		return err
	}
	_ = s.updateDbOfDnsResponse(ctx, firstRecord.ID, "ok")
	return nil
}

func (s *IpUpdateService) allowConcurrent(item *ent.PlatformConfig) bool {
	// EdgeOne 平台不支持并发，串行；其他并发
	if item.Platform == platformconfig.PlatformTencentEo {
		return false
	}
	return true
}

// 更新CDN记录的LastAPIResponse
func (s *IpUpdateService) updateDbOfCdnResponse(ctx context.Context, id int64, last_api_response string) error {
	if s.db != nil && id != 0 {
		return s.db.CDNRecord.UpdateOneID(id).
			SetLastAPIResponse(last_api_response).
			SetLastAPIResponseAt(time.Now()).
			Exec(ctx)
	}
	return nil
}

// 更新DNS记录的LastAPIResponse
func (s *IpUpdateService) updateDbOfDnsResponse(ctx context.Context, id int64, last_api_response string) error {
	if s.db != nil && id != 0 {
		return s.db.DNSRecord.UpdateOneID(id).
			SetLastAPIResponse(last_api_response).
			SetLastAPIResponseAt(time.Now()).
			Exec(ctx)
	}
	return nil
}
