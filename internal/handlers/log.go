package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Log(c *gin.Context) {
	// 请求参数
	var requestParams types.LogRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.LogResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetLogResponse]{
			List: []types.GetLogResponse{},
		},
	}

	// 查询全部操作日志
	allOperationLog, err := global.GvaEntLogDb.OperationLog.Query().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allOperationLog.Items {
			pageData.List = append(pageData.List, types.GetLogResponse{
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

	pageData.Page = allOperationLog.Page
	pageData.PageSize = allOperationLog.Size
	pageData.Total = allOperationLog.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "log.html"),
		types.BasePageData[types.LogResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "操作日志",
			Data:  pageData,
		},
	)
}
