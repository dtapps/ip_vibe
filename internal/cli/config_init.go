/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 模板定义
const basicTemplate = `providers:
  - type: dns # 类型(必填) dns/cdn
    id: dns1 # ID(必填、全局唯一)
    auto_update: false # 自动更新(默认: false、可选) false/true	
    provider: dnspod # dns 提供商(必填) dnspod/cloudflare
    ip_type: v6 # IP类型(必填) v4/v6
    domain: example.ddns.net # 域名(必填)
    config:
      key: value # dns 提供商特定配置
      secret_id: value
      secret_key: value
`

const fullTemplate = `providers:
  - type: dns # 类型(必填) dns/cdn
    id: dns1 # ID(必填、全局唯一)
    auto_update: false # 自动更新(默认: false、可选) false/true	
    provider: dnspod # dns 提供商(必填) dnspod/cloudflare
    ip_type: v6 # IP类型(必填) v4/v6
    domain: example.ddns.net # 域名(必填)
    config:
      key: value # dns 提供商特定配置
      secret_id: value
      secret_key: value

  - type: cdn # 类型(必填) dns/cdn
    id: cdn1 # ID(必填、全局唯一)
    auto_update: false # 自动更新(默认: false、可选) false/true	
    provider: edgeone # cdn 提供商(必填) edgeone
    ip_type: v6 # IP类型(必填) v4/v6
    domain: cdn.example.com # 域名(必填)
    config:
      key: value # cdn 提供商特定配置
      zone_id: value
      secret_id: value
      secret_key: value
    dns_provider:
      provider: dnspod # dns 提供商(必填) dnspod/cloudflare
      config:
        key: value # dns 提供商特定配置
        secret_id: value
        secret_key: value
`

var templateType string
var force bool

// configInitCmd represents the init command
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件 (config.yaml)",
	Long: `生成一个 config.yaml 模板，支持 basic（仅 DNS 提供商）或 full（DNS 和 CDN 提供商）两种类型。
go run main.go config init --type full --force
go run main.go config init --config ./config_copy_full_force.yaml --type full --force
`,
	Run: func(cmd *cobra.Command, args []string) {
		// 验证模板类型
		templateType = strings.ToLower(templateType)
		if templateType != "basic" && templateType != "full" {
			cmd.PrintErrf("错误: 无效的模板类型 '%s'，必须为 'basic' 或 'full' \n", templateType)
			os.Exit(1)
		}

		// 检查配置文件是否已存在
		configPath := viper.GetString("config_file")
		if configPath == "" {
			configPath = "config.yaml"
		}
		// if _, err := os.Stat(configPath); err == nil && !force {
		// 	cmd.Printf("配置文件 %s 已存在，是否覆盖？(y/N): \n", configPath)
		// 	var response string
		// 	fmt.Scanln(&response)
		// 	if strings.ToLower(response) != "y" {
		// 		cmd.Println("取消初始化")
		// 		return
		// 	}
		// }

		// 选择模板
		var templateContent string
		if templateType == "basic" {
			templateContent = basicTemplate
		} else {
			templateContent = fullTemplate
		}

		// 写入配置文件
		if err := os.WriteFile(configPath, []byte(templateContent), 0600); err != nil {
			cmd.PrintErrf("写入配置文件失败: %v \n", err)
			os.Exit(1)
		}

		// 更新 Viper
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			cmd.PrintErrf("加载新配置文件失败: %v \n", err)
			os.Exit(1)
		}

		cmd.Printf("已初始化配置文件 %s（类型: %s） \n", configPath, templateType)
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)

	configInitCmd.Flags().StringVar(&templateType, "type", "full", "模板类型 (basic 或 full)")
	configInitCmd.Flags().BoolVar(&force, "force", false, "强制覆盖现有配置文件")
}
