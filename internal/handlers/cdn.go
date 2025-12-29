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

func Cdn(c *gin.Context) {
	// 请求参数
	var requestParams types.CdnRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.CdnResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetCdnRecordResponse]{
			List: []types.GetCdnRecordResponse{},
		},
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 查询全部CDN记录
	allCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		WithConfig().
		WithIPGroup().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allCdnRecord.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allCdnRecord.Page
	pageData.PageSize = allCdnRecord.Size
	pageData.Total = allCdnRecord.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "cdn.html"),
		types.BasePageData[types.CdnResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "CDN记录",
			Data:  pageData,
		},
	)
}
