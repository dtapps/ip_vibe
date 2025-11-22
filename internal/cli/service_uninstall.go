/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/services/service"

	"github.com/spf13/cobra"
)

// serviceUninstallCmd represents the uninstall command
var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载服务",
	Long:  `卸载服务`,
	RunE: func(cmd *cobra.Command, args []string) error {

		s, err := service.NewSVC(service.NewProgram(), service.NewSVCConfig())
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}

		err = s.Uninstall()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("服务卸载成功")

		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceUninstallCmd)
}
