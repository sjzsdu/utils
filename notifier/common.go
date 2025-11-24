package notifier

import (
	"fmt"
	"strings"
)

// FormatNotificationTitle 格式化通知标题
func FormatNotificationTitle(items []MessageItem) string {
	total := len(items)
	if total == 0 {
		return "空的趋势雷达通知"
	}

	return fmt.Sprintf("📊 趋势雷达: %d条资讯", total)
}

// FormatNotificationSummary 格式化通知摘要
func FormatNotificationSummary(items []MessageItem) string {
	if len(items) == 0 {
		return "暂无新资讯"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("共 %d 条资讯", len(items)))

	return summary.String()
}

// truncateText 截断文本
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
