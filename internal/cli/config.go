/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  `配置管理`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("config called")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
