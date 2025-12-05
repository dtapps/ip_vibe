package backend

import (
	"context"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"

	"github.com/go-co-op/gocron/v2"
)

// IP更新
type IpUpdateService struct {
	db     *ent.Client      // 数据库客户端
	cron   gocron.Scheduler // 定时任务调度器
	dnsJob gocron.Job       // DNS刷新任务
	cdnJob gocron.Job       // CDN刷新任务
	log    *slog.Logger     // 日志记录器
}

// NewIpUpdateService 创建 IP 更新服务
func NewIpUpdateService(db *ent.Client, cron gocron.Scheduler) *IpUpdateService {
	return &IpUpdateService{
		db:   db,
		cron: cron,
		log: slog.With(
			slog.String("component", "ip_update"),
		),
	}
}

// Close 关闭服务，释放资源
func (s *IpUpdateService) Close() {
	if s.cron != nil && s.dnsJob != nil {
		_ = s.cron.RemoveJob(s.dnsJob.ID())
	}
	if s.cron != nil && s.cdnJob != nil {
		_ = s.cron.RemoveJob(s.cdnJob.ID())
	}
}

// 启动 DNS刷新 任务，延迟 1 分钟添加，确保 CDN 先跑
func (s *IpUpdateService) StartDnsJob(ctx context.Context) (err error) {
	// 判断是否存在DNS任务
	if s.dnsJob != nil {
		return nil
	}
	// 延时1分钟
	time.Sleep(1 * time.Minute)
	// 每10分钟执行一次
	s.dnsJob, err = s.cron.NewJob(
		gocron.DurationJob(10*time.Minute),                   // 每10分钟执行一次
		gocron.NewTask(s.runDnsTask),                         // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("DNS刷新"),                             // 任务名称
		gocron.WithTags("dns_update"),                        // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

// 启动 CDN刷新 任务
func (s *IpUpdateService) StartCdnJob(ctx context.Context) (err error) {
	// 判断是否存在CDN任务
	if s.cdnJob != nil {
		return nil
	}
	// 延时1分钟
	// time.Sleep(1 * time.Minute)
	// 每10分钟执行一次
	s.cdnJob, err = s.cron.NewJob(
		gocron.DurationJob(10*time.Minute),                   // 每10分钟执行一次
		gocron.NewTask(s.runCdnTask),                         // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("CDN刷新"),                             // 任务名称
		gocron.WithTags("cdn_update"),                        // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}
