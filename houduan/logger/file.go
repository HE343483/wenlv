package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// manager 负责日志文件的打开、按天轮转、容量清理与控制台输出。
// 所有写操作都在同一把锁内完成,保证多 goroutine(行程任务、定时抓取)下不串行错乱。
type manager struct {
	mu sync.Mutex

	dir           string
	console       bool
	consoleLevel  Level
	retentionDays int

	// day 为当前文件句柄所属日期;跨天时关闭旧句柄,下次写入自动新建当天文件
	day   string
	files map[Level]*os.File

	// 错误 ID 的当日序号
	errDay string
	errSeq uint64

	// 目录不可写时只提示一次,之后退化为仅控制台输出
	warned bool
}

var mgr = &manager{
	dir:           defaultDir,
	console:       true,
	consoleLevel:  ParseLevel(defaultLevel),
	retentionDays: defaultRetentionDays,
	files:         make(map[Level]*os.File),
}

// Configure 应用日志配置。可在启动时调用一次;重复调用会切换目录并重新打开文件。
func Configure(opts Options) {
	dir := strings.TrimSpace(opts.Dir)
	if dir == "" {
		dir = defaultDir
	}

	mgr.mu.Lock()
	if mgr.dir != dir {
		mgr.closeFilesLocked()
		mgr.dir = dir
	}
	mgr.console = !opts.DisableConsole
	mgr.consoleLevel = ParseLevel(opts.Level)
	mgr.retentionDays = opts.RetentionDays
	mgr.warned = false
	mgr.mu.Unlock()

	cleanupExpired(dir, opts.RetentionDays)
	Infof("日志模块已就绪: 目录=%s 控制台级别=%s 保留天数=%d", dir, ParseLevel(opts.Level).String(), opts.RetentionDays)
}

// Close 关闭所有日志文件句柄。
func Close() error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	return mgr.closeFilesLocked()
}

// ConfigureTool 供 crawler、cmd/* 等独立命令行工具使用:日志默认落到
// logs/tools/<name>/,与服务端日志分开,避免工具输出混进运行日志里。
// 保留期默认 7 天;LOG_DIR / LOG_LEVEL 环境变量可覆盖。
func ConfigureTool(name string) {
	dir := strings.TrimSpace(os.Getenv("LOG_DIR"))
	if dir == "" {
		dir = filepath.Join(defaultDir, "tools", name)
	}
	retention := 7
	if v := os.Getenv("LOG_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			retention = n
		}
	}
	Configure(Options{Dir: dir, Level: os.Getenv("LOG_LEVEL"), RetentionDays: retention})
}

// LogDir 返回当前日志目录。
func LogDir() string {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	return mgr.dir
}

// emit 将一行日志写入对应级别文件(必要时输出到控制台)。
func (m *manager) emit(now time.Time, level Level, entry string) {
	m.mu.Lock()
	day := now.Format(fileDayLayout)
	if m.day != day {
		// 跨天:关闭前一天句柄,写入时按需新建当天文件
		_ = m.closeFilesLocked()
		m.day = day
	}

	if f := m.fileLocked(level); f != nil {
		_, _ = f.WriteString(entry + "\n")
	}
	// 致命错误在两个文件里都能看到:error 文件用于汇总所有报错
	if level == LevelFatal {
		if f := m.fileLocked(LevelError); f != nil {
			_, _ = f.WriteString(entry + "\n")
		}
	}

	toConsole := m.console && level >= m.consoleLevel
	m.mu.Unlock()

	if toConsole {
		_, _ = os.Stdout.WriteString(levelColor[level] + entry + "\x1b[0m\n")
	}
}

// fileLocked 返回级别对应的文件句柄,首次使用时创建。调用方需持有锁。
func (m *manager) fileLocked(level Level) *os.File {
	if f, ok := m.files[level]; ok {
		return f
	}
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		m.warnOnceLocked(err)
		return nil
	}
	name := fmt.Sprintf("%s-%s.log", level.fileName(), m.day)
	f, err := os.OpenFile(filepath.Join(m.dir, name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		m.warnOnceLocked(err)
		return nil
	}
	m.files[level] = f

	if m.retentionDays > 0 {
		go cleanupExpired(m.dir, m.retentionDays)
	}
	return f
}

// warnOnceLocked 日志文件不可写时降级提示,避免刷屏。
func (m *manager) warnOnceLocked(err error) {
	if m.warned {
		return
	}
	m.warned = true
	_, _ = fmt.Fprintf(os.Stderr, "日志文件不可写,已降级为仅控制台输出: %v\n", err)
}

// closeFilesLocked 关闭并清空所有句柄。
func (m *manager) closeFilesLocked() error {
	var firstErr error
	for level, f := range m.files {
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(m.files, level)
	}
	return firstErr
}

// nextErrorID 生成错误 ID:ERR-YYYYMMDD-序号-严重程度,序号按天从 1 开始。
func nextErrorID(t time.Time, sev Severity) string {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	day := t.Format(idDayLayout)
	if mgr.errDay != day {
		mgr.errDay = day
		mgr.errSeq = 0
	}
	mgr.errSeq++
	return fmt.Sprintf("ERR-%s-%04d-%s", day, mgr.errSeq, sev)
}

// cleanupExpired 删除超过保留期的日志文件(仅匹配本模块生成的文件名,避免误删)。
func cleanupExpired(dir string, retentionDays int) {
	if retentionDays <= 0 {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		base := strings.TrimSuffix(name, ".log")
		// 形如 error-2026-09-20:末 10 位为日期,前缀需是已知级别
		if len(base) <= len("2006-01-02") || !strings.HasSuffix(name, ".log") {
			continue
		}
		datePart := base[len(base)-len(fileDayLayout):]
		prefix := strings.TrimSuffix(base[:len(base)-len(fileDayLayout)], "-")
		if !isLevelFile(prefix) {
			continue
		}
		day, err := time.ParseInLocation(fileDayLayout, datePart, time.Local)
		if err != nil || !day.Before(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, name))
	}
}

// isLevelFile 判断文件名前缀是否为已知日志级别。
func isLevelFile(prefix string) bool {
	for _, n := range levelFiles {
		if n == prefix {
			return true
		}
	}
	return false
}
