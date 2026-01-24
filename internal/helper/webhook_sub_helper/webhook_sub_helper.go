package webhook_sub_helper

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/global"
)

// getEntSlog: 获取业务数据库日志实例
func getEntSlog() *slog.Logger {
	logger := slog.Default()
	if global.GvaEntSlog != nil {
		logger = global.GvaEntSlog
	}
	return logger
}

// getEntLogSlog: 获取操作日志数据库日志实例
func getEntLogSlog() *slog.Logger {
	logger := slog.Default()
	if global.GvaEntLogSlog != nil {
		logger = global.GvaEntLogSlog
	}
	return logger
}

// 创建通知订阅（同步 webhook_ids 列表）
func Create(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	getEntSlog().DebugContext(ctx, "[创建通知订阅]准备",
		slog.Any("webhookIDs", webhookIDs),
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 批量创建
	createList := make([]*ent.WebhookSubscriptionCreate, 0, len(webhookIDs))
	for _, item := range webhookIDs {
		createList = append(createList,
			global.GvaEntDb.WebhookSubscription.Create().
				SetName(fmt.Sprintf("关联_%d_%d", objID, item)).
				SetObjType(objType).
				SetObjID(objID).
				SetWebhookID(item).
				SetIsActive(true),
		)
	}
	if len(createList) > 0 {
		getEntSlog().DebugContext(ctx, "[创建通知订阅]准备批量创建",
			slog.Any("webhookIDs", webhookIDs),
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Int("count", len(createList)),
		)
		err = global.GvaEntDb.WebhookSubscription.CreateBulk(createList...).
			Exec(ctx)
		if err != nil {
			getEntSlog().ErrorContext(ctx, "[创建通知订阅]批量失败",
				slog.Any("webhookIDs", webhookIDs),
				slog.String("objType", objType.String()),
				slog.Int64("objID", objID),
				slog.Any("err", err),
			)
			return err
		}
	}

	return nil
}

// 更新通知订阅（同步 webhook_ids 列表）
func Update(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	getEntSlog().DebugContext(ctx, "[更新通知订阅]准备",
		slog.Any("webhookIDs", webhookIDs),
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		getEntSlog().ErrorContext(ctx, "[更新通知订阅]查询已存在订阅失败",
			slog.Any("webhookIDs", webhookIDs),
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Any("err", err),
		)
		return err
	}

	// 构建已存在的 webhook_ids 集合
	existingWebhookIDMap := make(map[int64]bool)
	var existingWebhookIDs []int64
	for _, sub := range existingSubs {
		existingWebhookIDMap[(sub.WebhookID)] = true
		existingWebhookIDs = append(existingWebhookIDs, sub.WebhookID)
	}

	// 新增：webhook_ids 中有，但数据库中没有的
	createList := make([]*ent.WebhookSubscriptionCreate, 0, len(webhookIDs))
	for _, id := range webhookIDs {
		if !existingWebhookIDMap[id] {
			createList = append(createList,
				global.GvaEntDb.WebhookSubscription.Create().
					SetName(fmt.Sprintf("关联_%d_%d", objID, id)).
					SetObjType(objType).
					SetObjID(objID).
					SetWebhookID(id).
					SetIsActive(true),
			)
		}
	}
	if len(createList) > 0 {
		getEntSlog().DebugContext(ctx, "[更新通知订阅]准备批量创建",
			slog.Any("webhookIDs", webhookIDs),
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Int("count", len(createList)),
		)
		err = global.GvaEntDb.WebhookSubscription.CreateBulk(createList...).
			Exec(ctx)
		if err != nil {
			getEntSlog().ErrorContext(ctx, "[更新通知订阅]创建新增订阅失败",
				slog.Any("webhookIDs", webhookIDs),
				slog.String("objType", objType.String()),
				slog.Int64("objID", objID),
				slog.Any("err", err),
			)
		}
	}

	// 删除：数据库中有，但 webhook_ids 中没有的
	var toDeleteWebhookIDs []int64
	for _, wid := range existingWebhookIDs {
		found := false
		for _, reqWID := range webhookIDs {
			if wid == reqWID {
				found = true
				break
			}
		}
		if !found {
			toDeleteWebhookIDs = append(toDeleteWebhookIDs, wid)
		}
	}

	// 批量删除
	if len(toDeleteWebhookIDs) > 0 {
		getEntSlog().DebugContext(ctx, "[更新通知订阅]准备批量删除",
			slog.Any("webhookIDs", webhookIDs),
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Int("count", len(toDeleteWebhookIDs)),
		)
		_, err = global.GvaEntDb.WebhookSubscription.Delete().
			Where(webhooksubscription.ObjTypeEQ(objType)).
			Where(webhooksubscription.ObjIDEQ(objID)).
			Where(webhooksubscription.WebhookIDIn(toDeleteWebhookIDs...)).
			Exec(ctx)
		if err != nil {
			getEntSlog().ErrorContext(ctx, "[更新通知订阅]删除已存在订阅失败",
				slog.Any("webhookIDs", webhookIDs),
				slog.String("objType", objType.String()),
				slog.Int64("objID", objID),
				slog.Any("err", err),
			)
			return err
		}
	}

	return nil
}

// 删除通知订阅（同步 webhook_ids 列表）
func Delete(ctx context.Context, objType webhooksubscription.ObjType, objID int64) (err error) {
	getEntSlog().DebugContext(ctx, "[删除通知订阅]准备",
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		getEntSlog().ErrorContext(ctx, "[删除通知订阅]查询已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Any("err", err),
		)
		return err
	}

	// 构建已存在的 webhook_ids 集合
	var existingWebhookIDs []int64
	for _, sub := range existingSubs {
		existingWebhookIDs = append(existingWebhookIDs, sub.WebhookID)
	}

	// 批量删除
	_, err = global.GvaEntDb.WebhookSubscription.Delete().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		Where(webhooksubscription.WebhookIDIn(existingWebhookIDs...)).
		Exec(ctx)
	if err != nil {
		getEntSlog().ErrorContext(ctx, "[删除通知订阅]删除已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Any("err", err),
		)
		return err
	}

	return nil
}

// 删除通知订阅（同步 webhook_ids 列表）
func DeleteTx(ctx context.Context, tx *ent.Tx, objType webhooksubscription.ObjType, objID int64) (err error) {
	getEntSlog().DebugContext(ctx, "[删除通知订阅]准备",
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := tx.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		getEntSlog().ErrorContext(ctx, "[删除通知订阅]查询已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Any("err", err),
		)
		return err
	}

	// 构建已存在的 webhook_ids 集合
	var existingWebhookIDs []int64
	for _, sub := range existingSubs {
		existingWebhookIDs = append(existingWebhookIDs, sub.WebhookID)
	}

	// 批量删除
	_, err = tx.WebhookSubscription.Delete().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		Where(webhooksubscription.WebhookIDIn(existingWebhookIDs...)).
		Exec(ctx)
	if err != nil {
		getEntSlog().ErrorContext(ctx, "[删除通知订阅]删除已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Any("err", err),
		)
		return err
	}

	return nil
}
