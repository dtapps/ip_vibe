package dns_helper

import (
	"log/slog"
	"vibe_ip/internal/ent"
)

type ProcessService struct {
	Log      *slog.Logger // 日志
	Db       *ent.Client  // 数据库
	RecordID int64        // 记录ID
	Debug    bool         // 调试模式
}

type QueryDomainResponse struct {
	DomainID    string // 域名ID
	DomainName  string // 域名名称
	OriginType  string // 源站类型
	OriginValue string // 源站地址
}
