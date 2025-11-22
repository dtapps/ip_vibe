package service

import "github.com/kardianos/service"

func NewSVCConfig() *service.Config {
	return &service.Config{
		Name:        "vibe_ip",
		DisplayName: "动态域名解析",
		Description: `VibeIP 是一个动态域名解析（DDNS）工具，支持多个 DNS 和 CDN 提供商，简化动态 IP 的域名管理和 CDN 配置。`,
	}
}

func NewSVC(prg *program, conf *service.Config) (service.Service, error) {
	s, err := service.New(prg, conf)
	if err != nil {
		return nil, err
	}
	return s, nil
}
