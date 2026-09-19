package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestLimiter(cfg RateLimitConfig) *RateLimiter {
	return NewRateLimiter(cfg, nil)
}

func TestAllowSlidingWindow(t *testing.T) {
	rl := newTestLimiter(RateLimitConfig{})
	rule := limitRule{window: 50 * time.Millisecond, max: 2}
	for i := 0; i < 2; i++ {
		if !rl.allow("test", "k1", rule) {
			t.Fatalf("第 %d 次请求不应被限流", i+1)
		}
	}
	if rl.allow("test", "k1", rule) {
		t.Fatal("第 3 次请求应被限流")
	}
	// 不同身份键互不影响
	if !rl.allow("test", "k2", rule) {
		t.Fatal("不同身份键不应被限流")
	}
	// 窗口滑过后恢复
	time.Sleep(60 * time.Millisecond)
	if !rl.allow("test", "k1", rule) {
		t.Fatal("窗口滑过后应恢复放行")
	}
}

func TestAllowMultipleRules(t *testing.T) {
	rl := newTestLimiter(RateLimitConfig{})
	short := limitRule{window: time.Hour, max: 3}
	long := limitRule{window: 24 * time.Hour, max: 5}
	// 3 次后短窗口拒绝,即使长窗口未满
	for i := 0; i < 3; i++ {
		if !rl.allow("plan", "u1", short, long) {
			t.Fatalf("第 %d 次请求不应被限流", i+1)
		}
	}
	if rl.allow("plan", "u1", short, long) {
		t.Fatal("短窗口超限后应被限流")
	}
}

func TestPlanLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := newTestLimiter(RateLimitConfig{PlanPer10Min: 2})
	router := gin.New()
	router.POST("/api/trip/plan", rl.PlanLimit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	do := func() *httptest.ResponseRecorder {
		body := `{"user_id":"u1","city":"成都"}`
		req := httptest.NewRequest(http.MethodPost, "/api/trip/plan", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	for i := 0; i < 2; i++ {
		if w := do(); w.Code != http.StatusOK {
			t.Fatalf("第 %d 次请求应放行,实际 %d", i+1, w.Code)
		}
	}
	w := do()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("第 3 次请求应返回 429,实际 %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("429 响应应为 JSON: %v", err)
	}
	if resp["detail"] == "" {
		t.Fatal("429 响应应包含 detail 提示信息")
	}
}

func TestIdentityKeyFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Token 无效时应降级到 body user_id
	rl := NewRateLimiter(RateLimitConfig{}, func(ctx context.Context, token string) (uint, error) {
		return 0, context.DeadlineExceeded
	})
	req := httptest.NewRequest(http.MethodPost, "/api/trip/plan", strings.NewReader(`{"user_id":"anon-abc"}`))
	req.Header.Set("Authorization", "Bearer bad-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if got := rl.identityKey(c); got != "anon-anon-abc" {
		t.Fatalf("无效 Token 应回退到 body user_id,实际 %s", got)
	}

	// 无 user_id 时回退 IP
	req2 := httptest.NewRequest(http.MethodGet, "/api/map/weather?x=1", nil)
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = req2
	if got := rl.identityKey(c2); !strings.HasPrefix(got, "ip-") {
		t.Fatalf("无 user_id 应回退到 IP,实际 %s", got)
	}
}

func TestPlanConcurrencySlots(t *testing.T) {
	rl := newTestLimiter(RateLimitConfig{PlanMaxConcurrent: 2})
	if !rl.AcquirePlanSlot() || !rl.AcquirePlanSlot() {
		t.Fatal("前两次应能获取并发槽位")
	}
	if rl.AcquirePlanSlot() {
		t.Fatal("超过并发上限应获取失败")
	}
	rl.ReleasePlanSlot()
	if !rl.AcquirePlanSlot() {
		t.Fatal("释放后应能重新获取槽位")
	}
}
