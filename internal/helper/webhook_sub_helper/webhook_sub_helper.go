package webhook_sub_helper

import (
	"context"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/global"
)

// 创建通知订阅（同步 webhook_ids 列表）
func Create(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	if len(webhookIDs) == 0 {
		return nil
	}

	for _, item := range webhookIDs {
		err = global.GvaEntDb.WebhookSubscription.Create().
			SetObjType(objType).
			SetObjID(objID).
			SetWebhookID(item).
			SetIsActive(true).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// 更新通知订阅（同步 webhook_ids 列表）
func Update(ctx context.Context, webhookIDs []int64, objType webhooksubscription.ObjType, objID int64) (err error) {
	// if len(webhookIDs) == 0 {
	// 	return nil
	// }

	// 查询当前已存在的订阅（基于 objType + objID）
	existingSubs, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
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
	for _, id := range webhookIDs {
		if !existingWebhookIDMap[id] {
			global.GvaEntDb.WebhookSubscription.Create().
				SetObjType(objType).
				SetObjID(objID).
				SetWebhookID(id).
				SetIsActive(true).
				Exec(ctx)
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

	// 批量删除（更高效）
	if len(toDeleteWebhookIDs) > 0 {
		_, err = global.GvaEntDb.WebhookSubscription.Delete().
			Where(webhooksubscription.ObjTypeEQ(objType)).
			Where(webhooksubscription.ObjIDEQ(objID)).
			Where(webhooksubscription.WebhookIDIn(toDeleteWebhookIDs...)).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// 删除通知订阅（同步 webhook_ids 列表）
func Delete(ctx context.Context, objType webhooksubscription.ObjType, objID int64) (err error) {

	// 查询当前已存在的订阅（基于 objType + objID）
	existingSubs, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		All(ctx)
	if err != nil {
		return err
	}

	// 构建已存在的 webhook_ids 集合
	var existingWebhookIDs []int64
	for _, sub := range existingSubs {
		existingWebhookIDs = append(existingWebhookIDs, sub.WebhookID)
	}

	// 批量删除（更高效）
	_, err = global.GvaEntDb.WebhookSubscription.Delete().
		Where(webhooksubscription.ObjTypeEQ(objType)).
		Where(webhooksubscription.ObjIDEQ(objID)).
		Where(webhooksubscription.WebhookIDIn(existingWebhookIDs...)).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
