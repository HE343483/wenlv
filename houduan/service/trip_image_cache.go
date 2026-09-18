package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ============ 图片字节磁盘缓存(多内容平台共用) ============
// 行程模块的图片缓存统一落在 data/photo_cache 下,以 cacheKey 前缀区分来源(如 dy:kw:、kw:、url:)。
// 磁盘里存的是「图片字节」而非时效性直链,因此 TTL 可以放得比较长:
// 每次缓存过期都意味着要重新去内容平台搜一次图,而反复搜图正是触发平台风控的主因。

const (
	// 默认缓存有效期:7 天(可用 TRIP_IMAGE_CACHE_TTL_HOURS 覆盖)
	defaultImageCacheTTL = 7 * 24 * time.Hour
	// 默认缓存容量上限:512MB(可用 TRIP_IMAGE_CACHE_MAX_MB 覆盖)
	defaultImageCacheMaxMB = 512
	// 容量清理的最小间隔:避免每次写入都遍历目录
	imageCacheCleanupInterval = time.Hour
)

// imageDiskCache 图片字节的磁盘缓存。
type imageDiskCache struct {
	dir      string
	ttl      time.Duration
	maxBytes int64

	writeMu sync.Mutex // 写入串行化(临时文件 + rename)
	gcMu    sync.Mutex // 清理单飞 + 节流
	lastGC  time.Time
}

// newImageDiskCache 构造磁盘缓存;ttl/maxBytes 非正值时回退默认值。
func newImageDiskCache(dir string, ttl time.Duration, maxBytes int64) *imageDiskCache {
	if ttl <= 0 {
		ttl = defaultImageCacheTTL
	}
	if maxBytes <= 0 {
		maxBytes = int64(defaultImageCacheMaxMB) << 20
	}
	return &imageDiskCache{dir: dir, ttl: ttl, maxBytes: maxBytes}
}

// cacheDir 返回图片缓存目录。
func (c *imageDiskCache) cacheDir() string { return filepath.Join(c.dir, "photo_cache") }

func (c *imageDiskCache) paths(cacheKey string) (string, string) {
	sum := sha256.Sum256([]byte(cacheKey))
	digest := hex.EncodeToString(sum[:])
	dir := c.cacheDir()
	return filepath.Join(dir, digest+".img"), filepath.Join(dir, digest+".meta")
}

// Read 读取缓存;命中返回图片字节与 Content-Type。过期文件不再返回(但由清理任务负责删除)。
func (c *imageDiskCache) Read(cacheKey string) ([]byte, string, bool) {
	imgPath, metaPath := c.paths(cacheKey)
	info, err := os.Stat(imgPath)
	if err != nil || time.Since(info.ModTime()) >= c.ttl {
		return nil, "", false
	}
	content, err := os.ReadFile(imgPath)
	if err != nil {
		return nil, "", false
	}
	contentType := "image/jpeg"
	if raw, err := os.ReadFile(metaPath); err == nil {
		if t := strings.TrimSpace(string(raw)); t != "" {
			contentType = t
		}
	}
	return content, contentType, true
}

// Write 写入缓存(先写临时文件再重命名,避免并发读到半截内容),并按需触发容量清理。
func (c *imageDiskCache) Write(cacheKey string, content []byte, contentType string) {
	imgPath, metaPath := c.paths(cacheKey)
	c.writeMu.Lock()
	if err := os.MkdirAll(filepath.Dir(imgPath), 0o755); err == nil {
		if err := os.WriteFile(imgPath+".tmp", content, 0o644); err == nil {
			if err := os.Rename(imgPath+".tmp", imgPath); err == nil {
				_ = os.WriteFile(metaPath+".tmp", []byte(contentType), 0o644)
				_ = os.Rename(metaPath+".tmp", metaPath)
			}
		}
	}
	c.writeMu.Unlock()

	c.gcIfDue()
}

// gcIfDue 低频触发一次清理(默认每小时最多一次)。
func (c *imageDiskCache) gcIfDue() {
	c.gcMu.Lock()
	if time.Since(c.lastGC) < imageCacheCleanupInterval {
		c.gcMu.Unlock()
		return
	}
	c.lastGC = time.Now()
	c.gcMu.Unlock()

	c.cleanup()
}

// cleanup 先按 TTL 删除过期文件;若仍超出容量上限,再按最久未修改优先淘汰。
func (c *imageDiskCache) cleanup() {
	dir := c.cacheDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	type cacheItem struct {
		path string
		size int64
		mod  time.Time
	}
	items := make([]cacheItem, 0, len(entries))
	var total int64
	now := time.Now()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".img") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		imgPath := filepath.Join(dir, entry.Name())
		if now.Sub(info.ModTime()) >= c.ttl {
			c.remove(imgPath)
			continue
		}
		items = append(items, cacheItem{path: imgPath, size: info.Size(), mod: info.ModTime()})
		total += info.Size()
	}

	if total <= c.maxBytes {
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mod.Before(items[j].mod) })
	removed := 0
	for _, it := range items {
		if total <= c.maxBytes {
			break
		}
		c.remove(it.path)
		total -= it.size
		removed++
	}
	if removed > 0 {
		fmt.Printf("🧹 [图片缓存] 超出容量上限,已淘汰 %d 个最久未使用的缓存文件\n", removed)
	}
}

// remove 同时删除图片本体与对应的 Content-Type 元数据文件。
func (c *imageDiskCache) remove(imgPath string) {
	_ = os.Remove(imgPath)
	_ = os.Remove(strings.TrimSuffix(imgPath, ".img") + ".meta")
}
