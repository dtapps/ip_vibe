/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/helper/config_helper"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configShowCmd represents the show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "示所有 DNS 和 CDN 提供商配置",
	Long: `示所有 DNS 和 CDN 提供商配置
go run main.go config show
go run main.go config show --config ./config_copy.yaml
`,
	Run: func(cmd *cobra.Command, args []string) {

		// 读取配置
		var config config_helper.Config
		if err := viper.Unmarshal(&config); err != nil {
			cmd.PrintErrf("解析配置文件失败: %v \n", err)
			return
		}

		// 验证提供商
		for _, p := range config.Providers {
			if p.Type != "dns" && p.Type != "cdn" {
				cmd.PrintErrf("无效的提供商类型 %s: %s \n", p.Domain, p.Type)
				return
			}
			if p.ID == "" {
				cmd.PrintErrf("提供商缺少 ID: %s \n", p.Domain)
				return
			}
		}

		// 显示提供商列表
		cmd.Printf("找到 %d 个提供商 \n", len(config.Providers))
		if len(config.Providers) == 0 {
			cmd.Println("未配置任何提供商")
			return
		}

		cmd.Println("提供商列表:")
		for i, p := range config.Providers {
			if p.Type == "dns" {
				cmd.Printf("%d. 类型: %s, ID: %s, 提供商: %s, 域名: %s, 配置: %v \n",
					i+1, p.Type, p.ID, p.Provider, p.Domain, p.Config)
			} else {
				cmd.Printf("%d. 类型: %s, ID: %s, 提供商: %s, 域名: %s, 配置: %v, DNS 提供商: %s \n",
					i+1, p.Type, p.ID, p.Provider, p.Domain, p.Config, p.DNSProvider.Provider)
			}
		}

	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
}
