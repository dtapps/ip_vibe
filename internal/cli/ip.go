/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"github.com/spf13/cobra"
)

// ipCmd represents the ip command
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "IP管理",
	Long:  `IP管理`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("ip called")
	},
}

func init() {
	rootCmd.AddCommand(ipCmd)
}
