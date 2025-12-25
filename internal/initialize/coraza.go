package initialize

import (
	"log"
	"vibe_ip/internal/global"

	coreruleset "github.com/corazawaf/coraza-coreruleset"
	"github.com/corazawaf/coraza/v3"
)

// InitCoraza 初始化 Coraza 防火墙
func InitCoraza() {
	var err error
	global.GvaCoraza, err = coraza.NewWAF(
		coraza.NewWAFConfig().
			WithDirectives(`
				SecRuleEngine On
				SecRequestBodyAccess On
				SecResponseBodyAccess Off
			`).
			WithRootFS(coreruleset.FS),
	)
	if err != nil {
		log.Fatalf("[初始化 Coraza] 错误:%v", err)
	}
}
