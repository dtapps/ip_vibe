package initialize

import (
	"log"
	"log/slog"
	"runtime"
	"time"
	_ "time/tzdata" // 时区 https://www.itlangzi.com/s/1526209468730183680.html
)

// 初始化时间
func init() {
	if runtime.GOOS == "windows" {
		var err error
		time.Local, err = time.LoadLocation("Asia/Shanghai")
		if err != nil {
			slog.Error("[初始化时间] 错误",
				slog.Any("err", err),
			)
			log.Fatal(err)
		}
	}
}
