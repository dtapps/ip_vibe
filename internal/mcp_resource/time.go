package mcp_resource

import (
	"context"
	"time"
	"vibe_ip/internal/global"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 注册间资源
func RegisterTimeResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "current-time",
		Description: "获取当前时间",
		URI:         "time://current",
	}, getCurrentTimeResource)
}

// 时间相关响应
type GetCurrentTimeResponse struct {
	Timestamp int64  `json:"timestamp" jsonschema:"时间戳"`
	Datetime  string `json:"datetime" jsonschema:"日期时间"`
	Date      string `json:"date" jsonschema:"日期"`
	Time      string `json:"time" jsonschema:"时间"`
	Timezone  string `json:"timezone" jsonschema:"时区"`
	IsToday   bool   `json:"is_today" jsonschema:"是否为今天"`
}

func getCurrentTimeResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	now := time.Now()

	response := GetCurrentTimeResponse{
		Timestamp: now.Unix(),
		Datetime:  now.Format("2006-01-02 15:04:05"),
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		Timezone:  global.GvaTimeZone,
		IsToday:   now.Format("2006-01-02") == now.Format("2006-01-02"),
	}

	return successResponse(req.Params.URI, response)
}
