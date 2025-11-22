/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"vibe_ip/internal/services/service"

	"github.com/spf13/cobra"
)

// serviceInstallCmd represents the install command
var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装服务",
	Long:  `安装服务`,
	RunE: func(cmd *cobra.Command, args []string) error {

		svcConfig := service.NewSVCConfig()

		svcConfig.Arguments = []string{
			"service",
			"run",
		}

		s, err := service.NewSVC(service.NewProgram(), svcConfig)
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}

		err = s.Install()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("服务安装成功")

		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceInstallCmd)
}
