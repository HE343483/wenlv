package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// LogEntry 日志表模型(log_entries):所有级别的日志统一落库,按 level 字段区分,
// 报错带错误 ID 与严重程度,访问日志带请求 ID,可在数据库中直接检索定位。
type LogEntry struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"not null;index:idx_log_created;comment:记录时间" json:"created_at"`
	Level     string    `gorm:"size:10;not null;index:idx_log_level;comment:日志级别:DEBUG调试/INFO运行节点/ACCESS访问/WARN警告/ERROR报错/FATAL致命" json:"level"`
	Severity  string    `gorm:"size:10;not null;comment:报错严重程度:LOW轻微/MEDIUM一般/HIGH严重/CRITICAL致命(仅报错类,非报错为空串)" json:"severity"`
	ErrorID   string    `gorm:"size:40;not null;index:idx_log_error_id;comment:错误ID(ERR-日期-序号-严重程度),精确定位唯一一条报错(非报错为空串)" json:"error_id"`
	RequestID string    `gorm:"size:40;not null;index:idx_log_request_id;comment:请求ID(REQ-日期-序号),串联访问日志与报错(非请求链路日志为空串)" json:"request_id"`
	Module    string    `gorm:"size:32;not null;comment:来源模块:server=后端服务,其余为独立工具名(crawler/story-enrich等)" json:"module"`
	Source    string    `gorm:"size:128;not null;comment:调用位置(文件:行号)" json:"source"`
	Message   string    `gorm:"type:text;not null;comment:日志内容" json:"message"`
	Fields    string    `gorm:"type:text;not null;comment:附加字段(k=v 形式,如 err=... task_id=...;无附加字段为空串)" json:"fields"`
}

// TableName 表名。
func (LogEntry) TableName() string { return "log_entries" }

// entry 队列容量与批量写入参数。
const (
	entryQueueSize = 8192 // 日志队列容量,写库过慢时超出部分丢弃
	flushBatch     = 100   // 单次批量 INSERT 的最大条数
	flushInterval  = 500 * time.Millisecond
)

// writer 负责把日志异步批量写入 MySQL:
// 业务侧只往 channel 投递,后台 goroutine 攒批落库,避免每条日志一次 INSERT 拖慢业务。
type writer struct {
	mu      sync.Mutex
	db      *gorm.DB // 已静默 GORM 日志的会话,防止日志写库再次触发日志造成递归
	ch      chan *LogEntry
	quit    chan struct{} // 退出信号:通知 loop 排空后返回;ch 永不 close,避免业务侧发送到已关闭 channel panic
	done    chan struct{}
	closeCh   sync.Once
	closeDone sync.Once
	ready     bool

	retention int // 保留天数,<=0 不清理
	dropped   atomic.Uint64
	warnOnce  sync.Once
}

var w = &writer{
	ch:   make(chan *LogEntry, entryQueueSize),
	quit: make(chan struct{}),
	done: make(chan struct{}),
}

// dbReady 判断日志是否已接入数据库。
func dbReady() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ready
}

// SetDB 接入数据库:自动建 log_entries 表、续接当日错误 ID 序号,
// 并启动异步批量写入与过期清理。在 MySQL 连接成功后调用一次。
func SetDB(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db 为空")
	}
	// 用静默会话写库:GORM 的日志已接入本模块,否则写日志的 SQL 又会记一条日志,无限递归
	silent := db.Session(&gorm.Session{
		Logger: gormlogger.Discard,
		NewDB:  true,
	})
	if err := silent.AutoMigrate(&LogEntry{}); err != nil {
		return err
	}

	// 错误 ID 序号续接:读取当日最大序号,重启后不与已有报错重复
	seedErrorSeq(silent, time.Now())

	w.mu.Lock()
	if w.ready {
		w.mu.Unlock()
		return nil
	}
	w.db = silent
	w.ready = true
	w.mu.Unlock()

	go w.loop()
	go w.cleanupLoop()
	return nil
}

// Configure 应用日志配置(控制台级别/库内保留天数)。数据库接入前调用,只影响控制台输出。
func Configure(opts Options) {
	consoleMu.Lock()
	consoleEnabled = !opts.DisableConsole
	consoleLevel = ParseLevel(opts.Level)
	consoleMu.Unlock()

	w.mu.Lock()
	w.retention = opts.RetentionDays
	w.mu.Unlock()

	Infof("日志模块已就绪: 落库=log_entries 控制台级别=%s 保留天数=%d", ParseLevel(opts.Level).String(), opts.RetentionDays)
	if consoleOnly() {
		Infof("数据库尚未接入,当前日志暂仅输出控制台,接入后自动开始落库")
	}
}

// Close 通知后台写入 goroutine 排空队列后退出并等待完成,进程退出前调用
// (Fatalf 内部已自动调用)。未接入数据库时(纯控制台模式)没有后台 goroutine,直接关闭。
// 注意:不关闭 ch 本身,避免其它 goroutine 并发写日志时 send on closed channel panic。
func Close() error {
	w.mu.Lock()
	ready := w.ready
	w.mu.Unlock()
	w.closeCh.Do(func() {
		close(w.quit)
	})
	if !ready {
		w.closeDone.Do(func() {
			close(w.done)
		})
	}
	<-w.done
	return nil
}

// ConfigureTool 供 crawler、cmd/* 等独立命令行工具使用:日志统一落同一张
// log_entries 表,以 module 字段区分来源,与服务端日志互不混淆。
// LOG_LEVEL / LOG_RETENTION_DAYS 环境变量可覆盖默认值。
func ConfigureTool(name string) {
	retention := 7
	if v := os.Getenv("LOG_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			retention = n
		}
	}
	setModule(name)
	Configure(Options{Level: os.Getenv("LOG_LEVEL"), RetentionDays: retention})
}

// loop 后台攒批写库:满一批立即写入,否则按固定间隔刷一次。
// 收到退出信号后排空队列中现存日志再返回。
func (w *writer) loop() {
	defer close(w.done)
	buf := make([]LogEntry, 0, flushBatch)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case e := <-w.ch:
			buf = append(buf, *e)
			if len(buf) >= flushBatch {
				w.flush(&buf)
			}
		case <-ticker.C:
			w.flush(&buf)
		case <-w.quit:
			// 排空现存队列后退出(此时新到的日志可能漏 1-2 条,进程即将退出,可接受)
			for {
				select {
				case e := <-w.ch:
					buf = append(buf, *e)
					if len(buf) >= flushBatch {
						w.flush(&buf)
					}
				default:
					w.flush(&buf)
					return
				}
			}
		}
	}
}

// flush 批量写入,失败时降级提示(仅一次)并丢弃,避免阻塞业务。
func (w *writer) flush(buf *[]LogEntry) {
	if len(*buf) == 0 {
		return
	}
	w.mu.Lock()
	db := w.db
	w.mu.Unlock()
	if db == nil {
		*buf = (*buf)[:0]
		return
	}
	if err := db.CreateInBatches(*buf, flushBatch).Error; err != nil {
		w.warnOnce.Do(func() {
			_, _ = fmt.Fprintf(os.Stderr, "[logger] 日志落库失败,已降级为仅控制台输出: %v\n", err)
		})
	}
	*buf = (*buf)[:0]
}

// cleanupLoop 每小时清理一次超过保留期的日志行。
func (w *writer) cleanupLoop() {
	for {
		time.Sleep(time.Hour)
		w.mu.Lock()
		db, retention := w.db, w.retention
		w.mu.Unlock()
		if db == nil || retention <= 0 {
			continue
		}
		cutoff := time.Now().AddDate(0, 0, -retention)
		_ = db.Where("created_at < ?", cutoff).Delete(&LogEntry{}).Error
	}
}

// ===== 投递与控制台输出 =====

var (
	consoleMu      sync.Mutex
	consoleEnabled = true
	consoleLevel   = ParseLevel(defaultLevel)
)

// moduleVar 当前来源模块,默认 server;独立工具通过 ConfigureTool 设置。
// 用 atomic.Value 读写:每条日志都会读取,避免与 setModule 产生数据竞争。
var moduleVar atomic.Value

func init() { moduleVar.Store("server") }

func setModule(name string) {
	if name = strings.TrimSpace(name); name != "" {
		moduleVar.Store(name)
	}
}

// currentModule 返回当前来源模块名。
func currentModule() string {
	if v, ok := moduleVar.Load().(string); ok {
		return v
	}
	return "server"
}

// renderFields 把结构化 kv 序列化为「k=v」空格分隔的字符串,用于落库 Fields 列。
func renderFields(kv []any) string {
	var b strings.Builder
	for i := 0; i+1 < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if !ok || k == "" {
			continue
		}
		v := escapeValue(fmt.Sprintf("%v", kv[i+1]))
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}
	return b.String()
}

// buildEntry 组装落库记录:把 kv 中的 request_id 提取成独立列便于检索,其余留在 fields。
func buildEntry(t time.Time, level Level, errID string, sev Severity, source, msg string, kv []any) *LogEntry {
	e := &LogEntry{
		CreatedAt: t,
		Level:     level.String(),
		ErrorID:   errID,
		Module:    currentModule(),
		Source:    source,
		Message:   msg,
		Fields:    renderFields(kv),
	}
	if sev != "" {
		e.Severity = string(sev)
	}
	// request_id 是排查链路的关键字段,单独提取成列
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok && k == "request_id" {
			if v, ok := kv[i+1].(string); ok {
				e.RequestID = v
				break
			}
		}
	}
	return e
}

// emitEntry 投递到落库队列(未接入数据库时丢弃,只留控制台),并按配置输出控制台。
func emitEntry(now time.Time, level Level, e *LogEntry) {
	select {
	case w.ch <- e:
	default:
		// 队列满:丢弃并计数,避免反压阻塞业务 goroutine
		n := w.dropped.Add(1)
		if n == 1 || n%1000 == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "[logger] 日志队列已满,累计丢弃 %d 条\n", n)
		}
	}

	consoleMu.Lock()
	toConsole := consoleEnabled && level >= consoleLevel
	consoleMu.Unlock()
	if toConsole {
		_, _ = os.Stdout.WriteString(levelColor[level] + composeConsole(now, e) + "\x1b[0m\n")
	}
}

// composeConsole 拼装控制台日志行,格式与表字段一一对应:
// 时间 | 级别 | 错误ID(- 表示无) | 位置 | 消息 | 字段
func composeConsole(t time.Time, e *LogEntry) string {
	var b strings.Builder
	b.WriteString(t.Format(timeLayout))
	b.WriteString(" | ")
	b.WriteString(fmt.Sprintf("%-6s", e.Level))
	b.WriteString(" | ")
	if e.ErrorID == "" {
		b.WriteString("-")
	} else {
		b.WriteString(e.ErrorID)
	}
	b.WriteString(" | ")
	b.WriteString(e.Source)
	b.WriteString(" | ")
	b.WriteString(e.Message)
	if e.Fields != "" {
		b.WriteString(" | ")
		b.WriteString(e.Fields)
	}
	return b.String()
}

// escapeValue 规范化字段值:压缩换行避免破坏控制台「一行一条日志」的结构,
// 含空格时加引号保证 k=v 字段不被截断。
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

// ===== 错误 ID 序号 =====

var (
	errSeqMu  sync.Mutex
	errSeqDay string
	errSeq    uint64
)

// nextErrorSeq 返回指定日期的下一个错误序号(线程安全)。
func nextErrorSeq(day string) uint64 {
	errSeqMu.Lock()
	defer errSeqMu.Unlock()
	if errSeqDay != day {
		errSeqDay = day
		errSeq = 0
	}
	errSeq++
	return errSeq
}

// seedErrorSeq 从库里读取当日已用过的最大错误序号,重启后序号续接不重复。
func seedErrorSeq(db *gorm.DB, t time.Time) {
	day := t.Format(idDayLayout)
	var maxID string
	_ = db.Model(&LogEntry{}).
		Where("level IN ? AND error_id LIKE ?", []string{"ERROR", "FATAL"}, "ERR-"+day+"-%").
		Select("MAX(error_id)").
		Scan(&maxID).Error
	// ERR-20260920-0001-HIGH → 按 - 切分为 4 段,第 3 段是序号
	parts := strings.Split(maxID, "-")
	if len(parts) == 4 {
		if n, err := strconv.ParseUint(parts[2], 10, 64); err == nil {
			errSeqMu.Lock()
			if day != errSeqDay || n > errSeq {
				errSeqDay = day
				errSeq = n
			}
			errSeqMu.Unlock()
		}
	}
}
