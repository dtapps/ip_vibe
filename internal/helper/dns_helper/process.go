package dns_helper

import (
	"context"
	"encoding/json"
)

// 参数
type DNSProcessParams struct {
	DNSRecordID   string // DNS记录ID
	GroupRecordID string // 源站组记录ID

	ConfigJSON  json.RawMessage // 配置
	SubDomain   string          // 子域名
	OriginValue string          // 源站值
}

type DNSProcess interface {
	// 更新记录
	UpsertRecord(ctx context.Context, params *DNSProcessParams) error
}
