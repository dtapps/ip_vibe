package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return p
}

func TestReadLastLines(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5\n"
	p := writeTempFile(t, content)

	got, err := ReadLastLines(p, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "line4" || got[1] != "line5" {
		t.Errorf("ReadLastLines(,2)=%v want [line4 line5]", got)
	}

	// 请求行数超过实际行数 -> 返回全部
	got, err = ReadLastLines(p, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 5 || got[0] != "line1" || got[4] != "line5" {
		t.Errorf("ReadLastLines(,100)=%v want all 5 lines", got)
	}

	// 恰好等于全部行数
	got, err = ReadLastLines(p, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("ReadLastLines(,5) len=%d want 5", len(got))
	}
}

func TestReadLastLinesMissingFile(t *testing.T) {
	_, err := ReadLastLines(filepath.Join(t.TempDir(), "nope.txt"), 10)
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
