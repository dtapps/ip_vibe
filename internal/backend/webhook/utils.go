package webhook

import (
	"strings"
)

type TextItem struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

func StringToTextArray(str string) [][]TextItem {
	lines := strings.Split(str, "\n")
	var result []TextItem

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// 如果当前行次不是最后一行，添加换行符
			if line != lines[len(lines)-1] {
				line += "\n"
			}
			result = append(result, TextItem{
				Tag:  "text",
				Text: line,
			})
		}
	}

	return [][]TextItem{result}
}
