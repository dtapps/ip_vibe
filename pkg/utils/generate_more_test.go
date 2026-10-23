package utils

import (
	"regexp"
	"testing"
)

var hexRe = regexp.MustCompile(`^[0-9a-f]+$`)

func TestGenerateIDBasedHashDeterministic(t *testing.T) {
	// 同输入必须同输出（幂等）
	for _, id := range []int{1, 2, 100, 1000} {
		a := GenerateIDBasedHash(id)
		b := GenerateIDBasedHash(id)
		if a != b {
			t.Errorf("GenerateIDBasedHash(%d) not deterministic: %q != %q", id, a, b)
		}
		if len(a) != 64 {
			t.Errorf("GenerateIDBasedHash(%d) len=%d want 64", id, len(a))
		}
		if !hexRe.MatchString(a) {
			t.Errorf("GenerateIDBasedHash(%d)=%q not hex", id, a)
		}
	}

	// 不同输入不碰撞（小集合）
	seen := map[string]int{}
	for _, id := range []int{1, 2, 3, 4, 5, 100, 101, 1000, 9999} {
		h := GenerateIDBasedHash(id)
		if prev, ok := seen[h]; ok {
			t.Errorf("collision between id %d and %d", prev, id)
		}
		seen[h] = id
	}
}

func TestGenerateShortIDBasedHash(t *testing.T) {
	// 长度被正确截取
	for _, l := range []int{1, 8, 10, 32} {
		got := GenerateShortIDBasedHash(42, l)
		if len(got) != l {
			t.Errorf("length=%d got len=%d", l, len(got))
		}
	}

	// 前缀与全量哈希一致
	full := GenerateIDBasedHash(42)
	for _, l := range []int{1, 8, 10} {
		if got := GenerateShortIDBasedHash(42, l); got != full[:l] {
			t.Errorf("short(l=%d)=%q want prefix %q", l, got, full[:l])
		}
	}

	// 超过 64 被上限截断
	if got := GenerateShortIDBasedHash(42, 100); len(got) != 64 {
		t.Errorf("length>64 should cap to 64, got %d", len(got))
	}

	// 长度 0 -> 空串
	if got := GenerateShortIDBasedHash(42, 0); got != "" {
		t.Errorf("length=0 should be empty, got %q", got)
	}

	// 泛型 int64 分支
	if got := GenerateShortIDBasedHash[int64](42, 10); len(got) != 10 || !hexRe.MatchString(got) {
		t.Errorf("int64 variant wrong: %q", got)
	}
}
