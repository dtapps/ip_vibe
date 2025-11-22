package types

// IP组 响应结构
type GetIPGroupsResponse struct {
	ID          int64   `json:"id"`                     // IP组序号
	Name        string  `json:"name"`                   // IP组名称
	Description string  `json:"description"`            // IP组描述
	DeviceCount int     `json:"device_count,omitempty"` // 设备数
	Devices     []int64 `json:"devices,omitempty"`      // 设备列表
}

// 创建IP组 请求结构
type CreateIPGroupRequest struct {
	ID          int64   `json:"id,omitempty"`          // IP组序号
	Name        string  `json:"name"`                  // IP组名称
	Description string  `json:"description,omitempty"` // IP组描述
	Devices     []int64 `json:"devices,omitempty"`     // 设备列表
}
