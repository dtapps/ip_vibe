package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
)

// TaskService 任务调度服务接口
type TaskService interface {
	GetTasks(ctx context.Context) ([]types.JobData, error)
	GetTask(ctx context.Context, id string) (*types.JobData, error)
	RunTask(ctx context.Context, id string) error
	DeleteTask(ctx context.Context, id string) error
	StartTask(ctx context.Context) error
	StopTask(ctx context.Context)
	CleanTask(ctx context.Context) error
}

// TaskServiceImpl 任务调度服务实现
type TaskServiceImpl struct {
	db        *ent.Client
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
	scheduler gocron.Scheduler
}

// NewTaskService 创建任务调度服务实例
func NewTaskService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper, cron gocron.Scheduler) TaskService {
	return &TaskServiceImpl{
		db:        db,
		debug:     debug,
		logHelper: logHelper,
		scheduler: cron,
	}
}

// GetJobsHelper 辅助方法：创建 JobHelper 实例
func (s *TaskServiceImpl) GetJobsHelper() *gocron_helper.JobHelper {
	return gocron_helper.NewJobHelper(s.scheduler)
}

// GetTasks 获取全部任务
func (s *TaskServiceImpl) GetTasks(ctx context.Context) ([]types.JobData, error) {
	jobHelper := s.GetJobsHelper()
	// 保持 Handler 中使用的限制 5
	jobs := jobHelper.GetJobsData(5)
	return jobs, nil
}

// GetTask 获取单个任务
func (s *TaskServiceImpl) GetTask(ctx context.Context, id string) (*types.JobData, error) {
	jobHelper := s.GetJobsHelper()
	// 保持 Handler 中使用的限制 15
	jobs := jobHelper.GetJobsData(15)
	for _, job := range jobs {
		if job.ID == id {
			return &job, nil
		}
	}
	return nil, fmt.Errorf("任务不存在")
}

// RunTask 运行任务
func (s *TaskServiceImpl) RunTask(ctx context.Context, id string) error {
	jobHelper := s.GetJobsHelper()
	// 运行任务
	err := jobHelper.RunJob(id)
	if err != nil {
		if s.debug {
			fmt.Printf("运行任务: %s - %s\n", id, err.Error())
		}
		return err
	}
	return nil
}

// DeleteTask 删除任务
func (s *TaskServiceImpl) DeleteTask(ctx context.Context, id string) error {
	jobHelper := s.GetJobsHelper()
	// 删除任务
	err := jobHelper.DeleteJob(id)
	if err != nil {
		if s.debug {
			fmt.Printf("删除任务: %s - %s\n", id, err.Error())
		}
		return err
	}
	return nil
}

// StartTask 启动调度器
func (s *TaskServiceImpl) StartTask(ctx context.Context) error {
	jobHelper := s.GetJobsHelper()
	jobs := jobHelper.GetJobsData(5) // 获取任务列表

	// 启动 GoCron 调度器 (调用 initialize 包中的函数)
	initialize.StartGoCron()

	if len(jobs) == 0 {
		// 启动CDN刷新任务
		// 保持 Handler 逻辑：在 Service 内部创建并调用
		// 注意：这里的 context 应该从外部传入的 ctx 继承
		ipUpdate := NewIpUpdateService(s.db, s.scheduler)
		if err := ipUpdate.StartCdnJob(ctx); err != nil {
			if s.debug {
				fmt.Printf("启动CDN刷新任务失败: %s\n", err.Error())
			}
			return fmt.Errorf("启动CDN刷新任务失败: %w", err)
		}
	}
	return nil
}

// StopTask 停止调度器
func (s *TaskServiceImpl) StopTask(ctx context.Context) {
	// 停止 GoCron 调度器 (调用 initialize 包中的函数)
	initialize.StopGoCron()
}

// CleanTask 清理全部任务
func (s *TaskServiceImpl) CleanTask(ctx context.Context) error {
	jobHelper := s.GetJobsHelper()
	jobs := jobHelper.GetJobsData(5) // 获取全部任务 (保持 Handler 中使用的限制 5)

	// 清理任务
	for _, job := range jobs {
		err := jobHelper.DeleteJob(job.ID)
		if err != nil {
			if s.debug {
				fmt.Printf("清理全部任务: %s\n", err.Error())
			}
			return err
		}
	}
	return nil
}
