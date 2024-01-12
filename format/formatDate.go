package format

import (
	"fmt"
	"strings"
	"time"
)

// formatTime 用於將時間字符串轉換為更友好的格式
func FormatTime(timeStr string) string {
	return strings.Replace(timeStr, "2024-", "", 1) // 示例：移除年份
}

// formatTime 將時間字符串從 "01-11 12:00:00" 轉換為 "1月11日 12:00"
func FormatToCh(timeStr string) string {
	// 解析時間字符串，假設年份為2024
	parsedTime, err := time.Parse("01-02 15:04:05", timeStr)
	if err != nil {
		return timeStr // 如果解析失敗，返回原始字符串
	}
	return fmt.Sprintf("%d月%d日 %02d:%02d", parsedTime.Month(), parsedTime.Day(), parsedTime.Hour(), parsedTime.Minute())
}
