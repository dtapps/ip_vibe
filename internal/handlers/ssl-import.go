package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/sslimport"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
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

func SslImportDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.SSLImportDetailRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.SSLImportDetailResponse{
		OperationLogs: []ent_log.OperationLog{},
	}

	// 创建服务
	service := service.NewSSLImportService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 查询 SSL 导入记录
	firstSSLImport, err := global.GvaEntDb.SSLImport.Query().
		Where(sslimport.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询 SSL 导入记录失败: "+err.Error())
		return
	}
	item, err := service.BuildResponse(c, firstSSLImport)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	pageData.Item = *item

	// 解析证书
	parsedCert, err := utils.ParseCertificateFull([]byte(firstSSLImport.Certificate))
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "解析证书详情失败: "+err.Error())
		return
	}
	pageData.CertDetail = *parsedCert

	// 最近操作日志
	allOperationLogs, err := global.GvaEntLogDb.OperationLog.Query().
		Where(operationlog.SourceIn(operationlog.SourceUser)).
		Where(operationlog.ObjTypeEQ(operationlog.ObjTypeSSLImport)).
		WhereObjIDContains(firstSSLImport.ID).
		OrderByCreatedAtDesc().
		Limit(5).
		All(c)
	if err == nil {
		for _, item := range allOperationLogs {
			pageData.OperationLogs = append(pageData.OperationLogs, *item)
		}
	}

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "ssl_detail.html"),
		types.BasePageData[types.SSLImportDetailResponse]{
			Title: "SSL导入详情",
			Data:  pageData,
		},
		"ssl",
	)
}
