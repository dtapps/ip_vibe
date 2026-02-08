package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/provider/cdn"
	"vibe_ip/internal/provider/dns"
	"vibe_ip/internal/provider/tun"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

// 获取平台配置Schema
//
//	@Description	获取平台配置Schema
//	@Summary		获取平台配置Schema
//	@Tags			Schema
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_type	path		string																				true	"平台类型"
//	@Success		200			{object}	types.APIResponse{data=[]json.RawMessage{types.GetPlatformConfigSchemaResponse}}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}																	"参数错误"
//	@Failure		404			{object}	types.APIResponse{}																	"错误响应"
//	@Router			/schemas/{config_type}/platform-config [get]
func GetPlatformConfigSchema(c *gin.Context) {
	// 请求参数
	var requestParams types.GetPlatformConfigSchemaRequest
	requestParams.ConfigType = c.Param("config_type")
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备响应
	response := &types.GetPlatformConfigSchemaResponse{}

	successResponse(c, response)
}

// 获取CDN记录Schema
//
//	@Description	获取CDN记录Schema
//	@Summary		获取CDN记录Schema
//	@Tags			Schema
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_type	path		string																		true	"平台类型"
//	@Param			config_id	path		int64																		true	"平台配置ID"
//	@Success		200			{object}	types.APIResponse{data=[]json.RawMessage{types.GetCdnRecordSchemaResponse}}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}															"参数错误"
//	@Failure		404			{object}	types.APIResponse{}															"错误响应"
//	@Router			/schemas/{config_type}/cdn-record/{config_id} [get]
func GetCdnRecordSchema(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordSchemaRequest
	requestParams.ConfigType = c.Param("config_type")
	requestParams.ConfigID = cast.ToInt64(c.Param("config_id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备响应
	response := &types.GetCdnRecordSchemaResponse{}

	// 查询平台配置
	firstConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(requestParams.ConfigID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "配置不存在")
		return
	}

	// 判断平台类型是否匹配
	if !strings.EqualFold(requestParams.ConfigType, firstConfig.Platform.String()) {
		errorResponse(c, http.StatusBadRequest, "平台类型与配置类型不匹配")
		return
	}

	// 准备字段组
	fields := make([]any, 0)

	var cdnProviderOpts = make([]cdn.Option, 0, 1)
	if global.GvaHttpLog != nil {
		cdnProviderOpts = append(cdnProviderOpts, cdn.WithHttpTransport(global.GvaHttpLog))
	}

	switch firstConfig.Platform {
	case platformconfig.PlatformTencentEo:

		cdnProviderCfg := cdn.NewDefaultTencentEOConfig()
		err = cdnProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		cdnProvider, err := cdn.NewTencentEOProviderRaw(c, global.GvaGocronSlog.With(), cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		zoneOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 获取域名列表
			zones, err := cdnProvider.ZonesRaw(egCtx)
			if err != nil {
				return err
			}
			for _, zone := range zones {
				zoneOptions = append(zoneOptions, types.SCHEMAFieldOption{
					Value: utils.PtrToString(zone.ZoneId),
					Label: utils.TrimDotsPrt(zone.ZoneName),
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "zone",
			Label:       "站点",
			Type:        "select",
			Placeholder: "请选择站点",
			Required:    true,
			Options:     zoneOptions,
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "ipv6_status",
				Label:       "IPv6 状态",
				Type:        "select",
				Placeholder: "请选择 IPv6 状态",
				Required:    true,
				Options: []types.SCHEMAFieldOption{
					{
						Value: "follow",
						Label: "遵循站点",
					},
					{
						Value: "on",
						Label: "开启",
					},
					{
						Value: "off",
						Label: "关闭",
					},
				},
			},
			{
				ID:          "origin_protocol",
				Label:       "回源协议",
				Type:        "select",
				Placeholder: "请选择回源协议",
				Required:    true,
				Options: []types.SCHEMAFieldOption{
					{
						Value: "FOLLOW",
						Label: "协议跟随",
					},
					{
						Value: "HTTP",
						Label: "HTTP协议回源",
					},
					{
						Value: "HTTPS",
						Label: "HTTPS协议回源",
					},
				},
			},
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "http_port",
				Label:       "HTTP端口",
				Type:        "number",
				Placeholder: "如 80",
				Required:    true,
			},
			{
				ID:          "https_port",
				Label:       "HTTPS端口",
				Type:        "number",
				Placeholder: "如 443",
				Required:    true,
			},
		})

	case platformconfig.PlatformAliyunEsa:

		cdnProviderCfg := cdn.NewDefaultAliyunESAConfig()
		err = cdnProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		cdnProvider, err := cdn.NewAliyunESAProviderRaw(c, global.GvaGocronSlog.With(), cdnProviderCfg, cdnProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		siteOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 获取域名列表
			sites, err := cdnProvider.SitesRaw(egCtx)
			if err != nil {
				return err
			}
			for _, zonesite := range sites {
				siteOptions = append(siteOptions, types.SCHEMAFieldOption{
					Value: utils.PtrToInt64(zonesite.SiteId),
					Label: utils.TrimDotsPrt(zonesite.SiteName),
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "site",
			Label:       "站点",
			Type:        "select",
			ValueType:   "number",
			Placeholder: "请选择站点",
			Required:    true,
			Options:     siteOptions,
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "biz_name",
				Label:       "业务场景",
				Type:        "select",
				Placeholder: "请选择业务场景",
				Required:    true,
				Options: []types.SCHEMAFieldOption{
					{
						Value: "web",
						Label: "web网页",
					},
					{
						Value: "api",
						Label: "api接口",
					},
					{
						Value: "video_image",
						Label: "视频图片",
					},
				},
			},
			{
				ID:          "origin_scheme",
				Label:       "回源协议",
				Type:        "select",
				Placeholder: "请选择回源协议",
				Required:    true,
				Options: []types.SCHEMAFieldOption{
					{
						Value: "follow",
						Label: "协议跟随",
					},
					{
						Value: "http",
						Label: "HTTP协议回源",
					},
					{
						Value: "https",
						Label: "HTTPS协议回源",
					},
				},
			},
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "origin_http_port",
				Label:       "HTTP端口",
				Type:        "number",
				Placeholder: "如 80",
				Required:    true,
			},
			{
				ID:          "origin_https_port",
				Label:       "HTTPS端口",
				Type:        "number",
				Placeholder: "如 443",
				Required:    true,
			},
		})
		fields = append(fields, types.SCHEMAField{
			ID:          "comment",
			Label:       "备注",
			Type:        "text",
			Placeholder: "请输入备注",
			Required:    false,
		})

	default:
		errorResponse(c, http.StatusBadRequest, "未知的平台类型")
		return
	}

	// 把 multipleFields + singleField 整体 marshal 成一个 JSON 数组
	rawFields, err := json.Marshal(fields)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "序列化字段失败: "+err.Error())
		return
	}

	// 然后放进 []json.RawMessage
	response.Fields = []json.RawMessage{rawFields}

	successResponse(c, response)
}

// 获取DNS记录Schema
//
//	@Description	获取DNS记录Schema
//	@Summary		获取DNS记录Schema
//	@Tags			Schema
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_type	path		string																		true	"平台类型"
//	@Param			config_id	path		int64																		true	"平台配置ID"
//	@Success		200			{object}	types.APIResponse{data=[]json.RawMessage{types.GetDnsRecordSchemaResponse}}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}															"参数错误"
//	@Failure		404			{object}	types.APIResponse{}															"错误响应"
//	@Router			/schemas/{config_type}/dns-record/{config_id} [get]
func GetDnsRecordSchema(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDnsRecordSchemaRequest
	requestParams.ConfigType = c.Param("config_type")
	requestParams.ConfigID = cast.ToInt64(c.Param("config_id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备响应
	response := &types.GetDnsRecordSchemaResponse{}

	// 查询平台配置
	firstConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(requestParams.ConfigID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "配置不存在")
		return
	}

	// 判断平台类型是否匹配
	if !strings.EqualFold(requestParams.ConfigType, firstConfig.Platform.String()) {
		errorResponse(c, http.StatusBadRequest, "平台类型与配置类型不匹配")
		return
	}

	// 准备字段组
	fields := make([]any, 0)

	var dnsProviderOpts = make([]dns.Option, 0, 1)
	if global.GvaHttpLog != nil {
		dnsProviderOpts = append(dnsProviderOpts, dns.WithHttpTransport(global.GvaHttpLog))
	}

	switch firstConfig.Platform {
	case platformconfig.PlatformCloudflare:

		dnsProviderCfg := dns.NewDefaultCloudfareDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		dnsProvider, err := dns.NewCloudfareDNSProviderRaw(c, global.GvaGocronSlog.With(), dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		zoneOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 获取所有站点
			zones, err := dnsProvider.ZonesRaw(egCtx)
			if err != nil {
				return err
			}
			for _, zone := range zones {
				zoneOptions = append(zoneOptions, types.SCHEMAFieldOption{
					Value: zone.ID,
					Label: utils.TrimDots(zone.Name),
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "zone",
			Label:       "域名",
			Type:        "select",
			Placeholder: "请选择域名",
			Required:    true,
			Options:     zoneOptions,
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "comment",
				Label:       "描述",
				Type:        "text",
				Placeholder: "请输入描述",
				Required:    false,
			},
		})

	case platformconfig.PlatformTencentDNS:

		dnsProviderCfg := dns.NewDefaultTencentDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		dnsProvider, err := dns.NewTencentDNSProviderRaw(c, global.GvaGocronSlog.With(), dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		domainOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 获取域名列表
			domains, err := dnsProvider.DescribeDomainList(egCtx)
			if err != nil {
				return err
			}
			for _, domain := range domains {
				domainOptions = append(domainOptions, types.SCHEMAFieldOption{
					Value: utils.PtrToUint64(domain.DomainId),
					Label: utils.TrimDotsPrt(domain.Name),
				})
			}
			return nil
		})

		// 准备选项
		recordLineOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 查询线路列表
			lines, err := dnsProvider.ListLines(egCtx)
			if err != nil {
				return err
			}
			for _, line := range lines {
				recordLineOptions = append(recordLineOptions, types.SCHEMAFieldOption{
					Value: line.LineID,
					Label: line.LineName,
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "domain",
			Label:       "域名",
			Type:        "select",
			ValueType:   "number",
			Placeholder: "请选择域名",
			Required:    true,
			Options:     domainOptions,
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "record_line",
				Label:       "线路",
				Type:        "select",
				Placeholder: "请选择线路",
				Required:    true,
				Options:     recordLineOptions,
			},
			{
				ID:          "remark",
				Label:       "备注",
				Type:        "text",
				Placeholder: "请输入备注",
				Required:    false,
			},
		})

	case platformconfig.PlatformAliyunDNS:

		dnsProviderCfg := dns.NewDefaultAliyunDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		dnsProvider, err := dns.NewAliyunDNSProviderRaw(c, global.GvaGocronSlog.With(), dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		domainOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 获取域名列表
			lines, err := dnsProvider.DescribeDomains(egCtx)
			if err != nil {
				return err
			}
			for _, line := range lines {
				domainOptions = append(domainOptions, types.SCHEMAFieldOption{
					Value: utils.PtrToString(line.DomainId),
					Label: utils.TrimDotsPrt(line.DomainName),
				})
			}
			return nil
		})

		// 准备选项
		lineOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 查询线路列表
			lines, err := dnsProvider.ListLines(egCtx)
			if err != nil {
				return err
			}
			for _, line := range lines {
				lineOptions = append(lineOptions, types.SCHEMAFieldOption{
					Value: line.LineID,
					Label: line.LineName,
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "domain",
			Label:       "域名",
			Type:        "select",
			Placeholder: "请选择域名",
			Required:    true,
			Options:     domainOptions,
		})
		fields = append(fields, []types.SCHEMAField{{
			ID:          "line",
			Label:       "线路",
			Type:        "select",
			Placeholder: "请选择线路",
			Required:    true,
			Options:     lineOptions,
		}})

	case platformconfig.PlatformBaiduDNS:

		dnsProviderCfg := dns.NewDefaultBaiduDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		dnsProvider, err := dns.NewBaiduDNSProviderRaw(c, global.GvaGocronSlog.With(), dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		zoneOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 获取域名列表
			lines, err := dnsProvider.ListZone(egCtx)
			if err != nil {
				return err
			}
			for _, line := range lines {
				zoneOptions = append(zoneOptions, types.SCHEMAFieldOption{
					Value: line.Id,
					Label: utils.TrimDots(line.Name),
				})
			}
			return nil
		})

		// 准备选项
		lineOptions := make([]types.SCHEMAFieldOption, 0)
		lineOptions = append(lineOptions, types.SCHEMAFieldOption{
			Value: "default",
			Label: "默认",
		})
		lineOptions = append(lineOptions, types.SCHEMAFieldOption{
			Value: "ct",
			Label: "电信",
		})
		lineOptions = append(lineOptions, types.SCHEMAFieldOption{
			Value: "cmnet",
			Label: "移动",
		})
		lineOptions = append(lineOptions, types.SCHEMAFieldOption{
			Value: "cnc",
			Label: "联通",
		})
		lineOptions = append(lineOptions, types.SCHEMAFieldOption{
			Value: "edu",
			Label: "教育网",
		})
		lineOptions = append(lineOptions, types.SCHEMAFieldOption{
			Value: "search",
			Label: "搜索引擎(百度)",
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "zone",
			Label:       "域名",
			Type:        "select",
			Placeholder: "请选择域名",
			Required:    true,
			Options:     zoneOptions,
		})
		fields = append(fields, []types.SCHEMAField{{
			ID:          "line",
			Label:       "线路",
			Type:        "select",
			Placeholder: "请选择线路",
			Required:    true,
			Options:     lineOptions,
		}, {
			ID:          "description",
			Label:       "描述",
			Type:        "text",
			Placeholder: "请输入描述",
			Required:    false,
		}})

	case platformconfig.PlatformHuaweiDNS:

		dnsProviderCfg := dns.NewDefaultHuaweiDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		dnsProvider, err := dns.NewHuaweiDNSProviderRaw(c, global.GvaGocronSlog.With(), dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		zoneOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 查询域名列表
			zones, err := dnsProvider.ListPublicZones(egCtx)
			if err != nil {
				return err
			}
			for _, zone := range *zones {
				zoneOptions = append(zoneOptions, types.SCHEMAFieldOption{
					Value: zone.Id,
					Label: utils.TrimDotsPrt(zone.Name),
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "zone",
			Label:       "域名",
			Type:        "select",
			Placeholder: "请输入域名",
			Required:    true,
			Options:     zoneOptions,
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "description",
				Label:       "描述",
				Type:        "text",
				Placeholder: "请输入描述",
				Required:    false,
			},
		})

	case platformconfig.PlatformVolcengineDNS:

		dnsProviderCfg := dns.NewDefaultVolcengineDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		dnsProvider, err := dns.NewVolcengineDNSProviderRaw(c, global.GvaGocronSlog.With(), dnsProviderCfg, dnsProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备选项
		zoneOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 查询域名列表
			zones, err := dnsProvider.ListZones(egCtx)
			if err != nil {
				return err
			}
			for _, zone := range zones {
				zoneOptions = append(zoneOptions, types.SCHEMAFieldOption{
					Value: utils.PtrToInt32(zone.ZID),
					Label: utils.TrimDotsPrt(zone.ZoneName),
				})
			}
			return nil
		})

		// 准备选项
		lineOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 查询线路列表
			lines, err := dnsProvider.ListLines(egCtx)
			if err != nil {
				return err
			}
			for _, line := range lines {
				lineOptions = append(lineOptions, types.SCHEMAFieldOption{
					Value: line.LineID,
					Label: line.LineName,
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, types.SCHEMAField{
			ID:          "zone",
			Label:       "域名",
			Type:        "select",
			ValueType:   "number",
			Placeholder: "请输入域名",
			Required:    true,
			Options:     zoneOptions,
		})
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "line",
				Label:       "线路",
				Type:        "select",
				Placeholder: "请选择线路",
				Required:    true,
				Options:     lineOptions,
			},
			{
				ID:          "remark",
				Label:       "备注",
				Type:        "text",
				Placeholder: "请输入备注",
				Required:    false,
			},
		})

	default:
		errorResponse(c, http.StatusBadRequest, "未知的平台类型")
		return
	}

	// 把 multipleFields + singleField 整体 marshal 成一个 JSON 数组
	rawFields, err := json.Marshal(fields)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "序列化字段失败: "+err.Error())
		return
	}

	// 然后放进 []json.RawMessage
	response.Fields = []json.RawMessage{rawFields}

	successResponse(c, response)
}

// 获取TUN记录Schema
//
//	@Description	获取TUN记录Schema
//	@Summary		获取TUN记录Schema
//	@Tags			Schema
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_type	path		string																		true	"平台类型"
//	@Param			config_id	path		int64																		true	"平台配置ID"
//	@Success		200			{object}	types.APIResponse{data=[]json.RawMessage{types.GetTunRecordSchemaResponse}}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}															"参数错误"
//	@Failure		404			{object}	types.APIResponse{}															"错误响应"
//	@Router			/schemas/{config_type}/tun-record/{config_id} [get]
func GetTunRecordSchema(c *gin.Context) {
	// 请求参数
	var requestParams types.GetTunRecordSchemaRequest
	requestParams.ConfigType = c.Param("config_type")
	requestParams.ConfigID = cast.ToInt64(c.Param("config_id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备响应
	response := &types.GetTunRecordSchemaResponse{}

	// 查询平台配置
	firstConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(requestParams.ConfigID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "配置不存在")
		return
	}

	// 判断平台类型是否匹配
	if !strings.EqualFold(requestParams.ConfigType, firstConfig.Platform.String()) {
		errorResponse(c, http.StatusBadRequest, "平台类型与配置类型不匹配")
		return
	}

	// 准备字段组
	fields := make([]any, 0)

	var tunProviderOpts = make([]tun.Option, 0, 1)
	if global.GvaHttpLog != nil {
		tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(global.GvaHttpLog))
	}

	switch firstConfig.Platform {
	case platformconfig.PlatformCloudflare:

		tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
		err = tunProviderCfg.ParseCustomConfig(firstConfig.ConfigJSON)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		tunProvider, err := tun.NewCloudfareTUNProviderRaw(c, global.GvaGocronSlog.With(), tunProviderCfg, tunProviderOpts...)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 并发检查
		eg, egCtx := errgroup.WithContext(c)
		eg.SetLimit(10) // 最多 10 个并发

		// 准备隧道选项
		tunnelOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 查询隧道列表
			connectors, err := tunProvider.Connectors(egCtx)
			if err != nil {
				return err
			}
			for _, connector := range connectors {
				tunnelOptions = append(tunnelOptions, types.SCHEMAFieldOption{
					Value: connector.ID,
					Label: connector.Name,
				})
			}
			return nil
		})

		// 准备选项
		zoneOptions := make([]types.SCHEMAFieldOption, 0)
		eg.Go(func() error {
			// 查询域名列表
			zones, err := tunProvider.Zones(egCtx)
			if err != nil {
				return err
			}
			for _, zone := range zones {
				zoneOptions = append(zoneOptions, types.SCHEMAFieldOption{
					Value: zone.ID,
					Label: zone.Name,
				})
			}
			return nil
		})

		if err = eg.Wait(); err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 构造字段
		fields = append(fields, []types.SCHEMAField{
			{
				ID:          "tunnel_id",
				Label:       "隧道",
				Type:        "select",
				Placeholder: "请选择隧道",
				Required:    true,
				Options:     tunnelOptions,
			},
			{
				ID:          "zone_id",
				Label:       "域名",
				Type:        "select",
				Placeholder: "请选择域名",
				Required:    true,
				Options:     zoneOptions,
			},
		})
		fields = append(fields, types.SCHEMAField{
			ID:          "remark",
			Label:       "备注",
			Type:        "text",
			Placeholder: "请输入备注",
			Required:    false,
		})

	default:
		errorResponse(c, http.StatusBadRequest, "未知的平台类型")
		return
	}

	// 把 multipleFields + singleField 整体 marshal 成一个 JSON 数组
	rawFields, err := json.Marshal(fields)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "序列化字段失败: "+err.Error())
		return
	}

	// 然后放进 []json.RawMessage
	response.Fields = []json.RawMessage{rawFields}

	successResponse(c, response)
}
