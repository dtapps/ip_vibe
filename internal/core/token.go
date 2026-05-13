package core

import (
	"encoding/base64"
	"errors"
	"strings"
)

// GetBasicAuthToken 获取基本认证 token
func GetBasicAuthToken(username, password string) string {
	if username == "" || password == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

// GetBasicAuthHeader 获取基本认证 header
func GetBasicAuthHeader(token string) string {
	return "Basic " + token
}

// 解析基本认证 header 中的 token
func ParseBasicAuthHeader(header string) (string, error) {
	if len(header) < 6 || header[:6] != "Basic " {
		return "", errors.New("无效的基本认证头")
	}
	return header[6:], nil
}

// 解析基本认证 token 中的用户名和密码
func ParseBasicAuthToken(token string) (string, string, error) {
	// 判断 token 是否为空
	if token == "" {
		return "", "", errors.New("无效的基本认证令牌")
	}
	// 判断是否存在 Basic 前缀，如有则去掉
	token = strings.TrimPrefix(token, "Basic ")
	// 解码 token
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", errors.New("无效的基本认证令牌")
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("无效的基本认证令牌格式")
	}
	return parts[0], parts[1], nil
}

// 验证基本认证 token 是否匹配用户名和密码
func VerifyBasicAuthToken(token, username, password string) bool {
	parsedUsername, parsedPassword, err := ParseBasicAuthToken(token)
	if err != nil {
		return false
	}
	return parsedUsername == username && parsedPassword == password
}
