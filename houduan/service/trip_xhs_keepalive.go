package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============ 小红书 Cookie 保活 ============
// 定时携带当前 Cookie 访问小红书首页(模拟活跃用户),
// 把响应下发的 Set-Cookie 合并回运行时配置并持久化,从而延长 Cookie 有效期;
// 遭遇风控(461)时仅告警不改动 Cookie,由业务请求的降级逻辑兜底。

const (
	xhsKeepaliveInterval = 24 * time.Hour // 保活周期
	xhsKeepaliveDelay    = 2 * time.Minute
	xhsHomeURL           = "https://www.xiaohongshu.com/explore"
)

type XHSKeepalive struct {
	settings *TripSettings
	client   *http.Client

	startOnce sync.Once
}

// NewXHSKeepalive 构造 Cookie 保活服务。
func NewXHSKeepalive(settings *TripSettings) *XHSKeepalive {
	return &XHSKeepalive{
		settings: settings,
		client: &http.Client{
			Timeout:   20 * time.Second,
			Transport: &http.Transport{Proxy: nil},
		},
	}
}

// Start 启动保活后台任务(仅执行一次)。
func (k *XHSKeepalive) Start() {
	k.startOnce.Do(func() {
		go func() {
			time.Sleep(xhsKeepaliveDelay)
			k.runOnce()
			ticker := time.NewTicker(xhsKeepaliveInterval)
			defer ticker.Stop()
			for range ticker.C {
				k.runOnce()
			}
		}()
	})
}

func (k *XHSKeepalive) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := k.refresh(ctx); err != nil {
		fmt.Printf("🔄 [Cookie保活] %v\n", err)
	}
}

// refresh 访问小红书首页并合并 Set-Cookie。
func (k *XHSKeepalive) refresh(ctx context.Context) error {
	cookie := NormalizeXHSCookie(k.settings.Snapshot().XHSCookie)
	if cookie == "" {
		// 未配置 Cookie,保活无意义,静默跳过
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, xhsHomeURL, nil)
	if err != nil {
		return fmt.Errorf("保活请求构造失败: %w", err)
	}
	headers := map[string]string{
		"accept":                  "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"accept-language":         "zh-CN,zh;q=0.9,en;q=0.8",
		"cache-control":           "no-cache",
		"pragma":                  "no-cache",
		"sec-ch-ua":               `"Not A(Brand";v="99", "Microsoft Edge";v="121", "Chromium";v="121"`,
		"sec-ch-ua-mobile":        "?0",
		"sec-ch-ua-platform":      `"Windows"`,
		"sec-fetch-dest":          "document",
		"sec-fetch-mode":          "navigate",
		"sec-fetch-site":          "none",
		"sec-fetch-user":          "?1",
		"upgrade-insecure-requests": "1",
		"user-agent":              xhsUserAgent,
	}
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	req.Header.Set("Cookie", cookie)

	resp, err := k.client.Do(req)
	if err != nil {
		return fmt.Errorf("保活请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 461 {
		return fmt.Errorf("保活请求遭遇风控 (status=%d),保留原 Cookie 并跳过本次更新", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("保活请求异常 (status=%d),保留原 Cookie", resp.StatusCode)
	}

	// 合并服务端下发的 Set-Cookie(新值覆盖同名旧值,新增字段追加)
	updated := mergeSetCookie(cookie, resp.Cookies())
	if updated == "" {
		return nil
	}
	if updated == cookie {
		return nil
	}
	k.settings.Update(map[string]string{"xhs_cookie": updated})
	fmt.Printf("🔄 [Cookie保活] 已从小红书响应续期 Cookie 并持久化 (旧长度=%d 新长度=%d)\n", len(cookie), len(updated))
	return nil
}

// mergeSetCookie 把响应的 Set-Cookie 合并进原 Cookie 串,保持原有字段顺序、新字段追加在尾部。
func mergeSetCookie(orig string, setCookies []*http.Cookie) string {
	if len(setCookies) == 0 {
		return orig
	}
	// http.Cookie 的 Value 已经过解码,直接透传可能破坏原有转义;
	// 与浏览器行为一致,使用原值即可(小红书 Cookie 字段值不含分号/逗号)。
	values := map[string]string{}
	for _, c := range setCookies {
		if c == nil || strings.TrimSpace(c.Name) == "" {
			continue
		}
		values[c.Name] = c.Value
	}
	if len(values) == 0 {
		return orig
	}

	pairs := strings.Split(orig, "; ")
	seen := map[string]bool{}
	out := make([]string, 0, len(pairs)+len(values))
	for _, pair := range pairs {
		name := strings.TrimSpace(strings.SplitN(pair, "=", 2)[0])
		if name == "" {
			continue
		}
		if v, ok := values[name]; ok {
			pair = name + "=" + v
			seen[name] = true
		}
		out = append(out, pair)
	}
	for _, c := range setCookies {
		if c == nil || seen[c.Name] {
			continue
		}
		out = append(out, c.Name+"="+c.Value)
	}
	return strings.Join(out, "; ")
}
