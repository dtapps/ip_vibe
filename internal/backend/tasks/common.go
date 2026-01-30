package tasks

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
)

func allowConcurrent(item *ent.PlatformConfig) bool {
	// EdgeOne 平台不支持并发，串行；其他并发
	if item.Platform == platformconfig.PlatformTencentEo {
		return false
	}
	return true
}
