package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/provider/dns"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取平台配置Schema
//
//	@Description	获取平台配置Schema
//	@Summary		获取平台配置Schema
//	@Tags			Schema
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_type	path		string													true	"平台类型"
//	@Success		200			{object}	APIResponse{data=types.GetPlatformConfigSchemaResponse}	"成功响应"
//	@Failure		400			{object}	APIResponse{}											"参数错误"
//	@Failure		404			{object}	APIResponse{}											"错误响应"
//	@Router			/schemas/{config_type}/platform-config [get]
func GetPlatformConfigSchema(c *gin.Context) {
	// 请求参数
	var requestParams types.GetPlatformConfigSchemaRequest
	requestParams.ConfigType = c.Param("config_type")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
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
//	@Param			config_type	path		string												true	"平台类型"
//	@Param			config_id	path		int64												true	"平台配置ID"
//	@Success		200			{object}	APIResponse{data=types.GetCdnRecordSchemaResponse}	"成功响应"
//	@Failure		400			{object}	APIResponse{}										"参数错误"
//	@Failure		404			{object}	APIResponse{}										"错误响应"
//	@Router			/schemas/{config_type}/cdn-record/{config_id} [get]
func GetCdnRecordSchema(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordSchemaRequest
	requestParams.ConfigType = c.Param("config_type")
	requestParams.ConfigID = cast.ToInt64(c.Param("config_id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备响应
	response := &types.GetCdnRecordSchemaResponse{}

	successResponse(c, response)
}

// 获取DNS记录Schema
//
//	@Description	获取DNS记录Schema
//	@Summary		获取DNS记录Schema
//	@Tags			Schema
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_type	path		string												true	"平台类型"
//	@Param			config_id	path		int64												true	"平台配置ID"
//	@Success		200			{object}	APIResponse{data=types.GetDnsRecordSchemaResponse}	"成功响应"
//	@Failure		400			{object}	APIResponse{}										"参数错误"
//	@Failure		404			{object}	APIResponse{}										"错误响应"
//	@Router			/schemas/{config_type}/dns-record/{config_id} [get]
func GetDnsRecordSchema(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDnsRecordSchemaRequest
	requestParams.ConfigType = c.Param("config_type")
	requestParams.ConfigID = cast.ToInt64(c.Param("config_id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
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
	var fieldGroup any

	var dnsProviderOpts = make([]dns.DnsOption, 0)
	if global.GvaHttpLog != nil {
		dnsProviderOpts = append(dnsProviderOpts, dns.WithHttpTransport(global.GvaHttpLog))
	}

	switch firstConfig.Platform {
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
		// 查询线路
		lines, err := dnsProvider.ListLines(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		if len(lines) == 0 {
			errorResponse(c, http.StatusNotFound, "未查询到线路")
			return
		}
		// 准备 options
		options := make([]types.SCHEMAFieldOption, len(lines))
		for i, line := range lines {
			options[i] = types.SCHEMAFieldOption{
				Value: line.LineID,
				Label: line.LineName,
			}
		}
		// 构造字段组（内层数组）
		fieldGroup = []types.SCHEMAField{
			{
				ID:          "record_line",
				Label:       "线路",
				Type:        "select",
				Placeholder: "请选择线路",
				Required:    true,
				Options:     options,
			},
			{
				ID:          "remark",
				Label:       "备注",
				Type:        "text",
				Placeholder: "请输入备注",
				Required:    false,
			},
		}
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
		// 查询线路
		lines, err := dnsProvider.ListLines(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		if len(lines) == 0 {
			errorResponse(c, http.StatusNotFound, "未查询到线路")
			return
		}
		// 准备 options
		options := make([]types.SCHEMAFieldOption, len(lines))
		for i, line := range lines {
			options[i] = types.SCHEMAFieldOption{
				Value: line.LineID,
				Label: line.LineName,
			}
		}
		// 构造字段组
		fieldGroup = types.SCHEMAField{
			ID:          "line",
			Label:       "线路",
			Type:        "select",
			Placeholder: "请选择线路",
			Required:    true,
			Options:     options,
		}
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
		// 查询线路
		lines, err := dnsProvider.ListLines(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
		if len(lines) == 0 {
			errorResponse(c, http.StatusNotFound, "未查询到线路")
			return
		}
		// 准备 options
		options := make([]types.SCHEMAFieldOption, len(lines))
		for i, line := range lines {
			options[i] = types.SCHEMAFieldOption{
				Value: line.LineID,
				Label: line.LineName,
			}
		}
		// 构造字段组（内层数组）
		fieldGroup = []types.SCHEMAField{
			{
				ID:          "line",
				Label:       "线路",
				Type:        "select",
				Placeholder: "请选择线路",
				Required:    true,
				Options:     options,
			},
			{
				ID:          "remark",
				Label:       "备注",
				Type:        "text",
				Placeholder: "请输入备注",
				Required:    false,
			},
		}
	default:
		errorResponse(c, http.StatusBadRequest, "未知的平台类型")
		return
	}

	// 把 fieldGroup（[]SCHEMAField）整体 marshal 成一个 JSON 数组
	rawFields, err := json.Marshal(fieldGroup)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "序列化字段失败: "+err.Error())
		return
	}

	// 然后放进 []json.RawMessage
	response.Fields = []json.RawMessage{rawFields}

	successResponse(c, response)
}
