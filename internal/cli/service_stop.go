/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/services/service"

	"github.com/spf13/cobra"
)

// serviceStopCmd represents the stop command
var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止服务",
	Long:  `停止服务`,
	RunE: func(cmd *cobra.Command, args []string) error {

		s, err := service.NewSVC(service.NewProgram(), service.NewSVCConfig())
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}

		err = s.Stop()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("服务停止成功")

		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceStopCmd)
}
