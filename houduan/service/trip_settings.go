package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wenlv-backend/model"
)

// TripRuntimeSettings 与前端设置页字段一一对应(命名沿用 TripStar 原项目,便于前端零改动迁移)。
type TripRuntimeSettings struct {
	ViteAmapWebKey   string `json:"vite_amap_web_key"`
	ViteAmapWebJSKey string `json:"vite_amap_web_js_key"`
	GoogleMapsAPIKey string `json:"google_maps_api_key"`
	GoogleMapsProxy  string `json:"google_maps_proxy"`
	XHSCookie        string `json:"xhs_cookie"`
	DouyinCookie     string `json:"douyin_cookie"`
	OpenAIAPIKey     string `json:"openai_api_key"`
	OpenAIBaseURL    string `json:"openai_base_url"`
	OpenAIModel      string `json:"openai_model"`
}

const tripMaskChar = "\u2022"

// 机密字段以掩码返回前端,保存时若回传掩码则保持原值不变。
var tripSecretFields = map[string]bool{
	"openai_api_key":    true,
	"xhs_cookie":        true,
	"douyin_cookie":     true,
	"vite_amap_web_key": true,
}

var tripSettingKeys = []string{
	"vite_amap_web_key",
	"vite_amap_web_js_key",
	"google_maps_api_key",
	"google_maps_proxy",
	"xhs_cookie",
	"douyin_cookie",
	"openai_api_key",
	"openai_base_url",
	"openai_model",
}

// TripMemoryParams 用户偏好记忆模块参数(对应原项目 MEMORY_* 环境变量)。
type TripMemoryParams struct {
	DecayFactor      float64 // 权重惰性衰减因子
	WeightThreshold  float64 // 低于该权重视为遗忘
	MaxRecall        int     // TOP-K 召回上限
	MinInitWeight    float64 // 入库准入门槛
	MaxSingleContent int     // 单条偏好注入 Prompt 时的字数上限
}

// TripSettings 管理行程模块的运行时配置:默认值来自 .env,可在前端设置页修改并持久化(优先 MySQL,磁盘 JSON 兜底)。
type TripSettings struct {
	mu       sync.RWMutex
	defaults TripRuntimeSettings
	values   TripRuntimeSettings
	file     string
	dataDir  string
	db       *gorm.DB // 可选:接入后设置项落库 trip_settings,重启不丢失
	// 以下为不参与前端热更新的静态配置
	plannerTimeout   int
	llmTimeout       int
	enableUserMemory bool
	memory           TripMemoryParams
	// 图片磁盘缓存(景点图片字节)的有效期与容量上限
	imageCacheTTL      time.Duration
	imageCacheMaxBytes int64
}

// TripSettingsOptions 构造运行时配置管理器所需的参数。
type TripSettingsOptions struct {
	Defaults         TripRuntimeSettings
	DataDir          string
	PlannerTimeout   int // 秒
	LLMTimeout       int // 秒
	EnableUserMemory bool
	Memory           TripMemoryParams
	// ImageCacheTTLHours 图片缓存有效期(小时);<=0 时用默认值
	ImageCacheTTLHours int
	// ImageCacheMaxMB 图片缓存容量上限(MB);<=0 时用默认值
	ImageCacheMaxMB int
}

// NewTripSettings 构造运行时配置管理器,并加载磁盘上已有的覆盖项。
func NewTripSettings(opt TripSettingsOptions) *TripSettings {
	dataDir := opt.DataDir
	if dataDir == "" {
		dataDir = "data"
	}
	memory := opt.Memory
	if memory.DecayFactor <= 0 {
		memory.DecayFactor = 0.97
	}
	if memory.WeightThreshold <= 0 {
		memory.WeightThreshold = 2.0
	}
	if memory.MaxRecall <= 0 {
		memory.MaxRecall = 10
	}
	if memory.MinInitWeight <= 0 {
		memory.MinInitWeight = 4.0
	}
	if memory.MaxSingleContent <= 0 {
		memory.MaxSingleContent = 120
	}
	s := &TripSettings{
		defaults:         opt.Defaults,
		values:           opt.Defaults,
		file:             filepath.Join(dataDir, "runtime_settings.json"),
		dataDir:          dataDir,
		plannerTimeout:   opt.PlannerTimeout,
		llmTimeout:       opt.LLMTimeout,
		enableUserMemory: opt.EnableUserMemory,
		memory:           memory,
	}
	if opt.ImageCacheTTLHours > 0 {
		s.imageCacheTTL = time.Duration(opt.ImageCacheTTLHours) * time.Hour
	}
	if opt.ImageCacheMaxMB > 0 {
		s.imageCacheMaxBytes = int64(opt.ImageCacheMaxMB) << 20
	}
	s.load()
	return s
}

// AttachDB 接入 MySQL:启动时读取 trip_settings 覆盖项(优先级高于磁盘 JSON),保存时写库。
func (s *TripSettings) AttachDB(db *gorm.DB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db = db
	var rows []model.TripSetting
	if err := db.Find(&rows).Error; err != nil {
		return
	}
	if len(rows) == 0 {
		return
	}
	data := make(map[string]string, len(rows))
	for _, row := range rows {
		data[row.Key] = row.Value
	}
	s.apply(data)
}

// MemoryParams 返回用户偏好记忆模块参数。
func (s *TripSettings) MemoryParams() TripMemoryParams { return s.memory }

// DataDir 返回行程模块数据目录(任务持久化、图片缓存等)。
func (s *TripSettings) DataDir() string { return s.dataDir }

// ImageCacheTTL 图片磁盘缓存有效期(<=0 时由缓存自身回退默认值)。
func (s *TripSettings) ImageCacheTTL() time.Duration { return s.imageCacheTTL }

// ImageCacheMaxBytes 图片磁盘缓存容量上限(<=0 时由缓存自身回退默认值)。
func (s *TripSettings) ImageCacheMaxBytes() int64 { return s.imageCacheMaxBytes }

// PlannerTimeoutSeconds 规划阶段 LLM 超时(秒)。
func (s *TripSettings) PlannerTimeoutSeconds() int {
	if s.plannerTimeout <= 0 {
		return 180
	}
	return s.plannerTimeout
}

// LLMTimeoutSeconds 常规 LLM 调用超时(秒)。
func (s *TripSettings) LLMTimeoutSeconds() int {
	if s.llmTimeout <= 0 {
		return 120
	}
	return s.llmTimeout
}

// MemoryEnabled 是否启用用户偏好记忆模块。
func (s *TripSettings) MemoryEnabled() bool { return s.enableUserMemory }

func (s *TripSettings) load() {
	raw, err := os.ReadFile(s.file)
	if err != nil {
		return
	}
	var data map[string]string
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}
	s.apply(data)
}

func (s *TripSettings) apply(updates map[string]string) {
	for _, key := range tripSettingKeys {
		v, ok := updates[key]
		if !ok {
			continue
		}
		// 前端未修改的密码框可能原样回传掩码,此时保留原真实值。
		if strings.Contains(v, tripMaskChar) {
			continue
		}
		switch key {
		case "vite_amap_web_key":
			s.values.ViteAmapWebKey = v
		case "vite_amap_web_js_key":
			s.values.ViteAmapWebJSKey = v
		case "google_maps_api_key":
			s.values.GoogleMapsAPIKey = v
		case "google_maps_proxy":
			s.values.GoogleMapsProxy = v
		case "xhs_cookie":
			s.values.XHSCookie = v
		case "douyin_cookie":
			s.values.DouyinCookie = v
		case "openai_api_key":
			s.values.OpenAIAPIKey = v
		case "openai_base_url":
			s.values.OpenAIBaseURL = normalizeBaseURL(v)
		case "openai_model":
			s.values.OpenAIModel = strings.TrimSpace(v)
		}
	}
}

func normalizeBaseURL(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return v
	}
	return strings.TrimRight(v, "/")
}

// Snapshot 返回当前生效配置(内部使用,含真实密钥)。
func (s *TripSettings) Snapshot() TripRuntimeSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// 空值回退到默认值,避免设置页留空导致功能不可用。
	out := s.values
	if out.OpenAIBaseURL == "" {
		out.OpenAIBaseURL = s.defaults.OpenAIBaseURL
	}
	if out.OpenAIModel == "" {
		out.OpenAIModel = s.defaults.OpenAIModel
	}
	return out
}

// Masked 返回供前端展示的配置(机密字段掩码)。
func (s *TripSettings) Masked() map[string]string {
	cur := s.Snapshot()
	raw := map[string]string{
		"vite_amap_web_key":    cur.ViteAmapWebKey,
		"vite_amap_web_js_key": cur.ViteAmapWebJSKey,
		"google_maps_api_key":  cur.GoogleMapsAPIKey,
		"google_maps_proxy":    cur.GoogleMapsProxy,
		"xhs_cookie":           cur.XHSCookie,
		"douyin_cookie":        cur.DouyinCookie,
		"openai_api_key":       cur.OpenAIAPIKey,
		"openai_base_url":      cur.OpenAIBaseURL,
		"openai_model":         cur.OpenAIModel,
	}
	for k := range raw {
		if tripSecretFields[k] {
			raw[k] = maskSecret(raw[k])
		}
	}
	return raw
}

// Update 保存配置并立即写盘,返回更新后的掩码视图。
func (s *TripSettings) Update(updates map[string]string) map[string]string {
	s.mu.Lock()
	s.apply(updates)
	persist := map[string]string{
		"vite_amap_web_key":    s.values.ViteAmapWebKey,
		"vite_amap_web_js_key": s.values.ViteAmapWebJSKey,
		"google_maps_api_key":  s.values.GoogleMapsAPIKey,
		"google_maps_proxy":    s.values.GoogleMapsProxy,
		"xhs_cookie":           s.values.XHSCookie,
		"douyin_cookie":        s.values.DouyinCookie,
		"openai_api_key":       s.values.OpenAIAPIKey,
		"openai_base_url":      s.values.OpenAIBaseURL,
		"openai_model":         s.values.OpenAIModel,
	}
	_ = os.MkdirAll(filepath.Dir(s.file), 0o755)
	if data, err := json.MarshalIndent(persist, "", "  "); err == nil {
		tmp := s.file + ".tmp"
		if err := os.WriteFile(tmp, data, 0o600); err == nil {
			_ = os.Rename(tmp, s.file)
		}
	}
	// 同步落库 trip_settings(按 key 幂等覆盖),数据库优先于磁盘文件
	if s.db != nil {
		for k, v := range persist {
			row := model.TripSetting{Key: k, Value: v, UpdatedAt: time.Now()}
			if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
				break
			}
		}
	}
	s.mu.Unlock()
	return s.Masked()
}

// maskSecret 掩码机密值;多行配置(如多个小红书 Cookie,每行一个)逐行掩码,
// 便于前端确认已配置了几条,而不是糊成一团。
func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, "\n") {
		lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
		for i, line := range lines {
			lines[i] = maskSecretLine(line)
		}
		return strings.Join(lines, "\n")
	}
	return maskSecretLine(value)
}

func maskSecretLine(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return strings.Repeat(tripMaskChar, 8)
	}
	return value[:4] + strings.Repeat(tripMaskChar, 8) + value[len(value)-4:]
}
