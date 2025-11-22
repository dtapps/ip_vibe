/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"os"
	"path/filepath"
	"vibe_ip/internal/helper/config_helper"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vibe_ip",
	Short: "动态域名解析",
	Long:  `VibeIP 是一个动态域名解析（DDNS）工具，支持多个 DNS 和 CDN 提供商，简化动态 IP 的域名管理和 CDN 配置。`,
}

// configFile 存储配置文件路径
var configFile string

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
func init() {
	// 初始化 Viper
	viper.SetConfigType("yaml")                   // 明确指定 YAML 格式
	viper.SetConfigName("config")                 // 配置文件名（无扩展名）
	viper.AddConfigPath(".")                      // 当前工作目录
	viper.AddConfigPath(filepath.Dir(os.Args[0])) // 可执行文件目录
	viper.AddConfigPath(os.Getenv("HOME"))        // 用户主目录（备用）
	viper.SetConfigPermissions(0600)              // 限制文件权限

	// 设置默认值
	viper.SetDefault("providers", []config_helper.Config{})

	// 定义全局 --config 标志
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "配置文件路径 (默认: ./config.yaml)")
	viper.BindPFlag("config_file", rootCmd.PersistentFlags().Lookup("config"))

	// 自动加载环境变量（保护敏感数据）
	viper.AutomaticEnv()

	// PersistentPreRun 确保所有子命令加载配置文件
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if configFile != "" {
			viper.SetConfigFile(configFile)
			cmd.Printf("读取配置文件路径: %s \n", configFile)
		}

		if err := viper.ReadInConfig(); err != nil {
			cmd.Printf("读取配置文件路径: %v \n", viper.ConfigFileUsed())
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				cmd.Println("未找到配置文件")
			} else {
				cmd.PrintErrf("读取配置文件失败: %v \n", err)
				return
			}
		}

		// 监控配置文件变化
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			cmd.Println("配置文件已更改: ", e.Name)
		})

	}

}
