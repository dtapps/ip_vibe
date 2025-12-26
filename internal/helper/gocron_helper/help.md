gocron.DurationJob(
10\*time.Hour, // 每 10 小时执行一次
),

gocron.DurationRandomJob(
8*time.Minute, // 最小间隔 8 分钟
12*time.Minute, // 最大间隔 12 分钟
),

gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式

gocron.WithStartAt(gocron.WithStartImmediately()), // 立即开始

gocron.WithStartAt(gocron.WithStartDateTime(time.Now().Add(1\*time.Minute))), // 首次延迟 1 分钟

gocron.WithName(s.taskName), // 任务名称

gocron.WithEventListeners(
s.jobHelper.PanicPeriodicListener(),
s.jobHelper.BeforePeriodicListener(),
s.jobHelper.AfterPeriodicListener(),
), // 事件监听
