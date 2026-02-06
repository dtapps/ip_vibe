package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"
)

// TaskService 任务服务接口
type TaskService interface {
	BuildResponse(ctx context.Context, item *ent.CronJob) (*types.GetTaskResponse, error)
	GetTasks(ctx context.Context, request *types.GetTasksRequest) ([]*types.GetTaskResponse, error)
	GetTask(ctx context.Context, id int64) (*types.GetTaskResponse, error)
	UpdateTask(ctx context.Context, request *types.UpdateTaskRequest) error
	UpdateTaskActive(ctx context.Context, id int64, status string) error
	BatchUpdateTaskActive(ctx context.Context, ids []int64, status string) error
	DeleteTask(ctx context.Context, id int64) error
	BatchDeleteTask(ctx context.Context, ids []int64) error
}

// TaskServiceImpl 任务服务实现
type TaskServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式
}

// NewTaskService 创建任务服务实例
func NewTaskService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) TaskService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &TaskServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  debug,
	}
}

// SetDebug 设置调试模式
func (s *TaskServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *TaskServiceImpl) BuildResponse(ctx context.Context, item *ent.CronJob) (*types.GetTaskResponse, error) {
	resp := &types.GetTaskResponse{
		ID:            item.ID,
		EntryID:       item.EntryID,
		Name:          item.Name,
		HandlerName:   item.HandlerName,
		HandlerArgs:   item.HandlerArgs,
		TaskType:      item.TaskType.String(),
		JobType:       item.JobType.String(),
		JobCategory:   item.JobCategory.String(),
		BusinessType:  item.BusinessType.String(),
		BusinessID:    item.BusinessID,
		CronExpr:      item.CronExpr,
		WithSeconds:   item.WithSeconds,
		IntervalValue: item.IntervalValue,
		MinInterval:   item.MinInterval,
		MaxInterval:   item.MaxInterval,
		IntervalUnit:  item.IntervalUnit.String(),
		Status:        item.Status.String(),
		IsSingleton:   item.IsSingleton,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	return resp, nil
}

// GetTasks 获取全部任务
func (s *TaskServiceImpl) GetTasks(ctx context.Context, request *types.GetTasksRequest) ([]*types.GetTaskResponse, error) {
	// 查询全部任务
	allTask, err := s.db.CronJob.Query().
		WhereIf(request.Status != "", cronjob.StatusEQ(cronjob.Status(request.Status))).
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询记录失败: %w", err)
	}

	lists := make([]*types.GetTaskResponse, 0, len(allTask))
	for _, item := range allTask {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetTask 获取任务详情
func (s *TaskServiceImpl) GetTask(ctx context.Context, id int64) (*types.GetTaskResponse, error) {
	// 查询任务
	firstTask, err := s.db.CronJob.Query().
		Where(cronjob.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询任务失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstTask)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateTask 更新 任务
func (s *TaskServiceImpl) UpdateTask(ctx context.Context, request *types.UpdateTaskRequest) error {

	// 查询任务
	firstTask, err := s.db.CronJob.Query().
		Where(cronjob.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询任务失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}

	// 更新任务
	updateTask := s.db.CronJob.UpdateOne(firstTask)
	updateTask.SetJobType(cronjob.JobType(request.JobType))
	updateTask.SetCronExpr(request.CronExpr)
	updateTask.SetWithSeconds(request.WithSeconds)
	updateTask.SetIntervalValue(request.IntervalValue)
	updateTask.SetMinInterval(request.MinInterval)
	updateTask.SetMaxInterval(request.MaxInterval)
	updateTask.SetIntervalUnit(cronjob.IntervalUnit(request.IntervalUnit))
	updateTask.SetIsSingleton(request.IsSingleton)
	updateTask.SetRunOnStart(request.RunOnStart)
	err = updateTask.Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新任务失败: %w", err)
	}

	return nil
}

// UpdateTaskActive 更新 任务状态
func (s *TaskServiceImpl) UpdateTaskActive(ctx context.Context, id int64, status string) error {

	// 查询任务
	firstTask, err := s.db.CronJob.Query().
		Where(cronjob.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询任务失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}

	// 更新任务状态
	err = s.db.CronJob.UpdateOne(firstTask).
		SetStatus(cronjob.Status(status)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	return nil
}

// BatchUpdateTaskActive 批量更新 任务状态
func (s *TaskServiceImpl) BatchUpdateTaskActive(ctx context.Context, ids []int64, status string) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询任务
	allTask, err := tx.CronJob.Query().
		Where(cronjob.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询任务失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if len(allTask) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询任务失败: 数量 %d 与请求数量 %d 不匹配", len(allTask), len(ids))
	}

	// 更新任务状态
	err = tx.CronJob.Update().
		Where(cronjob.IDIn(ids...)).
		SetStatus(cronjob.Status(status)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// DeleteTask 删除 任务
func (s *TaskServiceImpl) DeleteTask(ctx context.Context, id int64) error {

	// 查询任务
	firstTask, err := s.db.CronJob.Query().
		Where(cronjob.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询任务失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}

	// 删除任务
	err = s.db.CronJob.DeleteOne(firstTask).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	return nil
}

// BatchDeleteTask 批量删除 任务
func (s *TaskServiceImpl) BatchDeleteTask(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询任务
	allTask, err := tx.CronJob.Query().
		Where(cronjob.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if len(allTask) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询任务失败: 数量 %d 与请求数量 %d 不匹配", len(allTask), len(ids))
	}

	// 检查是否存在启用的任务
	for _, item := range allTask {
		if item.Status == cronjob.StatusRunning {
			tx.Rollback()
			return fmt.Errorf("删除任务失败: 存在启用的任务")
		}
	}

	// 删除任务
	_, err = tx.CronJob.Delete().
		Where(cronjob.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除任务失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}
