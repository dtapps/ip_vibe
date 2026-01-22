package webhook_sub_helper

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/global"
)

// 创建通知订阅（同步 webhook_ids 列表）
func Create(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	slog.InfoContext(ctx, "[创建通知订阅]准备",
		slog.Any("webhookIDs", webhookIDs),
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 批量创建
	createList := make([]*ent.WebhookSubscriptionCreate, 0)
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
		err = global.GvaEntDb.WebhookSubscription.CreateBulk(createList...).
			Exec(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "[创建通知订阅]失败",
				slog.Any("webhookIDs", webhookIDs),
				slog.String("objType", objType.String()),
				slog.Int64("objID", objID),
				slog.String("error", err.Error()),
			)
			return err
		}
	}

	return nil
}

// 更新通知订阅（同步 webhook_ids 列表）
func Update(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	slog.InfoContext(ctx, "[更新通知订阅]准备",
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
		slog.ErrorContext(ctx, "[更新通知订阅]查询已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.String("error", err.Error()),
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
	createList := make([]*ent.WebhookSubscriptionCreate, 0)
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
		err = global.GvaEntDb.WebhookSubscription.CreateBulk(createList...).
			Exec(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "[更新通知订阅]创建新增订阅失败",
				slog.Any("webhookIDs", webhookIDs),
				slog.String("objType", objType.String()),
				slog.Int64("objID", objID),
				slog.String("error", err.Error()),
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
		_, err = global.GvaEntDb.WebhookSubscription.Delete().
			Where(webhooksubscription.ObjTypeEQ(objType)).
			Where(webhooksubscription.ObjIDEQ(objID)).
			Where(webhooksubscription.WebhookIDIn(toDeleteWebhookIDs...)).
			Exec(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "[更新通知订阅]删除已存在订阅失败",
				slog.String("objType", objType.String()),
				slog.Int64("objID", objID),
				slog.String("error", err.Error()),
			)
			return err
		}
	}

	return nil
}

// 删除通知订阅（同步 webhook_ids 列表）
func Delete(ctx context.Context, objType webhooksubscription.ObjType, objID int64) (err error) {
	slog.InfoContext(ctx, "[删除通知订阅]准备",
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "[删除通知订阅]查询已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.String("error", err.Error()),
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
		slog.ErrorContext(ctx, "[删除通知订阅]删除已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// 删除通知订阅（同步 webhook_ids 列表）
func DeleteTx(ctx context.Context, tx *ent.Tx, objType webhooksubscription.ObjType, objID int64) (err error) {
	slog.InfoContext(ctx, "[删除通知订阅]准备",
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := tx.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "[删除通知订阅]查询已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.String("error", err.Error()),
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
		slog.ErrorContext(ctx, "[删除通知订阅]删除已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}
