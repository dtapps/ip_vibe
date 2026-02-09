package cron_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/cronjoblog"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/provider/task"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.dtapp.net/library/contrib/http_log"
)

type CronServerManager struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	cron          gocron.Scheduler              // 定时任务调度器
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	taskRegistry  *provider.TaskTask            // 任务注册中心
	mu            sync.Mutex                    // 引入读写锁
	runningJobs   map[int64]gocron.Job          // 正在运行的任务映射
	jobConfigs    map[int64]string              // 任务配置映射
}

var (
	instanceCronServer *CronServerManager
	onceCronServer     sync.Once
)

// NewCronServerManager 创建定时任务服务管理器实例
func NewCronServerManager(ctx context.Context, db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) (*CronServerManager, error) {
	taskRegistry, err := provider.NewTaskTask(ctx, logger)
	if err != nil {
		return nil, err
	}
	onceCronServer.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instanceCronServer = &CronServerManager{
			db:            db,
			logDb:         logDb,
			cron:          cron,
			logger:        logger,
			httpTransport: httpTransport,
			taskRegistry:  taskRegistry,
			runningJobs:   make(map[int64]gocron.Job),
			jobConfigs:    make(map[int64]string),
		}
	})
	return instanceCronServer, nil
}

// GetCronServerInstance 获取定时任务服务管理器单例
func GetCronServerInstance() *CronServerManager {
	return instanceCronServer
}

// InitHandlers 初始化并注册所有的任务处理器
func (m *CronServerManager) InitHandlers(ctx context.Context) {

	// 注册 CDN同步 处理器
	cdnSyncHandler := task.NewCDNSyncTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.CDNSyncTaskName, cdnSyncHandler)

	// 注册 DNS同步 处理器
	dnsSynHandler := task.NewDNSSyncTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.DNSSyncTaskName, dnsSynHandler)

	// 注册 域名监控 处理器
	domainMonitorHandler := task.NewDomainMonitorTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.DomainMonitorTaskName, domainMonitorHandler)

	// 注册 SSL申请 处理器
	sslApplyHandler := task.NewSSLApplyTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.SSLApplyTaskName, sslApplyHandler)

	// 注册 SSL刷新 处理器
	sslRefreshHandler := task.NewSSLRefreshTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.SSLRefreshTaskName, sslRefreshHandler)

	// 注册 Cloudflare隧道同步 处理器
	cloudflareTunnelSyncHandler := task.NewCloudflareTunnelSyncTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.CloudflareTunnelSyncTaskName, cloudflareTunnelSyncHandler)

	// 注册 Cloudflare站点同步 处理器
	cloudflareZoneSyncHandler := task.NewCloudflareZoneSyncTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.CloudflareZoneSyncTaskName, cloudflareZoneSyncHandler)

	// 注册 EdgeOne站点同步 处理器
	edgeOneZoneSyncHandler := task.NewEdgeOneZoneSyncTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.EdgeOneZoneSyncTaskName, edgeOneZoneSyncHandler)

	// 注册 ESA站点 处理器
	esaSiteSyncHandler := task.NewESASiteSyncTaskProvider(m.db, m.logDb, m.logger, m.httpTransport)
	m.taskRegistry.RegisterTask(ctx, task.ESASiteSyncTaskName, esaSiteSyncHandler)

	// 注册 数据清除 处理器
	clearDataHandler := task.NewClearDataTaskProviderRaw(m.db, m.logDb, m.logger)
	m.taskRegistry.RegisterTask(ctx, task.ClearDataTaskName, clearDataHandler)

	m.logger.InfoContext(ctx, "所有任务处理器注册完成")
}

// LoadTasks 从数据库加载并同步所有状态为 running 的任务到 gocron
func (m *CronServerManager) LoadTasks(ctx context.Context) {

	// 加载 SSL刷新 任务
	sslRefreshHandler := task.NewSSLRefreshTaskProviderRaw(m.db, m.logDb, m.logger, m.httpTransport)
	sslRefreshHandlerTable := map[string]bool{
		ent.TypeSSLRecord: true,
		ent.TypeSSLImport: true,
	}
	for table := range sslRefreshHandlerTable {
		err := sslRefreshHandler.LoadRecurringTask(ctx, m.db, table)
		if err != nil {
			m.logger.ErrorContext(ctx, "加载 SSL刷新 周期任务失败",
				slog.String("type", table),
				slog.Any("err", err),
			)
		}
	}

	// 加载 数据清除 任务
	clearDataHandler := task.NewClearDataTaskProviderRaw(m.db, m.logDb, m.logger)
	clearDataHandlerTable := map[string]bool{
		ent_log.TypeHttpLog:              true,
		ent_log.TypeRestyLog:             true,
		ent_log.TypeReqLog:               true,
		ent_log.TypeDomainMonitorHistory: true,
		ent_log.TypeOperationLog:         true,
		ent_log.TypeWebhookHistory:       true,
		ent_log.TypeDeviceIpReport:       true,
		ent_log.TypePortForwardStats:     true,
	}
	for table := range clearDataHandlerTable {
		err := clearDataHandler.LoadRecurringTask(ctx, m.db, table)
		if err != nil {
			m.logger.ErrorContext(ctx, "加载 数据清除 周期任务失败",
				slog.String("type", table),
				slog.Any("err", err),
			)
		}
	}

	m.logger.InfoContext(ctx, "所有任务加载完成")
}

// RunHandler 这是一个通用的执行入口，供 gocron 调用
func (m *CronServerManager) RunHandler(jobID int64, handlerName string, args json.RawMessage, meta provider.TaskMetadata) {
	ctx := context.Background()
	startTime := time.Now()

	// 创建初始日志 (Running 状态)
	logRecord, err := m.logDb.CronJobLog.Create().
		SetCronJobID(jobID).
		SetHandlerName(handlerName).
		SetStartedAt(startTime).
		SetStatus(cronjoblog.StatusRunning).
		Save(ctx)
	if err != nil {
		m.logger.Error("创建运行日志失败", slog.Int64("job_id", jobID), slog.Any("err", err))
	}

	// 获取处理器并执行
	var result string
	var execErr error
	handler, ok := m.taskRegistry.GetTask(handlerName)
	if !ok {
		execErr = fmt.Errorf("未找到处理器: %s", handlerName)
	} else {
		result, execErr = handler.Process(ctx, args, meta)
	}

	// 计算结果
	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	// 回填日志
	if logRecord != nil {
		updaterLog := m.logDb.CronJobLog.UpdateOne(logRecord)
		updaterLog.SetFinishedAt(endTime)
		updaterLog.SetDuration(duration)
		updaterLog.SetOutput(result)

		if execErr != nil {
			updaterLog.SetStatus(cronjoblog.StatusFailed)
			updaterLog.SetErrorInfo(execErr.Error())
		} else {
			updaterLog.SetStatus(cronjoblog.StatusSuccess)
		}
		_ = updaterLog.Exec(ctx)
	}

	if execErr != nil {
		m.logger.Error("任务执行失败",
			slog.Int64("job_id", jobID),
			slog.Any("err", execErr),
		)
	}
}

// LoadAllJobs 从数据库加载所有状态为 running 的任务到 gocron
func (m *CronServerManager) LoadAllJobs(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 从数据库查询所有开启的任务
	jobs, err := m.db.CronJob.Query().
		Where(cronjob.StatusEQ(cronjob.StatusRunning)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, j := range jobs {
		err := m.addJobToCron(j)
		if err != nil {
			m.logger.Error("加载任务失败",
				slog.Int64("id", j.ID),
				slog.Any("err", err),
			)
		}
	}
	return nil
}

// SyncAllJobs 从数据库全量同步任务状态到内存调度器
func (m *CronServerManager) SyncAllJobs(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 从数据库获取所有当前应该运行的任务
	dbJobs, err := m.db.CronJob.Query().
		Where(cronjob.StatusEQ(cronjob.StatusRunning)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query running jobs failed: %w", err)
	}

	// 将数据库任务转为 Map 方便查找
	dbJobMap := make(map[int64]*ent.CronJob)
	for _, j := range dbJobs {
		dbJobMap[j.ID] = j
	}

	// 遍历内存中正在运行的任务
	for id := range m.runningJobs {
		dbJob, exists := dbJobMap[id]

		// 如果数据库里已经没有这个任务，或者状态不再是 running
		if !exists {
			m.logger.Info("任务已停用或删除，正在从调度器移除", slog.Int64("id", id))
			m.removeJobFromCron(id)
			continue
		}

		// 如果任务还在，对比指纹看参数是否变了
		currentFingerprint := m.generateFingerprint(dbJob)
		if oldFingerprint, ok := m.jobConfigs[id]; ok && oldFingerprint != currentFingerprint {
			m.logger.Info("任务配置已变更，正在重启计时器", slog.Int64("id", id))
			m.removeJobFromCron(id)
			if err := m.addJobToCron(dbJob); err == nil {
				m.jobConfigs[id] = currentFingerprint
			}
		}
	}

	// 遍历数据库任务，如果内存里没有，则添加
	for id, dbJob := range dbJobMap {
		if _, running := m.runningJobs[id]; !running {
			m.logger.Info("发现新任务，正在加入调度器", slog.Int64("id", id))
			if err := m.addJobToCron(dbJob); err == nil {
				m.jobConfigs[id] = m.generateFingerprint(dbJob)
			} else {
				m.logger.Error("启动新任务失败", slog.Int64("id", id), slog.Any("err", err))
			}
		}
	}

	return nil
}

// SyncJob 同步单个任务的状态（自动识别：新增/更新/删除）
func (m *CronServerManager) SyncJob(ctx context.Context, jobID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 尝试从数据库获取
	j, err := m.db.CronJob.Get(ctx, jobID)

	// 如果查不到 (NotFound)，视为删除动作
	if err != nil {
		if ent.IsNotFound(err) {
			m.removeJobFromCron(jobID)
			return nil
		}
		return err
	}

	// 如果状态不是运行中，同样执行移除
	if j.Status != cronjob.StatusRunning {
		m.removeJobFromCron(jobID)
		return nil
	}

	// 指纹比对逻辑（避免无意义的重启）
	newFingerprint := m.generateFingerprint(j)
	if oldFingerprint, exists := m.jobConfigs[jobID]; exists {
		if oldFingerprint == newFingerprint {
			return nil // 配置未变，不操作
		}
		m.removeJobFromCron(jobID) // 配置变了，先删
	}

	// 加入调度
	if err := m.addJobToCron(j); err != nil {
		return err
	}
	m.jobConfigs[jobID] = newFingerprint
	return nil
}

// 辅助方法：生成配置指纹
func (m *CronServerManager) generateFingerprint(j *ent.CronJob) string {
	return fmt.Sprintf("%v-%v-%d-%v-%d-%d-%s-%s-%t-%t",
		j.JobType,
		j.TaskType,
		j.IntervalValue,
		j.IntervalUnit,
		j.MinInterval,
		j.MaxInterval,
		j.CronExpr,
		string(j.HandlerArgs),
		j.RunOnStart,
		j.IsSingleton,
	)
}

// 内部方法：真正把任务塞进 gocron
func (m *CronServerManager) addJobToCron(j *ent.CronJob) error {

	// 转换时间单位
	var timeUnit time.Duration
	switch j.IntervalUnit {
	case cronjob.IntervalUnitSeconds:
		timeUnit = time.Second
	case cronjob.IntervalUnitMinutes:
		timeUnit = time.Minute
	case cronjob.IntervalUnitHours:
		timeUnit = time.Hour
	case cronjob.IntervalUnitDays:
		timeUnit = time.Hour * 24
	case cronjob.IntervalUnitWeeks:
		timeUnit = time.Hour * 24 * 7
	}

	// 定义执行逻辑
	task := gocron.NewTask(func() {
		m.RunHandler(
			j.ID,
			j.HandlerName,
			j.HandlerArgs,
			provider.TaskMetadata{
				JobID:        j.ID,
				JobName:      j.Name,
				TaskType:     j.TaskType,
				TaskCategory: j.JobCategory,
			},
		)
	})

	// 任务选项
	opts := make([]gocron.JobOption, 0)

	// 单例模式
	if j.IsSingleton {
		opts = append(
			opts,
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		)
	}

	// 立即开始
	if j.RunOnStart {
		opts = append(
			opts,
			gocron.WithStartAt(gocron.WithStartImmediately()),
		)
	}

	// 任务编号
	opts = append(
		opts,
		gocron.WithIdentifier(j.EntryID),
	)

	// 任务名称
	opts = append(
		opts,
		gocron.WithName(j.Name),
	)

	// 任务事件监听器
	opts = append(
		opts,
		gocron.WithEventListeners(
			m.PanicPeriodicListener(),
			m.BeforePeriodicListener(),
			m.AfterPeriodicListener(),
		),
	)

	// 根据你的随机时间/固定时间逻辑配置调度
	var sch gocron.JobDefinition
	switch j.JobType {
	case cronjob.JobTypeCron:
		sch = gocron.CronJob(j.CronExpr, j.WithSeconds)
	case cronjob.JobTypeDuration:
		sch = gocron.DurationJob(
			time.Duration(j.IntervalValue) * timeUnit,
		)
	case cronjob.JobTypeRandom:
		sch = gocron.DurationRandomJob(
			time.Duration(j.MinInterval)*timeUnit, // 最小间隔
			time.Duration(j.MaxInterval)*timeUnit, // 最大间隔
		)
	}

	newJob, err := m.cron.NewJob(sch, task, opts...)
	if err != nil {
		return err
	}

	m.runningJobs[j.ID] = newJob
	return nil
}

// 内部方法：从 gocron 移除
func (m *CronServerManager) removeJobFromCron(jobID int64) {
	if oldJob, ok := m.runningJobs[jobID]; ok {
		_ = m.cron.RemoveJob(oldJob.ID())
		delete(m.runningJobs, jobID)
		delete(m.jobConfigs, jobID)
	}
}

// BeforePeriodicListener 任务开始前的监听器
func (j *CronServerManager) BeforePeriodicListener() gocron.EventListener {
	return gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {

	})
}

// AfterPeriodicListener 任务执行完成后的监听器
func (m *CronServerManager) AfterPeriodicListener() gocron.EventListener {
	return gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
		ctx := context.Background()

		// 从数据库查询任务
		j, err := m.db.CronJob.Query().
			Where(cronjob.EntryIDEQ(jobID)).
			First(ctx)
		if err != nil {
			m.logger.ErrorContext(ctx, "[AfterPeriodicListener] 查询任务失败",
				slog.String("entry_id", jobID.String()),
				slog.Any("err", err),
			)
			return
		}

		// 如果是一次性任务，执行完后修改状态并从内存移除
		if j.TaskType == cronjob.TaskTypeOneOff {

			// 默认状态是已完成
			status := cronjob.StatusFinished

			// 如果任务是失败状态，更新状态为失败
			if j.Status == cronjob.StatusFailed {
				status = cronjob.StatusFailed
			}

			// 更新数据库状态
			_ = m.db.CronJob.UpdateOne(j).
				SetStatus(status).
				Exec(ctx)

			// 移除任务
			m.mu.Lock()
			m.removeJobFromCron(j.ID)
			m.mu.Unlock()

			m.logger.InfoContext(ctx, "[AfterPeriodicListener] 一次性任务执行完成并已停止",
				slog.String("entry_id", jobID.String()),
			)
		}

	})
}

// PanicPeriodicListener 任务异常时的监听器
func (m *CronServerManager) PanicPeriodicListener() gocron.EventListener {
	return gocron.AfterJobRunsWithPanic(func(jobID uuid.UUID, jobName string, recoverData any) {
		ctx := context.Background()

		// 从数据库查询任务
		j, err := m.db.CronJob.Query().
			Where(cronjob.EntryIDEQ(jobID)).
			First(ctx)
		if err != nil {
			m.logger.ErrorContext(ctx, "[PanicPeriodicListener] 查询任务失败",
				slog.String("entry_id", jobID.String()),
				slog.Any("err", err),
			)
			return
		}

		// 如果是一次性任务，执行完后修改状态并从内存移除
		if j.TaskType == cronjob.TaskTypeOneOff {

			// 更新数据库状态
			_ = m.db.CronJob.UpdateOne(j).
				SetStatus(cronjob.StatusFailed).
				Exec(ctx)

			// 移除任务
			m.mu.Lock()
			m.removeJobFromCron(j.ID)
			m.mu.Unlock()

			m.logger.InfoContext(ctx, "[PanicPeriodicListener] 一次性任务执行失败并已停止",
				slog.String("entry_id", jobID.String()),
			)
		}

	})
}
