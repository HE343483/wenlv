package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/pkg"
)

// ============ 行程图片 OSS 落库 ============
// 行程生成完成后,后台异步把景点图片上传到 OSS,并把公开链接写回行程 JSON
// (持久化到 trip_plans.plan_json),历史详情渲染时直连 OSS,不再走 /api/poi/image 代理。
// 直链时效短的内容平台图片(小红书/抖音)经 OSS 中转后变为永久可访问。

// ImageFetchFunc 按 (source, name, city) 取景点图片字节。
type ImageFetchFunc func(ctx context.Context, source, name, city string) ([]byte, string, error)

// TripImageEnricher 行程图片 OSS 上传器。
type TripImageEnricher struct {
	signer  *pkg.OssSigner
	fetcher ImageFetchFunc
}

// NewTripImageEnricher 构造上传器;OSS 四项配置不完整时返回 nil(整条链路静默跳过)。
func NewTripImageEnricher(endpoint, accessKey, secretKey, bucket string, fetcher ImageFetchFunc) *TripImageEnricher {
	signer := pkg.NewOssSigner(&pkg.OssConfig{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
	})
	if !signer.Configured() || fetcher == nil {
		return nil
	}
	return &TripImageEnricher{signer: signer, fetcher: fetcher}
}

// isOSSImage 判断链接是否已是 OSS 直链(幂等,重复进入不重复上传)。
func isOSSImage(url string) bool {
	return strings.Contains(url, "aliyuncs.com")
}

// Enrich 为行程中缺图的景点抓图并上传 OSS,直接就地写入 Attraction.ImageURL。
// 返回成功/失败条数;单图失败不阻断其他图片(前端对这些景点仍回退代理取图)。
func (e *TripImageEnricher) Enrich(ctx context.Context, plan *model.TripPlan) (okCount, failCount int) {
	if e == nil || plan == nil {
		return 0, 0
	}
	source := plan.AttractionSource
	if source == "" {
		source = "xhs"
	}
	city := plan.City

	// 去重收集缺图景点(不同天引用同名景点只取一次图)
	names := make([]string, 0, 16)
	seen := make(map[string]bool)
	for i := range plan.Days {
		for j := range plan.Days[i].Attractions {
			a := &plan.Days[i].Attractions[j]
			if a.Name == "" || isOSSImage(a.ImageURL) || seen[a.Name] {
				continue
			}
			seen[a.Name] = true
			names = append(names, a.Name)
		}
	}
	if len(names) == 0 {
		return 0, 0
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3) // 并发 3:与前端取图并发一致,降低风控压力
	for _, name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			url, err := e.uploadOne(ctx, source, name, city)
			mu.Lock()
			if err != nil {
				failCount++
				mu.Unlock()
				logger.Warnf("[行程图片OSS] %s 上传失败: %v", name, err)
				return
			}
			okCount++
			mu.Unlock()
			// 写回所有同名景点
			for i := range plan.Days {
				for j := range plan.Days[i].Attractions {
					if plan.Days[i].Attractions[j].Name == name {
						plan.Days[i].Attractions[j].ImageURL = url
					}
				}
			}
		}(name)
	}
	wg.Wait()
	return okCount, failCount
}

// uploadOne 取图字节并上传 OSS,失败重试 1 次。
func (e *TripImageEnricher) uploadOne(ctx context.Context, source, name, city string) (string, error) {
	content, contentType, err := e.fetcher(ctx, source, name, city)
	if err != nil {
		// 换一次机会:抓图偶发风控/超时时重试
		time.Sleep(2 * time.Second)
		content, contentType, err = e.fetcher(ctx, source, name, city)
		if err != nil {
			return "", fmt.Errorf("取图失败: %w", err)
		}
	}
	if len(content) == 0 {
		return "", fmt.Errorf("图片字节为空")
	}
	key := "trip/" + tripImageKey(source, name, city) + imageExt(contentType)
	return e.signer.PutObjectBytes(content, key, contentType)
}

// tripImageKey 同一景点+城市+来源生成稳定 key,跨行程复用同一对象(重复上传为覆盖写)。
func tripImageKey(source, name, city string) string {
	sum := sha256.Sum256([]byte(source + "|" + city + "|" + name))
	return hex.EncodeToString(sum[:8]) // 16 位十六进制,够防碰撞且路径短
}

// imageExt 按 Content-Type 推断扩展名,兜底 .jpg。
func imageExt(contentType string) string {
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	default:
		return ".jpg"
	}
}
