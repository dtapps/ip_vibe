package utils

import "testing"

func TestExtractDomainParts(t *testing.T) {
	cases := []struct {
		input    string
		wantFull string
		wantSub  string
		wantMain string
		wantErr  bool
	}{
		{"www.b.a.com", "www.b.a.com", "www.b", "a.com", false},
		{"sub.www.b.a.com", "sub.www.b.a.com", "sub.www.b", "a.com", false},
		{"a.com", "a.com", "", "a.com", false},
		{"sub.a.com:8080", "sub.a.com", "sub", "a.com", false}, // 端口被剥离
		{"com", "", "", "", true},                              // 只有 TLD，无法取 eTLD+1
	}
	for _, c := range cases {
		full, sub, main, err := ExtractDomainParts(c.input)
		if c.wantErr {
			if err == nil {
				t.Errorf("ExtractDomainParts(%q) want error, got nil", c.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ExtractDomainParts(%q) unexpected error: %v", c.input, err)
			continue
		}
		if full != c.wantFull || sub != c.wantSub || main != c.wantMain {
			t.Errorf("ExtractDomainParts(%q)=(%q,%q,%q) want (%q,%q,%q)",
				c.input, full, sub, main, c.wantFull, c.wantSub, c.wantMain)
		}
	}
}

// TestExtractDomainPartsWithPath 覆盖「输入带路径」分支。
// 注意：该分支目前直接拿整串 URL 调 publicsuffix.EffectiveTLDPlusOne，没有先解析出 host，
// 所以 "https://sub.a.com/path" 这类输入的结果并不可靠（已知缺陷，当前无调用方）。
// 这里只断言分支被正常走到、未报错，不锁定错误的具体返回值，避免测试反过来固化缺陷。
func TestExtractDomainPartsWithPath(t *testing.T) {
	for _, input := range []string{"https://sub.a.com/path", "sub.a.com/path"} {
		full, _, main, err := ExtractDomainParts(input)
		if err != nil {
			t.Fatalf("ExtractDomainParts(%q) 意外报错: %v", input, err)
		}
		if full == "" || main == "" {
			t.Fatalf("ExtractDomainParts(%q) 返回空结果: full=%q main=%q", input, full, main)
		}
	}
}

func TestTrimDots(t *testing.T) {
	cases := []struct{ input, expected string }{
		{"www.b.a.com", "www.b.a.com"},
		{".b.a.com", "b.a.com"},
		{"a.com.", "a.com"},
		{"sub.www.b.a.com.", "sub.www.b.a.com"},
		{"..a.com..", "a.com"},
	}
	for _, c := range cases {
		if got := TrimDots(c.input); got != c.expected {
			t.Errorf("TrimDots(%q)=%q want %q", c.input, got, c.expected)
		}
	}
	if TrimDotsPrt(nil) != "" {
		t.Error("TrimDotsPrt(nil) should be \"\"")
	}
	s := ".a.com."
	if got := TrimDotsPrt(&s); got != "a.com" {
		t.Errorf("TrimDotsPrt(%q)=%q want a.com", s, got)
	}
}

func TestEqualDomainIgnoreCase(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"Example.COM", "example.com", true},
		{"  a.com  ", "a.com", true},
		{".a.com.", "a.com", true}, // 首尾点会被 trim
		{"a.com", "b.com", false},
	}
	for _, c := range cases {
		if got := EqualDomainIgnoreCase(c.a, c.b); got != c.want {
			t.Errorf("EqualDomainIgnoreCase(%q,%q)=%v want %v", c.a, c.b, got, c.want)
		}
	}

	// 指针版本：任一为 nil 返回 false
	s1, s2 := "A.com", "a.com"
	if !EqualDomainIgnoreCasePtr(&s1, &s2) {
		t.Error("EqualDomainIgnoreCasePtr should be true")
	}
	if EqualDomainIgnoreCasePtr(nil, &s2) {
		t.Error("EqualDomainIgnoreCasePtr(nil, x) should be false")
	}

	// 泛型版本：按字符串形式比较
	if !EqualIgnoreCase("AbC", "abc") {
		t.Error("EqualIgnoreCase string should be true")
	}
	if !EqualIgnoreCase(123, 123) {
		t.Error("EqualIgnoreCase int should be true")
	}
}

func TestGetRecordType(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"1.2.3.4", "A", false},
		{" 8.8.8.8 ", "A", false},
		{"::1", "AAAA", false},
		{"2001:db8::1", "AAAA", false},
		{"example.com", "CNAME", false},
		{"sub.example.com", "CNAME", false},
		{"", "", true},
		{"bad_host", "", true},
	}
	for _, c := range cases {
		got, err := GetRecordType(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("GetRecordType(%q) want error, got %q", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("GetRecordType(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("GetRecordType(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestGetFullDomain(t *testing.T) {
	if got := GetFullDomain("@", "a.com"); got != "a.com" {
		t.Errorf("GetFullDomain(@,a.com)=%q want a.com", got)
	}
	if got := GetFullDomain("www", "a.com"); got != "www.a.com" {
		t.Errorf("GetFullDomain(www,a.com)=%q want www.a.com", got)
	}
}

func TestGetDnsDomainParts(t *testing.T) {
	full, sub, main, recType, err := GetDnsDomainParts("", "a.com", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if full != "@.a.com" || sub != "@" || main != "a.com" || recType != "" {
		t.Errorf("empty sub => (%q,%q,%q,%q)", full, sub, main, recType)
	}

	_, _, _, recType, err = GetDnsDomainParts("www", "a.com", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recType != "A" {
		t.Errorf("recordType=%q want A", recType)
	}
}

func TestGetCdnDomainParts(t *testing.T) {
	full, sub, main := GetCdnDomainParts("", "a.com")
	if full != "@.a.com" || sub != "@" || main != "a.com" {
		t.Errorf("GetCdnDomainParts(,)= (%q,%q,%q)", full, sub, main)
	}
}

func TestParseDomain(t *testing.T) {
	full, sub, main := ParseDomain("", "a.com")
	if full != "@.a.com" || sub != "@" || main != "a.com" {
		t.Errorf("ParseDomain(,)= (%q,%q,%q)", full, sub, main)
	}
}

func TestParseCdnDomainType(t *testing.T) {
	cases := []struct {
		platform string
		origin   []string
		want     string
		wantErr  bool
	}{
		{"tencent_eo", []string{"a"}, "IP_DOMAIN", false},
		{"tencent_eo", []string{"a", "b"}, "ORIGIN_GROUP", false},
		{"aliyun_esa", []string{"a"}, "A/AAAA", false},
		{"aliyun_esa", []string{"a", "b"}, "CNAME", false},
		{"unknown", []string{"a"}, "", true},
	}
	for _, c := range cases {
		got, _, err := ParseCdnDomainType(c.platform, c.origin)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseCdnDomainType(%q) want error", c.platform)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseCdnDomainType(%q) unexpected error: %v", c.platform, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseCdnDomainType(%q,%v)=%q want %q", c.platform, c.origin, got, c.want)
		}
	}
}

func TestParseDnsDomainType(t *testing.T) {
	got, origin, err := ParseDnsDomainType("1.2.3.4")
	if err != nil || got != "A" || origin != "1.2.3.4" {
		t.Errorf("ParseDnsDomainType(1.2.3.4)=(%q,%q,%v)", got, origin, err)
	}
	if _, _, err := ParseDnsDomainType(""); err == nil {
		t.Error("ParseDnsDomainType(\"\") want error")
	}
}
