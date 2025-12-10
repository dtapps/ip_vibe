package types

import (
	"time"
)

// 获取全部域名监控 请求结构
type GetDomainMonitorsRequest struct {
	IsActive *bool `json:"is_active" binding:"required" jsonschema:"title=是否活跃,description=指定是否只返回活跃的域名监控记录,example=true"` // 是否活跃
}

// 获取域名监控 响应结构
type GetDomainMonitorResponse struct {
	ID            int64     `json:"id" jsonschema:"title=域名监控记录序号,description=域名监控记录的唯一标识符"`                                  // 域名监控记录序号
	Name          string    `json:"name" jsonschema:"title=名称,description=域名监控的名称"`                                           // 名称
	Domain        string    `json:"domain" jsonschema:"title=域名,description=要监控的域名"`                                          // 域名
	CheckInterval int       `json:"check_interval" jsonschema:"title=检查间隔,description=检查间隔（秒）"`                               // 检查间隔（秒）
	CheckProtocol string    `json:"check_protocol" jsonschema:"title=检查协议,description=检查类型（HTTP/HTTPS）,enum=HTTP,enum=HTTPS"` // 检查类型（HTTP/HTTPS）
	IsActive      bool      `json:"is_active" jsonschema:"title=是否活跃,description=是否活跃,default=true"`                          // 是否活跃
	CurrentStatus string    `json:"current_status" jsonschema:"title=当前状态,description=当前运行状态"`                                // 当前运行状态
	LastSuccessAt time.Time `json:"last_success_at" jsonschema:"title=最后成功时间,description=最后一次检查成功时间"`                         // 最后一次检查成功时间
	LastFailureAt time.Time `json:"last_failure_at" jsonschema:"title=最后失败时间,description=最后一次检查失败时间"`                         // 最后一次检查失败时间

	Webhooks []int64 `json:"webhooks,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取域名监控 请求结构
type GetDomainMonitorRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=域名监控记录序号,description=域名监控记录的唯一标识符,example=1"` // 域名监控记录序号
}

// 获取域名监控 请求结构
type CreateDomainMonitorRequest struct {
	Name          string `json:"name" jsonschema:"title=名称,description=域名监控的名称,example=网站监控"`                                            // 名称
	Domain        string `json:"domain" jsonschema:"title=域名,description=要监控的域名,example=example.com"`                                    // 域名
	CheckInterval int    `json:"check_interval" jsonschema:"title=检查间隔,description=检查间隔（秒）,example=300"`                                 // 检查间隔（秒）
	CheckProtocol string `json:"check_protocol" jsonschema:"title=检查协议,description=检查类型（HTTP/HTTPS）,enum=HTTP,enum=HTTPS,example=HTTPS"` // 检查类型（HTTP/HTTPS）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,2,3]"` // 关联的 Webhook 列表
}

// 更新域名监控 请求结构
type UpdateDomainMonitorRequest struct {
	GetDomainMonitorRequest
	Name          string `json:"name" jsonschema:"title=名称,description=域名监控的名称,example=更新后的网站监控"`        // 名称
	CheckInterval int    `json:"check_interval" jsonschema:"title=检查间隔,description=检查间隔（秒）,example=600"` // 检查间隔（秒）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,3,4]"` // 关联的 Webhook 列表
}

// 更新域名监控启用状态 请求结构
type UpdateDomainMonitorActiveRequest struct {
	GetDomainMonitorRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"title=是否活跃,description=是否活跃,default=true,example=true"` // 是否活跃
}

// 删除域名监控 请求结构
type DeleteDomainMonitorRequest struct {
	GetDomainMonitorRequest
}

// 检查域名监控 请求结构
type CheckDomainMonitorRequest struct {
	GetDomainMonitorRequest
}
