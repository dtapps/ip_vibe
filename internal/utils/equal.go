package utils

import (
	"sort"
	"strings"
)

// EqualStringIgnoreCase 忽略大小写的字符串比较
func EqualStringIgnoreCase(a, b string) bool {
	// 去除前后空格
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	return strings.EqualFold(a, b)
}

// EqualStringSliceUnordered 比较两个字符串切片是否包含相同的唯一元素（忽略顺序和重复）
func EqualStringSliceUnordered(a, b []string) bool {
	// 辅助函数：返回去重且排序后的副本
	sortedUnique := func(ss []string) []string {
		if len(ss) == 0 {
			return []string{}
		}
		// 1. 复制并排序
		copySS := make([]string, len(ss))
		copy(copySS, ss)
		sort.Strings(copySS)

		// 2. 去重（保留有序）
		result := []string{copySS[0]}
		for i := 1; i < len(copySS); i++ {
			if copySS[i] != copySS[i-1] {
				result = append(result, copySS[i])
			}
		}
		return result
	}

	aClean := sortedUnique(a)
	bClean := sortedUnique(b)

	// 3. 比较两个去重排序后的切片是否完全一致
	if len(aClean) != len(bClean) {
		return false
	}
	for i := range aClean {
		if aClean[i] != bClean[i] {
			return false
		}
	}
	return true
}
