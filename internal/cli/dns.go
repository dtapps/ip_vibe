/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS管理",
	Long:  `DNS管理`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("dns called")
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)
}
