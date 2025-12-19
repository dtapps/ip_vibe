package dns_helper

import (
	"context"
	"encoding/json"
)

// 参数
type DNSProcessParams struct {
	DNSRecordID string // DNS记录ID

	ConfigJSON  json.RawMessage // 配置
	SubDomain   string          // 子域名
	OriginValue string          // 源站值
}

type DNSProcess interface {
	UpsertRecord(ctx context.Context, params *DNSProcessParams) error
}
