package gocron_helper

import (
	"fmt"
	"strings"
	"time"
	"vibe_ip/pkg/utils"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type JobData struct {
	ID             string   `json:"id"`              // 任务序号
	Name           string   `json:"name"`            // 任务名称
	Tags           []string `json:"tags"`            // 任务标签
	LastRun        string   `json:"last_run"`        // 上次运行时间
	NextRun        string   `json:"next_run"`        // 下次运行时间
	NextRuns       []string `json:"next_runs"`       // 下次运行时间列表
	Schedule       string   `json:"schedule"`        // 任务调度表达式
	ScheduleDetail string   `json:"schedule_detail"` // 任务调度表达式详情
}

type JobHelper struct {
	Scheduler gocron.Scheduler
}

func NewJobHelper(scheduler gocron.Scheduler) *JobHelper {
	return &JobHelper{
		Scheduler: scheduler,
	}
}

// 查询全部任务
func (j *JobHelper) GetJobsData(nextRunsLimit int) (jobs []JobData) {
	if j.Scheduler == nil {
		return jobs
	}
	for _, job := range j.Scheduler.Jobs() {
		nextRun, _ := job.NextRun()
		lastRun, _ := job.LastRun()

		// get next 5 runs
		nextRuns, _ := job.NextRuns(nextRunsLimit)

		// determine schedule info based on job name patterns or intervals
		schedule, scheduleDetail := j.inferSchedule(job, nextRuns)

		jobs = append(jobs, JobData{
			ID:             job.ID().String(),
			Name:           job.Name(),
			Tags:           job.Tags(),
			LastRun:        utils.FormatTime(lastRun),
			NextRun:        utils.FormatTime(nextRun),
			NextRuns:       utils.FormatTimes(nextRuns),
			Schedule:       schedule,
			ScheduleDetail: scheduleDetail,
		})
	}
	return jobs
}

// 运行任务
func (j *JobHelper) RunJob(jobID string) error {
	if j.Scheduler == nil {
		return fmt.Errorf("scheduler is nil")
	}
	id, err := uuid.Parse(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}
	jobs := j.Scheduler.Jobs()
	for _, job := range jobs {
		if job.ID() == id {
			if err = job.RunNow(); err != nil {
				return fmt.Errorf("job execution failed: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("job not found: %w", err)
}

// 删除任务
func (j *JobHelper) DeleteJob(jobID string) error {
	if j.Scheduler == nil {
		return fmt.Errorf("scheduler is nil")
	}
	id, err := uuid.Parse(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}
	if err := j.Scheduler.RemoveJob(id); err != nil {
		return fmt.Errorf("job not found: %w", err)
	}
	return nil
}

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
