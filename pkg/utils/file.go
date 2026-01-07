package utils

import (
	"bufio"
	"log/slog"
	"os"
)

// ReadLastLines 读取文件最后 N 行
func ReadLastLines(filePath string, maxLines int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("ReadLastLines Open Error",
			slog.String("filePath", filePath),
			slog.Any("err", err),
		)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// 设定一个 5MB 的缓冲区上限
	const maxCapacity = 5 * 1024 * 1024

	// 初始大小 64KB
	buf := make([]byte, 64*1024)
	// 设置缓冲区及其上限

	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > maxLines {
			lines = lines[1:] // 保持只存最后 N 行
		}
	}

	if err := scanner.Err(); err != nil {
		// 如果一行超过了 5MB，这里依然会报错，但 5MB 对日志来说足够了
		slog.Error("ReadLastLines Scanner Error",
			slog.String("filePath", filePath),
			slog.Any("err", err),
		)
		return nil, err
	}

	// 习惯上日志最新的在下面，如果前端需要最新的在最上面，可以 Reverse 这个 slice
	return lines, nil
}
