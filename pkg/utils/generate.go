package utils

import (
	"crypto/sha256"
	"fmt"
)

// GenerateIDBasedHash 根据数据库表的 id 生成一个固定且唯一的标识符
// 输出只包含 0-9 和 a-f，符合 a-z, A-Z, 0-9, _, - 的要求
func GenerateIDBasedHash(dbID int) string {
	idBytes := []byte(fmt.Sprintf("%d", dbID))
	hash := sha256.Sum256(idBytes)
	return fmt.Sprintf("%x", hash) // %x 输出小写十六进制
}

// GenerateShortIDBasedHash 截取 SHA256 哈希值的一部分
func GenerateShortIDBasedHash[T int | int64](dbID T, length int) string {
	idBytes := []byte(fmt.Sprintf("%d", dbID))
	hash := sha256.Sum256(idBytes)
	fullHash := fmt.Sprintf("%x", hash)

	// 确保截取长度不超过 64
	if length > 64 {
		length = 64
	}

	return fullHash[:length]
}
