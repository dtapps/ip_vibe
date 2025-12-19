package dns_helper

import (
	"log/slog"
	"vibe_ip/internal/ent"

	"go.dtapp.net/library/contrib/http_log"
)

type ProcessService struct {
	Log      *slog.Logger               // 日志
	Db       *ent.Client                // 数据库
	RecordID int64                      // 记录ID
	Debug    bool                       // 调试模式
	HttpLog  *http_log.LoggingTransport // HTTP日志记录器
}

type QueryDomainResponse struct {
	RecordID    string // 域名ID
	FullDomain  string // 完整域名
	RecordType  string // 记录类型
	OriginValue string // 源站地址
}
