package core

import (
	"encoding/base64"
	"testing"
)

func TestGetBasicAuthToken(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		want     string
	}{
		{name: "正常用户名密码", username: "admin", password: "123456", want: base64.StdEncoding.EncodeToString([]byte("admin:123456"))},
		{name: "密码含冒号", username: "admin", password: "a:b", want: base64.StdEncoding.EncodeToString([]byte("admin:a:b"))},
		{name: "用户名为空返回空串", username: "", password: "123456", want: ""},
		{name: "密码为空返回空串", username: "admin", password: "", want: ""},
		{name: "两者都为空返回空串", username: "", password: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBasicAuthToken(tt.username, tt.password); got != tt.want {
				t.Fatalf("GetBasicAuthToken(%q, %q) = %q, want %q", tt.username, tt.password, got, tt.want)
			}
		})
	}
}

func TestGetBasicAuthHeader(t *testing.T) {
	if got := GetBasicAuthHeader("abc"); got != "Basic abc" {
		t.Fatalf("GetBasicAuthHeader(\"abc\") = %q, want \"Basic abc\"", got)
	}
	if got := GetBasicAuthHeader(""); got != "Basic " {
		t.Fatalf("GetBasicAuthHeader(\"\") = %q, want \"Basic \"", got)
	}
}

func TestParseBasicAuthHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{name: "标准 header", header: "Basic dXNlcjpwYXNz", want: "dXNlcjpwYXNz"},
		{name: "仅前缀无 token", header: "Basic ", want: ""},
		{name: "长度不足", header: "Basic", wantErr: true},
		{name: "空前缀", header: "", wantErr: true},
		{name: "前缀大小写敏感", header: "basic dXNlcjpwYXNz", wantErr: true},
		{name: "非 Basic 前缀", header: "Bearer token", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBasicAuthHeader(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseBasicAuthHeader(%q) 期望报错，实际返回 %q", tt.header, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseBasicAuthHeader(%q) 意外报错: %v", tt.header, err)
			}
			if got != tt.want {
				t.Fatalf("ParseBasicAuthHeader(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestParseBasicAuthToken(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		wantUsername string
		wantPassword string
		wantErr      bool
	}{
		{name: "无 Basic 前缀", token: base64.StdEncoding.EncodeToString([]byte("admin:123456")), wantUsername: "admin", wantPassword: "123456"},
		{name: "带 Basic 前缀", token: "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:123456")), wantUsername: "admin", wantPassword: "123456"},
		{name: "密码含冒号只切第一段", token: base64.StdEncoding.EncodeToString([]byte("admin:a:b")), wantUsername: "admin", wantPassword: "a:b"},
		{name: "空 token", token: "", wantErr: true},
		{name: "非法 base64", token: "!!!not-base64!!!", wantErr: true},
		{name: "缺少冒号分隔", token: base64.StdEncoding.EncodeToString([]byte("adminonly")), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, password, err := ParseBasicAuthToken(tt.token)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseBasicAuthToken(%q) 期望报错，实际返回 (%q, %q)", tt.token, username, password)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseBasicAuthToken(%q) 意外报错: %v", tt.token, err)
			}
			if username != tt.wantUsername || password != tt.wantPassword {
				t.Fatalf("ParseBasicAuthToken(%q) = (%q, %q), want (%q, %q)",
					tt.token, username, password, tt.wantUsername, tt.wantPassword)
			}
		})
	}
}

func TestVerifyBasicAuthToken(t *testing.T) {
	valid := GetBasicAuthToken("admin", "123456")

	tests := []struct {
		name     string
		token    string
		username string
		password string
		want     bool
	}{
		{name: "完全匹配", token: valid, username: "admin", password: "123456", want: true},
		{name: "带 Basic 前缀也匹配", token: GetBasicAuthHeader(valid), username: "admin", password: "123456", want: true},
		{name: "用户名不符", token: valid, username: "root", password: "123456", want: false},
		{name: "密码不符", token: valid, username: "admin", password: "654321", want: false},
		{name: "token 非法", token: "not-a-token", username: "admin", password: "123456", want: false},
		{name: "token 为空", token: "", username: "admin", password: "123456", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyBasicAuthToken(tt.token, tt.username, tt.password); got != tt.want {
				t.Fatalf("VerifyBasicAuthToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
