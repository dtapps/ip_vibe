package sys_env

import (
	"fmt"
	"os"
)

// Env 环境变量
type Env struct {
	// [backend][reporter][forwarder] 是否开启调试模式
	DEBUG bool `json:"debug" env:"DEBUG"`

	// [backend] 服务地址
	SERVER_ADDR string `json:"server_addr" env:"SERVER_ADDR"`

	// [reporter][forwarder] 后端服务地址
	BACKEND_ENDPOINT string `json:"backend_endpoint" env:"BACKEND_ENDPOINT"`

	// [backend][reporter][forwarder] 基本认证用户名
	BASIC_AUTH_USERNAME string `json:"basic_auth_username" env:"BASIC_AUTH_USERNAME"`
	// [backend][reporter][forwarder] 基本认证密码
	BASIC_AUTH_PASSWORD string `json:"basic_auth_password" env:"BASIC_AUTH_PASSWORD"`
	// [backend][reporter][forwarder] 设备代码
	DEVICE_CODE string `json:"device_code" env:"DEVICE_CODE"`
	// [backend][reporter] 设备名称
	DEVICE_NAME string `json:"device_name" env:"DEVICE_NAME"`
}

// GvaEnv 全局环境变量
var GvaEnv *Env

// 初始化环境变量
func init() {
	// 内部辅助函数：查不到就返回默认值
	getVal := func(key string, defaultVal string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		return defaultVal
	}

	// 设置变量
	GvaEnv = new(Env)
	GvaEnv.DEBUG = getVal("DEBUG", "false") == "true"
	GvaEnv.SERVER_ADDR = getVal("SERVER_ADDR", "0.0.0.0:8080")
	GvaEnv.BACKEND_ENDPOINT = getVal("BACKEND_ENDPOINT", "127.0.0.1:8080")
	GvaEnv.BASIC_AUTH_USERNAME = getVal("BASIC_AUTH_USERNAME", "")
	GvaEnv.BASIC_AUTH_PASSWORD = getVal("BASIC_AUTH_PASSWORD", "")
	GvaEnv.DEVICE_CODE = getVal("DEVICE_CODE", "")
	GvaEnv.DEVICE_NAME = getVal("DEVICE_NAME", "")

	fmt.Println("sys_env:", GvaEnv)
}
