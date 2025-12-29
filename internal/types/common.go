package types

// 公共 页面数据 结构
type BasePageData[T any] struct {
	Debug      bool   // 是否开启调试模式
	ActiveMenu string // 当前激活的菜单
	Title      string // 页面标题
	Data       T      // 页面数据
}

// 设置是否开启调试模式
func (b *BasePageData[T]) SetDebug(debug bool) {
	b.Debug = debug
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

// 公共 页面分页 请求结构
type CommonPageRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"10"` // 每页数量
}

// 公共 页面分页 响应结构
type CommonPageResponse[T any] struct {
	List     []T   `json:"list"`      // 通知列表
	Page     int   `json:"page"`      // 当前页码
	PageSize int   `json:"page_size"` // 每页数量
	Total    int64 `json:"total"`     // 总数量
}
