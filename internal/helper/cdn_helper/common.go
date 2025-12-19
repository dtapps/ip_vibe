package cdn_helper

import (
	"encoding/json"
	"fmt"
)

func indentJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func printJSON(name string, request any, response any) {
	fmt.Printf("%s 请求参数:\n%s\n", name, indentJSON(request))
	fmt.Printf("%s 响应参数:\n%s\n", name, indentJSON(response))
}
