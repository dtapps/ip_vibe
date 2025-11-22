/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/services/service"

	"github.com/spf13/cobra"
)

// serviceRestartCmd represents the restart command
var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "重启服务",
	Long:  `重启服务`,
	RunE: func(cmd *cobra.Command, args []string) error {

		s, err := service.NewSVC(service.NewProgram(), service.NewSVCConfig())
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}

		err = s.Restart()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("服务重启成功")

		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceRestartCmd)
}
