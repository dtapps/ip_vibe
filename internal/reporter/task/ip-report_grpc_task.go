package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// IPReportGRPCTask IP上报任务
type IPReportGRPCTask struct {
	cron   gocron.Scheduler // 定时任务调度器
	job    gocron.Job       // IP上报任务
	logger *slog.Logger     // 日志记录器

	clientTarget   string                           // gRPC 客户端目标地址
	clientOpts     []grpc.DialOption                // gRPC 客户端连接选项
	client         *grpc.ClientConn                 // gRPC 客户端连接
	ipReportClient grpc_generated.OpenServiceClient // gRPC 业务服务客户端
}

// NewIPReportGRPCTask 创建一个新的服务实例
func NewIPReportGRPCTask(cron gocron.Scheduler) *IPReportGRPCTask {
	return &IPReportGRPCTask{
		cron:   cron,
		logger: slog.With("component", "ip-report"),
		clientOpts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

// Close 关闭服务，释放资源
func (s *IPReportGRPCTask) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetBaseURL 设置基础URL
func (s *IPReportGRPCTask) SetBaseURL(url string) *IPReportGRPCTask {
	s.clientTarget = url
	return s
}

// 设置基本认证
func (s *IPReportGRPCTask) SetBasicAuth(username, password string) *IPReportGRPCTask {
	if username != "" && password != "" {
		s.clientOpts = append(
			s.clientOpts,
			grpc.WithPerRPCCredentials(&grpc_credential.BasicAuth{
				Username: username,
				Password: password,
			}),
		)
	}
	return s
}

// StartPeriodicTask 启动任务
func (s *IPReportGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 连接gRPC服务器
	s.client, err = grpc.NewClient(s.clientTarget, s.clientOpts...)
	if err != nil {
		return fmt.Errorf("连接gRPC服务器失败: %w", err)
	}

	// 注册业务服务
	s.ipReportClient = grpc_generated.NewOpenServiceClient(s.client)

	// 每8到12分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			8*time.Minute,  // 最小间隔
			12*time.Minute, // 最大间隔
		),
		gocron.NewTask(s.RunPeriodicTask),                    // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("gRPC IP上报"),                         // 任务名称
	)
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, "IP上报gRPC任务已启动",
		slog.String("interval", "8-12分钟随机"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *IPReportGRPCTask) RunPeriodicTask() {

	var err error
	ctx := context.Background()

	// 请求参数
	deviceCode := global.GvaEnv.DEVICE_CODE
	params := types.DeviceReportDeviceIpRequest{
		DeviceName: global.GvaEnv.DEVICE_NAME,
	}
	// 检查设备代码是否为空
	if deviceCode == "" {
		// 获取设备ID
		deviceCode, err = machineid.ID()
		if err != nil {
			s.logger.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}

	// 获取设备IP
	iPHelper := ip_helper.NewIPHelper()
	ipv4, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV4)
	if err == nil {
		params.IPV4Addr = ipv4.IP
	}
	ipv6, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV6)
	if err == nil {
		params.IPV6Addr = ipv6.IP
	}
	s.logger.InfoContext(ctx, "IP上报开始",
		slog.String("device_code", deviceCode),
		slog.String("device_name", params.DeviceName),
		slog.String("ipv4_addr", params.IPV4Addr),
		slog.String("ipv6_addr", params.IPV6Addr),
	)
	err = s.reportIP(ctx, deviceCode, params)
	if err != nil {
		s.logger.ErrorContext(ctx, "IP上报失败",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.Any("err", err),
		)
		return
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		s.logger.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)
	} else {
		s.logger.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
		)
	}
}

// reportIP 发送IP上报请求
func (s *IPReportGRPCTask) reportIP(ctx context.Context, deviceCode string, params types.DeviceReportDeviceIpRequest) (err error) {

	// 构建gRPC请求
	req := &grpc_generated.ReportDeviceIPRequest{
		DeviceCode: deviceCode,
		DeviceName: params.DeviceName,
		Ipv4Addr:   params.IPV4Addr,
		Ipv6Addr:   params.IPV6Addr,
	}

	// 调用gRPC服务
	resp, err := s.ipReportClient.ReportDeviceIP(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC调用失败: %w", err)
	}

	// 记录响应
	s.logger.InfoContext(ctx, "gRPC IP上报响应",
		slog.Any("response", resp.String()),
	)

	return nil
}
