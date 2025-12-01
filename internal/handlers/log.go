package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Log(c *gin.Context) {
	// 准备页面数据
	pageData := types.LogResponse{
		List:     []types.GetOperationLogsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部操作日志
	allOperationLog, err := global.GvaDb.OperationLog.Query().
		Order(ent.Desc(operationlog.FieldCreatedAt)).
		Limit(15).
		All(c)
	if err == nil {
		for _, item := range allOperationLog {
			pageData.List = append(pageData.List, types.GetOperationLogsResponse{
				ID:        item.ID,
				OpType:    item.OpType.String(),
				ObjID:     item.ObjID,
				ObjType:   item.ObjType.String(),
				OldValue:  item.OldValue,
				NewValue:  item.NewValue,
				Status:    item.Status.String(),
				Message:   item.Message,
				CreatedAt: item.CreatedAt,
			})
		}
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("log pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/log.html", pageData)
}
