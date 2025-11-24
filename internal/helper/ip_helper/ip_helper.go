package ip_helper

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

type IPType string

const (
	IPTypeV4 IPType = "v4"
	IPTypeV6 IPType = "v6"
)

// IPHelper 用于检查公网 IP 地址
type IPHelper struct {
	// 用于 IPv4 检查的 URL 列表
	iPv4URLs []string
	// 用于 IPv6 检查的 URL 列表
	iPv6URLs []string
	// Resty HTTP 客户端
	client *resty.Client
	// 请求超时时间
	timeout time.Duration
}

// NewIPHelper 创建一个新的 IPHelper 实例
func NewIPHelper() *IPHelper {
	// 创建一个新的 Resty 客户端
	client := resty.New()

	// 设置默认超时
	client.SetTimeout(5 * time.Second)

	// 启用请求跟踪以获取延迟信息
	client.EnableTrace()

	return &IPHelper{
		iPv4URLs: []string{
			"https://eo-ip.dtapp.net",
			"https://eo-ip.dtapp.top",
			"https://eo-ip.liguangchun.cn",
			"https://eo-ip.liguangchun.com",
			"https://cf-ip.dtapp.net",
			"https://cf-ip.dtapp.top",
			"https://cf-ip.liguangchun.cn",
			"https://cf-ip.liguangchun.com",
			"https://ddns.oray.com/checkip",
			"http://v4.66666.host:66/ip",
			"https://myip.ipip.net",
			"http://v4.666666.host:66/ip",
			"https://4.ipw.cn",
			"https://ip.3322.net",
		},
		iPv6URLs: []string{
			"https://eo-ip.dtapp.net",
			"https://eo-ip.dtapp.top",
			"https://eo-ip.liguangchun.cn",
			"https://eo-ip.liguangchun.com",
			"https://cf-ip.dtapp.net",
			"https://cf-ip.dtapp.top",
			"https://cf-ip.liguangchun.cn",
			"https://cf-ip.liguangchun.com",
			"http://v6.66666.host:66/ip",
			"http://myip6.ipip.net",
			"https://6.ipw.cn",
			"http://v6.666666.host:66/ip",
		},
		client:  client,
		timeout: 5 * time.Second,
	}
}

// IPResult 表示从 IP 查询服务获取的结果
type IPResult struct {
	IP     string // IP 地址
	Source string // 查询服务的名称或 URL
	Error  error  // 错误信息（如果有）
}

// GetIP 并发获取 IP 地址
func (c *IPHelper) GetIP(ctx context.Context, ipType IPType) (*IPResult, error) {
	var urls []string

	// 根据 ipType 选择要使用的 URL 列表
	switch ipType {
	case IPTypeV4:
		urls = c.iPv4URLs
	case IPTypeV6:
		urls = c.iPv6URLs
	default:
		// 默认同时检查 IPv4 和 IPv6
		urls = append(urls, c.iPv4URLs...)
		urls = append(urls, c.iPv6URLs...)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("没有可用的 %s IP 地址查询服务", ipType)
	}

	// 创建一个带超时的上下文
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 创建一个通道来接收结果
	resultChan := make(chan *IPResult, len(urls))

	// 使用errgroup来管理goroutines
	g, gCtx := errgroup.WithContext(ctxWithTimeout)

	// 启动goroutine来并发请求所有URL
	for _, url := range urls {
		url := url // 创建局部变量避免闭包问题
		g.Go(func() error {
			// 检查上下文是否已取消
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				c.fetchIP(gCtx, url, ipType, resultChan)
				return nil
			}
		})
	}

	// 启动一个goroutine来等待所有请求完成并关闭通道
	go func() {
		_ = g.Wait() // 忽略错误，我们会在下面处理结果
		close(resultChan)
	}()

	// 收集结果，同时监视超时
	var firstError error
	for i := 0; i < len(urls); i++ {
		select {
		case result := <-resultChan:
			if result != nil && result.Error == nil && result.IP != "" {
				return result, nil // 成功!
			}
			if result != nil && firstError == nil && result.Error != nil {
				firstError = result.Error // 保存第一个错误
			}
		case <-ctxWithTimeout.Done():
			return nil, ctxWithTimeout.Err() // 超时!
		}
	}

	// 如果到这里，所有源都失败了
	if firstError != nil {
		return nil, fmt.Errorf("所有 %s IP 地址查询服务都失败，第一个错误: %w", ipType, firstError)
	}
	return nil, fmt.Errorf("没有可用的 %s IP 地址查询服务", ipType)
}

// GetAllIP 并发获取所有 IP 地址
func (c *IPHelper) GetAllIP(ctx context.Context, ipType IPType) ([]*IPResult, error) {
	var urls []string

	// 根据 ipType 选择要使用的 URL 列表
	switch ipType {
	case IPTypeV4:
		urls = c.iPv4URLs
	case IPTypeV6:
		urls = c.iPv6URLs
	default:
		// 默认同时检查 IPv4 和 IPv6
		urls = append(urls, c.iPv4URLs...)
		urls = append(urls, c.iPv6URLs...)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("没有可用的 %s IP 地址查询服务", ipType)
	}

	// 创建一个带超时的上下文
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 创建一个通道来接收结果
	resultChan := make(chan *IPResult, len(urls))

	// 使用errgroup来管理goroutines
	g, gCtx := errgroup.WithContext(ctxWithTimeout)

	// 启动goroutine来并发请求所有URL
	for _, url := range urls {
		url := url // 创建局部变量避免闭包问题
		g.Go(func() error {
			// 检查上下文是否已取消
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				c.fetchIP(gCtx, url, ipType, resultChan)
				return nil
			}
		})
	}

	// 启动一个goroutine来等待所有请求完成并关闭通道
	go func() {
		_ = g.Wait() // 忽略错误，我们会在下面处理结果
		close(resultChan)
	}()

	// 收集所有结果
	var results []*IPResult
	for i := 0; i < len(urls); i++ {
		select {
		case result := <-resultChan:
			if result != nil {
				results = append(results, result)
			}
		case <-ctxWithTimeout.Done():
			// 如果超时，我们应该仍然收集到目前为止收到的所有结果
			// 并返回超时错误
			// 排空通道中可能已经到达的任何剩余结果
			for len(resultChan) > 0 {
				result := <-resultChan
				if result != nil {
					results = append(results, result)
				}
			}
			return results, ctxWithTimeout.Err()
		}
	}

	return results, nil
}

// fetchIP 从指定的 URL 获取 IP 地址，并根据 ipType 验证 IP 类型
func (c *IPHelper) fetchIP(ctx context.Context, url string, ipType IPType, resultChan chan<- *IPResult) {
	result := &IPResult{
		Source: url,
	}

	// 使用 Resty 发送请求
	resp, err := c.client.R().
		SetContext(ctx).
		Get(url)

	// err != nil 意味着请求失败 (例如，DNS 查找失败，连接被拒绝等)
	if err != nil {
		result.Error = err
		resultChan <- result
		return
	}

	// IsError 检查状态码是否为 4xx 或 5xx
	if resp.IsError() {
		result.Error = fmt.Errorf("请求返回错误状态码: %s", resp.Status())
		resultChan <- result
		return
	}

	// resp.String() 返回响应体作为字符串
	ip := c.extractIP(resp.String())
	if ip == "" {
		result.Error = fmt.Errorf("无法从响应体中提取 IP 地址")
		resultChan <- result
		return
	}

	// 验证提取的 IP 地址是否符合请求的 IP 类型
	if !c.validateIPType(ip, ipType) {
		result.Error = fmt.Errorf("提取的 IP 地址类型与请求的类型不匹配: %s", ipType)
		resultChan <- result
		return
	}

	result.IP = ip
	resultChan <- result
}

// extractIP 从响应中提取 IP 地址
func (c *IPHelper) extractIP(response string) string {
	// 使用正则表达式匹配 IPv4 地址
	ipv4Regex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	matches := ipv4Regex.FindString(response)
	if matches != "" {
		return matches
	}

	// 如果需要，也可以添加 IPv6 地址的正则表达式匹配
	ipv6Regex := regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)
	matches = ipv6Regex.FindString(response)
	if matches != "" {
		return matches
	}

	return ""
}

// validateIPType 验证 IP 地址是否与指定的 IP 类型匹配
func (c *IPHelper) validateIPType(ip string, ipType IPType) bool {
	switch ipType {
	case IPTypeV4:
		return c.isIPv4(ip)
	case IPTypeV6:
		return c.isIPv6(ip)
	default:
		// 如果未指定类型，则接受任何类型
		return c.isIPv4(ip) || c.isIPv6(ip)
	}
}

// isIPv4 检查是否为 IPv4 地址
func (c *IPHelper) isIPv4(ip string) bool {
	ipv4Regex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	return ipv4Regex.MatchString(ip)
}

// isIPv6 检查是否为 IPv6 地址
func (c *IPHelper) isIPv6(ip string) bool {
	ipv6Regex := regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)
	return ipv6Regex.MatchString(ip)
}
