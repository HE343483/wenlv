// Package middleware 提供 Gin 中间件。
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 面向用户的限流参数(全部可通过 .env 覆盖,<=0 表示关闭该项限制)。
type RateLimitConfig struct {
	PlanPer10Min      int // LLM 行程生成:每 10 分钟每用户次数
	PlanPerDay        int // LLM 行程生成:每自然滑动 24h 每用户次数
	PlanMaxConcurrent int // LLM 行程生成:全局并发任务上限(长任务,防额度打爆)
	ChatPerMin        int // LLM 问答:每分钟每用户次数
	MapPerMin         int // 地图/POI 查询:每分钟每用户次数
	ImagePerMin       int // 小红书/抖音搜图代理:每分钟每用户次数
}

// limitRule 单条滑动窗口规则。
type limitRule struct {
	window time.Duration
	max    int
}

// RateLimiter 进程内滑动窗口限流器(单实例部署,无分布式需求)。
// 按"身份键"(登录用户 > 请求 user_id > 客户端 IP)分别计数,
// 防止大量用户耗尽共享的 LLM/高德额度或触发小红书风控。
type RateLimiter struct {
	cfg      RateLimitConfig
	validate func(ctx context.Context, token string) (uint, error)

	mu      sync.Mutex
	windows map[string]map[string][]time.Time // 类别 -> 身份键 -> 请求时间戳

	planConcurrent atomic.Int64
}

// NewRateLimiter 构造限流器,并启动后台清理协程防止内存泄漏。
func NewRateLimiter(cfg RateLimitConfig, validate func(ctx context.Context, token string) (uint, error)) *RateLimiter {
	r := &RateLimiter{
		cfg:      cfg,
		validate: validate,
		windows:  map[string]map[string][]time.Time{},
	}
	go r.cleanupLoop()
	return r
}

// PlanLimit 行程生成接口限流(频率限制;全局并发槽位由 handler 在任务生命周期内持有)。
func (r *RateLimiter) PlanLimit() gin.HandlerFunc {
	rules := r.planRules()
	return func(c *gin.Context) {
		if len(rules) == 0 {
			c.Next()
			return
		}
		if !r.allow("plan", r.identityKey(c), rules...) {
			tooManyRequests(c, "行程生成请求过于频繁，请稍后再试")
			return
		}
		c.Next()
	}
}

// ChatLimit LLM 问答接口限流。
func (r *RateLimiter) ChatLimit() gin.HandlerFunc {
	return r.minuteLimit("chat", r.cfg.ChatPerMin, "AI 问答请求过于频繁，请稍后再试")
}

// MapLimit 地图/POI 查询接口限流。
func (r *RateLimiter) MapLimit() gin.HandlerFunc {
	return r.minuteLimit("map", r.cfg.MapPerMin, "地图查询请求过于频繁，请稍后再试")
}

// ImageLimit 小红书/抖音搜图代理接口限流。
func (r *RateLimiter) ImageLimit() gin.HandlerFunc {
	return r.minuteLimit("image", r.cfg.ImagePerMin, "图片获取请求过于频繁，请稍后再试")
}

// AcquirePlanSlot 获取全局行程生成并发槽位(获取失败返回 false)。
func (r *RateLimiter) AcquirePlanSlot() bool {
	if r.cfg.PlanMaxConcurrent <= 0 {
		return true
	}
	for {
		cur := r.planConcurrent.Load()
		if cur >= int64(r.cfg.PlanMaxConcurrent) {
			return false
		}
		if r.planConcurrent.CompareAndSwap(cur, cur+1) {
			return true
		}
	}
}

// ReleasePlanSlot 释放全局行程生成并发槽位(任务结束时必须调用)。
func (r *RateLimiter) ReleasePlanSlot() {
	if r.cfg.PlanMaxConcurrent <= 0 {
		return
	}
	r.planConcurrent.Add(-1)
}

func (r *RateLimiter) planRules() []limitRule {
	rules := []limitRule{}
	if r.cfg.PlanPer10Min > 0 {
		rules = append(rules, limitRule{window: 10 * time.Minute, max: r.cfg.PlanPer10Min})
	}
	if r.cfg.PlanPerDay > 0 {
		rules = append(rules, limitRule{window: 24 * time.Hour, max: r.cfg.PlanPerDay})
	}
	return rules
}

func (r *RateLimiter) minuteLimit(category string, perMin int, msg string) gin.HandlerFunc {
	if perMin <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	rule := limitRule{window: time.Minute, max: perMin}
	return func(c *gin.Context) {
		if !r.allow(category, r.identityKey(c), rule) {
			tooManyRequests(c, msg)
			return
		}
		c.Next()
	}
}

// allow 滑动窗口计数:全部规则通过才放行,并记录本次请求时间戳。
func (r *RateLimiter) allow(category, key string, rules ...limitRule) bool {
	now := time.Now()
	var maxWindow time.Duration
	for _, rule := range rules {
		if rule.window > maxWindow {
			maxWindow = rule.window
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	bucket, ok := r.windows[category]
	if !ok {
		bucket = map[string][]time.Time{}
		r.windows[category] = bucket
	}
	times := bucket[key]
	kept := times[:0]
	for _, t := range times {
		if now.Sub(t) < maxWindow {
			kept = append(kept, t)
		}
	}
	times = kept
	for _, rule := range rules {
		count := 0
		for _, t := range times {
			if now.Sub(t) < rule.window {
				count++
			}
		}
		if count >= rule.max {
			bucket[key] = times
			return false
		}
	}
	bucket[key] = append(times, now)
	return true
}

// identityKey 身份键解析:登录用户 > 请求 user_id(query 或 JSON body) > 客户端 IP。
// Token 校验失败不拦截,降级到下一优先级。
func (r *RateLimiter) identityKey(c *gin.Context) string {
	if header := c.GetHeader("Authorization"); r.validate != nil && strings.HasPrefix(header, "Bearer ") {
		if uid, err := r.validate(c.Request.Context(), strings.TrimPrefix(header, "Bearer ")); err == nil && uid > 0 {
			return fmt.Sprintf("user-%d", uid)
		}
	}
	if uid := strings.TrimSpace(c.Query("user_id")); uid != "" {
		return "anon-" + uid
	}
	if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
		if body, ok := peekBody(c); ok {
			var payload struct {
				UserID string `json:"user_id"`
			}
			if json.Unmarshal(body, &payload) == nil && payload.UserID != "" {
				return "anon-" + payload.UserID
			}
		}
	}
	return "ip-" + c.ClientIP()
}

// peekBody 读取请求体供解析 user_id,并原样回填保证后续 handler 可再次读取。
func peekBody(c *gin.Context) ([]byte, bool) {
	if c.Request.Body == nil {
		return nil, false
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, false
	}
	_ = c.Request.Body.Close()
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, true
}

// cleanupLoop 定期清理过期时间戳与空键,防止长期运行内存膨胀。
func (r *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		r.mu.Lock()
		for category, bucket := range r.windows {
			for key, times := range bucket {
				kept := times[:0]
				for _, t := range times {
					// 24h 覆盖当前最长的限流窗口(PlanPerDay)
					if now.Sub(t) < 24*time.Hour {
						kept = append(kept, t)
					}
				}
				if len(kept) == 0 {
					delete(bucket, key)
				} else {
					bucket[key] = kept
				}
			}
			if len(bucket) == 0 {
				delete(r.windows, category)
			}
		}
		r.mu.Unlock()
	}
}

func tooManyRequests(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"detail":  msg,
		"message": msg,
	})
}
