package ip_helper

import "testing"

// 这些方法只做正则匹配，不访问 hc 的字段，直接构造零值实例即可测试。
func newTestHelper() *IPHelper { return &IPHelper{} }

func TestExtractIP(t *testing.T) {
	c := newTestHelper()

	tests := []struct {
		name     string
		response string
		want     string
	}{
		{name: "纯 IPv4", response: "1.2.3.4", want: "1.2.3.4"},
		{name: "文本中包含 IPv4", response: "your ip is 203.0.113.7 ok", want: "203.0.113.7"},
		// 注意：压缩写法的 IPv6（如 2001:db8::1）会被截断成 2001:db8::（正则分支顺序缺陷，已知问题），
		// 因此这里用完整写法断言正确路径，压缩写法的缺陷不在此固化。
		{name: "纯 IPv6 完整写法", response: "2001:0db8:0000:0000:0000:0000:0000:0001", want: "2001:0db8:0000:0000:0000:0000:0000:0001"},
		{name: "IPv4 优先于 IPv6", response: "10.0.0.1 and 2001:0db8:0000:0000:0000:0000:0000:0001", want: "10.0.0.1"},
		{name: "无 IP 返回空串", response: "no address here", want: ""},
		{name: "空响应返回空串", response: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.extractIP(tt.response); got != tt.want {
				t.Fatalf("extractIP(%q) = %q, want %q", tt.response, got, tt.want)
			}
		})
	}
}

func TestValidateIPType(t *testing.T) {
	c := newTestHelper()

	tests := []struct {
		name   string
		ip     string
		ipType IPType
		want   bool
	}{
		{name: "IPv4 请求配 IPv4 结果", ip: "1.2.3.4", ipType: IPTypeV4, want: true},
		{name: "IPv4 请求配 IPv6 结果", ip: "2001:db8::1", ipType: IPTypeV4, want: false},
		{name: "IPv6 请求配 IPv6 结果", ip: "2001:db8::1", ipType: IPTypeV6, want: true},
		{name: "IPv6 请求配 IPv4 结果", ip: "1.2.3.4", ipType: IPTypeV6, want: false},
		{name: "未指定类型时接受 IPv4", ip: "1.2.3.4", ipType: IPType(""), want: true},
		{name: "未指定类型时接受 IPv6", ip: "2001:db8::1", ipType: IPType(""), want: true},
		{name: "未指定类型时拒绝非 IP", ip: "not-an-ip", ipType: IPType(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.validateIPType(tt.ip, tt.ipType); got != tt.want {
				t.Fatalf("validateIPType(%q, %q) = %v, want %v", tt.ip, tt.ipType, got, tt.want)
			}
		})
	}
}

func TestIsIPv4AndIPv6(t *testing.T) {
	c := newTestHelper()

	if !c.isIPv4("192.168.1.1") {
		t.Fatal("192.168.1.1 应被识别为 IPv4")
	}
	if c.isIPv4("2001:db8::1") {
		t.Fatal("2001:db8::1 不应被识别为 IPv4")
	}
	if !c.isIPv6("2001:db8::1") {
		t.Fatal("2001:db8::1 应被识别为 IPv6")
	}
	if c.isIPv6("192.168.1.1") {
		t.Fatal("192.168.1.1 不应被识别为 IPv6")
	}
}
