// Package config 负责从环境变量读取应用配置。
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 汇总整个应用所需的运行配置。
type Config struct {
	Server struct {
		Port string
	}
	JWT struct {
		AccessSecret  string
		RefreshSecret string
		AccessTTL     int // 秒
		RefreshTTL    int // 秒
		Issuer        string
	}
	MySQL struct {
		Host, Port, User, Password, DBName string
	}
	Redis struct {
		Addr     string
		Username string
		Password string
		DB       int
	}
	OSS struct {
		Endpoint, AccessKey, SecretKey, Bucket string
	}
	// Trip 为「AI 行程规划」模块(移植自 TripStar)的默认配置,
	// 其中敏感项可在前端设置页热更新,并持久化到 Trip.DataDir/runtime_settings.json。
	Trip struct {
		LLMAPIKey        string
		LLMBaseURL       string
		LLMModel         string
		LLMTimeout       int
		PlannerTimeout   int
		AmapWebKey       string
		AmapWebJSKey     string
		GoogleMapsAPIKey string
		GoogleMapsProxy  string
		XHSCookie        string
		DouyinCookie     string
		EnableUserMemory bool
		DataDir          string
		// 景点图片磁盘缓存的有效期(小时)与容量上限(MB)
		ImageCacheTTLHours int
		ImageCacheMaxMB    int
		// ScenicCacheTTLHours 景点详情页实时数据(周边推荐/交通站点)的 Redis 缓存小时数。
		ScenicCacheTTLHours int

		// 用户偏好记忆模块参数(对应原项目 MEMORY_* 环境变量)
		MemoryDecayFactor      float64
		MemoryWeightThreshold  float64
		MemoryMaxRecall        int
		MemoryMinInitWeight    float64
		MemoryMaxSingleContent int
	}
}

// Load 加载配置:优先读取工区目录 .env 文件,未命中时回退到系统环境变量。
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件,将使用系统环境变量")
	}

	c := &Config{}
	c.Server.Port = getEnv("SERVER_PORT", "8080")

	c.JWT.AccessSecret = getEnv("JWT_ACCESS_SECRET", "dev-access-secret")
	c.JWT.RefreshSecret = getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret")
	c.JWT.AccessTTL = getEnvInt("JWT_ACCESS_TTL", 900)      // 15 分钟
	c.JWT.RefreshTTL = getEnvInt("JWT_REFRESH_TTL", 604800) // 7 天
	c.JWT.Issuer = getEnv("JWT_ISSUER", "wenlv-backend")

	c.MySQL.Host = getEnv("DB_HOST", "127.0.0.1")
	c.MySQL.Port = getEnv("DB_PORT", "3306")
	c.MySQL.User = getEnv("DB_USER", "root")
	c.MySQL.Password = getEnv("DB_PASSWORD", "")
	c.MySQL.DBName = getEnv("DB_NAME", "wenlv")

	c.Redis.Addr = getEnv("REDIS_ADDR", "127.0.0.1:6379")
	c.Redis.Username = getEnv("REDIS_USERNAME", "")
	c.Redis.Password = getEnv("REDIS_PASSWORD", "")
	c.Redis.DB = getEnvInt("REDIS_DB", 0)

	c.OSS.Endpoint = getEnv("OSS_ENDPOINT", "")
	c.OSS.AccessKey = getEnv("OSS_ACCESS_KEY", "")
	c.OSS.SecretKey = getEnv("OSS_SECRET_KEY", "")
	c.OSS.Bucket = getEnv("OSS_BUCKET", "")

	// ===== AI 行程规划模块 =====
	c.Trip.LLMAPIKey = getEnv("LLM_API_KEY", "")
	c.Trip.LLMBaseURL = getEnv("LLM_BASE_URL", "https://api.openai.com/v1")
	c.Trip.LLMModel = getEnv("LLM_MODEL_ID", "gpt-4")
	c.Trip.LLMTimeout = getEnvInt("LLM_TIMEOUT", 120)
	c.Trip.PlannerTimeout = getEnvInt("TRIP_PLANNER_TIMEOUT", 180)
	// 高德 Key 与旧版前端 .env 命名保持一致,便于直接迁移
	c.Trip.AmapWebKey = getEnv("AMAP_WEB_KEY", getEnv("VITE_AMAP_WEB_KEY", ""))
	c.Trip.AmapWebJSKey = getEnv("AMAP_WEB_JS_KEY", getEnv("VITE_AMAP_WEB_JS_KEY", ""))
	c.Trip.GoogleMapsAPIKey = getEnv("GOOGLE_MAPS_API_KEY", "")
	c.Trip.GoogleMapsProxy = getEnv("GOOGLE_MAPS_PROXY", "")
	c.Trip.XHSCookie = getEnv("XHS_COOKIE", "")
	c.Trip.DouyinCookie = getEnv("DOUYIN_COOKIE", "")
	c.Trip.EnableUserMemory = getEnv("ENABLE_USER_MEMORY", "false") == "true" ||
		getEnv("ENABLE_USER_MEMORY", "false") == "1"
	c.Trip.DataDir = getEnv("TRIP_DATA_DIR", "data")
	// 图片缓存存的是图片字节而非时效直链,TTL 放宽可显著降低上游搜图频率(风控主因)
	c.Trip.ImageCacheTTLHours = getEnvInt("TRIP_IMAGE_CACHE_TTL_HOURS", 168) // 7 天
	c.Trip.ImageCacheMaxMB = getEnvInt("TRIP_IMAGE_CACHE_MAX_MB", 512)
	c.Trip.ScenicCacheTTLHours = getEnvInt("SCENIC_CACHE_TTL_HOURS", 24)

	// 用户偏好记忆参数(与原项目 MEMORY_* 环境变量一一对应)
	c.Trip.MemoryDecayFactor = getEnvFloat("MEMORY_DECAY_FACTOR", 0.97)
	c.Trip.MemoryWeightThreshold = getEnvFloat("MEMORY_THRESHOLD", 2.0)
	c.Trip.MemoryMaxRecall = getEnvInt("MEMORY_MAX_RECALL", 10)
	c.Trip.MemoryMinInitWeight = getEnvFloat("MIN_INIT_WEIGHT", 4.0)
	c.Trip.MemoryMaxSingleContent = getEnvInt("MEMORY_MAX_SINGLE_CONTENT", 120)

	return c
}

// DSN 返回 MySQL 连接串。
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQL.User, c.MySQL.Password, c.MySQL.Host, c.MySQL.Port, c.MySQL.DBName)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n := 0
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		log.Printf("配置项 %s 解析失败,使用默认值 %d: %v", key, def, err)
		return def
	}
	return n
}

func getEnvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f := 0.0
	if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
		log.Printf("配置项 %s 解析失败,使用默认值 %v: %v", key, def, err)
		return def
	}
	return f
}
