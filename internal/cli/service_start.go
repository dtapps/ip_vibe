/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/services/service"

	"github.com/spf13/cobra"
)

// serviceStartCmd represents the start command
var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "启动服务",
	Long:  `启动服务`,
	RunE: func(cmd *cobra.Command, args []string) error {

		s, err := service.NewSVC(service.NewProgram(), service.NewSVCConfig())
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}

		err = s.Start()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("服务启动成功")

		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceStartCmd)
}
