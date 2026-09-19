package logger

import (
	"fmt"
	"strings"
)

// GormWriter 把 GORM 的 SQL 日志接入统一日志模块:
// 普通 SQL 进 debug 文件、慢 SQL 进 warn 文件、SQL 错误进 error 文件(严重程度 LOW)。
// 这样 SQL 细节不会再和控制台/业务日志混在一起,排查时按级别去对应文件找即可。
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
