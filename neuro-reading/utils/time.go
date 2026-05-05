package utils

import "time"

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func CurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func CurrentDate() string {
	return time.Now().Format("2006-01-02")
}