package gocron_helper

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gostring"
)

// ParseTaskTime 解析任务时间
func ParseTaskTime(taskTime string, timeUnit time.Duration) (jobDuration gocron.JobDefinition, isStart bool, err error) {
	taskTime = gostring.TrimSpace(taskTime)
	if taskTime == "" {
		return nil, false, fmt.Errorf("时间未配置")
	}
	if gostring.Contains(taskTime, "-") {
		// 解析时间范围
		times := gostring.Split(taskTime, "-")
		if len(times) != 2 {
			return nil, false, fmt.Errorf("时间格式错误：%s", taskTime)
		}
		jobDuration = gocron.DurationRandomJob(
			time.Duration(cast.ToInt64(times[0]))*timeUnit, // 最小间隔
			time.Duration(cast.ToInt64(times[1]))*timeUnit, // 最大间隔
		)
	} else {
		jobDuration = gocron.DurationJob(
			time.Duration(cast.ToInt64(taskTime)) * timeUnit,
		)
	}
	isStart = true
	return
}

// ParseTaskTimeInterval 解析任务时间间隔
// 如果 maxTime = 0 则表示只有 minTime 这个时间间隔
func ParseTaskTimeInterval(taskTime string) (minTime int64, maxTime int64) {
	taskTime = gostring.TrimSpace(taskTime)
	if taskTime == "" {
		return 0, 0
	}
	if gostring.Contains(taskTime, "-") {
		// 解析时间范围
		times := gostring.Split(taskTime, "-")
		if len(times) != 2 {
			return 0, 0
		}
		minTime = cast.ToInt64(times[0])
		maxTime = cast.ToInt64(times[1])
	} else {
		minTime = cast.ToInt64(taskTime)
		maxTime = 0
	}
	return
}
