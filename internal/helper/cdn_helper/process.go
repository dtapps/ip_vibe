package cdn_helper

import (
	"context"
	"encoding/json"
)

// 参数
type CDNProcessParams struct {
	CDNRecordID   string // CDN记录ID
	GroupRecordID string // 源站组记录ID

	ConfigJSON     json.RawMessage // 配置
	GroupName      string          // 源站组名称
	SubDomain      string          // 子域名
	OriginValues   []string        // 源站地址列表
	OriginProtocol string          // 回源协议
	HttpPort       int             // HTTP回源端口
	HttpsPort      int             // HTTPS回源端口

	originValue string // 源站地址
}

type CDNProcess interface {
	UpsertRecord(ctx context.Context, params *CDNProcessParams) error
}
