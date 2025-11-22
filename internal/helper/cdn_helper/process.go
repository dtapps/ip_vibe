package cdn_helper

import (
	"log/slog"
	"vibe_ip/ent"
)

type ProcessService struct {
	Log      *slog.Logger // 日志
	Db       *ent.Client  // 数据库
	RecordID int64        // 记录ID
	GroupID  int64        // 源站组ID
	Debug    bool         // 调试模式
}
