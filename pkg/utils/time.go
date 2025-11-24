package utils

import "time"

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}

func FormatTimes(times []time.Time) []string {
	result := make([]string, 0, len(times))
	for _, t := range times {
		result = append(result, FormatTime(t))
	}
	return result
}
