package services

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/ent/ipgroup"
	"vibe_ip/ent/operationlog"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
)

// IP上报
type IpReportService struct {
	db   *ent.Client      // 数据库客户端
	cron gocron.Scheduler // 定时任务调度器
	job  gocron.Job       // IP上报任务
	log  *slog.Logger     // 日志记录器
}

func NewIpReportService(db *ent.Client, cron gocron.Scheduler) *IpReportService {
	return &IpReportService{
		db:   db,
		cron: cron,
		log: slog.With(
			slog.String("component", "ip_report"),
		),
	}
}

func (s *IpReportService) StartJob(ctx context.Context) (err error) {
	// 每8到12分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			8*time.Minute,  // 最小间隔
			12*time.Minute, // 最大间隔
		),
		gocron.NewTask(s.runTask),                            // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("IP上报"),                              // 任务名称
		gocron.WithTags("ip_report"),                         // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

// 全局互斥锁
var globalIpReportMutex sync.Mutex

func (s *IpReportService) runTask() {
	globalIpReportMutex.Lock()
	defer globalIpReportMutex.Unlock()

	var err error
	ctx := context.Background()

	// 请求参数
	params := types.ReportDeviceIpRequest{
		DeviceCode: global.GvaEnv.DEVICE_CODE,
		DeviceName: global.GvaEnv.DEVICE_NAME,
	}

	// 检查设备代码是否为空
	if params.DeviceCode == "" {
		// 获取设备ID
		params.DeviceCode, err = machineid.ID()
		if err != nil {
			s.log.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}
	// 检查设备名称是否为空
	if params.DeviceName == "" {
		params.DeviceName = "本设备"
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
	// 判断IPV4和IPV6是否为空
	if params.IPV4Addr == "" && params.IPV6Addr == "" {
		s.log.ErrorContext(ctx, "IPV4和IPV6地址都为空")
		return
	}

	// 是否通知
	isNotify := false
	notifyData := utils.TemplateData{
		EventType:  "IP_UPDATE",
		EventTime:  time.Now().Format("2006-01-02 15:04:05"),
		EventBody:  "IP地址更新事件",
		DeviceCode: params.DeviceCode,
		DeviceName: params.DeviceName,
		IPv4Addr:   params.IPV4Addr,
		IPv6Addr:   params.IPV6Addr,
	}

	// 描述
	msg := ""

	// 开始事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "开始事务失败",
			slog.Any("err", err),
		)
		return
	}

	// 检查设备是否存在
	firstDevice, err := tx.Device.Query().
		Where(device.CodeEQ(params.DeviceCode)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			_ = tx.Rollback()
			s.log.ErrorContext(ctx, "查询设备失败",
				slog.Any("err", err),
			)
			return
		}

		// 创建设备
		saveDevice, err := tx.Device.Create().
			SetCode(params.DeviceCode).
			SetName(params.DeviceName).
			SetIPV4(params.IPV4Addr).
			SetIPV6(params.IPV6Addr).
			SetLastReportedAt(time.Now()).
			SetLastHeartbeatAt(time.Now()).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			s.log.ErrorContext(ctx, "创建设备失败",
				slog.Any("err", err),
			)
			return
		}
		params.ID = saveDevice.ID
		// 记录操作日志
		_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "自动创建设备")

		// 创建设备IP报告
		saveDeviceIPReport, err := tx.DeviceIpReport.Create().
			SetDevice(saveDevice).
			SetIPV4(params.IPV4Addr).
			SetIPV6(params.IPV6Addr).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			s.log.ErrorContext(ctx, "创建设备IP报告失败",
				slog.Any("err", err),
			)
			return
		}
		// 记录操作日志
		_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDeviceIPReport.ID, operationlog.ObjTypeDeviceIPReport, operationlog.StatusSuccess, "自动创建设备IP报告")

		// 创建设备IP组
		saveDeviceIpGroup, err := tx.IpGroup.Create().
			AddDevices(saveDevice).
			SetName(params.DeviceName).
			SetDescription(params.DeviceCode).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			s.log.ErrorContext(ctx, "创建IP组失败",
				slog.Any("err", err),
			)
			return
		}
		// 记录操作日志
		_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDeviceIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "自动创建设备IP组")

		isNotify = true

		msg = "创建设备IP成功"
	} else {
		params.ID = firstDevice.ID
		params.DeviceName = firstDevice.Name
		if firstDevice.IPV4 != params.IPV4Addr || firstDevice.IPV6 != params.IPV6Addr {
			// 更新设备IP
			isUpdateDeviceName := false
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetIPV4(params.IPV4Addr).
				SetIPV6(params.IPV6Addr).
				SetLastReportedAt(time.Now()).
				SetLastHeartbeatAt(time.Now())
			if params.DeviceName != "" && firstDevice.Name != params.DeviceName {
				isUpdateDeviceName = true
				updateDevice.SetName(params.DeviceName)
			}
			err = updateDevice.Exec(ctx)
			if err != nil {
				_ = tx.Rollback()
				s.log.ErrorContext(ctx, "更新设备IP失败",
					slog.Any("err", err),
				)
				return
			}
			// 记录操作日志
			_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataUpdate(ctx, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "自动更新设备")

			// 创建设备IP报告
			saveDeviceIPReport, err := tx.DeviceIpReport.Create().
				SetDevice(firstDevice).
				SetIPV4(params.IPV4Addr).
				SetIPV6(params.IPV6Addr).
				Save(ctx)
			if err != nil {
				_ = tx.Rollback()
				s.log.ErrorContext(ctx, "创建设备IP报告失败",
					slog.Any("err", err),
				)
				return
			}
			// 记录操作日志
			_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDeviceIPReport.ID, operationlog.ObjTypeDeviceIPReport, operationlog.StatusSuccess, "自动创建设备IP报告")
			if isUpdateDeviceName {
				err = tx.IpGroup.Update().
					Where(ipgroup.DescriptionEQ(params.DeviceCode)).
					SetName(params.DeviceName).
					Exec(ctx)
				if err != nil {
					_ = tx.Rollback()
					s.log.ErrorContext(ctx, "更新IP组名称失败",
						slog.Any("err", err),
					)
					return
				}
				// 记录操作日志
				// 	_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataUpdate(ctx, saveDeviceIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "自动创建设备IP组")
			}

			isNotify = true

			msg = "更新设备IP成功"
		} else {
			// 更新设备心跳时间
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetLastHeartbeatAt(time.Now())
			if params.DeviceName != "" && firstDevice.Name != params.DeviceName {
				updateDevice.SetName(params.DeviceName)
			}
			err = updateDevice.Exec(ctx)
			if err != nil {
				_ = tx.Rollback()
				s.log.ErrorContext(ctx, "更新设备心跳时间失败",
					slog.Any("err", err),
				)
				return
			}
			msg = "设备IP未改变，不添加IP报告"
		}
	}
	_ = tx.Commit()

	nextRun, err := s.job.NextRun()
	if err == nil {
		s.log.InfoContext(ctx, "IP上报成功",
			slog.String("msg", msg),
			slog.String("device_code", params.DeviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)
	} else {
		s.log.InfoContext(ctx, "IP上报成功",
			slog.String("msg", msg),
			slog.String("device_code", params.DeviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
		)
	}

	// 发送通知
	if isNotify {
		go func() {
			_ = NewWebhookService(global.GvaDb).Send(context.Background(), webhooksubscription.ObjTypeDevice, params.ID, notifyData)
		}()
	}

}
