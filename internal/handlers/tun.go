package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/tunrecord"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Tun(c *gin.Context) {
	// 请求参数
	var requestParams types.TunRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.TunResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetTunRecordResponse, types.TunRequest]{
			Filter: requestParams,
			List:   []types.GetTunRecordResponse{},
		},
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 查询全部TUN记录
	allTunRecord, err := global.GvaEntDb.TUNRecord.Query().
		WhereIf(requestParams.ID > 0, tunrecord.IDEQ(requestParams.ID)).
		WhereIf(requestParams.ConfigID > 0, tunrecord.ConfigIDEQ(requestParams.ConfigID)).
		WhereIf(requestParams.RecordID != "", tunrecord.RecordIDEQ(requestParams.RecordID)).
		WhereIf(requestParams.SubDomain != "", tunrecord.SubDomainContains(requestParams.SubDomain)).
		WhereIf(requestParams.ServeType != "", tunrecord.ServeTypeEQ(tunrecord.ServeType(requestParams.ServeType))).
		WhereIf(!requestParams.IsActive.IsEmpty(), tunrecord.IsActiveEQ(requestParams.IsActive.Value())).
		WithConfig().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allTunRecord.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allTunRecord.Page
	pageData.PageSize = allTunRecord.Size
	pageData.Total = allTunRecord.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "tun.html"),
		types.BasePageData[types.TunResponse]{
			Title: "TUN记录",
			Data:  pageData,
		},
	)
}
