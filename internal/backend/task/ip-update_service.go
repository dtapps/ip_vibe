package task

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
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/cdn_helper"
	"vibe_ip/internal/helper/dns_helper"
	"vibe_ip/internal/helper/gocron_helper"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/goarray"
	"go.dtapp.net/library/utils/gomd5"
	"go.dtapp.net/library/utils/gorandom"
	"golang.org/x/sync/errgroup"
)

// IP更新任务
type IPUpdateTask struct {
	db            *ent.Client                // 数据库客户端
	logDb         *ent_log.Client            // 日志数据库客户端
	cron          gocron.Scheduler           // 定时任务调度器
	dnsJob        gocron.Job                 // DNS刷新任务
	cdnJob        gocron.Job                 // CDN刷新任务
	jobHelper     *gocron_helper.JobHelper   // 任务助手
	logger        *slog.Logger               // 日志记录器
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 是否开启调试模式
}

// 创建实例
func NewIPUpdateTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingTransport, debug bool) *IPUpdateTask {
	return &IPUpdateTask{
		db:        db,
		logDb:     logDb,
		cron:      cron,
		jobHelper: gocron_helper.NewJobHelper(cron, logger),
		logger: logger.With(
			slog.String("任务", "IP更新"),
		),
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// Close 关闭服务，释放资源
func (s *IPUpdateTask) Close() {
	if s.cron != nil && s.dnsJob != nil {
		_ = s.cron.RemoveJob(s.dnsJob.ID())
	}
	if s.cron != nil && s.cdnJob != nil {
		_ = s.cron.RemoveJob(s.cdnJob.ID())
	}
}

// 自定义 logger
func (s *IPUpdateTask) setLogger(logger *slog.Logger) *IPUpdateTask {
	s.logger = logger
	return s
}

// 启动DNS任务，延迟 1 分钟添加，确保CDN先跑
func (s *IPUpdateTask) StartDnsJob(ctx context.Context) (err error) {
	// 判断是否存在DNS任务
	if s.dnsJob != nil {
		return nil
	}
	// 延时1分钟
	time.Sleep(1 * time.Minute)
	// 创建周期任务
	s.dnsJob, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			10*time.Minute, // 最小间隔 10 分钟
			12*time.Minute, // 最大间隔 12 分钟
		),
		gocron.NewTask(s.RunDnsTask),                         // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("DNS记录刷新"),                           // 任务名称
		gocron.WithEventListeners(
			s.jobHelper.PanicPeriodicListener(),
			s.jobHelper.BeforePeriodicListener(),
			s.jobHelper.AfterPeriodicListener(),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

// 启动CDN任务
func (s *IPUpdateTask) StartCdnJob(ctx context.Context) (err error) {
	// 判断是否存在CDN任务
	if s.cdnJob != nil {
		return nil
	}
	// 延时1分钟
	// time.Sleep(1 * time.Minute)
	// 创建周期任务
	s.cdnJob, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			10*time.Minute, // 最小间隔 10 分钟
			12*time.Minute, // 最大间隔 12 分钟
		),
		gocron.NewTask(s.RunCdnTask),                         // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("CDN记录刷新"),                           // 任务名称
		gocron.WithEventListeners(
			s.jobHelper.PanicPeriodicListener(),
			s.jobHelper.BeforePeriodicListener(),
			s.jobHelper.AfterPeriodicListener(),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

// 运行DNS任务
func (s *IPUpdateTask) RunDnsTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger.With(
		slog.String("类型", "DNS"),
	), s.dnsJob.ID().String(), s.dnsJob.Name())
	logger.InfoContext(ctx, "开始执行")

	// 分组记录
	var concurrentRecords, serialRecords []*PendingDnsRecord

	// 查询全部DNS记录
	allDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IsActiveEQ(true)).
		// 预加载 配置
		WithConfig().
		// 预加载 CDN记录
		WithCdnRecord().
		// 预加载 IP组
		WithIPGroup(func(withIPGroup *ent.IpGroupQuery) {
			// 在 IP组 里查询设备
			withIPGroup.WithDevices(func(withDevices *ent.DeviceQuery) {
				// 设备心跳时间在 12 分钟内
				withDevices.Where(device.LastHeartbeatAtGTE(time.Now().Add(-12 * time.Minute)))
				withDevices.OrderByIDAsc()
			})
		}).
		All(ctx)
	if err == nil {

		// 按策略分组
		for _, dnsRecord := range allDnsRecord {

			// 创建待处理记录
			pendingRecord := &PendingDnsRecord{
				item:     dnsRecord,
				platform: dnsRecord.Edges.Config.Platform,
				params: &dns_helper.DNSProcessParams{
					DNSRecordID: cast.ToString(dnsRecord.ID),

					ConfigJSON: dnsRecord.Edges.Config.ConfigJSON,
					SubDomain:  dnsRecord.SubDomain,
				},
			}

			if dnsRecord.Type == dnsrecord.TypeCNAME {

				// 获取关联的 CNAME 记录
				cdnRecord := dnsRecord.Edges.CdnRecord
				if cdnRecord == nil {
					// err = fmt.Errorf("记录 %d 没有关联的 CNAME 记录", dnsRecord.ID)
					// _ = s.updateDbOfDnsResponse(ctx, dnsRecord.ID, err.Error())
					continue
				}
				pendingRecord.params.OriginValue = cdnRecord.CdnDomain

			} else {

				// 获取关联的 IPGroup
				ipGroup := dnsRecord.Edges.IPGroup
				if ipGroup == nil {
					// err = fmt.Errorf("记录 %d 没有关联的 IP组", dnsRecord.ID)
					// _ = s.updateDbOfDnsResponse(ctx, dnsRecord.ID, err.Error())
					continue
				}
				pendingRecord.params.GroupRecordID = cast.ToString(ipGroup.ID)

				// 获取关联的设备列表
				devices := ipGroup.Edges.Devices
				if len(devices) == 0 {
					// err = fmt.Errorf("记录 %d 没有可用设备", ipGroup.ID)
					// _ = s.updateDbOfDnsResponse(ctx, dnsRecord.ID, err.Error())
					continue
				}

				// 收集该记录的源IP
				for _, device := range devices {
					if device.IPV4 != "" && dnsRecord.Type == dnsrecord.TypeA {
						pendingRecord.params.OriginValue = device.IPV4
					}
					if device.IPV6 != "" && dnsRecord.Type == dnsrecord.TypeAAAA {
						pendingRecord.params.OriginValue = device.IPV6
					}
				}

			}

			if pendingRecord.params.OriginValue == "" {
				// err = fmt.Errorf("记录 %d 需要更新的源为空", dnsRecord.ID)
				// _ = s.updateDbOfDnsResponse(ctx, dnsRecord.ID, err.Error())
				continue
			}

			// 按策略分组
			if allowConcurrent(dnsRecord.Edges.Config) {
				concurrentRecords = append(concurrentRecords, pendingRecord)
			} else {
				serialRecords = append(serialRecords, pendingRecord)
			}
		}

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					sonLogger := logger.With(
						slog.String("处理方式", "并发"),
					)

					if err = s.DnsUpdate(egCtx, sonLogger, item); err != nil {
						logger.ErrorContext(egCtx, "处理失败",
							slog.Any("error", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				logger.ErrorContext(ctx, "并发处理错误",
					slog.Any("error", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			sonLogger := logger.With(
				slog.String("处理方式", "串行"),
			)

			if err = s.DnsUpdate(ctx, sonLogger, item); err != nil {
				logger.ErrorContext(ctx, "处理失败",
					slog.Any("error", err),
				)
			}

		}

	}

	nextRun, err := s.dnsJob.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
			slog.String("下次运行时间", nextRun.Format(time.DateTime)),
		)
		// 判断CDN任务是否存在
		// if s.cdnJob == nil {
		// 	go s.StartCdnJob(context.Background())
		// }
	} else {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
		)
	}
}

// 运行CDN任务
func (s *IPUpdateTask) RunCdnTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger.With(
		slog.String("类型", "CDN"),
	), s.cdnJob.ID().String(), s.cdnJob.Name())
	logger.InfoContext(ctx, "开始执行")

	// 分组记录
	var concurrentRecords, serialRecords []*PendingCdnRecord

	// 查询全部CDN记录
	allCdnRecord, err := s.db.CDNRecord.Query().
		Where(cdnrecord.IsActiveEQ(true)).
		// 预加载 配置
		WithConfig().
		// 预加载 IP组
		WithIPGroup(func(withIPGroup *ent.IpGroupQuery) {
			// 在 IP组 里查询设备
			withIPGroup.WithDevices(func(withDevices *ent.DeviceQuery) {
				// 设备心跳时间在 12 分钟内
				withDevices.Where(device.LastHeartbeatAtGTE(time.Now().Add(-12 * time.Minute)))
				withDevices.OrderByIDAsc()
			})
		}).
		All(ctx)
	if err == nil {

		// 按策略分组
		for _, cdnRecord := range allCdnRecord {

			// 创建待处理记录
			pendingRecord := &PendingCdnRecord{
				item:     cdnRecord,
				platform: cdnRecord.Edges.Config.Platform,
				params: &cdn_helper.CDNProcessParams{
					CDNRecordID: cast.ToString(cdnRecord.ID),

					ConfigJSON:     cdnRecord.Edges.Config.ConfigJSON,
					SubDomain:      cdnRecord.SubDomain,
					OriginProtocol: strings.ToUpper(cdnRecord.OriginProtocol.String()),
					HttpPort:       cdnRecord.HTTPPort,
					HttpsPort:      cdnRecord.HTTPSPort,
				},
			}

			// 获取关联的 IPGroup
			ipGroup := cdnRecord.Edges.IPGroup
			if ipGroup == nil {
				// err = fmt.Errorf("记录 %d 没有关联的 IP组", cdnRecord.ID)
				// _ = s.updateDbOfCdnResponse(ctx, cdnRecord.ID, err.Error())
				continue
			}
			pendingRecord.params.GroupRecordID = cast.ToString(ipGroup.ID)
			pendingRecord.params.GroupName = gomd5.Get16MD5Encode(pendingRecord.params.GroupRecordID)

			// 获取关联的设备列表
			devices := ipGroup.Edges.Devices
			if len(devices) == 0 {
				// err = fmt.Errorf("记录 %d 没有可用设备", cdnRecord.ID)
				// _ = s.updateDbOfCdnResponse(ctx, cdnRecord.ID, err.Error())
				continue
			}

			// 收集该记录的源IP
			for _, device := range devices {
				if device.IPV4 != "" && cdnRecord.Type == cdnrecord.TypeA {
					pendingRecord.params.OriginValues = append(pendingRecord.params.OriginValues, device.IPV4)
				}
				if device.IPV6 != "" && cdnRecord.Type == cdnrecord.TypeAAAA {
					pendingRecord.params.OriginValues = append(pendingRecord.params.OriginValues, device.IPV6)
				}
			}

			// 去重
			pendingRecord.params.OriginValues = goarray.RemoveDuplicateElement(pendingRecord.params.OriginValues)
			if len(pendingRecord.params.OriginValues) == 0 {
				// err = fmt.Errorf("记录 %d 需要更新的源为空", cdnRecord.ID)
				// _ = s.updateDbOfCdnResponse(ctx, cdnRecord.ID, err.Error())
				continue
			}

			// 按策略分组
			if allowConcurrent(cdnRecord.Edges.Config) {
				concurrentRecords = append(concurrentRecords, pendingRecord)
			} else {
				serialRecords = append(serialRecords, pendingRecord)
			}
		}

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					sonLogger := logger.With(
						slog.String("处理方式", "并发"),
					)

					if err = s.CdnUpdate(egCtx, sonLogger, item); err != nil {
						logger.ErrorContext(egCtx, "处理失败",
							slog.Any("error", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				logger.ErrorContext(ctx, "并发处理错误",
					slog.Any("error", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			sonLogger := logger.With(
				slog.String("处理方式", "串行"),
			)

			if err = s.CdnUpdate(ctx, sonLogger, item); err != nil {
				logger.ErrorContext(ctx, "处理失败",
					slog.Any("error", err),
				)
			}

		}

	}

	nextRun, err := s.cdnJob.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
			slog.String("下次运行时间", nextRun.Format(time.DateTime)),
		)
		// 判断DNS任务是否存在
		if s.dnsJob == nil {
			go s.StartDnsJob(context.Background())
		}
	} else {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
		)
	}
}

// 待处理
type PendingCdnRecord struct {
	item     *ent.CDNRecord               // CDN记录
	platform platformconfig.Platform      // 平台
	params   *cdn_helper.CDNProcessParams // 处理参数
}

// 更新 CDN 记录
func (s *IPUpdateTask) CdnUpdate(ctx context.Context, logger *slog.Logger, pendingRecord *PendingCdnRecord) (err error) {

	// 日志
	logger = logger.With(
		slog.String("处理编号", gorandom.Alphanumeric(10)),
		slog.String("记录", pendingRecord.params.CDNRecordID),
		slog.String("提供商", pendingRecord.item.Type.String()),
		slog.String("平台", pendingRecord.platform.String()),
		slog.String("子域名", pendingRecord.item.SubDomain),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	var cdnProcess cdn_helper.CDNProcess

	switch pendingRecord.platform {
	case platformconfig.PlatformTencentEo:
		cdnProcess = cdn_helper.NewCDNProcessTencentEO(
			s.db,
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformAliyunEsa:
		cdnProcess = cdn_helper.NewCDNProcessAliyunEsa(
			s.db,
			logger,
			s.httpTransport,
			s.debug,
		)
	default:
		err = fmt.Errorf("暂不支持该平台: %s", pendingRecord.platform.String())
		_ = s.updateDbOfCdnResponse(ctx, pendingRecord.item.ID, err.Error())
		return err
	}

	// 更新CDN记录
	err = cdnProcess.UpsertRecord(ctx, pendingRecord.params)
	if err != nil {
		_ = s.updateDbOfCdnResponse(ctx, pendingRecord.item.ID, err.Error())
		return err
	}
	_ = s.updateDbOfCdnResponse(ctx, pendingRecord.item.ID, "ok")
	return nil
}

// 待处理
type PendingDnsRecord struct {
	item     *ent.DNSRecord               // DNS记录
	platform platformconfig.Platform      // 平台
	params   *dns_helper.DNSProcessParams // 处理参数
}

// 更新DNS记录
func (s *IPUpdateTask) DnsUpdate(ctx context.Context, logger *slog.Logger, pendingRecord *PendingDnsRecord) (err error) {

	// 日志
	logger = logger.With(
		slog.String("处理编号", gorandom.Alphanumeric(10)),
		slog.String("记录", pendingRecord.params.DNSRecordID),
		slog.String("提供商", pendingRecord.item.Type.String()),
		slog.String("平台", pendingRecord.platform.String()),
		slog.String("子域名", pendingRecord.item.SubDomain),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	var dnsProcess dns_helper.DNSProcess

	switch pendingRecord.platform {
	case platformconfig.PlatformCloudflare:
		dnsProcess = dns_helper.NewDNSProcessCloudflare(
			s.db,
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformTencentDNS:
		dnsProcess = dns_helper.NewDNSProcessTencentDns(
			s.db,
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformAliyunDNS:
		dnsProcess = dns_helper.NewDNSProcessAliyunDns(
			s.db,
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformBaiduDNS:
		dnsProcess = dns_helper.NewDNSProcessBaiduDns(
			s.db,
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformHuaweiDNS:
		dnsProcess = dns_helper.NewDNSProcessHuaweiDns(
			s.db,
			logger,
			s.httpTransport,
			s.debug,
		)
	default:
		err = fmt.Errorf("暂不支持该平台: %s", pendingRecord.platform.String())
		_ = s.updateDbOfDnsResponse(ctx, pendingRecord.item.ID, err.Error())
		return err
	}

	// 更新DNS记录
	err = dnsProcess.UpsertRecord(ctx, pendingRecord.params)
	if err != nil {
		_ = s.updateDbOfDnsResponse(ctx, pendingRecord.item.ID, err.Error())
		return err
	}
	_ = s.updateDbOfDnsResponse(ctx, pendingRecord.item.ID, "ok")
	return nil
}

// 更新CDN记录的LastAPIResponse
func (s *IPUpdateTask) updateDbOfCdnResponse(ctx context.Context, id int64, last_api_response string) error {
	if s.db != nil && id != 0 {
		return s.db.CDNRecord.UpdateOneID(id).
			SetLastAPIResponse(last_api_response).
			SetLastAPIResponseAt(time.Now()).
			Exec(ctx)
	}
	return nil
}

// 更新DNS记录的LastAPIResponse
func (s *IPUpdateTask) updateDbOfDnsResponse(ctx context.Context, id int64, last_api_response string) error {
	if s.db != nil && id != 0 {
		return s.db.DNSRecord.UpdateOneID(id).
			SetLastAPIResponse(last_api_response).
			SetLastAPIResponseAt(time.Now()).
			Exec(ctx)
	}
	return nil
}
