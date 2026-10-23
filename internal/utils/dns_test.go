package utils

import "testing"

func TestDNSValuesEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"都为空", "", "", true},
		{"一空一非空", "", "1.2.3.4", false},
		{"空白 trim 后相等", "  1.2.3.4  ", "1.2.3.4", true},
		{"相同 IPv4", "1.2.3.4", "1.2.3.4", true},
		{"不同 IPv4", "1.2.3.4", "1.2.3.5", false},
		{"IPv6 压缩形式不同但等价", "2001:db8::1", "2001:0db8:0:0:0:0:0:1", true},
		{"不同 IPv6", "2001:db8::1", "2001:db8::2", false},
		{"域名大小写", "Example.COM", "example.com", true},
		{"域名末尾点", "example.com.", "example.com", true},
		{"不同域名", "example.com", "example.net", false},
	}
	for _, c := range cases {
		if got := DNSValuesEqual(c.a, c.b); got != c.want {
			t.Errorf("%s: DNSValuesEqual(%q,%q)=%v want %v", c.name, c.a, c.b, got, c.want)
		}
	}
}
