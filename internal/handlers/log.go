package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent_log/operationlog"
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
	if requestParams.Source == "" {
		requestParams.Source = operationlog.SourceUser.String()
	}

	// 准备页面数据
	pageData := types.LogResponse{
		CommonPageResponse: types.CommonPageResponse[any, types.LogRequest]{
			Filter: requestParams,
			List:   []any{},
		},
	}

	// 查询全部操作日志
	allOperationLog, err := global.GvaEntLogDb.OperationLog.Query().
		WhereIf(requestParams.ID > 0, operationlog.IDEQ(requestParams.ID)).
		WhereIf(requestParams.Source != "", operationlog.SourceEQ(operationlog.Source(requestParams.Source))).
		WhereIf(requestParams.OpType != "", operationlog.OpTypeEQ(operationlog.OpType(requestParams.OpType))).
		WhereIf(requestParams.ObjType != "", operationlog.ObjTypeEQ(operationlog.ObjType(requestParams.ObjType))).
		WhereIf(requestParams.ObjID > 0, operationlog.ObjIDEQ(requestParams.ObjID)).
		WhereIf(requestParams.Status != "", operationlog.StatusEQ(operationlog.Status(requestParams.Status))).
		WhereIf(requestParams.Message != "", operationlog.MessageContains(requestParams.Message)).
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allOperationLog.Items {
			pageData.List = append(pageData.List, &item)
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
