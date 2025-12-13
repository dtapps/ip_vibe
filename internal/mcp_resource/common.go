package mcp_resource

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"vibe_ip/internal/global"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// MIMETypeJSON JSON 内容类型
	MIMETypeJSON = "application/json"
	// MIMETypeText 文本内容类型
	MIMETypeText = "text/plain"
)

// errorResponse 错误响应（支持格式化）
func errorResponse(format string, args ...any) (*mcp.ReadResourceResult, error) {
	return nil, fmt.Errorf(format, args...)
}

// successResponse 成功响应（自动处理多种数据类型）
func successResponse(uri string, data any) (*mcp.ReadResourceResult, error) {
	var content string

	switch v := data.(type) {
	case string:
		content = v
	case []byte:
		content = string(v)
	case json.RawMessage:
		content = string(v)
	default:
		// 尝试JSON序列化
		if jsonData, err := json.Marshal(data); err == nil {
			content = string(jsonData)
		} else {
			content = fmt.Sprintf("%v", data)
		}
		// if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
		// 	content = string(jsonData)
		// } else {
		// 	content = fmt.Sprintf("%v", data)
		// }
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      uri,
				MIMEType: detectContentType(content),
				Text:     content,
			},
		},
	}, nil
}

// detectContentType 自动检测内容类型
func detectContentType(content string) string {
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return MIMETypeJSON
	}
	return MIMETypeText
}

// 扩展方法：自动把 tpl 注入 handler，不需要写两次 tpl
// -----------------------------------------------------------------------
// go-sdk 的 AddResourceTemplate 只接受 handler(ctx, req)
// 但我们的业务 handler 需要知道 tpl（URITemplate）才能解析参数
//
// 因此这里封装一层 AddResourceTemplateEx：
//   - 外部只写一次 tpl
//   - 内部自动把 tpl 传给业务 handler
//   - 保证业务 handler 直接用 tpl，无需重复写模板
//
// -----------------------------------------------------------------------
func AddResourceTemplateEx(
	s *mcp.Server,
	tpl *mcp.ResourceTemplate,
	handler func(context.Context, *mcp.ResourceTemplate, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error),
) {

	// 包装成 go-sdk 期望的 handler 形式 (ctx, req) -> (result)
	wrapped := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return handler(ctx, tpl, req) // 自动注入 tpl
	}

	// 调用原生 AddResourceTemplate（真正注册资源模板）
	s.AddResourceTemplate(tpl, wrapped)
}

var paramPattern = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// compileURITemplate 将 URI 模板转换为可用于参数提取的正则表达式
// -----------------------------------------------------------------------
// 输入模板:   cdn://records/{id}
// 输出正则:  ^cdn://records/(?P<id>[^/]+)$
//
// uritemplate 的原始实现并不会为 {id} 生成命名捕获组，因此我们必须自己扩展。
// (?P<id>) 为 Go 自带正则支持的命名捕获形式，可从 SubexpNames 取回参数名。
// -----------------------------------------------------------------------
func compileURITemplate(templ string) (*regexp.Regexp, []string, error) {
	var params []string

	// 将 {id} 替换成命名捕获组 (?P<id>xxx)
	regexStr := paramPattern.ReplaceAllStringFunc(templ, func(m string) string {
		name := paramPattern.FindStringSubmatch(m)[1] // 提取参数名 id
		params = append(params, name)
		return fmt.Sprintf("(?P<%s>[^/]+)", name) // 支持命名捕获
	})

	// 加上 ^ 和 $ 保证完整匹配
	regexStr = "^" + regexStr + "$"

	// 编译为正则
	re, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, nil, err
	}

	return re, params, nil
}

// ExtractURIParams 从实际 URI 中提取参数
// -----------------------------------------------------------------------
// 示例：
// tpl.URITemplate = cdn://records/{id}
// uri            = cdn://records/123
//
// 返回: map[string]string{"id": "123"}
//
// 逻辑流程：
//  1. 将模板编译成正则（含命名捕获）
//  2. 用正则匹配 URI
//  3. 遍历 SubexpNames 获取参数名
//  4. 根据匹配结果组装成 map 返回
//
// uritemplate 原生实现无法获得参数名，这里通过正则增强解决。
// -----------------------------------------------------------------------
func ExtractURIParams(tpl *mcp.ResourceTemplate, uri string) (map[string]string, error) {

	// 调试输出
	if global.GvaEnv.DEBUG {
		slog.Debug("ExtractURIParams",
			slog.String("tpl.URITemplate", tpl.URITemplate),
			slog.String("uri", uri),
		)
	}

	// 编译模板为可解析参数的正则表达式
	re, _, err := compileURITemplate(tpl.URITemplate)
	if err != nil {
		return nil, err
	}

	// 通过正则从 URI 中提取匹配项
	match := re.FindStringSubmatch(uri)
	if match == nil {
		return nil, fmt.Errorf("URI %s 不匹配模板 %s", uri, tpl.URITemplate)
	}

	// 调试输出匹配结果
	if global.GvaEnv.DEBUG {
		slog.Debug("ExtractURIParams",
			slog.Any("match", match),
		)
	}

	// 从命名捕获组中提取参数
	out := map[string]string{}
	for i, name := range re.SubexpNames() {
		if i > 0 && name != "" { // 第0位是整个匹配，不用管
			out[name] = match[i]
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("没有从URI %s 中提取到参数", uri)
	}

	return out, nil
}
