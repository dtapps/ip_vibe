package utils

import (
	"net/url"
	"sort"
	"strings"
)

// EqualStringIgnoreCase 忽略大小写的字符串比较
func EqualStringIgnoreCase(s1, s2 string) bool {
	// 去除前后空格
	s1 = strings.TrimSpace(s1)
	s2 = strings.TrimSpace(s2)
	return strings.EqualFold(s1, s2)
}

// EqualStringSliceUnordered 比较两个字符串切片是否包含相同的唯一元素（忽略顺序和重复）
func EqualStringSliceUnordered(sm1, sm2 []string) bool {
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

	s1Clean := sortedUnique(sm1)
	s2Clean := sortedUnique(sm2)

	// 3. 比较两个去重排序后的切片是否完全一致
	if len(s1Clean) != len(s2Clean) {
		return false
	}
	for i := range s1Clean {
		if s1Clean[i] != s2Clean[i] {
			return false
		}
	}
	return true
}

// EqualUrlIgnoreCase 忽略大小写的URL比较
func EqualUrlIgnoreCase(s1, s2 string) bool {
	u1, _ := url.Parse(s1)
	u2, _ := url.Parse(s2)
	return EqualStringIgnoreCase(u1.String(), u2.String())
}

// EqualUrlSliceUnordered 比较两个URL切片是否包含相同的唯一元素（忽略顺序和重复）
func EqualUrlSliceUnordered(sm1, sm2 []string) bool {
	um1 := make([]string, 0)
	for _, s1 := range sm1 {
		u1, _ := url.Parse(s1)
		um1 = append(um1, u1.String())
	}
	um2 := make([]string, 0)
	for _, s2 := range sm2 {
		u2, _ := url.Parse(s2)
		um2 = append(um2, u2.String())
	}
	return EqualStringSliceUnordered(um1, um2)
}
