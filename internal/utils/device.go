package utils

import "net"

func GetDeviceID() string {
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) != 0 && iface.Flags&net.FlagUp != 0 {
			return iface.HardwareAddr.String() // 例如 "52:54:00:12:34:56"
		}
	}
	// 兜底
	return ""
}
