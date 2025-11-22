package config_helper

import (
	"fmt"

	"github.com/spf13/viper"
)

// QueryCronProviders 查询自动更新的提供商配置
func QueryCronProviders() ([]*Provider, error) {

	// 读取配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 查找指定 ID 的 DNS 提供商配置
	var targetProviders []*Provider
	for i := range config.Providers {
		if config.Providers[i].AutoUpdate {
			targetProviders = append(targetProviders, &config.Providers[i])
		}
	}

	return targetProviders, nil
}

// QueryDnsProvider 查询指定的 DNS 提供商配置
func QueryDnsProvider(id string) (*Provider, error) {

	// 检查是否提供了 ID
	if id == "" {
		return nil, fmt.Errorf("请指定 ID")
	}

	// 读取配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 查找指定 ID 的 DNS 提供商配置
	var targetProvider *Provider
	for i := range config.Providers {
		if config.Providers[i].Type == "dns" && config.Providers[i].ID == id {
			targetProvider = &config.Providers[i]
			break
		}
	}

	if targetProvider == nil {
		return nil, fmt.Errorf("未找到 ID 为 '%s' 的 DNS 提供商配置", id)
	}

	return targetProvider, nil
}

// QueryCdnProvider 查询指定的 CDN 提供商配置
func QueryCdnProvider(id string) (*Provider, error) {

	// 检查是否提供了 ID
	if id == "" {
		return nil, fmt.Errorf("请指定 ID")
	}

	// 读取配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 查找指定 ID 的 CDN 提供商配置
	var targetProvider *Provider
	for i := range config.Providers {
		if config.Providers[i].Type == "cdn" && config.Providers[i].ID == id {
			targetProvider = &config.Providers[i]
			break
		}
	}

	if targetProvider == nil {
		return nil, fmt.Errorf("未找到 ID 为 '%s' 的 CDN 提供商配置", id)
	}

	return targetProvider, nil
}
