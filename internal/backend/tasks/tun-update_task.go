package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent/tunrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/provider/tun"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"golang.org/x/sync/errgroup"
)

// TUN更新任务
type TUNUpdateTask struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	cron          gocron.Scheduler              // 定时任务调度器
	job           gocron.Job                    // TUN更新任务
	jobHelper     *gocron_helper.JobHelper      // 任务助手
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 是否开启调试模式
}

// 创建实例
func NewTUNUpdateTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *TUNUpdateTask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &TUNUpdateTask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		jobHelper:     gocron_helper.NewJobHelper(cron, logger),
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// Close 关闭服务，释放资源
func (s *TUNUpdateTask) Close() {
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetDebug 设置调试模式
func (s *TUNUpdateTask) SetDebug(debug bool) *TUNUpdateTask {
	s.debug = debug
	return s
}

// 自定义 logger
func (s *TUNUpdateTask) setLogger(logger *slog.Logger) *TUNUpdateTask {
	s.logger = logger
	return s
}

// 启动TUN任务
func (s *TUNUpdateTask) StartJob(ctx context.Context) (err error) {
	// 判断是否存在TUN任务
	if s.job != nil {
		return nil
	}
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.TUNRefreshTaskTime,
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
	s.job, err = s.cron.NewJob(
		jobDuration,
		gocron.NewTask(s.RunTunTask), // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("TUN记录刷新"),                           // 任务名称
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

// 运行TUN任务
func (s *TUNUpdateTask) RunTunTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())
	logger.InfoContext(ctx, "开始执行")

	// 分组记录
	var concurrentRecords, serialRecords []*TUNUpdateParams

	// 查询全部TUN记录
	allTUNRecord, err := s.db.TUNRecord.Query().
		Where(tunrecord.IsActiveEQ(true)).
		// 预加载 配置
		WithConfig().
		// 排序
		OrderByCreatedAtAsc().
		All(ctx)
	if err == nil {

		// 按策略分组
		for _, tunRecord := range allTUNRecord {

			// 创建待处理记录
			params := &TUNUpdateParams{
				platform: tunRecord.Edges.Config,
				record:   tunRecord,

				TUNRecordID: cast.ToString(tunRecord.ID),
			}

			// 按策略分组
			if allowConcurrent(tunRecord.Edges.Config) {
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

					if err = s.TUNUpdate(egCtx, sonLogger, item); err != nil {
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

			if err = s.TUNUpdate(ctx, sonLogger, item); err != nil {
				logger.ErrorContext(ctx, "处理失败",
					slog.Any("error", err),
				)
			}

		}

	}

}

// 更新 TUN 记录 参数
type TUNUpdateParams struct {
	platform *ent.PlatformConfig // 平台信息
	record   *ent.TUNRecord      // TUN记录

	TUNRecordID string // TUN记录ID
}

// 更新 TUN 记录
func (s *TUNUpdateTask) TUNUpdate(ctx context.Context, logger *slog.Logger, params *TUNUpdateParams) (err error) {

	// 日志
	logger = logger.With(
		slog.String("记录", params.TUNRecordID),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	var tunProvider provider.TUNProvider
	var tunProviderOpts = make([]tun.Option, 0, len(schematype.GetConfigPlatforms))
	if s.httpTransport != nil {
		tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(s.httpTransport))
	}

	switch params.platform.Platform {
	case platformconfig.PlatformCloudflare:
		tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
		err = tunProviderCfg.ParseCustomConfig(params.platform.ConfigJSON)
		if err != nil {
			return fmt.Errorf("解析自定义配置失败: %w", err)
		}
		tunProvider, err = tun.NewCloudfareTUNProvider(ctx, logger, tunProviderCfg, tunProviderOpts...)
		if err != nil {
			return fmt.Errorf("初始提供程序失败: %w", err)
		}
	default:
		return fmt.Errorf("暂不支持该平台: %s", params.platform.String())
	}

	// 创建 TUN 更新任务
	tunTask, err := provider.NewTUNTask(
		logger,
		tunProvider,
		params.platform.ConfigJSON,
		params.record.ConfigJSON,
	)
	if err != nil {
		return fmt.Errorf("创建DNS任务失败: %w", err)
	}

	// 配置 TUN 参数
	tunTask.Config(params.record.SubDomain, params.record.DomainPath, params.record.ServeType.String(), params.record.ServeURI)

	// 执行 TUN 更新
	if err = tunTask.Execute(ctx); err != nil {
		// 保存 TUN 记录响应
		_ = tunTask.SaveResponse(ctx, s.db, params.record.ID, err.Error())
		return err
	}
	defer func() {
		// 保存 TUN 记录响应
		_ = tunTask.SaveResponse(ctx, s.db, params.record.ID, "ok")
	}()

	// 保存 TUN 记录刷新
	_ = tunTask.SaveRefresh(ctx, s.db, params.record.ID)

	// 执行 TUN 通知
	if err := tunTask.Notify(ctx, s.db, params.record.ID); err != nil {
		return err
	}

	return nil
}

// 更新CDN记录的LastAPIResponse
func (s *TUNUpdateTask) updateDbOfCdnResponse(ctx context.Context, id int64, last_api_response string) error {
	if s.db != nil && id != 0 {
		return s.db.CDNRecord.UpdateOneID(id).
			SetLastAPIResponse(last_api_response).
			SetLastAPIResponseAt(time.Now()).
			Exec(ctx)
	}
	return nil
}
