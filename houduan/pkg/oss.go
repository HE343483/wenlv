package pkg

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// NewOssSigner 构造签名器。
func NewOssSigner(cfg *OssConfig) *OssSigner {
	return &OssSigner{
		Endpoint:  cfg.Endpoint,
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