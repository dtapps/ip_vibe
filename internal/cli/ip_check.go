/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/helper/ip_helper"

	"github.com/spf13/cobra"
)

// 用于存储 flag 值的变量
var ipCheckType string

// ipCheckCmd represents the check command
var ipCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "查看当前IP状态",
	Long: `查看当前IP状态
go run main.go ip check -type v4
go run main.go ip check -type v6
`,
	Run: func(cmd *cobra.Command, args []string) {

		checker := ip_helper.NewIPChecker()
		ipInfo, err := checker.CheckIP(cmd.Context(), ipCheckType)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		if ipInfo.IP == "" {
			cmd.Println("找不到IP")
			return
		}
		cmd.Printf("Current public IP: %s \n", ipInfo.IP)

	},
}

func init() {
	ipCmd.AddCommand(ipCheckCmd)

	ipCheckCmd.Flags().StringVarP(&ipCheckType, "type", "t", "", "Specify IP type (e.g., v4 or v6)")
}
