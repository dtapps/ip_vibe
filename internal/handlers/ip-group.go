package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/ipgroup"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func IPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.IPGroupRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.IPGroupResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetIPGroupResponse, types.IPGroupRequest]{
			Filter: requestParams,
			List:   []types.GetIPGroupResponse{},
		},
	}

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
	)

	// 查询全部IP组
	allIpGroup, err := global.GvaEntDb.IpGroup.Query().
		WhereIf(requestParams.ID > 0, ipgroup.IDEQ(requestParams.ID)).
		WhereIf(requestParams.RecordID != "", ipgroup.RecordIDEQ(requestParams.RecordID)).
		WhereIf(requestParams.Name != "", ipgroup.NameContains(requestParams.Name)).
		WithDevices().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allIpGroup.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allIpGroup.Page
	pageData.PageSize = allIpGroup.Size
	pageData.Total = allIpGroup.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "ip-group.html"),
		types.BasePageData[types.IPGroupResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "IP组记录",
			Data:  pageData,
		},
	)
}
