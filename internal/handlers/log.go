package handlers

import (
	"fmt"
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
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
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

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "log.html"), pageData)
}
