package schematype

import (
	"encoding/json"
	"fmt"
)

type ParsePCJODomainNameStr struct {
	DomainName string `json:"domain_name"`
}

// 解析 PlatformConfig 的 config_json 字段 的 域名 信息
func ParsePCJODomainName(configJSON []byte) (ParsePCJODomainNameStr, error) {
	var configStr ParsePCJODomainNameStr
	if err := json.Unmarshal(configJSON, &configStr); err != nil {
		return ParsePCJODomainNameStr{}, err
	}
	return configStr, nil
}

// 拼接 PlatformConfig 的 config_json 字段 的 域名 字符串
func SplicingPCJODomainName(subDomain string, domainName string) string {
	return fmt.Sprintf("%s.%s", subDomain, domainName)
}
