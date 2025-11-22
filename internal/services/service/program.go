package service

import (
	"context"
	"fmt"
	"time"
	"vibe_ip/internal/helper/config_helper"
	"vibe_ip/internal/helper/ip_helper"

	"github.com/go-co-op/gocron/v2"
	"github.com/kardianos/service"
)

type program struct {
	logger service.Logger
}

func NewProgram() *program {
	return &program{}
}

func (p *program) Start(s service.Service) error {
	// Logger an't be obtained until Start is called.
	errs := make(chan error, 5)
	logger, err := s.Logger(errs)
	if err != nil {
		return err
	}
	p.logger = logger

	if service.Interactive() {
		p.logger.Info("在交互模式下运行")
	} else {
		p.logger.Info("在服务模式下运行")
	}
	go p.run()
	return nil
}

func (p *program) run() {
	p.logger.Infof("服务正在运行...")
	defer p.logger.Infof("服务已停止")

	// 初始化调度器
	s, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local),
		gocron.WithLogger(gocron.NewLogger(gocron.LogLevelDebug)),
	)
	if err != nil {
		p.logger.Error(err)
		return
	}
	defer s.Shutdown()

	_, err = s.NewJob(
		gocron.DurationJob(time.Duration(10)*time.Minute), // 任务周期
		gocron.NewTask(func() {

			var logPrefix = "cron"

			// 查询自动更新的提供商配置
			targetProviders, err := config_helper.QueryCronProviders()
			if err != nil {
				p.logger.Errorf("[%s]读取配置失败: %s \n",
					logPrefix,
					err,
				)
				return
			}
			if len(targetProviders) == 0 {
				p.logger.Infof("[%s]无需自动更新 \n",
					logPrefix,
				)
				return
			}
			for _, targetProvider := range targetProviders {

				var logPrefix = fmt.Sprintf("cron-%s-%s", targetProvider.Type, targetProvider.ID)

				// 获取IP
				checker := ip_helper.NewIPChecker()
				ipInfo, err := checker.CheckIP(context.TODO(), targetProvider.IpType)
				if err != nil {
					p.logger.Errorf("[%s]获取IP失败: %s \n",
						logPrefix,
						err,
					)
					continue
				}

				if ipInfo.IP == "" {
					p.logger.Infof("[%s]找不到IP",
						logPrefix,
					)
					continue
				}

			}

		}), // 任务方法
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
	)
	if err != nil {
		p.logger.Error(err)
		return
	}

	// 启动调度器
	s.Start()

	// 阻止调度器退出
	select {}

}

func (p *program) Stop(s service.Service) error {
	p.logger.Info("服务正在停止...")
	return nil
}
