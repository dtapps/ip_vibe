package cdn_helper

import (
	"log/slog"
	"vibe_ip/internal/ent"

	"go.dtapp.net/library/contrib/http_log"
)

const (
	OriginProtocolHTTP   = "HTTP"
	OriginProtocolHTTPS  = "HTTPS"
	OriginProtocolFOLLOW = "FOLLOW"
)

type ProcessService struct {
	Log      *slog.Logger               // 日志
	Db       *ent.Client                // 数据库
	RecordID int64                      // 记录ID
	GroupID  int64                      // 源站组ID
	Debug    bool                       // 调试模式
	HttpLog  *http_log.LoggingTransport // HTTP日志记录器
}

// 查询加速域名列表
type EdgeOneQueryDomainResponse struct {
	DomainID        string // 域名ID
	DomainName      string // 域名名称
	DomainCname     string // 域名别名
	OriginType      string // 源站类型
	OriginValue     string // 源站值
	OriginProtocol  string // 回源协议
	HttpOriginPort  int    // HTTP回源端口
	HttpsOriginPort int    // HTTPS回源端口
}

// 获取源站组列表
type EdgeOneQueryOriginGroupResponse struct {
	GroupID      string   // 源站组ID
	GroupName    string   // 源站组名称
	GroupType    string   // 源站组类型
	GroupOrigins []string // 源站组类型值
}

// 查询加速域名列表
type EdgeSecurityAccelerationResponseQueryDomain struct {
	DomainID    string // 域名ID
	DomainName  string // 域名名称
	DomainCname string // 域名别名
	OriginType  string // 源站类型
	OriginValue string // 源站值
}
