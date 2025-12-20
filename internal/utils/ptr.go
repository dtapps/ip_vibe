package utils

// Uint64ToPtr 将 uint64 转为 *uint64（返回新分配的内存，跨函数安全）
func Uint64ToPtr(v uint64) *uint64 {
	p := new(uint64)
	*p = v
	return p
}

// PtrToUint64 将 *uint64 转为 uint64，nil 安全（nil 返回 0）
func PtrToUint64(a *uint64) uint64 {
	if a == nil {
		return 0
	}
	return *a
}

// Int32ToPtr 将 int32 转为 *int32（返回新分配的内存，跨函数安全）
func Int32ToPtr(v int32) *int32 {
	p := new(int32)
	*p = v
	return p
}

// PtrToInt32 将 *int32 转为 int32，nil 安全（nil 返回 0）
func PtrToInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// BoolToPtr 将 bool 转为 *bool（返回新分配的内存，跨函数安全）
func BoolToPtr(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// PtrToBool 将 *bool 转为 bool，nil 安全（nil 返回 false）
func PtrToBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

// StringToPtr 将 string 转为 *string（返回新分配的内存，跨函数安全）
func StringToPtr(s string) *string {
	if s == "" {
		return nil // 可按业务需求改为 return &""，表示显式空值
	}
	p := new(string)
	*p = s
	return p
}

// PtrToString 将 *string 转为 string，nil 安全
func PtrToString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// StringsToPtr 将 string 转为 *string（返回新分配的内存，跨函数安全）
func StringsToPtr(s []string) *[]string {
	if len(s) == 0 {
		return nil // 可按业务需求改为 return &[]string{}，表示显式空值
	}
	p := new([]string)
	*p = s
	return p
}

// PtrToStrings 将 *[]string 转为 []string（nil 安全）
func PtrToStrings(p *[]string) []string {
	if p == nil {
		return nil
	}
	return *p
}
