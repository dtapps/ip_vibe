package initialize

import (
	"fmt"
	"os"
	"vibe_ip/internal/global"
)

func init() {
	// 从环境变量获取 [backend][reporter][forwarder] 是否开启调试模式
	global.GvaEnv.DEBUG = os.Getenv("DEBUG") == "true"

	// 从环境变量获取 [backend] gRPC 端口
	global.GvaEnv.GRPC_PORT = os.Getenv("GRPC_PORT")
	// if global.GvaEnv.GRPC_PORT == "" {
	// 	global.GvaEnv.GRPC_PORT = ":50051" // grpc默认端口
	// }
	if global.GvaEnv.GRPC_PORT != "" && global.GvaEnv.GRPC_PORT[0] != ':' {
		global.GvaEnv.GRPC_PORT = ":" + global.GvaEnv.GRPC_PORT // 判断GRPC_PORT前面是否有冒号
	}

	// 从环境变量获取 [backend] HTTP 端口
	global.GvaEnv.HTTP_PORT = os.Getenv("HTTP_PORT")
	if global.GvaEnv.HTTP_PORT == "" {
		global.GvaEnv.HTTP_PORT = ":8080" // gin默认端口
	}
	if global.GvaEnv.HTTP_PORT != "" && global.GvaEnv.HTTP_PORT[0] != ':' {
		global.GvaEnv.HTTP_PORT = ":" + global.GvaEnv.HTTP_PORT // 判断HTTP_PORT前面是否有冒号
	}

	// 从环境变量获取 [reporter][forwarder] HTTP URL
	global.GvaEnv.HTTP_URL = os.Getenv("HTTP_URL")

	// 从环境变量获取 [backend][reporter][forwarder] 基本认证用户名和密码
	global.GvaEnv.BASIC_AUTH_USERNAME = os.Getenv("BASIC_AUTH_USERNAME")
	global.GvaEnv.BASIC_AUTH_PASSWORD = os.Getenv("BASIC_AUTH_PASSWORD")

	// 打印用户名和密码
	if global.GvaEnv.DEBUG {
		fmt.Printf("BasicAuth 用户名: %s, 密码: %s\n", global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
	}

	// 从环境变量获取 [backend][reporter][forwarder] 设备代码
	global.GvaEnv.DEVICE_CODE = os.Getenv("DEVICE_CODE")

	// 从环境变量获取 [backend][reporter] 设备名称
	global.GvaEnv.DEVICE_NAME = os.Getenv("DEVICE_NAME")

	// 打印设备代码和设备名称
	if global.GvaEnv.DEBUG {
		fmt.Printf("设备代码: %s, 设备名称: %s\n", global.GvaEnv.DEVICE_CODE, global.GvaEnv.DEVICE_NAME)
	}
}
