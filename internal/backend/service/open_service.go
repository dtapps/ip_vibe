package service

import (
	"context"
	"fmt"
	"time"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/deviceportforward"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/deviceipreport"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/utils/gotime"
)

// DeviceOpenService 设备服务接口
type DeviceOpenService interface {
	// DeviceReportDeviceIp 处理设备上报IP的复杂事务和异步逻辑
	DeviceReportDeviceIp(ctx context.Context, request *types.DeviceReportDeviceIpRequest) (*types.DeviceReportDeviceIpResponse, error)
	// DevicePortForwards 查询设备的端口转发配置
	DevicePortForwards(ctx context.Context, deviceCode string) ([]types.GetPortForwardResponse, error)
	// CreatePortForwardStats 创建设备端口转发统计信息
	CreatePortForwardStats(ctx context.Context, deviceCode string, portForwardID int64, req *types.CreatePortForwardStatsRequest) error
}

// DeviceOpenServiceImpl 设备服务实现
type DeviceOpenServiceImpl struct {
	db             *ent.Client          // 数据库客户端
	logDb          *ent_log.Client      // 日志数据库客户端
	webhookService *webhookService.Send // Webhook服务
	debug          bool                 // 调试模式
}

// NewDeviceOpenService 创建设备服务实例
func NewDeviceOpenService(db *ent.Client, logDb *ent_log.Client, debug bool) DeviceOpenService {
	webhookService := webhookService.NewSend(db, logDb)
	return &DeviceOpenServiceImpl{
		db:             db,
		logDb:          logDb,
		webhookService: webhookService,
		debug:          debug,
	}
}

// DeviceReportDeviceIp 处理设备上报IP
func (s *DeviceOpenServiceImpl) DeviceReportDeviceIp(ctx context.Context, request *types.DeviceReportDeviceIpRequest) (*types.DeviceReportDeviceIpResponse, error) {
	// 判断IPV4和IPV6是否为空 (保持原始逻辑)
	if request.IPV4Addr == "" && request.IPV6Addr == "" {
		return nil, fmt.Errorf("IPV4和IPV6不能同时为空")
	}

	// 异步处理所需数据
	isNotify := false
	var deviceID int64

	// 通知模板数据
	notifyData := utils.TemplateData{
		EventTitle: "IP地址更新",
		EventType:  "IP_UPDATE",
		EventTime:  time.Now().Format("2006-01-02 15:04:05"),
		EventBody:  "IP地址更新事件",
		DeviceCode: request.DeviceCode,
		DeviceName: request.DeviceName,
		IPv4Addr:   request.IPV4Addr,
		IPv6Addr:   request.IPV6Addr,
	}

	msg := ""

	// 开始事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}

	// 辅助函数，用于简化事务回滚和错误返回
	rollbackAndError := func(msg string, err error) (*types.DeviceReportDeviceIpResponse, error) {
		_ = tx.Rollback()
		return nil, fmt.Errorf(msg+": %w", err)
	}

	// 检查设备是否存在
	firstDevice, err := tx.Device.Query().
		Where(device.CodeEQ(request.DeviceCode)).
		First(ctx)

	if err != nil {
		if !ent.IsNotFound(err) {
			return rollbackAndError("查询设备失败", err)
		}

		// 如果设备不存在：创建设备 (保持原始逻辑)
		if request.DeviceName == "" {
			request.DeviceName = request.DeviceCode
		}

		saveDevice, err := tx.Device.Create().
			SetCode(request.DeviceCode).
			SetName(request.DeviceName).
			SetIPV4(request.IPV4Addr).
			SetIPV6(request.IPV6Addr).
			SetLastReportedAt(time.Now()).
			SetLastHeartbeatAt(time.Now()).
			Save(ctx)

		if err != nil {
			if s.debug {
				fmt.Printf("创建设备: %s\n", err.Error())
			}
			return rollbackAndError("创建设备失败", err)
		}

		deviceID = saveDevice.ID
		isNotify = true
		msg = "创建设备IP成功"

	} else {
		// 设备存在：检查 IP 是否变化 (保持原始逻辑)
		deviceID = firstDevice.ID
		request.DeviceName = firstDevice.Name // 使用数据库中的设备名进行通知

		if firstDevice.IPV4 != request.IPV4Addr || firstDevice.IPV6 != request.IPV6Addr {
			// IP 发生变化
			isNotify = true

			// 更新设备IP和心跳时间
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetIPV4(request.IPV4Addr).
				SetIPV6(request.IPV6Addr).
				SetLastReportedAt(time.Now()).
				SetLastHeartbeatAt(time.Now())

			// 如果设备名有更新
			if request.DeviceName != "" && firstDevice.Name != request.DeviceName {
				updateDevice.SetName(request.DeviceName)
				notifyData.DeviceName = request.DeviceName // 确保异步通知使用新的设备名
			}

			err = updateDevice.Exec(ctx)
			if err != nil {
				if s.debug {
					fmt.Printf("更新设备IP: %s\n", err.Error())
				}
				return rollbackAndError("更新设备IP失败", err)
			}

			msg = "更新设备IP成功"

		} else {
			// IP 未改变：仅更新心跳时间
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetLastHeartbeatAt(time.Now())

			// 如果设备名有更新
			if request.DeviceName != "" && firstDevice.Name != request.DeviceName {
				updateDevice.SetName(request.DeviceName)
				notifyData.DeviceName = request.DeviceName
			}

			err = updateDevice.Exec(ctx)
			if err != nil {
				if s.debug {
					fmt.Printf("更新设备心跳时间: %s\n", err.Error())
				}
				return rollbackAndError("更新设备心跳时间失败", err)
			}
			msg = "设备IP未改变，不添加IP报告"
		}
	}

	// 提交核心事务：立即释放独占锁，这是解决卡死的关键！ (保持原始事务逻辑)
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// --- 异步处理非核心写入和通知 ---
	// 使用独立的 Go Routine 和 Background Context，不阻塞 HTTP 响应
	go func(asyncCtx context.Context, s *DeviceOpenServiceImpl) {

		// 记录 IP 报告并发送通知 (保持原始逻辑)
		if isNotify {
			// 创建设备IP报告
			err := s.logDb.DeviceIpReport.Create().
				SetDeviceID(deviceID).
				SetIPV4(request.IPV4Addr).
				SetIPV6(request.IPV6Addr).
				SetReportType(deviceipreport.ReportType(request.ReportType)).
				Exec(asyncCtx)
			if err != nil {
				// 异步写入失败，仅打印日志
				fmt.Printf("[异步写入失败] 创建设备IP报告: %s\n", err.Error())
			}

			// 发送通知 (使用注入的 webhookService)
			_ = s.webhookService.Send(asyncCtx, webhooksubscription.ObjTypeDevice, deviceID, notifyData)
		}
	}(context.WithoutCancel(ctx), s)

	return &types.DeviceReportDeviceIpResponse{
		Msg: msg,
	}, nil
}

// DevicePortForwards 查询设备的端口转发
func (s *DeviceOpenServiceImpl) DevicePortForwards(ctx context.Context, deviceCode string) ([]types.GetPortForwardResponse, error) {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.CodeEQ(deviceCode)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	portForwards := make([]types.GetPortForwardResponse, 0)

	// 查询端口转发
	allPortForward, err := firstDevice.QueryPortForwards().
		Where(deviceportforward.IsActiveEQ(true)).
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}

	// 组装响应数据
	for _, item := range allPortForward {
		portForwards = append(portForwards, types.GetPortForwardResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			LocalPort:   item.LocalPort,
			RemoteHost:  item.RemoteHost,
			RemotePort:  item.RemotePort,
			Protocols:   item.Protocols,
			AutoForward: item.AutoForward,
			IsActive:    item.IsActive,
		})
	}

	return portForwards, nil
}

// CreatePortForwardStats 创建设备端口转发统计信息
func (s *DeviceOpenServiceImpl) CreatePortForwardStats(ctx context.Context, deviceCode string, portForwardID int64, request *types.CreatePortForwardStatsRequest) error {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.CodeEQ(deviceCode)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询设备失败: %w", err)
	}

	// 查询端口转发
	firstPortForward, err := firstDevice.QueryPortForwards().
		Where(deviceportforward.IDEQ(portForwardID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询端口转发失败: %w", err)
	}

	// 创建端口转发统计信息
	err = s.logDb.PortForwardStats.Create().
		SetDeviceID(firstDevice.ID).
		SetPortForwardID(firstPortForward.ID).
		SetStartTime(gotime.SetCurrentUnix(request.StartTime).Time).
		SetActiveConn(request.ActiveConn).
		SetTotalConn(request.TotalConn).
		SetBytesIn(request.BytesIn).
		SetBytesOut(request.BytesOut).
		SetUDPPacketsIn(request.UDPPacketsIn).
		SetUDPPacketsOut(request.UDPPacketsOut).
		SetErrors(request.Errors).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建设备端口转发统计信息失败: %w", err)
	}

	return nil
}
