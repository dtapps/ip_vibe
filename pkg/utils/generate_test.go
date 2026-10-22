package utils

import (
	"fmt"
	"testing"
)

func Test_GenerateIDBasedHash(t *testing.T) {
	databaseIDs := []int{1, 2, 3, 100, 1000}

	for _, dbID := range databaseIDs {
		uniqueID := GenerateIDBasedHash(dbID)
		fmt.Printf("Database ID: %d -> Generated Unique ID (Base64): %s\n", dbID, uniqueID)
	}

	// 验证
	sameIDCheck := GenerateIDBasedHash(1)
	fmt.Printf("Re-generating for ID 1 (Base64): %s\n", sameIDCheck) // 应该与第一次生成的相同
}

func Test_GenerateShortIDBasedHash(t *testing.T) {
	databaseIDs := []int{1, 2, 3, 100, 1000}

	for _, dbID := range databaseIDs {
		uniqueID := GenerateShortIDBasedHash(dbID, 10)
		fmt.Printf("Database ID: %d -> Generated Unique ID (Base64): %s\n", dbID, uniqueID)
	}

	// 验证
	sameIDCheck := GenerateShortIDBasedHash(1, 10)
	fmt.Printf("Re-generating for ID 1 (Base64): %s\n", sameIDCheck) // 应该与第一次生成的相同
}
