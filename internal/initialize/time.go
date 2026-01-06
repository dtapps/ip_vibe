package initialize

import (
	"log/slog"
	"runtime"
	"time"
	_ "time/tzdata" // 时区 https://www.itlangzi.com/s/1526209468730183680.html
	"vibe_ip/internal/global"

	"github.com/dromara/carbon/v2"
)

// 初始化时间
func init() {
	if runtime.GOOS == "windows" {
		var err error
		time.Local, err = time.LoadLocation(global.DefaultTimezone)
		if err != nil {
			slog.Error("[初始化 时间] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}
	}
	carbon.SetTimezone(carbon.Shanghai)
	carbon.SetLocale("zh-CN")
}
