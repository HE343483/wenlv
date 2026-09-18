package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============ 小红书 Cookie 池 ============
// 支持在设置页配置多个 Cookie(每行一个),请求时轮换使用;
// 某个 Cookie 被风控/失效后进入冷却期并自动切换到下一个,
// 避免单 Cookie 失效导致整个景点推荐与搜图功能不可用。

const (
	// xhsCookieCooldown 被风控的 Cookie 冷却时长(冷却期内不再选用)
	xhsCookieCooldown = 30 * time.Minute
)

// xhsCookiePool 小红书 Cookie 轮换池。
type xhsCookiePool struct {
	mu       sync.Mutex
	cooldown map[string]time.Time // cookie → 冷却到期时间
	next     int                  // 下一个待选下标
}

func newXHSCookiePool() *xhsCookiePool {
	return &xhsCookiePool{cooldown: map[string]time.Time{}}
}

// splitXHSCookies 把配置值拆成 Cookie 列表(每行一个,忽略空行)。
func splitXHSCookies(raw string) []string {
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		if normalized := NormalizeXHSCookie(line); normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}

// replaceXHSCookieAt 用新值替换多行配置中的第 idx 个非空行(其余行保持不变)。
// 用于 Cookie 保活续期时只回写发生变化的那一条。
func replaceXHSCookieAt(raw string, idx int, value string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	seen := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		seen++
		if seen == idx {
			lines[i] = value
			break
		}
	}
	return strings.Join(lines, "\n")
}

// pick 轮换取用可用的 Cookie:优先返回未冷却的;若全部处于冷却期,
// 则返回最早解除冷却的那一个,保证仍有请求机会(而不是直接判定不可用)。
func (p *xhsCookiePool) pick(cookies []string) string {
	if len(cookies) == 0 {
		return ""
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for i := 0; i < len(cookies); i++ {
		idx := (p.next + i) % len(cookies)
		cookie := cookies[idx]
		if until, cooling := p.cooldown[cookie]; !cooling || now.After(until) {
			delete(p.cooldown, cookie)
			p.next = (idx + 1) % len(cookies)
			return cookie
		}
	}

	// 全部冷却中:选最早解除冷却的一个
	bestIdx, bestUntil := p.next%len(cookies), time.Time{}
	for i := 0; i < len(cookies); i++ {
		idx := (p.next + i) % len(cookies)
		until := p.cooldown[cookies[idx]]
		if bestUntil.IsZero() || until.Before(bestUntil) {
			bestIdx, bestUntil = idx, until
		}
	}
	p.next = (bestIdx + 1) % len(cookies)
	return cookies[bestIdx]
}

// markBad 把 Cookie 标记为冷却(风控/Cookie 失效)。
func (p *xhsCookiePool) markBad(cookie string) {
	if cookie == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cooldown[cookie] = time.Now().Add(xhsCookieCooldown)
}

// markGood 清除 Cookie 的冷却标记(请求成功)。
func (p *xhsCookiePool) markGood(cookie string) {
	if cookie == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.cooldown, cookie)
}

// cookies 返回当前配置的全部 Cookie。
func (s *XHSService) cookies() ([]string, error) {
	list := splitXHSCookies(s.settings.Snapshot().XHSCookie)
	if len(list) == 0 {
		return nil, &XHSNotConfiguredError{Msg: "小红书 Cookie 未配置,请先在前端设置页完成配置"}
	}
	return list, nil
}

// withXHSCookie 依次取用 Cookie 池中的 Cookie 执行 fn:
// 当前 Cookie 被风控时冷却并自动换下一个重试;非风控类错误直接返回,避免无意义重试放大请求量。
func withXHSCookie[T any](s *XHSService, ctx context.Context, fn func(cookie string) (T, error)) (T, error) {
	var zero T
	cookies, err := s.cookies()
	if err != nil {
		return zero, err
	}

	tried := make(map[string]bool, len(cookies))
	var lastErr error
	for attempt := 0; attempt < len(cookies); attempt++ {
		cookie := s.pool.pick(cookies)
		if cookie == "" || tried[cookie] {
			break
		}
		tried[cookie] = true

		result, err := fn(cookie)
		if err == nil {
			s.pool.markGood(cookie)
			return result, nil
		}
		lastErr = err
		if !isXHSCookieExpired(err) {
			return zero, err
		}
		// 风控/Cookie 失效:冷却当前 Cookie,换下一个重试
		s.pool.markBad(cookie)
		fmt.Printf("⚠️  [Cookie轮换] 当前小红书 Cookie 不可用,已冷却 %v(共 %d 个,尝试第 %d 个)\n",
			xhsCookieCooldown, len(cookies), attempt+1)
	}

	if lastErr == nil {
		lastErr = &XHSFetchError{Msg: "小红书 Cookie 均不可用,请更换后重试"}
	}
	return zero, lastErr
}

// isXHSCookieExpired 判断错误链中是否包含小红书 Cookie 过期/风控异常。
func isXHSCookieExpired(err error) bool {
	var target *XHSCookieExpiredError
	return asCookieExpired(err, &target)
}
