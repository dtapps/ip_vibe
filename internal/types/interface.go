package types

// 接口页面 请求结构
type InterfaceRequest struct {
	Proto string `json:"proto,omitempty" form:"proto,omitempty" default:"api"` // 当前选中的协议
}

// 接口页面 响应结构
type InterfaceResponse struct {
	CurrentProto string `json:"current_proto,omitempty"` // 当前选中的协议
	CodeSnippet  string `json:"code_snippet,omitempty"`  // 代码片段

	Token string `json:"token,omitempty"` // 认证 token

	HttpAddr   string `json:"http_addr,omitempty"`   // HTTP 地址
	GrpcAddr   string `json:"grpc_addr,omitempty"`   // gRPC 地址
	ThriftAddr string `json:"thrift_addr,omitempty"` // Thrift 地址
	DubboAddr  string `json:"dubbo_addr,omitempty"`  // Dubbo 地址
	TarsAddr   string `json:"tars_addr,omitempty"`   // TARS 地址
}
