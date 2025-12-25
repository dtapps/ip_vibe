package initialize

import (
	"log"
	"vibe_ip/internal/global"

	coreruleset "github.com/corazawaf/coraza-coreruleset"
	"github.com/corazawaf/coraza/v3"
)

/**

# 响应状态码
SecRule RESPONSE_STATUS "@ge 400" "id:200001,phase:4,log,pass,msg:'Detect response status >= 400'"

# 响应头
SecRule RESPONSE_HEADERS|RESPONSE_HEADERS_NAMES "@rx .*" "id:200002,phase:4,log,pass,msg:'Detect response headers'"

# 响应 body
SecRule RESPONSE_BODY "@rx sensitive|password|secret" "id:200003,phase:4,log,pass,msg:'Detect sensitive content in body'"

**/

// InitCoraza 初始化 Coraza 防火墙
func InitCoraza() {
	var err error
	global.GvaCoraza, err = coraza.NewWAF(
		coraza.NewWAFConfig().
			WithDirectives(`
# 开启 WAF
SecRuleEngine On

# 请求体检测（POST/JSON/Form等）
SecRequestBodyAccess On
		`).
			WithRootFS(coreruleset.FS),
	)
	if err != nil {
		log.Fatalf("[初始化 Coraza] 错误:%v", err)
	}
}

/**

# 测试请求头拦截
SecRule REQUEST_HEADERS:User-Agent "@contains BadBot" "id:900001,phase:1,deny,msg:'Blocked BadBot'"

# 测试请求 body 拦截
SecRule REQUEST_BODY "@contains test123" "id:900002,phase:2,deny,msg:'Blocked test123 content'"

# 测试 Cookie 拦截
SecRule REQUEST_COOKIES:session "@contains evil" "id:900003,phase:2,deny,msg:'Blocked evil cookie'"

**/
