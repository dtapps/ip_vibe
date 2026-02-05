package schemadefault

import (
	"go.dtapp.net/library/utils/gomd5"
	"go.dtapp.net/library/utils/gorandom"
)

// DefaultCode 生成随机验证码
func DefaultCode() string {
	data := gorandom.Alphanumeric(10)
	return gomd5.Get16MD5Encode(data)
}
