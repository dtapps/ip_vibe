package handlers

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/sslcertificatehistory"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func Ssl(c *gin.Context) {
	// 请求参数
	var requestParams types.SSLRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.SSLResponse{
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
		CommonPageResponse: types.CommonPageResponse[types.GetSSLRecordResponse]{
			List: []types.GetSSLRecordResponse{},
		},
	}

	// 创建SSL更新任务
	sslUpdateTask := task.NewSSLUpdateTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 创建服务
	service := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		sslUpdateTask,
		global.GvaEnv.DEBUG,
	)

	// 查询全部SSL记录
	allSSLRecord, err := global.GvaEntDb.SSLRecord.Query().
		WithConfig().
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

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "ssl.html"), pageData)
}

func SslDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.SSLDetailRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.SSLDetailResponse{
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
	}

	// 查询 SSL 记录
	firstSSLRecord, err := global.GvaEntDb.SSLRecord.Query().
		Where(sslrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询 SSL 记录失败: "+err.Error())
		return
	}
	pageData.Item = *firstSSLRecord

	pageData.CertDetail = *ParseSSLCertificate(firstSSLRecord.CertCertificate)

	// 查询 SSL 证书申请历史记录
	allSSLCertificateHistory, err := global.GvaEntLogDb.SSLCertificateHistory.Query().
		Where(sslcertificatehistory.IDEQ(requestParams.ID)).
		OrderByCreatedAtDesc().
		Limit(5).
		All(c)
	if err == nil {
		for _, item := range allSSLCertificateHistory {
			pageData.Historys = append(pageData.Historys, *item)
		}
	}

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "ssl_detail.html"), pageData)
}

func SslHistory(c *gin.Context) {
	// 请求参数
	var requestParams types.SSLHistoryRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.SSLHistoryResponse{
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
		CommonPageResponse: types.CommonPageResponse[ent_log.SSLCertificateHistory]{
			List: []ent_log.SSLCertificateHistory{},
		},
	}

	// 查询 SSL 记录
	firstSSLRecord, err := global.GvaEntDb.SSLRecord.Query().
		Where(sslrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询 SSL 记录失败: "+err.Error())
		return
	}
	pageData.Item = *firstSSLRecord

	// 查询 SSL 证书申请历史记录
	allSSLCertificateHistory, err := global.GvaEntLogDb.SSLCertificateHistory.Query().
		Where(sslcertificatehistory.IDEQ(requestParams.ID)).
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allSSLCertificateHistory.Items {
			pageData.List = append(pageData.List, *item)
		}
	}

	pageData.Page = allSSLCertificateHistory.Page
	pageData.PageSize = allSSLCertificateHistory.Size
	pageData.Total = allSSLCertificateHistory.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "ssl_history.html"), pageData)
}

func ParseSSLCertificate(certPEM string) *types.SSLDetailCertDetail {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	now := time.Now()

	// 判断是否包含通配符
	hasWildcard := false
	for _, name := range cert.DNSNames {
		if strings.Contains(name, "*") {
			hasWildcard = true
			break
		}
	}

	// 获取密钥长度
	keyLen := 0
	isECC := false
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		keyLen = pub.N.BitLen()
	case *ecdsa.PublicKey:
		isECC = true
		keyLen = pub.Curve.Params().BitSize
	}

	// 计算百分比
	total := cert.NotAfter.Sub(cert.NotBefore).Seconds()
	elapsed := now.Sub(cert.NotBefore).Seconds()
	percent := int((elapsed / total) * 100)
	if percent > 100 {
		percent = 100
	} else if percent < 0 {
		percent = 0
	}

	return &types.SSLDetailCertDetail{
		Subject:            cert.Subject.CommonName,
		Issuer:             cert.Issuer.CommonName,
		DNSNames:           cert.DNSNames,
		SerialNumber:       cert.SerialNumber.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		KeyLength:          keyLen,
		IsECC:              isECC,
		IsWildcard:         hasWildcard,
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		ExpiryDays:         int(cert.NotAfter.Sub(now).Hours() / 24),
		ExpiryPercent:      percent,
		IsExpired:          now.After(cert.NotAfter),
	}
}
