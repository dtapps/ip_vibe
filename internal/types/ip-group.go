package types

// IP组页面 请求结构
type IPGroupRequest struct {
	CommonPageRequest
	ID       int64  `json:"id,omitempty" form:"id,omitempty"`               // IP组序号
	RecordID string `json:"record_id,omitempty" form:"record_id,omitempty"` // 平台侧记录ID
	Name     string `json:"name,omitempty" form:"name,omitempty"`           // IP组名称
}

// IP组页面 响应结构
type IPGroupResponse struct {
	CommonPageResponse[GetIPGroupResponse, IPGroupRequest] // IP组列表
}
