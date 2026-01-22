package types

// IP组页面 请求结构
type IPGroupRequest struct {
	CommonPageRequest
	ID   int64  `json:"id,omitempty" form:"id,omitempty"`     // IP组序号
	Code string `json:"code,omitempty" form:"code,omitempty"` // IP组编码
	Name string `json:"name,omitempty" form:"name,omitempty"` // IP组名称
}

// IP组页面 响应结构
type IPGroupResponse struct {
	CommonPageResponse[GetIPGroupResponse, IPGroupRequest] // IP组列表
}
