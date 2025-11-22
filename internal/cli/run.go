/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"
	"time"
	"vibe_ip/internal/helper/config_helper"
	"vibe_ip/internal/helper/ip_helper"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"
)

// 用于存储 flag 值的变量
var runInterval int

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "运行",
	Long:  `运行`,
	Run: func(cmd *cobra.Command, args []string) {

		// 初始化调度器
		s, err := gocron.NewScheduler(
			gocron.WithLocation(time.Local),
			gocron.WithLogger(gocron.NewLogger(gocron.LogLevelDebug)),
		)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		defer s.Shutdown()

		_, err = s.NewJob(
			gocron.DurationJob(time.Duration(runInterval)*time.Minute), // 任务周期
			gocron.NewTask(func() {

				var logPrefix = "cron"

				// 查询自动更新的提供商配置
				targetProviders, err := config_helper.QueryCronProviders()
				if err != nil {
					cmd.PrintErrf("[%s]读取配置失败: %s \n",
						logPrefix,
						err,
					)
					return
				}
				if len(targetProviders) == 0 {
					cmd.Printf("[%s]无需自动更新 \n",
						logPrefix,
					)
					return
				}
				for _, targetProvider := range targetProviders {

					var logPrefix = fmt.Sprintf("cron-%s-%s", targetProvider.Type, targetProvider.ID)

					// 获取IP
					checker := ip_helper.NewIPChecker()
					ipInfo, err := checker.CheckIP(cmd.Context(), targetProvider.IpType)
					if err != nil {
						cmd.PrintErrf("[%s]获取IP失败: %s \n",
							logPrefix,
							err,
						)
						continue
					}
					if ipInfo.IP == "" {
						cmd.Printf("[%s]找不到IP \n",
							logPrefix,
						)
						continue
					}

				}

			}), // 任务方法
			gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
			gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		// 启动调度器
		s.Start()

		// 阻止调度器退出
		select {}

	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().IntVarP(&runInterval, "interval", "i", 10, "自动更新间隔（分钟）默认为10分钟")
}
