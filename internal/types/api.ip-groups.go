package types

// 获取IP组 响应结构
type GetIPGroupResponse struct {
	ID          int64   `json:"id" jsonschema:"title=IP组序号,description=IP组的唯一标识符"`                   // IP组序号
	Name        string  `json:"name" jsonschema:"title=IP组名称,description=IP组的名称"`                    // IP组名称
	Description string  `json:"description,omitempty" jsonschema:"title=IP组描述,description=IP组的描述信息"` // IP组描述
	DeviceCount int     `json:"device_count,omitempty" jsonschema:"title=设备数量,description=关联的设备数量"`  // 设备数
	Devices     []int64 `json:"devices,omitempty" jsonschema:"title=设备列表,description=关联的设备ID列表"`     // 设备列表

	Webhooks []int64 `json:"webhooks,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取IP组 请求结构
type GetIPGroupRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=IP组序号,description=IP组的唯一标识符,example=1"` // IP组序号
}

// 创建IP组 请求结构
type CreateIPGroupRequest struct {
	Name        string  `json:"name" binding:"required" jsonschema:"title=IP组名称,description=IP组的名称,example=办公网络"`         // IP组名称
	Description string  `json:"description,omitempty" jsonschema:"title=IP组描述,description=IP组的描述信息,example=公司办公网络IP组"`    // IP组描述
	Devices     []int64 `json:"devices" binding:"required" jsonschema:"title=设备列表,description=关联的设备ID列表,example=[1,2,3]"` // 设备列表

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,2,3]"` // 关联的 Webhook 列表
}

// 更新IP组 请求结构
type UpdateIPGroupRequest struct {
	GetIPGroupRequest
	Name        string  `json:"name" binding:"required" jsonschema:"title=IP组名称,description=IP组的名称,example=家庭网络"`         // IP组名称
	Description string  `json:"description,omitempty" jsonschema:"title=IP组描述,description=IP组的描述信息,example=家庭网络IP组"`      // IP组描述
	Devices     []int64 `json:"devices" binding:"required" jsonschema:"title=设备列表,description=关联的设备ID列表,example=[1,3,4]"` // 设备列表

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,3,4]"` // 关联的 Webhook 列表
}

// 删除IP组 请求结构
type DeleteIPGroupRequest struct {
	GetIPGroupRequest
}

// 获取IP组的设备列表 请求结构
type GetIPGroupDevicesRequest struct {
	GetIPGroupRequest
}

// 检查IP组健康状态 请求结构
type HealthcheckIPGroupRequest struct {
	GetIPGroupRequest
}
