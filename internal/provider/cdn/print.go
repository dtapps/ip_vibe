package cdn

import (
	"encoding/json"
	"vibe_ip/internal/helper/sys_env"
)

func indentJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func printJSON(name string, request any, response any) {
	if !sys_env.GvaEnv.DEBUG {
		return
	}
	// if request != nil {
	// 	fmt.Printf("%s 请求参数:\n%s\n", name, indentJSON(request))
	// }
	// if response != nil {
	// 	fmt.Printf("%s 响应参数:\n%s\n", name, indentJSON(response))
	// }
}
