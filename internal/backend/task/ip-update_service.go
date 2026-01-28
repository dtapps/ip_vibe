package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/provider/cdn"
	"vibe_ip/internal/provider/dns"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/goarray"
	"golang.org/x/sync/errgroup"
)

// IP更新任务
type IPUpdateTask struct {
	db             *ent.Client                   // 数据库客户端
	logDb          *ent_log.Client               // 日志数据库客户端
	cron           gocron.Scheduler              // 定时任务调度器
	dnsJob         gocron.Job                    // DNS刷新任务
	cdnJob         gocron.Job                    // CDN刷新任务
	jobHelper      *gocron_helper.JobHelper      // 任务助手
	logger         *slog.Logger                  // 日志记录器
	httpTransport  *http_log.LoggingRoundTripper // HTTP日志记录器
	webhookService *webhookService.Send          // Webhook服务
	debug          bool                          // 是否开启调试模式
}

// 创建实例
func NewIPUpdateTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *IPUpdateTask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	webhookService := webhookService.NewSend(db, logDb)
	return &IPUpdateTask{
		db:             db,
		logDb:          logDb,
		cron:           cron,
		jobHelper:      gocron_helper.NewJobHelper(cron, logger),
		logger:         logger,
		httpTransport:  httpTransport,
		webhookService: webhookService,
		debug:          debug,
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

// SetDebug 设置调试模式
func (s *IPUpdateTask) SetDebug(debug bool) *IPUpdateTask {
	s.debug = debug
	return s
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
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.DNSRefreshTaskTime,
		time.Minute,
	)
	if !isStart {
		return nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "解析设置的时间",
			slog.Any("err", err),
		)
		return err
	}
	// 创建周期任务
	s.dnsJob, err = s.cron.NewJob(
		jobDuration,
		gocron.NewTask(s.RunDnsTask), // 执行任务
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
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.CDNRefreshTaskTime,
		time.Minute,
	)
	if !isStart {
		return nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "解析设置的时间",
			slog.Any("err", err),
		)
		return err
	}
	// 创建周期任务
	s.cdnJob, err = s.cron.NewJob(
		jobDuration,
		gocron.NewTask(s.RunCdnTask), // 执行任务
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
	logger := s.jobHelper.SlogWith(s.logger, s.dnsJob.ID().String(), s.dnsJob.Name())
	logger.InfoContext(ctx, "开始执行")

	// 分组记录
	var concurrentRecords, serialRecords []*DnsUpdateParams

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
				invalidTime := db_cache.GvaSetting.DeviceOfflineInvalidTime // 单位：分钟
				if invalidTime <= 0 {
					return
				}
				// 设备心跳时间在 12 分钟内
				withDevices.Where(device.LastHeartbeatAtGTE(time.Now().Add(-time.Duration(invalidTime) * time.Minute)))
				withDevices.OrderByIDAsc()
			})
		}).
		// 排序
		OrderByCreatedAtAsc().
		All(ctx)
	if err == nil {

		// 按策略分组
		for _, dnsRecord := range allDnsRecord {

			// 创建待处理记录
			params := &DnsUpdateParams{
				platform: dnsRecord.Edges.Config,
				record:   dnsRecord,

				DNSRecordID: cast.ToString(dnsRecord.ID),
			}

			if dnsRecord.Type == dnsrecord.TypeCNAME {

				// 获取关联的 CNAME 记录
				cdnRecord := dnsRecord.Edges.CdnRecord
				if cdnRecord == nil {
					// err = fmt.Errorf("记录 %d 没有关联的 CNAME 记录", dnsRecord.ID)
					continue
				}
				params.OriginValue = cdnRecord.CdnDomain

			} else {

				// 获取关联的 IPGroup
				ipGroup := dnsRecord.Edges.IPGroup
				if ipGroup == nil {
					// err = fmt.Errorf("记录 %d 没有关联的 IP组", dnsRecord.ID)
					continue
				}
				params.GroupRecordID = cast.ToString(ipGroup.ID)

				// 获取关联的设备列表
				devices := ipGroup.Edges.Devices
				if len(devices) == 0 {
					// err = fmt.Errorf("记录 %d 没有可用设备", ipGroup.ID)
					continue
				}

				// 收集该记录的源IP
				for _, device := range devices {
					if device.IPV4 != "" && dnsRecord.Type == dnsrecord.TypeA {
						params.OriginValue = device.IPV4
					}
					if device.IPV6 != "" && dnsRecord.Type == dnsrecord.TypeAAAA {
						params.OriginValue = device.IPV6
					}
				}

			}

			if params.OriginValue == "" {
				// err = fmt.Errorf("记录 %d 需要更新的源为空", dnsRecord.ID)
				continue
			}

			// 按策略分组
			if allowConcurrent(dnsRecord.Edges.Config) {
				concurrentRecords = append(concurrentRecords, params)
			} else {
				serialRecords = append(serialRecords, params)
			}
		}

		logger.InfoContext(ctx, "数量",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
		)

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

}

// 运行CDN任务
func (s *IPUpdateTask) RunCdnTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.cdnJob.ID().String(), s.cdnJob.Name())
	logger.InfoContext(ctx, "开始执行")

	// 分组记录
	var concurrentRecords, serialRecords []*CdnUpdateParams

	// 查询全部CDN记录
	allCdnRecord, err := s.db.CDNRecord.Query().
		Where(cdnrecord.IsActiveEQ(true)).
		// 预加载 配置
		WithConfig().
		// 预加载 IP组
		WithIPGroup(func(withIPGroup *ent.IpGroupQuery) {
			// 在 IP组 里查询设备
			withIPGroup.WithDevices(func(withDevices *ent.DeviceQuery) {
				invalidTime := db_cache.GvaSetting.DeviceOfflineInvalidTime // 单位：分钟
				if invalidTime <= 0 {
					return
				}
				// 设备心跳时间在 12 分钟内
				withDevices.Where(device.LastHeartbeatAtGTE(time.Now().Add(-time.Duration(invalidTime) * time.Minute)))
				withDevices.OrderByIDAsc()
			})
		}).
		// 排序
		OrderByCreatedAtAsc().
		All(ctx)
	if err == nil {

		// 按策略分组
		for _, cdnRecord := range allCdnRecord {

			// 解析CDN记录配置
			var cdnRecordConfigJson schematype.CDNRecordConfigJSON
			err = json.Unmarshal([]byte(cdnRecord.ConfigJSON), &cdnRecordConfigJson)
			if err != nil {
				logger.ErrorContext(ctx, "解析CDN记录配置失败",
					slog.Any("error", err),
				)
				continue
			}

			// 创建待处理记录
			pendingRecord := &CdnUpdateParams{
				platform: cdnRecord.Edges.Config,
				record:   cdnRecord,

				CDNRecordID: cast.ToString(cdnRecord.ID),
			}

			// 获取关联的 IPGroup
			ipGroup := cdnRecord.Edges.IPGroup
			if ipGroup == nil {
				// err = fmt.Errorf("记录 %d 没有关联的 IP组", cdnRecord.ID)
				continue
			}
			pendingRecord.GroupRecordID = cast.ToString(ipGroup.ID)
			// pendingRecord.GroupName = gomd5.Get16MD5Encode(pendingRecord.params.GroupRecordID)

			// 获取关联的设备列表
			devices := ipGroup.Edges.Devices
			if len(devices) == 0 {
				// err = fmt.Errorf("记录 %d 没有可用设备", cdnRecord.ID)
				continue
			}

			// 收集该记录的源IP
			for _, device := range devices {
				if device.IPV4 != "" && cdnRecord.Type == cdnrecord.TypeA {
					pendingRecord.OriginValues = append(pendingRecord.OriginValues, device.IPV4)
				}
				if device.IPV6 != "" && cdnRecord.Type == cdnrecord.TypeAAAA {
					pendingRecord.OriginValues = append(pendingRecord.OriginValues, device.IPV6)
				}
			}

			// 去重
			pendingRecord.OriginValues = goarray.RemoveDuplicateElement(pendingRecord.OriginValues)
			if len(pendingRecord.OriginValues) == 0 {
				// err = fmt.Errorf("记录 %d 需要更新的源为空", cdnRecord.ID)
				continue
			}

			// 按策略分组
			if allowConcurrent(cdnRecord.Edges.Config) {
				concurrentRecords = append(concurrentRecords, pendingRecord)
			} else {
				serialRecords = append(serialRecords, pendingRecord)
			}
		}

		logger.InfoContext(ctx, "数量",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
		)

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

	// 判断DNS任务是否存在
	if s.dnsJob == nil {
		go s.StartDnsJob(context.Background())
	}
}

// 更新 CDN 记录 参数
type CdnUpdateParams struct {
	platform *ent.PlatformConfig // 平台信息
	record   *ent.CDNRecord      // CDN信息

	CDNRecordID   string   // CDN记录ID
	GroupRecordID string   // 源站组记录ID
	OriginValues  []string // 源站值列表
}

// 更新 CDN 记录
func (s *IPUpdateTask) CdnUpdate(ctx context.Context, logger *slog.Logger, params *CdnUpdateParams) (err error) {

	// 日志
	logger = logger.With(
		slog.String("记录", params.CDNRecordID),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	var cdnProvider provider.CDNProvider
	var cdnProviderOpts = make([]cdn.Option, 0, len(schematype.GetConfigPlatforms))
	if s.httpTransport != nil {
		cdnProviderOpts = append(cdnProviderOpts, cdn.WithHttpTransport(s.httpTransport))
	}

	switch params.platform.Platform {
	case platformconfig.PlatformTencentCdn:
		cdnProviderCfg := cdn.NewDefaultTencentCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewTencentCDNProvider(ctx, logger, cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformTencentEo:
		cdnProviderCfg := cdn.NewDefaultTencentEOConfig()
		err = cdnProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewTencentEOProvider(ctx, logger, cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformAliyunCdn:
		cdnProviderCfg := cdn.NewDefaultAliyunCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewAliyunCDNProvider(ctx, logger, cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformAliyunEsa:
		cdnProviderCfg := cdn.NewDefaultAliyunESAConfig()
		err = cdnProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewAliyunESAProvider(ctx, logger, cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformBaiduCdn:
		cdnProviderCfg := cdn.NewDefaultBaiduCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewBaiduCDNProvider(ctx, logger, cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformHuaweiCdn:
		cdnProviderCfg := cdn.NewDefaultHuaweiCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewHuaweiCDNProvider(ctx, logger, cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformQiniuCdn:
		cdnProviderCfg := cdn.NewDefaultQiniuCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewQiniuCDNProvider(ctx, logger, cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	default:
		return fmt.Errorf("暂不支持该平台: %s", params.platform.String())
	}

	// 创建 CDN 更新任务
	cdnTask, err := provider.NewCDNTask(
		logger,
		cdnProvider,
		params.platform.ConfigJSON,
		params.record.ConfigJSON,
	)
	if err != nil {
		return fmt.Errorf("创建CDN任务失败: %w", err)
	}

	// 配置 CDN 参数
	cdnTask.Config(params.record.SubDomain, params.record.Type.String(), params.OriginValues)

	// 执行 DNS 更新
	if err = cdnTask.Execute(ctx); err != nil {
		// 保存 DNS 记录响应
		_ = cdnTask.SaveResponse(ctx, s.db, params.record.ID, err.Error())
		return err
	}
	defer func() {
		// 保存 DNS 记录响应
		_ = cdnTask.SaveResponse(ctx, s.db, params.record.ID, "ok")
	}()

	// 保存 DNS 记录刷新
	_ = cdnTask.SaveRefresh(ctx, s.db, params.record.ID)

	// 执行 DNS 通知
	if err := cdnTask.Notify(ctx, s.webhookService, params.record.ID); err != nil {
		return err
	}

	return nil
}

// 更新 DNS 记录 参数
type DnsUpdateParams struct {
	platform *ent.PlatformConfig // 平台信息
	record   *ent.DNSRecord      // DNS信息

	DNSRecordID   string // DNS记录ID
	GroupRecordID string // 源站组记录ID
	OriginValue   string // 源站值
}

// 更新 DNS 记录
func (s *IPUpdateTask) DnsUpdate(ctx context.Context, logger *slog.Logger, params *DnsUpdateParams) (err error) {

	// 日志
	logger = logger.With(
		slog.String("记录", params.DNSRecordID),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	var dnsProvider provider.DNSProvider
	var dnsProviderOpts = make([]dns.Option, 0, len(schematype.GetConfigPlatforms))
	if s.httpTransport != nil {
		dnsProviderOpts = append(dnsProviderOpts, dns.WithHttpTransport(s.httpTransport))
	}

	switch params.platform.Platform {
	case platformconfig.PlatformCloudflare:
		dnsProviderCfg := dns.NewDefaultCloudfareDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewCloudfareDNSProvider(ctx, logger, dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformTencentDNS:
		dnsProviderCfg := dns.NewDefaultTencentDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewTencentDNSProvider(ctx, logger, dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformAliyunDNS:
		dnsProviderCfg := dns.NewDefaultAliyunDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewAliyunDNSProvider(ctx, logger, dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformBaiduDNS:
		dnsProviderCfg := dns.NewDefaultBaiduDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewBaiduDNSProvider(ctx, logger, dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformHuaweiDNS:
		dnsProviderCfg := dns.NewDefaultHuaweiDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewHuaweiDNSProvider(ctx, logger, dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformVolcengineDNS:
		dnsProviderCfg := dns.NewDefaultVolcengineDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewVolcengineDNSProvider(ctx, logger, dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	default:
		return fmt.Errorf("暂不支持该平台: %s", params.platform.String())
	}

	// 创建 DNS 更新任务
	dnsTask, err := provider.NewDNSTask(
		logger,
		dnsProvider,
		params.platform.ConfigJSON,
		params.record.ConfigJSON,
	)
	if err != nil {
		return fmt.Errorf("创建DNS任务失败: %w", err)
	}

	// 配置 DNS 参数
	dnsTask.Config(params.record.SubDomain, params.record.Type.String(), params.OriginValue)

	// 执行 DNS 更新
	if err = dnsTask.Execute(ctx); err != nil {
		// 保存 DNS 记录响应
		_ = dnsTask.SaveResponse(ctx, s.db, params.record.ID, err.Error())
		return err
	}
	defer func() {
		// 保存 DNS 记录响应
		_ = dnsTask.SaveResponse(ctx, s.db, params.record.ID, "ok")
	}()

	// 保存 DNS 记录刷新
	_ = dnsTask.SaveRefresh(ctx, s.db, params.record.ID)

	// 执行 DNS 通知
	if err := dnsTask.Notify(ctx, s.webhookService, params.record.ID); err != nil {
		return err
	}

	return nil
}
