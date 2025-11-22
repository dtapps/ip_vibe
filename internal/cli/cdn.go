/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"github.com/spf13/cobra"
)

// cdnCmd represents the cdn command
var cdnCmd = &cobra.Command{
	Use:   "cdn",
	Short: "CDN管理",
	Long:  `CDN管理`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("cdn called")
	},
}

func init() {
	rootCmd.AddCommand(cdnCmd)
}
