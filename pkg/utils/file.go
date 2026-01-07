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
		slog.Error("ReadLastLines",
			slog.String("filePath", filePath),
			slog.Any("err", err),
		)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > maxLines {
			lines = lines[1:] // 保持只存最后 N 行
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("ReadLastLines",
			slog.String("filePath", filePath),
			slog.Any("err", err),
		)
		return nil, err
	}

	// 习惯上日志最新的在下面，如果前端需要最新的在最上面，可以 Reverse 这个 slice
	return lines, nil
}
