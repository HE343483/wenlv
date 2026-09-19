package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// httpStatusError 表示下载返回了非 200 状态码,便于上层按状态码(如 429)做退避重试。
type httpStatusError struct{ Code int }

func (e *httpStatusError) Error() string { return fmt.Sprintf("HTTP %d", e.Code) }

// isRateLimitedError 判断错误是否为 HTTP 429(Too Many Requests)。
func isRateLimitedError(err error) bool {
	var se *httpStatusError
	return errors.As(err, &se) && se.Code == http.StatusTooManyRequests
}

// fetchBytes 下载外部图片,返回内容与 Content-Type。
// 限制 20MB,超时 20s;高德图片 CDN 可能要求 Referer,这里统一带普通 UA。
func fetchBytes(ctx context.Context, rawURL string) ([]byte, string, error) {
	return fetchBytesWithClient(ctx, http.DefaultClient, rawURL)
}

// fetchBytesWithClient 用指定客户端下载图片,供需要走代理的来源(如 Commons)复用。
func fetchBytesWithClient(ctx context.Context, client *http.Client, rawURL string) ([]byte, string, error) {
	if rawURL == "" {
		return nil, "", fmt.Errorf("空 URL")
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; wenlv-enricher/1.0)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", &httpStatusError{Code: resp.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return nil, "", err
	}
	if len(body) == 0 {
		return nil, "", fmt.Errorf("空响应体")
	}
	return body, resp.Header.Get("Content-Type"), nil
}
