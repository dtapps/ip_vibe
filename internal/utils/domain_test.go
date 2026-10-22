package utils

import (
	"fmt"
	"testing"
)

func Test_ExtractDomainParts(t *testing.T) {
	tests := []string{"www.b.a.com", "b.a.com", "a.com", "sub.www.b.a.com"}

	for _, domainInput := range tests {
		domainFull, domainSub, domainMain, _ := ExtractDomainParts(domainInput)
		fmt.Printf("输入: %-15s → 完整: %-15s 子域: %-10s 主域: %s\n", domainInput, domainFull, domainSub, domainMain)
	}
}

func Test_TrimDots(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"www.b.a.com", "www.b.a.com"},
		{".b.a.com", "b.a.com"},
		{"a.com.", "a.com"},
		{"sub.www.b.a.com.", "sub.www.b.a.com"},
	}

	for _, test := range tests {
		result := TrimDots(test.input)
		if result != test.expected {
			t.Errorf("TrimDots(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}
