/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"
	"vibe_ip/internal/helper/config_helper"
	"vibe_ip/internal/helper/ip_helper"

	"github.com/spf13/cobra"
)

// 用于存储 flag 值的变量
var dnsUpdateID string

// dnsUpdateCmd represents the update command
var dnsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "DNS 更新",
	Long: `DNS 更新
go run main.go dns update --id dns1
`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		var logPrefix = fmt.Sprintf("dns-%s", dnsUpdateID)

		// 读取配置
		targetProvider, err := config_helper.QueryDnsProvider(dnsUpdateID)
		if err != nil {
			cmd.PrintErrf("[%s]读取配置失败: %s \n",
				logPrefix,
				err,
			)
			return err
		}
		cmd.Printf("[%s]找到配置: type=%s provider=%s domain=%s ip_type=%s auto_update=%v \n",
			logPrefix,
			targetProvider.Type,
			targetProvider.Provider,
			targetProvider.Domain,
			targetProvider.IpType,
			targetProvider.AutoUpdate,
		)

		// 获取IP
		checker := ip_helper.NewIPChecker()
		ipInfo, err := checker.CheckIP(cmd.Context(), targetProvider.IpType)
		if err != nil {
			cmd.PrintErrf("[%s]获取IP失败: %s \n",
				logPrefix,
				err,
			)
			return err
		}
		if ipInfo.IP == "" {
			cmd.PrintErrf("[%s]找不到IP \n",
				logPrefix,
			)
			return fmt.Errorf("找不到IP")
		}

		return nil
	},
}

func init() {
	dnsCmd.AddCommand(dnsUpdateCmd)

	dnsUpdateCmd.Flags().StringVar(&dnsUpdateID, "id", "", "Specify ID of DNS provider")
}
