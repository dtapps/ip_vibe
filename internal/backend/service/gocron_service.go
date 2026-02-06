package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/backend/tasks"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/contrib/http_log"
)

// GocronService 任务调度服务接口
type GocronService interface {
	GetTasks(ctx context.Context) ([]*types.GocronJob, error)
	GetTask(ctx context.Context, id string) (*types.GocronJob, error)
	RunTask(ctx context.Context, id string) error
	DeleteTask(ctx context.Context, id string) error
	CleanAllTask(ctx context.Context) error
	StartScheduler(ctx context.Context) error
	StopScheduler(ctx context.Context)
}

// GocronServiceImpl 任务调度服务实现
type GocronServiceImpl struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	cron          gocron.Scheduler              // 定时任务调度器
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewGocronService 创建任务调度服务实例
func NewGocronService(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) GocronService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &GocronServiceImpl{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *GocronServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// getJobsHelper 辅助方法：创建 JobHelper 实例
func (s *GocronServiceImpl) getJobsHelper() *gocron_helper.JobHelper {
	return gocron_helper.NewJobHelper(s.cron, s.logger)
}

// GetTasks 获取全部任务
func (s *GocronServiceImpl) GetTasks(ctx context.Context) ([]*types.GocronJob, error) {
	jobHelper := s.getJobsHelper()

	// 获取任务列表
	jobs := jobHelper.GetJobsData(5)
	return jobs, nil
}

// GetTask 获取单个任务
func (s *GocronServiceImpl) GetTask(ctx context.Context, id string) (*types.GocronJob, error) {
	jobHelper := s.getJobsHelper()

	// 获取任务列表
	jobs := jobHelper.GetJobsData(15)
	for _, job := range jobs {
		if job.ID == id {
			return job, nil
		}
	}
	return nil, fmt.Errorf("任务不存在")
}

// RunTask 运行任务
func (s *GocronServiceImpl) RunTask(ctx context.Context, id string) error {
	jobHelper := s.getJobsHelper()

	// 运行任务
	err := jobHelper.RunJob(id)
	if err != nil {
		return fmt.Errorf("运行任务失败: %w", err)
	}
	return nil
}

// DeleteTask 删除任务
func (s *GocronServiceImpl) DeleteTask(ctx context.Context, id string) error {
	jobHelper := s.getJobsHelper()

	// 删除任务
	err := jobHelper.DeleteJob(id)
	if err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}
	return nil
}

// CleanAllTask 清理全部任务
func (s *GocronServiceImpl) CleanAllTask(ctx context.Context) error {
	jobHelper := s.getJobsHelper()

	// 获取任务列表
	jobs := jobHelper.GetJobsData(5)

	// 清理任务
	for _, job := range jobs {
		err := jobHelper.DeleteJob(job.ID)
		if err != nil {
			return fmt.Errorf("清理全部任务失败: %w", err)
		}
	}
	return nil
}

// StartScheduler 启动调度器
func (s *GocronServiceImpl) StartScheduler(ctx context.Context) error {
	jobHelper := s.getJobsHelper()

	// 获取任务列表
	jobs := jobHelper.GetJobsData(5)

	initialize.StartGoCron(ctx, s.logger)

	if len(jobs) == 0 {
		// 启动CDN刷新任务
		ipUpdateTask := tasks.NewIPUpdateTask(
			s.db,
			s.logDb,
			s.cron,
			s.logger,
			s.httpTransport,
		)
		if err := ipUpdateTask.StartCdnJob(ctx); err != nil {
			return fmt.Errorf("启动CDN、DNS刷新任务失败: %w", err)
		}
	}
	return nil
}

// StopScheduler 停止调度器
func (s *GocronServiceImpl) StopScheduler(ctx context.Context) {
	initialize.StopGoCron(ctx, s.logger)
}
