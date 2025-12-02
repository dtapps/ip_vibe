package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/ipgroup"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_param"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func IPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.IPGroupRequest
	if err := gin_param.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.IPGroupResponse{
		List:     []types.GetIPGroupsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部IP组
	allIpGroup, err := global.GvaDb.IpGroup.Query().
		WithDevices().
		Order(ent.Desc(ipgroup.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allIpGroup.Items {
			pageData.List = append(pageData.List, types.GetIPGroupsResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				DeviceCount: len(item.Edges.Devices),
			})
		}
	}

	pageData.Page = allIpGroup.Page
	pageData.PageSize = allIpGroup.Size
	pageData.Total = allIpGroup.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, "pages/ip-group.html", pageData)
}
