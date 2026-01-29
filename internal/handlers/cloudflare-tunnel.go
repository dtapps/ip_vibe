package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent/cloudflaretunnel"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func CloudflareTunnel(c *gin.Context) {
	// 请求参数
	var requestParams types.CloudflareTunnelRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.CloudflareTunnelResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetCloudflareTunnelResponse, types.CloudflareTunnelRequest]{
			Filter: requestParams,
			List:   []types.GetCloudflareTunnelResponse{},
		},
	}

	// 查询全部Cloudflare配置
	allCloudflareConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		Where(platformconfig.CapabilityIn(platformconfig.CapabilityAll, platformconfig.CapabilityTun)).
		OrderByCreatedAtAsc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allCloudflareConfig.Items {
			tunnels := make([]types.GetCloudflareTunnelTunnel, 0)
			// 查询隧道列表
			allTunnels, err := global.GvaEntDb.CloudflareTunnel.Query().
				Where(cloudflaretunnel.ConfigID(item.ID)).
				WithIngresses().
				OrderByNameAsc().
				All(c)
			if err == nil {
				for _, tunnel := range allTunnels {
					ingresses := make([]types.GetCloudflareTunnelIngressDetailResponse, 0, len(tunnel.Edges.Ingresses))
					for _, ingress := range tunnel.Edges.Ingresses {
						ingresses = append(ingresses, types.GetCloudflareTunnelIngressDetailResponse{
							ID:            ingress.ID,
							Hostname:      ingress.Hostname,
							Service:       ingress.Service,
							Path:          ingress.Path,
							OriginRequest: ingress.OriginRequest,
						})
					}
					tunnels = append(tunnels, types.GetCloudflareTunnelTunnel{
						ID:              tunnel.ID,
						TunnelID:        tunnel.TunnelID,
						Name:            tunnel.Name,
						Status:          tunnel.Status.String(),
						Connections:     tunnel.Connections,
						ConnsActiveAt:   tunnel.ConnsActiveAt,
						RemoteCreatedAt: tunnel.RemoteCreatedAt,
						Ingresses:       ingresses,
					})
				}
			}
			pageData.List = append(pageData.List, types.GetCloudflareTunnelResponse{
				ID:          item.ID,
				Capability:  item.Capability.String(),
				Name:        item.Name,
				Description: item.Description,
				Platform:    item.Platform.String(),
				Tunnels:     tunnels,
			})
		}
	}

	pageData.Page = allCloudflareConfig.Page
	pageData.PageSize = allCloudflareConfig.Size
	pageData.Total = allCloudflareConfig.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "cloudflare-tunnel.html"),
		types.BasePageData[types.CloudflareTunnelResponse]{
			Title: "CF 隧道管理",
			Data:  pageData,
		},
	)
}
