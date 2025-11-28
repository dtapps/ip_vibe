package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/ent/device"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/cdn_helper"
	"vibe_ip/internal/helper/dns_helper"
	"vibe_ip/pkg/utils"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/utils/goarray"
	"golang.org/x/sync/errgroup"
)

// IP更新
type IpUpdateService struct {
	db     *ent.Client      // 数据库客户端
	cron   gocron.Scheduler // 定时任务调度器
	dnsJob gocron.Job       // DNS刷新任务
	cdnJob gocron.Job       // CDN刷新任务
	log    *slog.Logger     // 日志记录器
}

// NewIpUpdateService 创建 IP 更新服务
func NewIpUpdateService(db *ent.Client, cron gocron.Scheduler) *IpUpdateService {
	return &IpUpdateService{
		db:   db,
		cron: cron,
		log: slog.With(
			slog.String("component", "ip_update"),
		),
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

// 全局互斥锁
var globalIpUpdateMutex sync.Mutex

func (s *IpUpdateService) runDnsTask() {
	globalIpUpdateMutex.Lock()
	defer globalIpUpdateMutex.Unlock()

	ctx := context.Background()

	// 查询全部DNS记录
	allDnsRecord, err := global.GvaDb.DNSRecord.Query().
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

			for _, vRecord := range concurrentRecords {
				item := vRecord // 创建副本以避免闭包问题
				eg.Go(func() error {

					newLog := s.log.With(
						slog.Int64("id", item.ID),
						slog.String("record", "DNS"),
						slog.String("type", item.Type.String()),
						slog.String("domain", item.Domain),
					)

					if err = s.DnsUpdate(egCtx, newLog, item); err != nil {
						newLog.ErrorContext(egCtx, "并发处理失败",
							slog.Any("err", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				s.log.ErrorContext(ctx, "并发处理错误",
					slog.Any("err", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			newLog := s.log.With(
				slog.Int64("id", item.ID),
				slog.String("record", "DNS"),
				slog.String("type", item.Type.String()),
				slog.String("domain", item.Domain),
			)

			if err = s.DnsUpdate(ctx, newLog, item); err != nil {
				newLog.ErrorContext(ctx, "串行处理失败",
					slog.Any("err", err),
				)
			}

		}

	}

	nextRun, err := s.dnsJob.NextRun()
	if err == nil {
		s.log.InfoContext(ctx, "DNS更新成功",
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)

		// 判断CDN任务是否存在
		// if s.cdnJob == nil {
		// 	go s.StartCdnJob(context.Background())
		// }
	} else {
		s.log.InfoContext(ctx, "DNS更新成功")
	}

}

func (s *IpUpdateService) runCdnTask() {
	globalIpUpdateMutex.Lock()
	defer globalIpUpdateMutex.Unlock()

	ctx := context.Background()

	// 查询全部CDN记录
	allCdnRecord, err := global.GvaDb.CDNRecord.Query().
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

			for _, vRecord := range concurrentRecords {
				item := vRecord // 创建副本以避免闭包问题
				eg.Go(func() error {

					newLog := s.log.With(
						slog.Int64("id", item.ID),
						slog.String("record", "CDN"),
						slog.String("type", item.Type.String()),
						slog.String("domain", item.Domain),
					)

					if err = s.CdnUpdate(egCtx, newLog, item); err != nil {
						newLog.ErrorContext(egCtx, "并发处理失败",
							slog.Any("err", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				s.log.ErrorContext(ctx, "并发处理错误",
					slog.Any("err", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			newLog := s.log.With(
				slog.Int64("id", item.ID),
				slog.String("record", "CDN"),
				slog.String("type", item.Type.String()),
				slog.String("domain", item.Domain),
			)

			if err = s.CdnUpdate(ctx, newLog, item); err != nil {
				newLog.ErrorContext(ctx, "串行处理失败",
					slog.Any("err", err),
				)
			}

		}

	}

	nextRun, err := s.cdnJob.NextRun()
	if err == nil {
		s.log.InfoContext(ctx, "CDN更新成功",
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)

		// 判断DNS任务是否存在
		if s.dnsJob == nil {
			go s.StartDnsJob(context.Background())
		}
	} else {
		s.log.InfoContext(ctx, "CDN更新成功")
	}

}

// 更新CDN记录
// WithIPGroup().
func (s *IpUpdateService) CdnUpdate(ctx context.Context, newLog *slog.Logger, firstRecord *ent.CDNRecord) (err error) {

	// 准备源
	origins := make([]string, 0)

	// 查询IP组
	firstIPGroup, err := firstRecord.QueryIPGroup().
		First(ctx)
	if err != nil {
		return fmt.Errorf("找不到IP组: %w", err)
	}

	// 查询设备
	allDevice, err := firstIPGroup.QueryDevices().
		Order(ent.Asc(device.FieldID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("找不到设备: %w", err)
	}
	for _, item := range allDevice {
		if item.IPV4 != "" && firstRecord.Type == cdnrecord.TypeA {
			origins = append(origins, item.IPV4)
		}
		if item.IPV6 != "" && firstRecord.Type == cdnrecord.TypeAAAA {
			origins = append(origins, item.IPV6)
		}
	}

	// 去重源
	origins = goarray.RemoveDuplicateElement(origins)

	// 检查源是否有效
	if len(origins) == 0 {
		return fmt.Errorf("需要更新的源为空")
	}

	switch firstRecord.Edges.Config.Platform {
	case platformconfig.PlatformEdgeone:

		if len(origins) == 1 {
			// 处理单IP
			newLog.InfoContext(ctx, "处理单IP",
				slog.String("domain_name", firstRecord.Domain),
				slog.String("domain_origin", origins[0]),
			)
			// 清理不需要的端口
			switch firstRecord.OriginProtocol {
			case cdnrecord.OriginProtocolHTTP:
				firstRecord.HTTPSPort = 0
			case cdnrecord.OriginProtocolHTTPS:
				firstRecord.HTTPPort = 0
			}
			err = cdn_helper.EdgeOneProcessSingleIP(
				ctx,
				cdn_helper.ProcessService{
					Log:      newLog,
					Db:       global.GvaDb,
					RecordID: firstRecord.ID,
					Debug:    global.GvaEnv.DEBUG,
				},
				cdn_helper.EdgeOneProcessSingleIPParams{
					ConfigJSON:     firstRecord.Edges.Config.ConfigJSON,
					DomainName:     firstRecord.Domain,
					DomainOrigin:   origins[0],
					OriginProtocol: firstRecord.OriginProtocol.String(),
					HttpPort:       firstRecord.HTTPPort,
					HttpsPort:      firstRecord.HTTPSPort,
				},
			)
			if err != nil {
				return fmt.Errorf("处理单IP失败: %w", err)
			}
		} else {
			// 处理多IP
			newLog.InfoContext(ctx, "处理多IP",
				slog.String("domain_name", firstRecord.Domain),
				slog.Any("domain_origins", origins),
			)
			// 清理不需要的端口
			switch firstRecord.OriginProtocol {
			case cdnrecord.OriginProtocolHTTP:
				firstRecord.HTTPSPort = 0
			case cdnrecord.OriginProtocolHTTPS:
				firstRecord.HTTPPort = 0
			}
			err = cdn_helper.EdgeOneProcessMultiIP(
				ctx,
				cdn_helper.ProcessService{
					Log:      newLog,
					Db:       global.GvaDb,
					RecordID: firstRecord.ID,
					GroupID:  firstIPGroup.ID,
					Debug:    global.GvaEnv.DEBUG,
				},
				cdn_helper.EdgeOneProcessMultiIPParams{
					ConfigJSON:     firstRecord.Edges.Config.ConfigJSON,
					DomainName:     firstRecord.Domain,
					DomainOrigins:  origins,
					GroupName:      utils.GenerateShortIDBasedHash(firstIPGroup.ID, 10),
					OriginProtocol: firstRecord.OriginProtocol.String(),
					HttpPort:       firstRecord.HTTPPort,
					HttpsPort:      firstRecord.HTTPSPort,
				},
			)
			if err != nil {
				return fmt.Errorf("处理多IP失败: %w", err)
			}
		}

		return nil
	default:
		return fmt.Errorf("不支持类型: %s", firstRecord.Edges.Config.Platform)
	}

}

// 更新DNS录
// WithIPGroup().
// WithCdnRecord()
func (s *IpUpdateService) DnsUpdate(ctx context.Context, newLog *slog.Logger, firstRecord *ent.DNSRecord) (err error) {

	// 准备源
	origin := ""

	if firstRecord.Type == dnsrecord.TypeCNAME {

		// 查询CDN记录
		firstCdnRecord, err := firstRecord.QueryCdnRecord().
			First(ctx)
		if err != nil {
			return fmt.Errorf("找不到CDN记录: %w", err)
		}

		// 准备源
		origin = firstCdnRecord.CdnDomain

	} else {
		// 查询IP组
		firstIPGroup, err := firstRecord.QueryIPGroup().
			First(ctx)
		if err != nil {
			return fmt.Errorf("找不到IP组: %w", err)
		}

		// 查询设备
		allDevice, err := firstIPGroup.QueryDevices().
			Order(ent.Asc(device.FieldID)).
			All(ctx)
		if err != nil {
			return fmt.Errorf("找不到设备: %w", err)
		}
		for _, item := range allDevice {
			if item.IPV4 != "" && firstRecord.Type == dnsrecord.TypeA {
				origin = item.IPV4
			}
			if item.IPV6 != "" && firstRecord.Type == dnsrecord.TypeAAAA {
				origin = item.IPV6
			}
		}
	}

	// 检查源是否有效
	if origin == "" {
		return fmt.Errorf("需要更新的源为空")
	}

	switch firstRecord.Edges.Config.Platform {
	case platformconfig.PlatformCloudflare:

		err = dns_helper.CloudflareProcessSingleIP(
			ctx,
			dns_helper.ProcessService{
				Log:      newLog,
				Db:       global.GvaDb,
				RecordID: firstRecord.ID,
				Debug:    global.GvaEnv.DEBUG,
			},
			dns_helper.CloudflareProcessSingleIPParams{
				ConfigJSON:   firstRecord.Edges.Config.ConfigJSON,
				DomainName:   firstRecord.Domain,
				DomainOrigin: origin,
			},
		)
		if err != nil {
			return fmt.Errorf("处理单IP失败: %w", err)
		}

		return nil
	case platformconfig.PlatformDnspod:

		err = dns_helper.DnsPodProcessSingleIP(
			ctx,
			dns_helper.ProcessService{
				Log:      newLog,
				Db:       global.GvaDb,
				RecordID: firstRecord.ID,
				Debug:    global.GvaEnv.DEBUG,
			},
			dns_helper.DnsPodProcessSingleIPParams{
				ConfigJSON:   firstRecord.Edges.Config.ConfigJSON,
				DomainName:   firstRecord.Domain,
				DomainOrigin: origin,
			},
		)
		if err != nil {
			return fmt.Errorf("处理单IP失败: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("不支持类型: %s", firstRecord.Edges.Config.Platform)
	}

}
func (s *IpUpdateService) allowConcurrent(item *ent.PlatformConfig) bool {
	// EdgeOne 平台不支持并发，串行；其他并发
	if item.Platform == platformconfig.PlatformEdgeone {
		return false
	}
	return true
}
