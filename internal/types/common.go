package types

import (
	"encoding/json"
	"strconv"
)

// API响应结构
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// 公共 页面数据 结构
type BasePageData[T any] struct {
	ActiveMenu           string // 当前激活的菜单
	Title                string // 页面标题
	Data                 T      // 页面数据
	CurrentPath          string // 当前路径，如 /dashboard
	CurrentPathWithQuery string // 当前路径 + 查询字符串，如 /dashboard?id=1&name=test
}

// 设置当前激活的菜单
func (b *BasePageData[T]) SetActiveMenu(menu string) {
	b.ActiveMenu = menu
}

// 设置页面标题
func (b *BasePageData[T]) SetTitle(title string) {
	b.Title = title
}

// 设置页面数据
func (b *BasePageData[T]) SetData(data T) {
	b.Data = data
}

// 设置当前路径
func (b *BasePageData[T]) SetCurrentPath(path string) {
	b.CurrentPath = path
}

// 设置当前路径 + 查询字符串
func (b *BasePageData[T]) SetCurrentPathWithQuery(pathWithQuery string) {
	b.CurrentPathWithQuery = pathWithQuery
}

// 公共 页面分页 请求结构
type CommonPageRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"10"` // 每页数量
}

// 公共 页面分页 响应结构
type CommonPageResponse[ListT any, FilterT any] struct {
	Filter   FilterT `json:"filter"`    // 筛选参数
	List     []ListT `json:"list"`      // 通知列表
	Page     int     `json:"page"`      // 当前页码
	PageSize int     `json:"page_size"` // 每页数量
	Total    int64   `json:"total"`     // 总数量
}

// 公共 页面分页带单个项 响应结构
type CommonPageWithItemResponse[ItemT any, ListT any, FilterT any] struct {
	Item ItemT `json:"item"` // 单个项
	CommonPageResponse[ListT, FilterT]
}

// 公共 页面分页带单个项带统计 响应结构
type CommonPageWithItemAndStatResponse[statT any, ItemT any, ListT any, FilterT any] struct {
	Stat statT `json:"stat"` // 统计数据
	Item ItemT `json:"item"` // 单个项
	CommonPageResponse[ListT, FilterT]
}

type NullableBool struct {
	boolVal *bool
	Set     bool // 仅当传了有效 true/false 时为 true
}

func (nb *NullableBool) UnmarshalJSON(data []byte) (err error) {

	// 处理空输入
	if len(data) == 0 {
		nb.Set = false
		nb.boolVal = nil
		return nil
	}

	// 空值、"null" 视为未设置
	s := string(data)
	if s == `""` || s == "null" {
		nb.Set = false
		nb.boolVal = nil
		return nil
	}

	// 尝试解析为布尔值
	var b bool
	err = json.Unmarshal(data, &b)
	if err != nil {
		return err
	}

	nb.Set = true
	nb.boolVal = &b

	return nil
}

func (nb *NullableBool) UnmarshalText(text []byte) (err error) {

	// 空值、"null" 视为未设置
	s := string(text)
	if s == "" || s == "null" {
		nb.Set = false
		nb.boolVal = nil
		return nil
	}

	// 尝试解析为布尔值
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	nb.Set = true
	nb.boolVal = &b

	return nil
}

func (nb *NullableBool) UnmarshalParam(param string) (err error) {

	// 空值、"null" 视为未设置
	if param == "" || param == "null" {
		nb.Set = false
		nb.boolVal = nil
		return nil
	}

	// 尝试解析为布尔值
	b, err := strconv.ParseBool(param)
	if err != nil {
		return err
	}

	nb.Set = true
	nb.boolVal = &b

	return nil
}

// IsEmpty 表示：字段未传，或传了 null / ""（即用户不想设置该字段）
func (nb NullableBool) IsEmpty() bool {
	return !nb.Set
}

// IsTrue 表示：明确传了 true
func (nb NullableBool) IsTrue() bool {
	return nb.Set && nb.boolVal != nil && *nb.boolVal
}

// IsFalse 表示：明确传了 false
func (nb NullableBool) IsFalse() bool {
	return nb.Set && nb.boolVal != nil && !*nb.boolVal
}

// Get 返回内部的 *bool 值，如果未设置或为 null，则返回 nil
func (nb NullableBool) Get() *bool {
	if nb.IsEmpty() {
		return nil
	}
	return nb.boolVal
}

// Value 默认将未设置视为 false
func (nb NullableBool) Value() bool {
	return nb.ValueOr(false)
}

// ValueOr 返回布尔值，如果未设置则返回默认值
func (nb NullableBool) ValueOr(def bool) bool {
	if nb.IsEmpty() || nb.boolVal == nil {
		return def
	}
	return *nb.boolVal
}
