package utils

import (
	"testing"
	"time"
)

func TestInt64PtrRoundTrip(t *testing.T) {
	for _, v := range []int64{0, 1, -1, 1 << 40} {
		if got := PtrToInt64(Int64ToPtr(v)); got != v {
			t.Errorf("Int64 round-trip %d -> %d", v, got)
		}
	}
	if PtrToInt64(nil) != 0 {
		t.Error("PtrToInt64(nil) should be 0")
	}
}

func TestUint64PtrRoundTrip(t *testing.T) {
	for _, v := range []uint64{0, 1, 1 << 63} {
		if got := PtrToUint64(Uint64ToPtr(v)); got != v {
			t.Errorf("Uint64 round-trip %d -> %d", v, got)
		}
	}
	if PtrToUint64(nil) != 0 {
		t.Error("PtrToUint64(nil) should be 0")
	}
}

func TestIntPtrRoundTrip(t *testing.T) {
	for _, v := range []int{0, 1, -5, 1000} {
		if got := PtrToInt(IntToPtr(v)); got != v {
			t.Errorf("Int round-trip %d -> %d", v, got)
		}
	}
	if PtrToInt(nil) != 0 {
		t.Error("PtrToInt(nil) should be 0")
	}
}

func TestInt32PtrRoundTrip(t *testing.T) {
	for _, v := range []int32{0, 1, -1, 12345} {
		if got := PtrToInt32(Int32ToPtr(v)); got != v {
			t.Errorf("Int32 round-trip %d -> %d", v, got)
		}
	}
	if PtrToInt32(nil) != 0 {
		t.Error("PtrToInt32(nil) should be 0")
	}
}

func TestBoolPtrRoundTrip(t *testing.T) {
	for _, v := range []bool{true, false} {
		if got := PtrToBool(BoolToPtr(v)); got != v {
			t.Errorf("Bool round-trip %v -> %v", v, got)
		}
	}
	if PtrToBool(nil) != false {
		t.Error("PtrToBool(nil) should be false")
	}
}

func TestStringToPtr(t *testing.T) {
	if StringToPtr("") != nil {
		t.Error("StringToPtr(\"\") should be nil")
	}
	p := StringToPtr("hi")
	if p == nil || *p != "hi" {
		t.Errorf("StringToPtr(\"hi\") wrong: %v", p)
	}
	if PtrToString(nil) != "" {
		t.Error("PtrToString(nil) should be \"\"")
	}
	if got := PtrToString(StringToPtr("x")); got != "x" {
		t.Errorf("PtrToString round-trip: %q", got)
	}
}

func TestStringsToPtr(t *testing.T) {
	if StringsToPtr([]string{}) != nil {
		t.Error("StringsToPtr(empty) should be nil")
	}
	if PtrToStrings(nil) != nil {
		t.Error("PtrToStrings(nil) should be nil")
	}
	got := PtrToStrings(StringsToPtr([]string{"a", "b"}))
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("StringsToPtr round-trip wrong: %v", got)
	}
}

func TestStringToTimePtr(t *testing.T) {
	if StringToTimePtr("") != nil {
		t.Error("StringToTimePtr(empty) should be nil")
	}
	if StringToTimePtr("not-a-time") != nil {
		t.Error("StringToTimePtr(invalid) should be nil")
	}

	const layout = time.DateTime
	s := "2026-10-23 12:34:56"
	p := StringToTimePtr(s)
	if p == nil {
		t.Fatalf("StringToTimePtr(%q) returned nil", s)
	}
	if want, _ := time.Parse(layout, s); !p.Equal(want) {
		t.Errorf("StringToTimePtr(%q)=%v want %v", s, *p, want)
	}

	if !PtrToTime(nil).IsZero() {
		t.Error("PtrToTime(nil) should be zero time")
	}
}
