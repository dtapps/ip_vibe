package task

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/thrift_generated/open"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gostring"
)

// PortForwardThriftTask  端口转发任务
type PortForwardThriftTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 端口转发任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	jobOne    sync.Once                // 确保只创建一次任务
	logger    *slog.Logger             // 日志记录器
	config    *PortForwardTaskConfig   // 配置

	service      *service.PortForwardServer // 端口转发服务
	forwardAgent *forward.Agent             // 端口转发器

	clientBase        *open.Base              // Thrift 客户端基础
	thriftClient      *thrift.TStandardClient // Thrift 客户端
	openServiceClient *open.OpenServiceClient // Thrift 业务服务
}

// NewPortForwardThriftTask  创建端口转发实例
func NewPortForwardThriftTask(cron gocron.Scheduler, logger *slog.Logger, service *service.PortForwardServer, forwardAgent *forward.Agent, thriftClient *thrift.TStandardClient, clientBase *open.Base) *PortForwardThriftTask {
	return &PortForwardThriftTask{
		cron:              cron,
		taskName:          "Thrift 端口转发",
		jobHelper:         gocron_helper.NewJobHelper(cron, logger),
		logger:            logger.With("component", "port-forward-thrift"),
		service:           service,
		forwardAgent:      forwardAgent,
		clientBase:        clientBase,
		thriftClient:      thriftClient,
		openServiceClient: open.NewOpenServiceClient(thriftClient),
	}
}

// Close 关闭端口转发实例，释放资源
func (s *PortForwardThriftTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// SetConfig 设置任务配置
func (s *PortForwardThriftTask) SetConfig(config *PortForwardTaskConfig) *PortForwardThriftTask {
	s.config = config
	return s
}

// parseTaskTime 解析任务时间
func (s *PortForwardThriftTask) parseTaskTime(taskTime string, timeUnit time.Duration) (jobDuration gocron.JobDefinition, isStart bool, err error) {
	if taskTime == "" {
		return nil, false, fmt.Errorf("时间未配置")
	}
	if strings.Contains(taskTime, "-") {
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

// StartPeriodicTask 启动周期任务
func (s *PortForwardThriftTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 获取设置
	getSetting, err := s.openServiceClient.GetSetting(ctx, s.clientBase)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取设置",
			slog.Any("err", err),
		)
		return err
	}
	// 解析设置的时间
	jobDuration, isStart, err := s.parseTaskTime(
		getSetting.PortForwardReloadTime,
		time.Minute,
	)
	if !isStart {
		return nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "解析设置的时间",
			slog.Any("err", err),
		)
		return err
	}
	// 确保只创建一次任务
	s.jobOne.Do(func() {
		// 创建周期任务
		s.job, err = s.cron.NewJob(
			jobDuration,
			gocron.NewTask(s.RunPeriodicTask), // 执行任务
			gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
			gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
			gocron.WithName(s.taskName),                          // 任务名称
			gocron.WithEventListeners(
				s.jobHelper.PanicPeriodicListener(),
				s.jobHelper.BeforePeriodicListener(),
				s.jobHelper.AfterPeriodicListener(),
			), // 事件监听
		)
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "创建定时任务失败",
			slog.Any("err", err),
		)
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, s.taskName+"任务已启动",
		slog.String("执行间隔", getSetting.PortForwardReloadTime),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *PortForwardThriftTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 检查 forwardAgent 是否初始化
	if s.forwardAgent == nil {
		logger.ErrorContext(ctx, "forwardAgent 未初始化")
		return
	}

	// 获取设备数据
	deviceData, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		logger.ErrorContext(ctx, "获取设备数据失败",
			slog.String("设备编号", s.config.DeviceCode),
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger = logger.With(
		slog.String("设备编号", deviceData.DeviceCode),
		slog.String("设备名称", deviceData.DeviceName),
	)

	// 获取数据
	list, err := s.getApi(ctx, logger, deviceData)
	if err != nil {
		return
	}

	// 转换类型：types.GetPortForwardResponse -> forward.Rule
	var forwardRules []forward.Rule
	for _, item := range list {

		// 构建监听地址
		listenAddr := ":" + string(rune(item.LocalPort))

		// 更准确的构建方式：
		listenAddr = net.JoinHostPort("", fmt.Sprintf("%d", item.LocalPort))

		// 构建目标地址
		targetAddr := net.JoinHostPort(item.RemoteHost, fmt.Sprintf("%d", item.RemotePort))

		rule := forward.Rule{
			ID:         item.ID,    // 唯一标识
			ListenAddr: listenAddr, // 本地监听地址
			TargetAddr: targetAddr, // 目标地址
			Protocols:  item.Protocols,
		}
		forwardRules = append(forwardRules, rule)
	}

	// 同步端口转发规则，热更新
	s.forwardAgent.Sync(ctx, forwardRules)
	logger.InfoContext(ctx, "同步端口转发规则，热更新",
		slog.Int("规则数量", len(forwardRules)),
	)
}

// getApi 请求
func (s *PortForwardThriftTask) getApi(ctx context.Context, logger *slog.Logger, deviceData *service.GetDeviceDataResult) (list []*PortForwardTaskApiList, err error) {

	// 构建Thrift请求
	req := &open.GetDevicePortForwardsRequest{
		DeviceCode: deviceData.DeviceCode,
	}

	// 调用Thrift服务
	resp, err := s.openServiceClient.GetDevicePortForwards(ctx, req, s.clientBase)
	if err != nil {
		logger.ErrorContext(ctx, s.taskName+" 调用失败",
			slog.Any("err", err),
		)
		return nil, fmt.Errorf(s.taskName+" 调用失败: %w", err)
	}

	// 转换响应
	for _, rule := range resp.Rules {
		list = append(list, &PortForwardTaskApiList{
			ID:         cast.ToString(rule.ID),
			LocalPort:  cast.ToInt(rule.LocalPort),
			RemoteHost: rule.RemoteHost,
			RemotePort: cast.ToInt(rule.RemotePort),
			Protocols:  rule.Protocols,
		})
	}

	// 记录响应
	logger.InfoContext(ctx, s.taskName+" 响应",
		slog.Any("响应", resp.String()),
	)

	return list, nil
}
