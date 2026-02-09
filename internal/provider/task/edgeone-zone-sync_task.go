package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/backend/syncs"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/provider"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
)

// EdgeOneZoneSyncTaskProvider 实现 provider.TaskProvider 接口
type EdgeOneZoneSyncTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db              *ent.Client                   // 数据库客户端
	logDb           *ent_log.Client               // 日志数据库客户端
	logger          *slog.Logger                  // 日志记录器
	httpTransport   *http_log.LoggingRoundTripper // HTTP日志记录器
	edgeOneZoneSync *syncs.EdgeOneZoneSync        // EdgeOne站点同步

	businessType cronjob.BusinessType // 业务类型
}

// NewEdgeOneZoneSyncTaskProvider 创建一个新的 EdgeOne站点同步 提供者
func NewEdgeOneZoneSyncTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewEdgeOneZoneSyncTaskProviderRaw(db, logDb, logger, httpTransport)
}

// NewEdgeOneZoneSyncTaskProvider 创建一个新的 EdgeOne站点同步 提供者
func NewEdgeOneZoneSyncTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *EdgeOneZoneSyncTaskProvider {
	edgeOneZoneSync := syncs.NewEdgeOneZoneSync(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &EdgeOneZoneSyncTaskProvider{
		db:              db,
		logDb:           logDb,
		logger:          logger.With(slog.String("handler", "edgeone-zone-sync")),
		httpTransport:   httpTransport,
		edgeOneZoneSync: edgeOneZoneSync,

		businessType: cronjob.BusinessTypeEdgeone,
	}
}

// EdgeOneZoneSyncTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	EdgeOneZoneSyncTaskName = "edgeone_zone_sync"
)

// EdgeOneZoneSyncTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type EdgeOneZoneSyncTaskArgs struct {
	ConfigID int64 `json:"config_id,omitempty"` // 平台配置序号
	ZoneID   int64 `json:"zone_id,omitempty"`   // 站点序号
	GroupID  int64 `json:"group_id,omitempty"`  // 源站组序号
}

// Process 执行 EdgeOne站点同步 逻辑
func (p *EdgeOneZoneSyncTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params EdgeOneZoneSyncTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.ConfigID == 0 && params.ZoneID == 0 && params.GroupID == 0 {

		// 调用 同步 全部账号 所有站点
		err := p.edgeOneZoneSync.SyncAllZones(ctx)
		if err != nil {
			return "", err
		}

		// 调用 同步 全部账号 所有站点 所有源站组
		err = p.edgeOneZoneSync.SyncAllZoneOriginGroups(ctx)
		if err != nil {
			return "", err
		}

		// 调用 同步 全部账号 所有站点 所有记录
		err = p.edgeOneZoneSync.SyncAllZoneRecords(ctx)
		if err != nil {
			return "", err
		}

	} else if params.ConfigID != 0 && params.ZoneID == 0 && params.GroupID == 0 {

		// 同步 某个账号 所有站点
		err := p.edgeOneZoneSync.SyncThisZones(ctx, params.ConfigID)
		if err != nil {
			return "", err
		}

		// 同步 某个账号 所有站点 所有源站组
		err = p.edgeOneZoneSync.SyncThisZoneOriginGroups(ctx, params.ConfigID)
		if err != nil {
			return "", err
		}

		// 同步 某个账号 所有站点 所有记录
		err = p.edgeOneZoneSync.SyncThisZoneRecords(ctx, params.ConfigID)
		if err != nil {
			return "", err
		}

	} else if params.ConfigID != 0 && params.ZoneID != 0 && params.GroupID == 0 {

		// 同步 某个账号 某个站点
		err := p.edgeOneZoneSync.SyncThisZone(ctx, params.ConfigID, params.ZoneID)
		if err != nil {
			return "", err
		}

		// 同步 某个账号 某个站点 所有源站组
		err = p.edgeOneZoneSync.SyncThisZoneOriginGroup(ctx, params.ConfigID, params.ZoneID)
		if err != nil {
			return "", err
		}

		// 同步 某个账号 某个站点 所有记录
		err = p.edgeOneZoneSync.SyncThisZoneRecord(ctx, params.ConfigID, params.ZoneID)
		if err != nil {
			return "", err
		}

	} else if params.ConfigID != 0 && params.ZoneID != 0 && params.GroupID != 0 {

	} else {
		return "", fmt.Errorf("参数错误")
	}

	return fmt.Sprintf("同步成功"), nil
}

// LoadTask 加载 EdgeOne站点同步 周期任务
func (p *EdgeOneZoneSyncTaskProvider) LoadTask(ctx context.Context, db *ent.Client, record *ent.PlatformConfig) error {

	var (
		taskType    = cronjob.TaskTypeRecurring
		jobType     = cronjob.JobTypeRandom
		jobCategory = cronjob.JobCategorySystem
	)

	// 查找任务
	_, err := db.CronJob.Query().
		Where(cronjob.TaskTypeEQ(taskType)).
		Where(cronjob.BusinessTypeEQ(p.businessType)).
		Where(cronjob.BusinessIDEQ(record.ID)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf("查找任务失败: %w", err)
		}

		// 创建任务
		err = db.CronJob.Create().
			SetName("EdgeOne站点同步: " + record.Name).
			SetHandlerName(EdgeOneZoneSyncTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(EdgeOneZoneSyncTaskArgs{
				ConfigID: record.ID,
			}))).
			SetTaskType(taskType).
			SetJobType(jobType).
			SetJobCategory(jobCategory).
			SetBusinessType(p.businessType).
			SetBusinessID(record.ID).
			SetMinInterval(6).
			SetMaxInterval(8).
			SetIntervalUnit(cronjob.IntervalUnitHours).
			SetStatus(cronjob.StatusRunning).
			SetIsSingleton(true). // 单例模式
			SetRunOnStart(false). // 启动时不立即运行
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("创建任务失败: %w", err)
		}

		return nil
	}

	return nil
}

// AddTask 添加 EdgeOne站点同步 一次性任务
func (p *EdgeOneZoneSyncTaskProvider) AddTask(ctx context.Context, db *ent.Client, configID int64, zoneID int64, groupID int64) error {

	var (
		taskType    = cronjob.TaskTypeOneOff
		jobType     = cronjob.JobTypeDuration
		jobCategory = cronjob.JobCategoryCustom
	)

	// 任务名称
	var name string

	// 业务键
	var businessID int64

	if configID != 0 && zoneID == 0 && groupID == 0 {
		name = fmt.Sprintf("EdgeOne站点同步: %d", configID)
		businessID = configID
	} else if configID != 0 && zoneID != 0 && groupID == 0 {
		name = fmt.Sprintf("EdgeOne站点同步: %d-%d", configID, zoneID)
		businessID = zoneID
	} else if configID != 0 && zoneID != 0 && groupID != 0 {
		name = fmt.Sprintf("EdgeOne站点同步: %d-%d-%d", configID, zoneID, groupID)
		businessID = groupID
	} else {
		return fmt.Errorf("参数错误")
	}

	// 创建任务
	err := db.CronJob.Create().
		SetName(name).
		SetHandlerName(EdgeOneZoneSyncTaskName).
		SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(EdgeOneZoneSyncTaskArgs{
			ConfigID: configID,
			ZoneID:   zoneID,
			GroupID:  groupID,
		}))).
		SetTaskType(taskType).
		SetJobType(jobType).
		SetJobCategory(jobCategory).
		SetBusinessType(p.businessType).
		SetBusinessID(businessID).
		SetIntervalValue(1).
		SetIntervalUnit(cronjob.IntervalUnitSeconds).
		SetStatus(cronjob.StatusRunning).
		SetIsSingleton(true). // 单例模式
		SetRunOnStart(true).  // 启动时立即运行
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	return nil
}
