package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func SslImport(c *gin.Context) {
	// 请求参数
	var requestParams types.SSLImportRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.SSLImportResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetSSLImportResponse, types.SSLImportRequest]{
			Filter: requestParams,
			List:   []types.GetSSLImportResponse{},
		},
	}

	// 创建服务
	service := service.NewSSLImportService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 查询全部SSL导入记录
	allSSLRecord, err := global.GvaEntDb.SSLImport.Query().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allSSLRecord.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allSSLRecord.Page
	pageData.PageSize = allSSLRecord.Size
	pageData.Total = allSSLRecord.Total

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "ssl-import.html"),
		types.BasePageData[types.SSLImportResponse]{
			Title: "SSL导入",
			Data:  pageData,
		},
		"ssl",
	)
}
