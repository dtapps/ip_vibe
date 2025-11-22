/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/services/service"

	"github.com/spf13/cobra"
)

// serviceRunCmd represents the run command
var serviceRunCmd = &cobra.Command{
	Use:   "run",
	Short: "运行服务",
	Long:  `运行服务`,
	RunE: func(cmd *cobra.Command, args []string) error {

		s, err := service.NewSVC(service.NewProgram(), service.NewSVCConfig())
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}

		err = s.Run()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("服务运行成功")

		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceRunCmd)
}
