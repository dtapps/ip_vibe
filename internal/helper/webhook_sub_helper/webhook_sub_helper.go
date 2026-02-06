package webhook_sub_helper

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhooksubscription"
)

type WebhookSubscription struct {
	db     *ent.Client
	logger *slog.Logger
}

func NewWebhookSubscription(db *ent.Client, logger *slog.Logger) *WebhookSubscription {
	return &WebhookSubscription{
		db:     db,
		logger: logger,
	}
}

// 创建通知订阅（同步 webhook_ids 列表）
func (ws *WebhookSubscription) Create(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	ws.logger.DebugContext(ctx, "[创建通知订阅]准备",
		slog.Any("webhookIDs", webhookIDs),
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 批量创建
	createList := make([]*ent.WebhookSubscriptionCreate, 0, len(webhookIDs))
	for _, item := range webhookIDs {
		createList = append(createList,
			ws.db.WebhookSubscription.Create().
				SetName(fmt.Sprintf("关联_%d_%d", objID, item)).
				SetObjType(objType).
				SetObjID(objID).
				SetWebhookID(item).
				SetIsActive(true),
		)
	}
	if len(createList) > 0 {
		ws.logger.DebugContext(ctx, "[创建通知订阅]准备批量创建",
			slog.Any("webhookIDs", webhookIDs),
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Int("count", len(createList)),
		)
		err = ws.db.WebhookSubscription.CreateBulk(createList...).
			Exec(ctx)
		if err != nil {
			ws.logger.ErrorContext(ctx, "[创建通知订阅]批量失败",
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
func (ws *WebhookSubscription) Update(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	ws.logger.DebugContext(ctx, "[更新通知订阅]准备",
		slog.Any("webhookIDs", webhookIDs),
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := ws.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		ws.logger.ErrorContext(ctx, "[更新通知订阅]查询已存在订阅失败",
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
				ws.db.WebhookSubscription.Create().
					SetName(fmt.Sprintf("关联_%d_%d", objID, id)).
					SetObjType(objType).
					SetObjID(objID).
					SetWebhookID(id).
					SetIsActive(true),
			)
		}
	}
	if len(createList) > 0 {
		ws.logger.DebugContext(ctx, "[更新通知订阅]准备批量创建",
			slog.Any("webhookIDs", webhookIDs),
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Int("count", len(createList)),
		)
		err = ws.db.WebhookSubscription.CreateBulk(createList...).
			Exec(ctx)
		if err != nil {
			ws.logger.ErrorContext(ctx, "[更新通知订阅]创建新增订阅失败",
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
		ws.logger.DebugContext(ctx, "[更新通知订阅]准备批量删除",
			slog.Any("webhookIDs", webhookIDs),
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Int("count", len(toDeleteWebhookIDs)),
		)
		_, err = ws.db.WebhookSubscription.Delete().
			Where(webhooksubscription.ObjTypeEQ(objType)).
			Where(webhooksubscription.ObjIDEQ(objID)).
			Where(webhooksubscription.WebhookIDIn(toDeleteWebhookIDs...)).
			Exec(ctx)
		if err != nil {
			ws.logger.ErrorContext(ctx, "[更新通知订阅]删除已存在订阅失败",
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
func (ws *WebhookSubscription) Delete(ctx context.Context, objType webhooksubscription.ObjType, objID int64) (err error) {
	ws.logger.DebugContext(ctx, "[删除通知订阅]准备",
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := ws.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		ws.logger.ErrorContext(ctx, "[删除通知订阅]查询已存在订阅失败",
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
	_, err = ws.db.WebhookSubscription.Delete().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		Where(webhooksubscription.WebhookIDIn(existingWebhookIDs...)).
		Exec(ctx)
	if err != nil {
		ws.logger.ErrorContext(ctx, "[删除通知订阅]删除已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Any("err", err),
		)
		return err
	}

	return nil
}

// 删除通知订阅（同步 webhook_ids 列表）
func (ws *WebhookSubscription) DeleteTx(ctx context.Context, tx *ent.Tx, objType webhooksubscription.ObjType, objID int64) (err error) {
	ws.logger.DebugContext(ctx, "[删除通知订阅]准备",
		slog.String("objType", objType.String()),
		slog.Int64("objID", objID),
	)

	// 查询当前已存在的订阅
	existingSubs, err := tx.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		ws.logger.ErrorContext(ctx, "[删除通知订阅]查询已存在订阅失败",
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
		ws.logger.ErrorContext(ctx, "[删除通知订阅]删除已存在订阅失败",
			slog.String("objType", objType.String()),
			slog.Int64("objID", objID),
			slog.Any("err", err),
		)
		return err
	}

	return nil
}
