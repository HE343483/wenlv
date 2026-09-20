package logger

import (
	"fmt"
	"strings"
)

// GormWriter 把 GORM 的 SQL 日志接入统一日志模块:
// 普通 SQL 进 DEBUG 级别、慢 SQL 进 WARN 级别、SQL 错误进 ERROR 级别(严重程度 LOW)。
// 这样 SQL 细节不会和控制台/业务日志混在一起,排查时按 level 字段到 log_entries 表里筛选即可。
type GormWriter struct{}

// Printf 实现 gorm logger.Writer 接口。
func (GormWriter) Printf(format string, args ...any) {
	// 折叠换行:保证「一行一条日志」
	msg := strings.Join(strings.Fields(fmt.Sprintf(format, args...)), " ")
	if msg == "" {
		return
	}
	switch {
	case strings.Contains(msg, "[error]"), strings.Contains(msg, "[ERROR]"):
		std.logAt(LevelError, SeverityLow, nil, "gorm", 0, "%s", msg)
	case strings.Contains(msg, "SLOW SQL"):
		std.logAt(LevelWarn, "", nil, "gorm", 0, "%s", msg)
	default:
		std.logAt(LevelDebug, "", nil, "gorm", 0, "%s", msg)
	}
}
