package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
)

// GocronService 任务调度服务接口
type GocronService interface {
	GetTasks(ctx context.Context) ([]*types.GocronJob, error)
	GetTask(ctx context.Context, id string) (*types.GocronJob, error)
	RunTask(ctx context.Context, id string) error
	DeleteTask(ctx context.Context, id string) error
	CleanAllTask(ctx context.Context) error
}

// GocronServiceImpl 任务调度服务实现
type GocronServiceImpl struct {
	db            *ent.Client              // 数据库客户端
	logDb         *ent_log.Client          // 日志数据库客户端
	cron          gocron.Scheduler         // 定时任务调度器
	jobHelper     *gocron_helper.JobHelper // 任务助手
	nextRunsLimit int                      // 下次执行时间限制
	logger        *slog.Logger             // 日志记录器
	debug         bool                     // 调试模式

}

// NewGocronService 创建任务调度服务实例
func NewGocronService(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger) GocronService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	jobHelper := gocron_helper.NewJobHelper(cron, logger)
	return &GocronServiceImpl{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		jobHelper:     jobHelper,
		nextRunsLimit: 15,
		logger:        logger,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *GocronServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// GetTasks 获取全部任务
func (s *GocronServiceImpl) GetTasks(ctx context.Context) ([]*types.GocronJob, error) {

	// 获取任务列表
	jobs := s.jobHelper.GetJobsData(s.nextRunsLimit)

	return jobs, nil
}

// GetTask 获取单个任务
func (s *GocronServiceImpl) GetTask(ctx context.Context, id string) (*types.GocronJob, error) {

	// 获取任务列表
	jobs := s.jobHelper.GetJobsData(s.nextRunsLimit)
	for _, job := range jobs {
		if job.ID == id {
			return job, nil
		}
	}

	return nil, fmt.Errorf("任务不存在")
}

// RunTask 运行任务
func (s *GocronServiceImpl) RunTask(ctx context.Context, id string) error {

	// 运行任务
	err := s.jobHelper.RunJob(id)
	if err != nil {
		return fmt.Errorf("运行任务失败: %w", err)
	}

	return nil
}

// DeleteTask 删除任务
func (s *GocronServiceImpl) DeleteTask(ctx context.Context, id string) error {

	// 删除任务
	err := s.jobHelper.DeleteJob(id)
	if err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	return nil
}

// CleanAllTask 清理全部任务
func (s *GocronServiceImpl) CleanAllTask(ctx context.Context) error {

	// 获取任务列表
	jobs := s.jobHelper.GetJobsData(5)

	// 清理任务
	for _, job := range jobs {
		err := s.jobHelper.DeleteJob(job.ID)
		if err != nil {
			return fmt.Errorf("清理全部任务失败: %w", err)
		}
	}

	return nil
}
