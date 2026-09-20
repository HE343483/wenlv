package pkg

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OssConfig OSS 连接配置。
type OssConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	// PublicDomain 可选:公网访问域名(如中转为 https://img.example.com),用于拼接图片 URL。
	PublicDomain string
}

// OssSigner 负责生成浏览器直传所需的 POST 签名 Policy。
type OssSigner struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	CdnHost   string
}

// normalizeEndpoint 去掉 Endpoint 可能带的协议前缀与尾斜杠,
// 统一按虚拟主机风格 <bucket>.<endpoint> 拼接,保证请求 URL 与返回 URL 一致。
func normalizeEndpoint(host string) string {
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	return strings.TrimSuffix(host, "/")
}

// NewOssSigner 构造签名器。
func NewOssSigner(cfg *OssConfig) *OssSigner {
	return &OssSigner{
		Endpoint:  normalizeEndpoint(cfg.Endpoint),
		Bucket:    cfg.Bucket,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		CdnHost:   cfg.PublicDomain,
	}
}

// UploadPolicy 返回给前端完成 POST 直传所需的字段。
type UploadPolicy struct {
	AccessKeyID string `json:"access_key_id"`
	Policy      string `json:"policy"`
	Signature   string `json:"signature"`
	Key         string `json:"key"`
	Host        string `json:"host"`
	Callback    string `json:"callback,omitempty"`
}

// signPolicy 用 HMAC-SHA1 对 base64 policy 签名。
func (s *OssSigner) signPolicy(policy string) string {
	mac := hmac.New(sha1.New, []byte(s.SecretKey))
	mac.Write([]byte(policy))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// GeneratePolicy 生成指定过期时间的上传签名。
// key 为 OSS 存储路径(如 uploads/1/1720000000_abc.jpg)。
func (s *OssSigner) GeneratePolicy(key string, expireMinutes int, callbacks map[string]string) (*UploadPolicy, error) {
	expire := time.Now().Add(time.Duration(expireMinutes) * time.Minute).Unix()

	cond := map[string]any{
		"expiration": time.Unix(expire, 0).UTC().Format("2006-01-02T15:04:05.000Z"),
		"conditions": []any{
			[]any{"starts-with", "$key", key},
			[]any{"content-length-range", 0, 20 << 20}, // 上限 20MB
		},
	}
	if len(callbacks) > 0 {
		// 可选服务端回调(需配置回调地址),非必填
		cbJSON, _ := json.Marshal(callbacks)
		cond["conditions"] = append(cond["conditions"].([]any), []any{"content-type", callbacks["contentType"]})
		_ = cbJSON
	}

	payload, err := json.Marshal(cond)
	if err != nil {
		return nil, err
	}
	policy := base64.StdEncoding.EncodeToString(payload)

	host := fmt.Sprintf("https://%s.%s", s.Bucket, s.Endpoint)
	if s.CdnHost != "" {
		host = s.CdnHost
	}

	return &UploadPolicy{
		AccessKeyID: s.AccessKey,
		Policy:      policy,
		Signature:   s.signPolicy(policy),
		Key:         key,
		Host:        host,
	}, nil
}

// ResolveURL 将存储 key 解析为可访问 URL。
func (s *OssSigner) ResolveURL(key string) string {
	if key == "" {
		return ""
	}
	base := s.CdnHost
	if base == "" {
		base = fmt.Sprintf("https://%s.%s", s.Bucket, s.Endpoint)
	}
	return fmt.Sprintf("%s/%s", base, key)
}

// Configured OSS 配置是否完整(四项必填)。
func (s *OssSigner) Configured() bool {
	return s.Endpoint != "" && s.AccessKey != "" && s.SecretKey != "" && s.Bucket != ""
}

// endpointHost 返回归一化后的 Endpoint(与 NewOssSigner 共用同一套归一化逻辑)。
func (s *OssSigner) endpointHost() string {
	return normalizeEndpoint(s.Endpoint)
}

// PutObject 以 OSS V1 Header 签名直传本地文件,返回公开访问 URL。
// 与上传接口的前端直传 Policy 不同,这里是服务端批处理(爬虫/采集任务)使用。
func (s *OssSigner) PutObject(localPath, key string) (string, error) {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(localPath)))
	if ct == "" {
		ct = "application/octet-stream"
	}
	return s.PutObjectBytes(data, key, ct)
}

// DeleteObject 删除 OSS 上的单个对象(用于替换配图后清理旧图,避免残留无用文件)。
func (s *OssSigner) DeleteObject(key string) error {
	if !s.Configured() {
		return errors.New("OSS 未配置")
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	// StringToSign = VERB \n Content-MD5 \n Content-Type \n Date \n /Bucket/Key
	stringToSign := fmt.Sprintf("DELETE\n\n\n%s\n/%s/%s", date, s.Bucket, key)
	mac := hmac.New(sha1.New, []byte(s.SecretKey))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	u := fmt.Sprintf("https://%s.%s/%s", s.Bucket, s.endpointHost(), key)
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", "OSS "+s.AccessKey+":"+signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("OSS 返回 HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// 上传前压缩参数:维基等来源原图常达 10MB+,直传既占桶空间又拖慢前端加载,
// 统一在入 OSS 前缩到 1920px 宽(前端详情大图展示宽度 ≤1280,留足高分屏余量)。
const (
	uploadImageMaxWidth = 1920
	uploadImageQuality  = 85
)

// compressImageForUpload 上传前压缩图片字节,返回压缩后内容与 Content-Type。
// 仅对图片类内容(或类型未知的二进制)尝试,非图片会因解码失败自动跳过;
// 压缩后输出格式与原图一致,故调用方按原扩展名拼的 key 无需改动。
func compressImageForUpload(data []byte, contentType string) ([]byte, string) {
	if !strings.HasPrefix(contentType, "image/") && contentType != "application/octet-stream" {
		return data, contentType
	}
	resized, ct, ok := ResizeImage(data, uploadImageMaxWidth, uploadImageQuality)
	if !ok {
		return data, contentType
	}
	return resized, ct
}

// PutObjectBytes 直传字节内容,contentType 为空时按 key 后缀推断。
// 图片类内容会先压缩(见 compressImageForUpload);非图片或压缩无收益时按原样上传。
func (s *OssSigner) PutObjectBytes(data []byte, key, contentType string) (string, error) {
	if !s.Configured() {
		return "", errors.New("OSS 未配置")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	data, contentType = compressImageForUpload(data, contentType)
	date := time.Now().UTC().Format(http.TimeFormat)
	// StringToSign = VERB \n Content-MD5 \n Content-Type \n Date \n /Bucket/Key
	stringToSign := fmt.Sprintf("PUT\n\n%s\n%s\n/%s/%s", contentType, date, s.Bucket, key)
	mac := hmac.New(sha1.New, []byte(s.SecretKey))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	u := fmt.Sprintf("https://%s.%s/%s", s.Bucket, s.endpointHost(), key)
	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Date", date)
	req.Header.Set("Content-Type", contentType)
	// 浏览器缓存 24h:看过的图片不再重复回源 OSS
	req.Header.Set("Cache-Control", "public, max-age=86400")
	req.Header.Set("Authorization", "OSS "+s.AccessKey+":"+signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("OSS 返回 HTTP %d: %s", resp.StatusCode, string(body))
	}
	return s.ResolveURL(key), nil
}
