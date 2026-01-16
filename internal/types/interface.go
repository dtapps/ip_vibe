package types

// 接口页面 请求结构
type InterfaceRequest struct {
	Proto string `json:"proto,omitempty" form:"proto,omitempty" default:"api"` // 当前选中的协议
}

// 接口页面 响应结构
type InterfaceResponse struct {
	CurrentProto string `json:"current_proto,omitempty"` // 当前选中的协议

	Token string `json:"token,omitempty"` // 认证 token

	HttpAddr string `json:"http_addr,omitempty"` // HTTP 地址
	GrpcAddr string `json:"grpc_addr,omitempty"` // gRPC 地址
}
