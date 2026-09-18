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

// ============ 小红书请求签名(JS 引擎桥接) ============
// 原项目通过 PyExecJS 调用本地 Node.js 执行签名 JS,
// 这里以同样的方式在 Go 中调用 node 子进程生成 x-s / x-t / x-s-common / x-xray-traceid。
// JS 文件位于 xhs_sign 目录(由 config 指定,默认 ./xhs_sign)。

const xhsSignRunner = "sign_runner.js"

// XHSSigner 签名器。
type XHSSigner struct {
	dir string

	mu        sync.Mutex
	available bool
	lastErr   string
}

// NewXHSSigner 构造签名器。
func NewXHSSigner(dir string) *XHSSigner {
	if dir == "" {
		dir = "xhs_sign"
	}
	return &XHSSigner{dir: dir}
}

// Available 检查 node 与签名脚本是否就绪。
func (s *XHSSigner) Available() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.available {
		return true
	}
	if _, err := exec.LookPath("node"); err != nil {
		s.lastErr = "未检测到 Node.js 运行时,小红书签名引擎不可用"
		return false
	}
	s.available = true
	return true
}

type signRequest struct {
	Mode   string `json:"mode"`
	API    string `json:"api,omitempty"`
	Data   string `json:"data,omitempty"`
	A1     string `json:"a1,omitempty"`
	Method string `json:"method,omitempty"`
}

type signResponse struct {
	XS       string `json:"xs"`
	XT       int64  `json:"xt"`
	XSCommon string `json:"xs_common"`
	TraceID  string `json:"traceId"`
	Error    string `json:"error"`
}

func (s *XHSSigner) run(ctx context.Context, payload signRequest) (*signResponse, error) {
	if !s.Available() {
		return nil, fmt.Errorf("%s", s.lastErr)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "node", filepath.Join(s.dir, xhsSignRunner))
	cmd.Stdin = bytes.NewReader(body)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("小红书签名生成失败: %s", truncateText(msg, 300))
	}
	out, ok := extractSignResult(stdout.String())
	if !ok {
		return nil, fmt.Errorf("小红书签名结果解析失败: %s", truncateText(stdout.String(), 200))
	}
	if out.Error != "" {
		return nil, fmt.Errorf("小红书签名引擎报错: %s", out.Error)
	}
	return out, nil
}

// extractSignResult 从 node stdout 中提取最后一行合法 JSON。
// 签名 JS(webpack 打包产物)会向 stdout 打印模块加载日志,需要跳过噪声行。
func extractSignResult(stdout string) (*signResponse, bool) {
	lines := strings.Split(stdout, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var out signResponse
		if err := json.Unmarshal([]byte(line), &out); err == nil {
			return &out, true
		}
	}
	return nil, false
}

// Sign 生成 x-s / x-t / x-s-common。
func (s *XHSSigner) Sign(ctx context.Context, api string, data any, a1, method string) (string, int64, string, error) {
	dataStr := ""
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			return "", 0, "", err
		}
		dataStr = string(raw)
	}
	resp, err := s.run(ctx, signRequest{Mode: "sign", API: api, Data: dataStr, A1: a1, Method: method})
	if err != nil {
		return "", 0, "", err
	}
	return resp.XS, resp.XT, resp.XSCommon, nil
}

// TraceID 生成 x-xray-traceid。
func (s *XHSSigner) TraceID(ctx context.Context) (string, error) {
	resp, err := s.run(ctx, signRequest{Mode: "traceid"})
	if err != nil {
		return "", err
	}
	return resp.TraceID, nil
}