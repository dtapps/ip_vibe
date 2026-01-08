package task

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/reporter/service"
	"vibe_ip/internal/thrift_generated/open"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/go-co-op/gocron/v2"
)

// IPReportThriftTask IP上报任务
type IPReportThriftTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // IP上报任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	jobOne    sync.Once                // 确保只创建一次任务
	logger    *slog.Logger             // 日志记录器
	config    *IPReportTaskConfig      // 配置

	service *service.IPReportService // IP报告服务

	clientBase        *open.Base              // Thrift 客户端基础
	thriftClient      *thrift.TStandardClient // Thrift 客户端
	openServiceClient *open.OpenServiceClient // Thrift 业务服务
}

// NewIPReportThriftTask 创建IP上报实例
func NewIPReportThriftTask(cron gocron.Scheduler, logger *slog.Logger, service *service.IPReportService, thriftClient *thrift.TStandardClient, clientBase *open.Base) *IPReportThriftTask {
	return &IPReportThriftTask{
		cron:              cron,
		taskName:          "Thrift IP上报",
		jobHelper:         gocron_helper.NewJobHelper(cron, logger),
		logger:            logger.With("component", "ip-report-thrift"),
		service:           service,
		clientBase:        clientBase,
		thriftClient:      thriftClient,
		openServiceClient: open.NewOpenServiceClient(thriftClient),
	}
}

// Close 关闭IP上报实例，释放资源
func (s *IPReportThriftTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// SetConfig 设置任务配置
func (s *IPReportThriftTask) SetConfig(config *IPReportTaskConfig) *IPReportThriftTask {
	s.config = config
	return s
}

// StartPeriodicTask 启动周期任务
func (s *IPReportThriftTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 确保只创建一次任务
	s.jobOne.Do(func() {
		// 创建周期任务
		s.job, err = s.cron.NewJob(
			gocron.DurationRandomJob(
				8*time.Minute,  // 最小间隔 8 分钟
				12*time.Minute, // 最大间隔 12 分钟
			),
			gocron.NewTask(s.RunPeriodicTask),                    // 执行任务
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
		slog.String("执行间隔", "8-12分钟随机"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *IPReportThriftTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 获取设备数据
	deviceData, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		logger.ErrorContext(ctx, "获取设备数据失败",
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger = logger.With(
		slog.String("设备编号", deviceData.DeviceCode),
		slog.String("设备名称", deviceData.DeviceName),
		slog.String("IPv4地址", deviceData.IPV4Addr),
		slog.String("IPv6地址", deviceData.IPV6Addr),
	)

	// 上传数据
	_ = s.postApi(ctx, logger, deviceData)
}

// postApi 请求
func (s *IPReportThriftTask) postApi(ctx context.Context, logger *slog.Logger, deviceData *service.GetDeviceDataResult) (err error) {

	// 构建Thrift请求
	req := &open.ReportDeviceIPRequest{
		DeviceCode: deviceData.DeviceCode,
		DeviceName: &deviceData.DeviceName,
		Ipv4Addr:   &deviceData.IPV4Addr,
		Ipv6Addr:   &deviceData.IPV6Addr,
	}

	// 调用Thrift服务
	resp, err := s.openServiceClient.ReportDeviceIP(ctx, req, s.clientBase)
	if err != nil {
		logger.ErrorContext(ctx, s.taskName+" 调用失败",
			slog.Any("err", err),
		)
		return fmt.Errorf(s.taskName+" 调用失败: %w", err)
	}

	// 记录响应
	logger.InfoContext(ctx, s.taskName+" 响应",
		slog.Any("响应", resp.String()),
	)

	return nil
}
