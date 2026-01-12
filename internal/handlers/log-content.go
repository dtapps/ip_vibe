package handlers

import (
	"net/http"
	"path/filepath"
	"strings"
	"vibe_ip/internal/ent_log/httplog"
	"vibe_ip/internal/ent_log/reqlog"
	"vibe_ip/internal/ent_log/restylog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"
	"vibe_ip/pkg/utils"

	"github.com/gin-gonic/gin"
)

func LogContent(c *gin.Context) {
	// 请求参数
	var requestParams types.LogContentRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.LogContentResponse{
		CommonPageResponse: types.CommonPageResponse[any, types.LogContentRequest]{
			Filter: requestParams,
			List:   []any{},
		},
	}

	switch requestParams.Type {
	case "http_log":

		// 查询全部日志
		allHttpLog, err := global.GvaEntLogDb.HttpLog.Query().
			WhereIf(requestParams.Hostname != "", httplog.HostnameEQ(requestParams.Hostname)).
			WhereIf(requestParams.Method != "", httplog.MethodEQ(requestParams.Method)).
			WhereIf(requestParams.StatusCode != 0, httplog.StatusCodeEQ(requestParams.StatusCode)).
			OrderByCreatedAtDesc().
			Paginate(c, requestParams.Page, requestParams.PageSize)
		if err == nil {
			for _, item := range allHttpLog.Items {
				pageData.List = append(pageData.List, &item)
			}
		}

		pageData.Page = allHttpLog.Page
		pageData.PageSize = allHttpLog.Size
		pageData.Total = allHttpLog.Total

	case "resty_log":

		// 查询全部日志
		allRestyLog, err := global.GvaEntLogDb.RestyLog.Query().
			WhereIf(requestParams.Hostname != "", restylog.HostnameEQ(requestParams.Hostname)).
			WhereIf(requestParams.Method != "", restylog.MethodEQ(requestParams.Method)).
			WhereIf(requestParams.StatusCode != 0, restylog.StatusCodeEQ(requestParams.StatusCode)).
			OrderByCreatedAtDesc().
			Paginate(c, requestParams.Page, requestParams.PageSize)
		if err == nil {
			for _, item := range allRestyLog.Items {
				pageData.List = append(pageData.List, &item)
			}
		}

		pageData.Page = allRestyLog.Page
		pageData.PageSize = allRestyLog.Size
		pageData.Total = allRestyLog.Total

	case "req_log":

		// 查询全部日志
		allReqLog, err := global.GvaEntLogDb.ReqLog.Query().
			WhereIf(requestParams.Hostname != "", reqlog.HostnameEQ(requestParams.Hostname)).
			WhereIf(requestParams.Method != "", reqlog.MethodEQ(requestParams.Method)).
			WhereIf(requestParams.StatusCode != 0, reqlog.StatusCodeEQ(requestParams.StatusCode)).
			OrderByCreatedAtDesc().
			Paginate(c, requestParams.Page, requestParams.PageSize)
		if err == nil {
			for _, item := range allReqLog.Items {
				pageData.List = append(pageData.List, &item)
			}
		}

		pageData.Page = allReqLog.Page
		pageData.PageSize = allReqLog.Size
		pageData.Total = allReqLog.Total

	default:

		lists := make([]string, 0)
		if requestParams.FileName != "" {
			// 强制提取纯文件名，防止 ../.. 攻击
			safeFileName := filepath.Base(requestParams.FileName)

			// 拼接成绝对路径
			absLogDir, _ := filepath.Abs("logs")

			// 获取 logs 目录的绝对路径
			targetPath := filepath.Join(absLogDir, safeFileName)

			// 二次检查：确保最终生成的路径依然在 logs 目录下
			if !strings.HasPrefix(targetPath, absLogDir) {
				lists = []string{"[Security] 拒绝跨目录访问"}
				return
			}

			// 日志存储在项目根目录的 logs 文件夹下
			// targetPath := filepath.Join("logs", requestParams.FileName)

			// 读取文件末尾行 (例如读取最后 200 行)
			lines, err := utils.ReadLastLines(targetPath, 200)
			if err != nil {
				// 如果文件不存在或读取失败，返回空列表或错误提示
				lists = []string{"[System] 暂无日志文件或读取失败: " + err.Error()}
			}
			lists = lines
		} else {
			lists = []string{"[System] 请输入日志文件名"}
		}

		for _, line := range lists {
			pageData.List = append(pageData.List, line)
		}

		// 文件模式下，分页数据可以设为固定或 0
		pageData.Total = int64(len(lists))
		pageData.PageSize = 0

	}

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "log-content.html"),
		types.BasePageData[types.LogContentResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "日志内容",
			Data:  pageData,
		},
	)
}
