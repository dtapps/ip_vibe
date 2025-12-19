package dns_helper

import (
	"context"
	"encoding/json"
)

// 参数
type DNSProcessParams struct {
	ConfigJSON  json.RawMessage // 配置
	RecordID    string          // 记录ID
	SubDomain   string          // 子域名
	OriginValue string          // 源站地址
}

type DNSProcess interface {
	UpsertRecord(ctx context.Context, params *DNSProcessParams) error
}
