package types

import "encoding/json"

// 获取配置 响应结构
type GetConfigResponse struct {
	ID          int64           `json:"id,omitempty" jsonschema:"title=配置序号,description=配置的唯一标识符"`                                                              // 配置序号
	Capability  string          `json:"capability" jsonschema:"title=平台服务能力,description=平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）,enum=dns,enum=cdn,enum=all"` // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string          `json:"name" jsonschema:"title=配置名称,description=配置的名称"`                                                                         // 配置名称
	Description string          `json:"description,omitempty" jsonschema:"title=配置描述,description=配置的描述信息"`                                                      // 配置描述
	Platform    string          `json:"platform" jsonschema:"title=配置平台,description=配置所属的平台"`                                                                   // 配置平台
	ConfigJSON  json.RawMessage `json:"config_json,omitempty" jsonschema:"title=配置信息,description=配置平台密钥等敏感信息（JSON格式）"`                                          // 配置平台密钥等敏感信息（JSON）
}

// 获取配置 请求结构
type GetConfigRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=配置序号,description=配置的唯一标识符,example=1"` // 配置序号
}

// 创建配置 请求结构
type CreateConfigRequest struct {
	Capability  string `json:"capability" binding:"required" jsonschema:"title=平台服务能力,description=平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）,enum=dns,enum=cdn,enum=all,example=dns"`                                       // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string `json:"name" binding:"required" jsonschema:"title=配置名称,description=配置的名称,example=阿里云DNS配置"`                                                                                                          // 配置名称
	Description string `json:"description,omitempty" jsonschema:"title=配置描述,description=配置的描述信息,example=用于管理阿里云DNS记录"`                                                                                                      // 配置描述
	Platform    string `json:"platform" binding:"required" jsonschema:"title=配置平台,description=配置所属的平台,example=aliyun"`                                                                                                      // 配置平台
	ConfigJSON  string `json:"config_json" binding:"required" jsonschema:"title=配置信息,description=配置平台密钥等敏感信息（JSON格式）,example='{\"accessKeyId\": \"your-access-key-id\", \"accessKeySecret\": \"your-access-key-secret\"}'"` // 配置平台密钥等敏感信息（JSON）
}

// 更新配置 请求结构
type UpdateConfigRequest struct {
	GetConfigRequest
	Capability  string `json:"capability" binding:"required" jsonschema:"title=平台服务能力,description=平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）,enum=dns,enum=cdn,enum=all,example=cdn"`                   // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string `json:"name" binding:"required" jsonschema:"title=配置名称,description=配置的名称,example=腾讯云CDN配置"`                                                                                      // 配置名称
	Description string `json:"description,omitempty" jsonschema:"title=配置描述,description=配置的描述信息,example=用于管理腾讯云CDN记录"`                                                                                  // 配置描述
	ConfigJSON  string `json:"config_json" binding:"required" jsonschema:"title=配置信息,description=配置平台密钥等敏感信息（JSON格式）,example='{\"secretId\": \"your-secret-id\", \"secretKey\": \"your-secret-key\"}'"` // 配置平台密钥等敏感信息（JSON）
}

// 删除配置 请求结构
type DeleteConfigRequest struct {
	GetConfigRequest
}
