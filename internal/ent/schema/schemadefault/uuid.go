package schemadefault

import (
	"github.com/google/uuid"
)

// DefaultUUID 生成 UUID v7
func DefaultUUID() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		// v7 生成失败通常是因为系统时间回退，极少发生
		// 备选方案：可以回退到 v4 或 panic，取决于业务容忍度
		return uuid.New()
	}
	return u
}
