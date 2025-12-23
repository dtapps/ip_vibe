package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
)

// TaskService 任务调度服务接口
type TaskService interface {
	GetTasks(ctx context.Context) ([]*types.JobData, error)
	GetTask(ctx context.Context, id string) (*types.JobData, error)
	RunTask(ctx context.Context, id string) error
	DeleteTask(ctx context.Context, id string) error
	CleanAllTask(ctx context.Context) error
	StartScheduler(ctx context.Context) error
	StopScheduler(ctx context.Context)
}

// TaskServiceImpl 任务调度服务实现
type TaskServiceImpl struct {
	db        *ent.Client     // 数据库客户端
	dbLog     *ent_log.Client // 日志数据库客户端
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
	cron      gocron.Scheduler // 定时任务调度器
	logger    *slog.Logger     // 日志记录器
}

// NewTaskService 创建任务调度服务实例
func NewTaskService(db *ent.Client, dbLog *ent_log.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper, cron gocron.Scheduler, logger *slog.Logger) TaskService {
	return &TaskServiceImpl{
		db:        db,
		dbLog:     dbLog,
		debug:     debug,
		logHelper: logHelper,
		cron:      cron,
		logger:    logger,
	}
}

// getJobsHelper 辅助方法：创建 JobHelper 实例
func (s *TaskServiceImpl) getJobsHelper() *gocron_helper.JobHelper {
	return gocron_helper.NewJobHelper(s.cron)
}

// GetTasks 获取全部任务
func (s *TaskServiceImpl) GetTasks(ctx context.Context) ([]*types.JobData, error) {
	jobHelper := s.getJobsHelper()

	// 获取任务列表
	jobs := jobHelper.GetJobsData(5)
	return jobs, nil
}

// GetTask 获取单个任务
func (s *TaskServiceImpl) GetTask(ctx context.Context, id string) (*types.JobData, error) {
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
func (s *TaskServiceImpl) RunTask(ctx context.Context, id string) error {
	jobHelper := s.getJobsHelper()

	// 运行任务
	err := jobHelper.RunJob(id)
	if err != nil {
		if s.debug {
			slog.ErrorContext(ctx, "运行任务失败",
				slog.String("id", id),
				slog.Any("err", err),
			)
		}
		return err
	}
	return nil
}

// DeleteTask 删除任务
func (s *TaskServiceImpl) DeleteTask(ctx context.Context, id string) error {
	jobHelper := s.getJobsHelper()

	// 删除任务
	err := jobHelper.DeleteJob(id)
	if err != nil {
		if s.debug {
			slog.ErrorContext(ctx, "删除任务失败",
				slog.String("id", id),
				slog.Any("err", err),
			)
		}
		return err
	}
	return nil
}

// CleanAllTask 清理全部任务
func (s *TaskServiceImpl) CleanAllTask(ctx context.Context) error {
	jobHelper := s.getJobsHelper()

	// 获取任务列表
	jobs := jobHelper.GetJobsData(5)

	// 清理任务
	for _, job := range jobs {
		err := jobHelper.DeleteJob(job.ID)
		if err != nil {
			if s.debug {
				slog.ErrorContext(ctx, "清理全部任务失败",
					slog.String("id", job.ID),
					slog.Any("err", err),
				)
			}
			return err
		}
	}
	return nil
}

// StartScheduler 启动调度器
func (s *TaskServiceImpl) StartScheduler(ctx context.Context) error {
	jobHelper := s.getJobsHelper()

	// 获取任务列表
	jobs := jobHelper.GetJobsData(5)

	initialize.StartGoCron()

	if len(jobs) == 0 {
		// 启动CDN刷新任务
		ipUpdate := task.NewIPUpdateTask(s.db, s.dbLog, s.cron, s.logger)
		if err := ipUpdate.StartCdnJob(ctx); err != nil {
			if s.debug {
				slog.ErrorContext(ctx, "启动CDN、DNS刷新任务失败",
					slog.Any("err", err),
				)
			}
			return fmt.Errorf("启动CDN、DNS刷新任务失败: %w", err)
		}
	}
	return nil
}

// StopScheduler 停止调度器
func (s *TaskServiceImpl) StopScheduler(ctx context.Context) {
	initialize.StopGoCron()
}
