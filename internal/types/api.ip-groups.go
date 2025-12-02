package types

// 获取全部IP组 响应结构
type GetIPGroupResponse struct {
	ID          int64   `json:"id"`                     // IP组序号
	Name        string  `json:"name"`                   // IP组名称
	Description string  `json:"description,omitempty"`  // IP组描述
	DeviceCount int     `json:"device_count,omitempty"` // 设备数
	Devices     []int64 `json:"devices,omitempty"`      // 设备列表
}

// 获取IP组 请求结构
type GetIPGroupRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // IP组序号
}

// 创建IP组 请求结构
type CreateIPGroupRequest struct {
	Name        string  `json:"name" binding:"required"`    // IP组名称
	Description string  `json:"description,omitempty"`      // IP组描述
	Devices     []int64 `json:"devices" binding:"required"` // 设备列表
}

// 更新IP组 请求结构
type UpdateIPGroupRequest struct {
	GetIPGroupRequest
	Name        string  `json:"name" binding:"required"`    // IP组名称
	Description string  `json:"description,omitempty"`      // IP组描述
	Devices     []int64 `json:"devices" binding:"required"` // 设备列表
}

// 删除IP组 请求结构
type DeleteIPGroupRequest struct {
	GetIPGroupRequest
}

// 获取IP组的设备列表 请求结构
type GetIPGroupDevicesRequest struct {
	GetIPGroupRequest
}
