package service

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"wenlv-backend/pkg"
)

// UploadService 处理 OSS 直传签名与 URL 解析。
type UploadService struct {
	signer     *pkg.OssSigner
	expireMins int
}

// NewUploadService 构造上传服务。
func NewUploadService(signer *pkg.OssSigner, expireMins int) *UploadService {
	return &UploadService{signer: signer, expireMins: expireMins}
}

// Policy 生成上传直传签名。fileName 用于保留扩展名。
func (s *UploadService) Policy(userID uint, fileName string) (*pkg.UploadPolicy, error) {
	key := buildObjectKey(userID, fileName)
	return s.signer.GeneratePolicy(key, s.expireMins, nil)
}

// ResolveURL 将 OSS 存储 key 转成可访问 URL。
func (s *UploadService) ResolveURL(key string) string {
	return s.signer.ResolveURL(key)
}

// buildObjectKey 构造存储路径,只保留扩展名与时间戳,避免目录穿越。
func buildObjectKey(uid uint, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = ".jpg"
	}
	return fmt.Sprintf("uploads/%d/%d%s", uid, time.Now().UnixNano(), ext)
}