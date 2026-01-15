package global

// Env 环境变量
type Env struct {
	DEBUG bool // [backend][reporter][forwarder] 是否开启调试模式

	HTTP_ADDR   string // [backend] HTTP 地址
	GRPC_ADDR   string // [backend] gRPC 地址
	THRIFT_ADDR string // [backend] Thrift 地址

	HTTP_URL   string // [reporter][forwarder] HTTP URL
	GRPC_URL   string // [reporter][forwarder] gRPC URL
	THRIFT_URL string // [reporter][forwarder] Thrift URL

	BASIC_AUTH_USERNAME string // [backend][reporter][forwarder] 基本认证用户名
	BASIC_AUTH_PASSWORD string // [backend][reporter][forwarder] 基本认证密码
	DEVICE_CODE         string // [backend][reporter][forwarder] 设备代码
	DEVICE_NAME         string // [backend][reporter] 设备名称
}
