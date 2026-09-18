package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ============ 抖音 a_bogus 签名(JS 引擎桥接) ============
// 抖音网页搜索/详情接口强制要求 a_bogus 签名(由混淆 JSVMP JS 生成,作用同小红书的 x-s),
// 这里以与小红书签名一致的方式,调用本地 Node.js 执行 douyin_sign/sign_runner.js 生成签名。
// 真正的签名实现需自行放置到 douyin_sign/douyin.js,接口约定见 sign_runner.js 头部注释。

const douyinSignRunner = "sign_runner.js"

// DouyinSigner 抖音签名器。
type DouyinSigner struct {
	dir string

	mu        sync.Mutex
	available bool
	lastErr   string
}

// NewDouyinSigner 构造抖音签名器。
func NewDouyinSigner(dir string) *DouyinSigner {
	if dir == "" {
		dir = "douyin_sign"
	}
	return &DouyinSigner{dir: dir}
}

// Available 检查 node 与签名脚本是否就绪。
func (s *DouyinSigner) Available() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.available {
		return true
	}
	if _, err := exec.LookPath("node"); err != nil {
		s.lastErr = "未检测到 Node.js 运行时,抖音签名引擎不可用"
		return false
	}
	s.available = true
	return true
}

type douyinSignRequest struct {
	Mode      string `json:"mode"`
	URL       string `json:"url,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
}

type douyinSignResponse struct {
	ABogus string `json:"a_bogus"`
	Error  string `json:"error"`
}

// SignABogus 为「路径 + 查询串」生成 a_bogus 签名。
func (s *DouyinSigner) SignABogus(ctx context.Context, urlWithQuery, userAgent string) (string, error) {
	if !s.Available() {
		return "", fmt.Errorf("%s", s.lastErr)
	}
	body, err := json.Marshal(douyinSignRequest{Mode: "sign", URL: urlWithQuery, UserAgent: userAgent})
	if err != nil {
		return "", err
	}
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "node", filepath.Join(s.dir, douyinSignRunner))
	cmd.Stdin = bytes.NewReader(body)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("抖音签名生成失败: %s", truncateText(msg, 300))
	}
	out, ok := extractDouyinSignResult(stdout.String())
	if !ok {
		return "", fmt.Errorf("抖音签名结果解析失败: %s", truncateText(stdout.String(), 200))
	}
	if out.Error != "" {
		return "", fmt.Errorf("抖音签名引擎报错: %s", out.Error)
	}
	if out.ABogus == "" {
		return "", fmt.Errorf("抖音签名结果为空")
	}
	return out.ABogus, nil
}

// extractDouyinSignResult 从 node stdout 中提取最后一行合法 JSON。
func extractDouyinSignResult(stdout string) (*douyinSignResponse, bool) {
	lines := strings.Split(stdout, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var out douyinSignResponse
		if err := json.Unmarshal([]byte(line), &out); err == nil {
			return &out, true
		}
	}
	return nil, false
}
