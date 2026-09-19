// Package logger 是后端统一的日志模块:按级别分文件落盘、按天轮转,
// 并给每条报错生成「含日期 + 严重程度」的错误 ID,便于精确定位线上问题。
//
// 目录结构(默认 logs/,可用 LOG_DIR 覆盖):
//
//	logs/
//	  debug-2026-09-20.log   调试细节(GORM SQL、LLM 原始响应等)
//	  info-2026-09-20.log    程序运行节点(启动、任务阶段、关键流程)
//	  access-2026-09-20.log  HTTP 访问日志(方法/路径/状态码/耗时)
//	  warn-2026-09-20.log    警告(可恢复异常、降级、上游风控)
//	  error-2026-09-20.log   报错(含错误 ID 与严重程度)
//	  fatal-2026-09-20.log   致命错误(进程退出前写入,同时会记入 error 文件)
//
// 日志行格式:
//
//	时间 | 级别 | 错误ID(- 表示无) | 位置 | 消息 | 字段
//	2026-09-20 14:25:31.456 | ERROR | ERR-20260920-0001-HIGH | handler/trip_handler.go:94 | 规划失败 | request_id=REQ-...
//
// 错误 ID 形如 ERR-20260920-0001-HIGH:日期 + 当日序号 + 严重程度,
// 直接 grep `ERR-20260920-0001` 即可定位到唯一一条报错。
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultDir           = "logs"
	defaultLevel         = "info"
	defaultRetentionDays = 30
	// timeLayout 落盘时间格式(毫秒精度,便于排查并发时序)。
	timeLayout = "2006-01-02 15:04:05.000"
	// idDayLayout 错误 ID 中的日期格式。
	idDayLayout = "20060102"
	// fileDayLayout 日志文件名中的日期格式。
	fileDayLayout = "2006-01-02"
)

// Level 日志级别,每个级别对应一个独立的日志文件。
type Level int

const (
	// LevelDebug 调试细节,仅落盘(控制台需 LOG_LEVEL=debug)。
	LevelDebug Level = iota
	// LevelInfo 程序运行节点。
	LevelInfo
	// LevelAccess HTTP 访问日志。
	LevelAccess
	// LevelWarn 警告。
	LevelWarn
	// LevelError 报错。
	LevelError
	// LevelFatal 致命错误。
	LevelFatal
)

var (
	levelNames = [...]string{"DEBUG", "INFO", "ACCESS", "WARN", "ERROR", "FATAL"}
	levelFiles = [...]string{"debug", "info", "access", "warn", "error", "fatal"}
	levelColor = [...]string{"\x1b[90m", "\x1b[36m", "\x1b[35m", "\x1b[33m", "\x1b[31m", "\x1b[41;97m"}
)

// String 返回级别的可读名称。
func (l Level) String() string {
	if l < LevelDebug || int(l) >= len(levelNames) {
		return "UNKNOWN"
	}
	return levelNames[l]
}

// fileName 返回该级别日志的文件名前缀。
func (l Level) fileName() string {
	if l < LevelDebug || int(l) >= len(levelFiles) {
		return "unknown"
	}
	return levelFiles[l]
}

// ParseLevel 解析级别名(大小写不敏感),无法识别时按 INFO 处理。
func ParseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return LevelDebug
	case "ACCESS":
		return LevelAccess
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Severity 报错严重程度,取值 LOW / MEDIUM / HIGH / CRITICAL。
type Severity string

const (
	// SeverityLow 轻微:不影响本次请求结果(如个别缓存写入失败)。
	SeverityLow Severity = "LOW"
	// SeverityMedium 一般:单个请求失败,主流程其它功能正常(默认)。
	SeverityMedium Severity = "MEDIUM"
	// SeverityHigh 严重:某个功能整体不可用(如依赖的上游接口全部失败)。
	SeverityHigh Severity = "HIGH"
	// SeverityCritical 致命:进程级故障,需立即处理。
	SeverityCritical Severity = "CRITICAL"
)

// ParseSeverity 解析严重程度,无法识别时按 MEDIUM 处理。
func ParseSeverity(s string) Severity {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "LOW", "L1":
		return SeverityLow
	case "HIGH", "L3":
		return SeverityHigh
	case "CRITICAL", "L4":
		return SeverityCritical
	default:
		return SeverityMedium
	}
}

// Options 日志模块配置。
type Options struct {
	// Dir 日志根目录,默认 logs。
	Dir string
	// Level 控制台输出的最低级别(debug/info/access/warn/error/fatal),默认 info。
	// 文件始终按级别完整落盘,不受该配置影响。
	Level string
	// DisableConsole 关闭控制台输出(仅落盘)。
	DisableConsole bool
	// RetentionDays 日志保留天数,<=0 表示不清理。
	RetentionDays int
}

// Logger 携带固定字段的日志器,便于在一条链路中带上 request_id/task_id 等上下文。
type Logger struct {
	kv []any
}

// std 全局默认日志器(无固定字段)。
var std = &Logger{}

// With 创建带固定字段的日志器,字段以 key, value 成对传入。
func With(kv ...any) *Logger { return &Logger{kv: kv} }

// With 在现有字段基础上追加字段,返回新的日志器。
func (l *Logger) With(kv ...any) *Logger {
	merged := make([]any, 0, len(l.kv)+len(kv))
	merged = append(merged, l.kv...)
	merged = append(merged, kv...)
	return &Logger{kv: merged}
}

// ===== 全局日志函数 =====

// Debugf 调试细节(含 GORM SQL),写入 debug 文件。
func Debugf(format string, args ...any) { std.logAt(LevelDebug, "", nil, "", 0, format, args...) }

// Infof 记录程序运行节点。
func Infof(format string, args ...any) { std.logAt(LevelInfo, "", nil, "", 0, format, args...) }

// Accessf 记录 HTTP 访问日志。
func Accessf(format string, args ...any) { std.logAt(LevelAccess, "", nil, "", 0, format, args...) }

// Warnf 记录警告。
func Warnf(format string, args ...any) { std.logAt(LevelWarn, "", nil, "", 0, format, args...) }

// Errorf 记录报错(默认 MEDIUM),返回错误 ID。
func Errorf(format string, args ...any) string {
	return std.logAt(LevelError, SeverityMedium, nil, "", 0, format, args...)
}

// ErrorSev 记录指定严重程度的报错,并附带原始 error,返回错误 ID。
func ErrorSev(sev Severity, err error, format string, args ...any) string {
	return std.logAt(LevelError, sev, err, "", 0, format, args...)
}

// Fatalf 记录致命错误(CRITICAL)后退出进程。
func Fatalf(format string, args ...any) {
	std.logAt(LevelFatal, SeverityCritical, nil, "", 0, format, args...)
	_ = Close()
	os.Exit(1)
}

// ===== 带上下文字段的日志方法 =====

// Debugf 调试细节。
func (l *Logger) Debugf(format string, args ...any) {
	l.logAt(LevelDebug, "", nil, "", 0, format, args...)
}

// Infof 程序运行节点。
func (l *Logger) Infof(format string, args ...any) {
	l.logAt(LevelInfo, "", nil, "", 0, format, args...)
}

// Accessf HTTP 访问日志。
func (l *Logger) Accessf(format string, args ...any) {
	l.logAt(LevelAccess, "", nil, "", 0, format, args...)
}

// Warnf 警告。
func (l *Logger) Warnf(format string, args ...any) {
	l.logAt(LevelWarn, "", nil, "", 0, format, args...)
}

// Errorf 报错(默认 MEDIUM),返回错误 ID。
func (l *Logger) Errorf(format string, args ...any) string {
	return l.logAt(LevelError, SeverityMedium, nil, "", 0, format, args...)
}

// ErrorSev 指定严重程度的报错,返回错误 ID。
func (l *Logger) ErrorSev(sev Severity, err error, format string, args ...any) string {
	return l.logAt(LevelError, sev, err, "", 0, format, args...)
}

// Fatalf 致命错误后退出进程。
func (l *Logger) Fatalf(format string, args ...any) {
	l.logAt(LevelFatal, SeverityCritical, nil, "", 0, format, args...)
	_ = Close()
	os.Exit(1)
}

// ===== 内部实现 =====

// logAt 是唯一的落盘入口。callerFile 为空时自动取调用方位置(调用栈第 2 层)。
func (l *Logger) logAt(level Level, sev Severity, err error, callerFile string, callerLine int, format string, args ...any) string {
	now := time.Now()
	msg := fmt.Sprintf(format, args...)

	file, line := callerFile, callerLine
	if file == "" {
		if _, f, ln, ok := runtime.Caller(2); ok {
			file, line = trimPath(f), ln
		}
	}

	// 报错级别统一生成错误 ID,并把严重程度写进 ID 便于单独摘出来定位
	errID := ""
	if level >= LevelError {
		if level == LevelFatal {
			sev = SeverityCritical
		}
		if sev == "" {
			sev = SeverityMedium
		}
		errID = nextErrorID(now, sev)
	}

	kv := l.kv
	if err != nil {
		if len(kv) == 0 {
			kv = []any{"err", err.Error()}
		} else {
			kv = append(append([]any{}, kv...), "err", err.Error())
		}
	}

	entry := compose(now, level, errID, formatCaller(file, line), msg, kv)
	mgr.emit(now, level, entry)
	return errID
}

// formatCaller 拼装调用位置;无行号时(如 GORM 适配器)只输出来源名。
func formatCaller(file string, line int) string {
	if line <= 0 {
		return file
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// compose 拼装日志行。
func compose(t time.Time, level Level, errID, caller, msg string, kv []any) string {
	var b strings.Builder
	b.WriteString(t.Format(timeLayout))
	b.WriteString(" | ")
	b.WriteString(fmt.Sprintf("%-6s", level.String()))
	b.WriteString(" | ")
	if errID == "" {
		b.WriteString("-")
	} else {
		b.WriteString(errID)
	}
	b.WriteString(" | ")
	b.WriteString(caller)
	b.WriteString(" | ")
	b.WriteString(msg)
	if len(kv) > 0 {
		b.WriteString(" | ")
		b.WriteString(renderFields(kv))
	}
	return b.String()
}

// renderFields 将 key/value 对渲染为 k=v 形式。
func renderFields(kv []any) string {
	var b strings.Builder
	for i := 0; i < len(kv); i += 2 {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(fmt.Sprint(kv[i]))
		b.WriteByte('=')
		if i+1 < len(kv) {
			b.WriteString(escapeValue(fmt.Sprint(kv[i+1])))
		} else {
			b.WriteString(`""`)
		}
	}
	return b.String()
}

// escapeValue 压缩值中的换行,避免破坏「一行一条日志」的结构。
func escapeValue(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "\\n")
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}

// trimPath 取相对包路径(保留最后两级),如 service/trip_xhs.go。
func trimPath(p string) string {
	p = filepath.ToSlash(p)
	parts := strings.Split(p, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return p
}
