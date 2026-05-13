package gocron_helper

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
	"vibe_ip/internal/types"

	"github.com/dromara/carbon/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

// 任务助手
type JobHelper struct {
	Scheduler gocron.Scheduler // 定时任务调度器
	logger    *slog.Logger     // 日志记录器
}

// 新建任务助手
func NewJobHelper(scheduler gocron.Scheduler, logger *slog.Logger) *JobHelper {
	return &JobHelper{
		Scheduler: scheduler,
		logger:    logger,
	}
}

// 查询全部任务
func (j *JobHelper) GetJobsData(nextRunsLimit int) (jobs []*types.GocronJob) {
	if j.Scheduler == nil {
		return jobs
	}
	if nextRunsLimit <= 0 {
		nextRunsLimit = 1
	}

	// 获取当前时间
	now := time.Now()

	for _, job := range j.Scheduler.Jobs() {
		nextRun, _ := job.NextRun()
		lastRun, _ := job.LastRunStartedAt()
		nextRuns, _ := job.NextRuns(nextRunsLimit)

		// 确保 NextRun 是未来时间
		// 条件1: nextRun 为零值（无下一次运行）
		// 条件2: nextRun 已过期（小于当前时间）
		if nextRun.IsZero() || nextRun.Before(now) {
			// 从 nextRuns 中查找第一个 >= now 的时间
			for _, nr := range nextRuns {
				if nr.After(now) || nr.Equal(now) {
					nextRun = nr
					break
				}
			}
		}

		// 推断任务调度信息
		schedule, scheduleDetail := j.inferSchedule(job, nextRuns)

		jobs = append(jobs, &types.GocronJob{
			ID:                     job.ID().String(),
			Name:                   job.Name(),
			LastRun:                lastRun,
			NextRun:                nextRun,
			ExecutionTime:          nextRuns,
			HumanizedExecutionTime: j.humanizeTimeList(nextRuns),
			Schedule:               schedule,
			ScheduleDetail:         scheduleDetail,
		})
	}
	return jobs
}

// 判断任务，传 id 或 name
func (j *JobHelper) IsJob(jobID string, jobName string) bool {
	if j.Scheduler == nil {
		return false
	}
	// 查询全部任务
	jobs := j.GetJobsData(5)
	for _, job := range jobs {
		if jobID != "" && job.ID == jobID {
			return true
		}
		if jobName != "" && job.Name == jobName {
			return true
		}
	}
	return false
}

// 获取任务详情
func (j *JobHelper) GetJob(jobID string) (job *types.GocronJobInfo, err error) {
	if j.Scheduler == nil {
		return nil, fmt.Errorf("调度器未初始化")
	}
	// 查询全部任务
	jobs := j.GetJobsData(5)
	for _, job := range jobs {
		if job.ID == jobID {
			return &types.GocronJobInfo{
				ID:      job.ID,
				Name:    job.Name,
				LastRun: job.LastRun,
				NextRun: job.NextRun,
			}, nil
		}
	}
	return nil, fmt.Errorf("任务不存在: %w", err)
}

// 运行任务
func (j *JobHelper) RunJob(jobID string) (err error) {
	if j.Scheduler == nil {
		return fmt.Errorf("调度器未初始化")
	}
	jobs := j.Scheduler.Jobs()
	for _, job := range jobs {
		if job.ID().String() == jobID {
			err = job.RunNow()
			if err != nil {
				return fmt.Errorf("任务执行失败: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("任务不存在: %w", err)
}

// 删除任务
func (j *JobHelper) DeleteJob(jobID string) error {
	if j.Scheduler == nil {
		return fmt.Errorf("调度器未初始化")
	}
	id, err := uuid.Parse(jobID)
	if err != nil {
		return fmt.Errorf("无效的任务ID: %w", err)
	}
	if err := j.Scheduler.RemoveJob(id); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}
	return nil
}

// 推断任务调度信息
func (j *JobHelper) inferSchedule(job gocron.Job, nextRuns []time.Time) (string, string) {
	name := job.Name()

	// try to infer from job name
	if len(name) > 0 {
		// check for common patterns in name
		if strings.Contains(name, "every") || strings.Contains(name, "interval") {
			if strings.Contains(name, "10s") || strings.Contains(name, "10-s") {
				return "Every 10 seconds", "Duration: 10s"
			}
			if strings.Contains(name, "5s") {
				return "Every 5 seconds", "Duration: 5s"
			}
			if strings.Contains(name, "minute") {
				return "Every minute", "Duration: 1m"
			}
		}

		if strings.Contains(name, "cron") {
			return "Cron schedule", "Cron: * * * * *"
		}

		if strings.Contains(name, "daily") {
			return "Daily", "Daily schedule"
		}

		if strings.Contains(name, "weekly") {
			return "Weekly", "Weekly: Mon, Wed, Fri"
		}

		if strings.Contains(name, "random") {
			return "Random interval", "Random: 5-15s"
		}

		if strings.Contains(name, "singleton") {
			return "Every 5 seconds (singleton)", "Duration: 5s, Mode: Singleton"
		}

		if strings.Contains(name, "limited") {
			return "Every 7 seconds (limited)", "Duration: 7s, Max runs: 3"
		}

		if strings.Contains(name, "parameter") {
			return "Every 12 seconds", "Duration: 12s"
		}

		if strings.Contains(name, "context") {
			return "Every 8 seconds", "Duration: 8s"
		}

		if strings.Contains(name, "one-time") || strings.Contains(name, "onetime") {
			return "One time only", "OneTime job"
		}
	}

	// try to infer from next runs interval
	if len(nextRuns) >= 2 {
		interval := nextRuns[1].Sub(nextRuns[0])

		if interval < time.Minute {
			seconds := int(interval.Seconds())
			return fmt.Sprintf("Every %d seconds", seconds), fmt.Sprintf("Duration: %ds", seconds)
		}
		if interval < time.Hour {
			minutes := int(interval.Minutes())
			return fmt.Sprintf("Every %d minutes", minutes), fmt.Sprintf("Duration: %dm", minutes)
		}
		if interval < 24*time.Hour {
			hours := int(interval.Hours())
			return fmt.Sprintf("Every %d hours", hours), fmt.Sprintf("Duration: %dh", hours)
		}
		days := int(interval.Hours() / 24)
		return fmt.Sprintf("Every %d days", days), fmt.Sprintf("Duration: %dd", days)
	}

	return "Scheduled", "Custom schedule"
}

func (j *JobHelper) SlogWith(logger *slog.Logger, jobID string, jobName string) *slog.Logger {
	return logger.With(
		slog.String("任务编号", jobID),
		slog.String("任务名称", jobName),
	)
}

// [周期性]任务开始前
func (j *JobHelper) BeforePeriodicListener() gocron.EventListener {
	return gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
		// 日志
		logger := j.SlogWith(j.logger, jobID.String(), jobName)
		// 获取任务
		jobInfo, err := j.GetJob(jobID.String())
		if err != nil {
			logger.Info("[周期性] 开始执行")
		} else {
			logger.Info("[周期性] 开始执行",
				slog.Time("上次运行时间", jobInfo.LastRun),
				slog.Time("下次运行时间", jobInfo.NextRun),
			)
		}
	})
}

// [周期性]任务结束后
func (j *JobHelper) AfterPeriodicListener() gocron.EventListener {
	return gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
		// 日志
		logger := j.SlogWith(j.logger, jobID.String(), jobName)
		// 获取任务
		jobInfo, err := j.GetJob(jobID.String())
		if err != nil {
			logger.Info("[周期性] 执行结束")
		} else {
			logger.Info("[周期性] 执行结束",
				slog.Time("上次运行时间", jobInfo.LastRun),
				slog.Time("下次运行时间", jobInfo.NextRun),
			)
		}
	})
}

// [周期性]任务异常时
func (j *JobHelper) PanicPeriodicListener() gocron.EventListener {
	return gocron.AfterJobRunsWithPanic(func(jobID uuid.UUID, jobName string, recoverData any) {
		// 日志
		logger := j.SlogWith(j.logger, jobID.String(), jobName)
		// 获取任务
		jobInfo, err := j.GetJob(jobID.String())
		if err != nil {
			logger.Error("[周期性] 执行异常",
				slog.Any("err", err),
			)
		}
		if jobInfo != nil {
			logger.Error("[周期性] 执行异常",
				slog.Time("上次运行时间", jobInfo.LastRun),
				slog.Time("下次运行时间", jobInfo.NextRun),
				slog.Any("err", recoverData),
			)
		}
	})
}

// [一次性]任务开始前
func (j *JobHelper) BeforeDisposableListener() gocron.EventListener {
	return gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
		// 日志
		logger := j.SlogWith(j.logger, jobID.String(), jobName)
		logger.Info("[一次性] 开始执行")
	})
}

// [一次性]任务结束后
func (j *JobHelper) AfterDisposableListener() gocron.EventListener {
	return gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
		// 日志
		logger := j.SlogWith(j.logger, jobID.String(), jobName)
		logger.Info("[一次性] 执行结束")
		// 删除任务
		err := j.DeleteJob(jobID.String())
		if err != nil {
			logger.Error("[一次性] 删除任务失败",
				slog.Any("err", err),
			)
		}
	})
}

// [一次性]任务异常时
func (j *JobHelper) PanicDisposableListener() gocron.EventListener {
	return gocron.AfterJobRunsWithPanic(func(jobID uuid.UUID, jobName string, recoverData any) {
		// 日志
		logger := j.SlogWith(j.logger, jobID.String(), jobName)
		logger.Error("[一次性] 执行异常",
			slog.Any("err", recoverData),
		)
		// 删除任务
		err := j.DeleteJob(jobID.String())
		if err != nil {
			logger.Error("[一次性] 删除任务失败",
				slog.Any("err", err),
			)
		}
	})
}

func (j *JobHelper) humanizeTimeList(times []time.Time) []string {
	var humanizedTimes []string
	for _, t := range times {
		humanizedTimes = append(humanizedTimes, carbon.NewCarbon(t).DiffForHumans())
	}
	return humanizedTimes
}
