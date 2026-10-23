package utils

import "testing"

func TestEqualStringIgnoreCase(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"完全相同", "abc", "abc", true},
		{"大小写不同", "Hello", "hELLo", true},
		{"前后空格", "  abc  ", "abc", true},
		{"两侧空格与大小写", "  Foo ", " foo  ", true},
		{"不同内容", "abc", "abd", false},
		{"空 vs 非空", "", "a", false},
		{"都为空白", "   ", " ", true},
	}
	for _, c := range cases {
		if got := EqualStringIgnoreCase(c.a, c.b); got != c.want {
			t.Errorf("%s: EqualStringIgnoreCase(%q,%q)=%v want %v", c.name, c.a, c.b, got, c.want)
		}
	}
}

func TestEqualStringSliceUnordered(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"同序", []string{"a", "b", "c"}, []string{"a", "b", "c"}, true},
		{"乱序", []string{"a", "b", "c"}, []string{"c", "a", "b"}, true},
		{"含重复去重", []string{"a", "a", "b"}, []string{"b", "a"}, true},
		{"长度不同", []string{"a", "b"}, []string{"a"}, false},
		{"元素不同", []string{"a", "b"}, []string{"a", "c"}, false},
		{"都为空", []string{}, []string{}, true},
		{"都 nil", nil, nil, true},
		{"空 vs 非空", []string{}, []string{"a"}, false},
	}
	for _, c := range cases {
		if got := EqualStringSliceUnordered(c.a, c.b); got != c.want {
			t.Errorf("%s: EqualStringSliceUnordered(%v,%v)=%v want %v", c.name, c.a, c.b, got, c.want)
		}
	}
}

func TestEqualUrlIgnoreCase(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"相同 URL", "https://a.com/x", "https://a.com/x", true},
		{"host 大小写", "https://A.com/x", "https://a.com/x", true},
		{"不同路径", "https://a.com/x", "https://a.com/y", false},
		{"不同 host", "https://a.com", "https://b.com", false},
	}
	for _, c := range cases {
		if got := EqualUrlIgnoreCase(c.a, c.b); got != c.want {
			t.Errorf("%s: EqualUrlIgnoreCase(%q,%q)=%v want %v", c.name, c.a, c.b, got, c.want)
		}
	}
}

func TestEqualUrlSliceUnordered(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"乱序相同", []string{"https://a.com", "https://b.com"}, []string{"https://b.com", "https://a.com"}, true},
		// 注意：切片版按 URL 字符串精确比较（区分大小写），与 EqualUrlIgnoreCase 语义不同
		{"大小写 host 视为不同", []string{"https://A.com"}, []string{"https://a.com"}, false},
		{"不同集合", []string{"https://a.com"}, []string{"https://b.com"}, false},
	}
	for _, c := range cases {
		if got := EqualUrlSliceUnordered(c.a, c.b); got != c.want {
			t.Errorf("%s: EqualUrlSliceUnordered(%v,%v)=%v want %v", c.name, c.a, c.b, got, c.want)
		}
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"A", "b", "c"}
	cases := []struct {
		target string
		want   bool
	}{
		{"A", true},
		{"a", false}, // 区分大小写
		{"b", true},
		{"z", false},
	}
	for _, c := range cases {
		if got := ContainsString(slice, c.target); got != c.want {
			t.Errorf("ContainsString(%v,%q)=%v want %v", slice, c.target, got, c.want)
		}
	}
}

func TestContainsStringIgnoreCase(t *testing.T) {
	slice := []string{"Hello", "World"}
	cases := []struct {
		target string
		want   bool
	}{
		{"hello", true},
		{"  WORLD ", true}, // 忽略大小写且 trim
		{"wor", false},     // 子串不算
		{"nope", false},
	}
	for _, c := range cases {
		if got := ContainsStringIgnoreCase(slice, c.target); got != c.want {
			t.Errorf("ContainsStringIgnoreCase(%v,%q)=%v want %v", slice, c.target, got, c.want)
		}
	}
}
