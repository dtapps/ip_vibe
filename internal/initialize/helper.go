package initialize

import (
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/operation_logs_helper"
)

// InitOperationLogHelper 初始化操作日志助手
func InitOperationLogHelper() {
	global.GvaOperationLogHelper = operation_logs_helper.NewOperationLogHelper()
}
