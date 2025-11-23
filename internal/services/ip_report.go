package services

import (
	"context"
	"log/slog"
	"time"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
)

// IP上报
type IpReportService struct {
	db   *ent.Client
	cron gocron.Scheduler
	job  gocron.Job
	log  *slog.Logger
}

func NewIpReportService(db *ent.Client, cron gocron.Scheduler) *IpReportService {
	return &IpReportService{db: db, cron: cron, log: slog.With(
		slog.String("service", "IpReportService"),
	)}
}

func (s *IpReportService) StartJob(ctx context.Context) (err error) {
	// 每8到12分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			8*time.Minute,  // 最小间隔
			12*time.Minute, // 最大间隔
		),
		gocron.NewTask(func() {
			// 请求参数
			params := types.ReportDeviceIpRequest{
				DeviceName: "本设备",
			}
			// 获取设备ID
			params.DeviceCode, err = machineid.ID()
			if err != nil {
				s.log.ErrorContext(ctx, "获取设备ID失败",
					slog.Any("err", err),
				)
				return
			}
			// 获取设备IP
			checker := ip_helper.NewIPChecker()
			ipv4, err := checker.CheckIP(ctx, "v4")
			if err == nil {
				params.IPV4Addr = ipv4.IP
			}
			ipv6, err := checker.CheckIP(ctx, "v6")
			if err == nil {
				params.IPV6Addr = ipv6.IP
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
					SetLastHeartbeatAt(time.Now()).
					Save(ctx)
				if err != nil {
					_ = tx.Rollback()
					s.log.ErrorContext(ctx, "创建设备失败",
						slog.Any("err", err),
					)
					return
				}
				// 记录操作日志
				err = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "自动创建设备")
				if err != nil {
					_ = tx.Rollback()
					s.log.ErrorContext(ctx, "记录操作日志失败",
						slog.Any("err", err),
					)
					return
				}
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
				err = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDeviceIPReport.ID, operationlog.ObjTypeDeviceIPReport, operationlog.StatusSuccess, "自动创建设备IP报告")
				if err != nil {
					_ = tx.Rollback()
					s.log.ErrorContext(ctx, "记录操作日志失败",
						slog.Any("err", err),
					)
					return
				}
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
				err = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDeviceIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "自动创建设备IP组")
				if err != nil {
					_ = tx.Rollback()
					s.log.ErrorContext(ctx, "记录操作日志失败",
						slog.Any("err", err),
					)
					return
				}
				params.ID = saveDevice.ID
				msg = "创建设备IP成功"
			} else {
				params.ID = firstDevice.ID
				if firstDevice.IPV4 != params.IPV4Addr || firstDevice.IPV6 != params.IPV6Addr {
					// 更新设备IP
					err = tx.Device.UpdateOne(firstDevice).
						SetIPV4(params.IPV4Addr).
						SetIPV6(params.IPV6Addr).
						SetLastHeartbeatAt(time.Now()).
						Exec(ctx)
					if err != nil {
						_ = tx.Rollback()
						s.log.ErrorContext(ctx, "更新设备IP失败",
							slog.Any("err", err),
						)
						return
					}
					// 记录操作日志
					err = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "自动更新设备")
					if err != nil {
						_ = tx.Rollback()
						s.log.ErrorContext(ctx, "记录操作日志失败",
							slog.Any("err", err),
						)
						return
					}
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
					err = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(ctx, saveDeviceIPReport.ID, operationlog.ObjTypeDeviceIPReport, operationlog.StatusSuccess, "自动创建设备IP报告")
					if err != nil {
						_ = tx.Rollback()
						s.log.ErrorContext(ctx, "记录操作日志失败",
							slog.Any("err", err),
						)
						return
					}
					msg = "更新设备IP成功"
				} else {
					// 更新设备心跳时间
					err = tx.Device.UpdateOne(firstDevice).
						SetLastHeartbeatAt(time.Now()).
						Exec(ctx)
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
			}
		}),
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
