package mcp_tool

import (
	"context"
	"fmt"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 注册时间工具
func RegisterTimeTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getCurrentTime",
		Description: "获取当前时间信息",
		InputSchema: utils.AutoSchema(GetCurrentTimeParams{}),
	}, getCurrentTime)
}

// 时间相关参数
type GetCurrentTimeParams struct {
	Format string `json:"format,omitempty" jsonschema:"时间格式"`
}

// 时间相关响应
type GetCurrentTimeResponse struct {
	Timestamp int64  `json:"timestamp" jsonschema:"时间戳"`
	Datetime  string `json:"datetime" jsonschema:"日期时间"`
	Date      string `json:"date" jsonschema:"日期"`
	Time      string `json:"time" jsonschema:"时间"`
	Timezone  string `json:"timezone" jsonschema:"时区"`
}

// 获取当前时间
func getCurrentTime(ctx context.Context, req *mcp.CallToolRequest, params *GetCurrentTimeParams) (*mcp.CallToolResult, any, error) {
	now := time.Now()

	if params.Format == "" {
		params.Format = "2006-01-02 15:04:05"
	}

	formattedTime := now.Format(params.Format)

	response := &GetCurrentTimeResponse{
		Timestamp: now.Unix(),
		Datetime:  formattedTime,
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		Timezone:  global.DefaultTimezone,
	}

	return successResponse(fmt.Sprintf("当前时间: %s", formattedTime), response)
}
