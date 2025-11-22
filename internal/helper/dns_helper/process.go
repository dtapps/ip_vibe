package dns_helper

import (
	"log/slog"
	"vibe_ip/ent"
)

type ProcessService struct {
	Log      *slog.Logger // 日志
	Db       *ent.Client  // 数据库
	RecordID int64        // 记录ID
	Debug    bool         // 调试模式
}
