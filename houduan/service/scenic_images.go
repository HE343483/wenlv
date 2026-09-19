// 景点图片多来源候选采集:高德相册 → 小红书 → Wikimedia Commons。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"wenlv-backend/model"
	"wenlv-backend/pkg"
)

// ScenicImageCandidate 一张待落库的景点图片及其来源(amap/xhs/commons)。
type ScenicImageCandidate struct {
	URL    string
	Source string
}

// 图片来源标记,用于日志统计与人工核对。
const (
	scenicImageSourceAmap    = "amap"
	scenicImageSourceXHS     = "xhs"
	scenicImageSourceCommons = "commons"
)

const (
	// scenicGalleryWant 每个景点目标收录的图片张数(上限)。
	scenicGalleryWant = 6
	// scenicCommonsMinWidth Commons 图片最小宽度(px)。
	scenicCommonsMinWidth = 800
	// scenicCommonsSearchLimit 单次 Commons 搜索请求返回的结果条数。
	scenicCommonsSearchLimit = 10
	// scenicCommonsThumbWidth Commons 缩略图宽度(px):原图可达 8MB 易触发 429,统一改用缩略图 URL。
	scenicCommonsThumbWidth = 1600
	// scenicCommonsRequestInterval Commons 相邻搜索请求的最小间隔,降低被限流的概率。
	scenicCommonsRequestInterval = 300 * time.Millisecond
	// xhsEmptyStreakLimit 小红书连续空结果达到该次数即判定整体失效,本次运行不再调用小红书。
	xhsEmptyStreakLimit = 2
)

// xhsShortCircuit 小红书链路短路状态:Cookie 失效后每次调用都要等接口超时,
// 连续空结果达到阈值后本次运行直接跳过后续景点的小红书抓取。
// 采集目前是串行批处理,加锁仅为并发安全。
type xhsShortCircuit struct {
	mu       sync.Mutex
	empty    int
	disabled bool
}

// available 返回小红书链路是否仍可调用。
func (x *xhsShortCircuit) available() bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	return !x.disabled
}

// noteEmpty 记录一次空结果,返回是否因此禁用小红书。
func (x *xhsShortCircuit) noteEmpty() bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.empty++
	if x.empty >= xhsEmptyStreakLimit {
		x.disabled = true
		return true
	}
	return false
}

// noteHit 记录一次命中,重置连续空结果计数。
func (x *xhsShortCircuit) noteHit() {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.empty = 0
}

// collectImageCandidates 按「高德 → 小红书 → Wikimedia Commons」优先级收集景点图片候选,
// 最多 want 张,按 URL 去重;任一来源失败只记日志并继续,返回顺序与来源优先级一致。
func (e *ScenicEnricher) collectImageCandidates(ctx context.Context, s model.ScenicSpot, amapPhotos []string, want int) []ScenicImageCandidate {
	if want <= 0 {
		return nil
	}
	out := make([]ScenicImageCandidate, 0, want)
	seen := make(map[string]bool, want)
	add := func(rawURL, source string) {
		u := strings.TrimSpace(rawURL)
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		out = append(out, ScenicImageCandidate{URL: u, Source: source})
	}

	// 1) 高德相册图
	for _, u := range amapPhotos {
		if len(out) >= want {
			break
		}
		add(u, scenicImageSourceAmap)
	}

	// 2) 小红书:按关键词变体依次取图;连续空结果即短路,避免后续景点反复等待超时
	if len(out) < want && e.xhs != nil && e.xhsState.available() {
		for _, kw := range []string{s.NameZH, s.NameZH + " 风景", s.NameZH + " 旅游", s.NameZH + " 攻略"} {
			if len(out) >= want {
				break
			}
			u := strings.TrimSpace(e.xhs.PhotoURL(ctx, kw, "成都"))
			if u == "" {
				if e.xhsState.noteEmpty() {
					log.Printf("    小红书连续 %d 次无结果,判定 Cookie 失效,本次运行后续景点跳过小红书", xhsEmptyStreakLimit)
					break
				}
				continue
			}
			e.xhsState.noteHit()
			add(u, scenicImageSourceXHS)
		}
	}

	// 3) Wikimedia Commons:中文名与英文名分别搜索,合并后按宽度从大到小补足
	if len(out) < want {
		names := []string{s.NameZH}
		nameEN := strings.TrimSpace(s.NameEN)
		if nameEN != "" {
			names = append(names, nameEN)
		}
		var hits []commonsImageResult
		seenTitle := map[string]bool{}
		commonsRequests := 0
		for _, q := range []string{s.NameZH, nameEN} {
			if q == "" || len(out)+len(hits) >= want {
				continue
			}
			if commonsRequests > 0 {
				time.Sleep(scenicCommonsRequestInterval)
			}
			commonsRequests++
			results, err := e.searchCommonsImages(ctx, q, names, want)
			if err != nil {
				log.Printf("    Commons 搜索失败 (%s): %v", q, err)
				continue
			}
			for _, r := range results {
				if seenTitle[r.Title] {
					continue
				}
				seenTitle[r.Title] = true
				hits = append(hits, r)
			}
		}
		sort.SliceStable(hits, func(i, j int) bool { return hits[i].Width > hits[j].Width })
		for _, r := range hits {
			if len(out) >= want {
				break
			}
			add(r.URL, scenicImageSourceCommons)
		}
	}

	return out
}

// commonsImageResult Wikimedia Commons 单条搜索结果(带文件名,便于日志核对)。
type commonsImageResult struct {
	Title string
	URL   string
	Width int
}

// searchCommonsImages 保留原方法签名,内部委托包级函数,供景点采集继续使用。
func (e *ScenicEnricher) searchCommonsImages(ctx context.Context, keyword string, names []string, limit int) ([]commonsImageResult, error) {
	return searchCommonsImagesFor(ctx, e.commonsClient, keyword, names, limit)
}

// searchCommonsImagesFor 按关键词搜索 Commons,过滤后按宽度从大到小返回最多 limit 张。
func searchCommonsImagesFor(ctx context.Context, client *http.Client, keyword string, names []string, limit int) ([]commonsImageResult, error) {
	if client == nil {
		return nil, fmt.Errorf("Commons 客户端未初始化")
	}
	apiURL := fmt.Sprintf(
		"https://commons.wikimedia.org/w/api.php?action=query&generator=search&gsrsearch=%s&gsrnamespace=6&gsrlimit=%d&prop=imageinfo&iiprop=url%%7Csize&iiurlwidth=%d&format=json",
		url.QueryEscape(keyword), scenicCommonsSearchLimit, scenicCommonsThumbWidth)
	var res struct {
		Query struct {
			Pages map[string]struct {
				Title     string `json:"title"`
				ImageInfo []struct {
					URL      string `json:"url"`
					ThumbURL string `json:"thumburl"`
					Width    int    `json:"width"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := commonsGetJSON(ctx, client, apiURL, &res); err != nil {
		return nil, err
	}
	var out []commonsImageResult
	for _, p := range res.Query.Pages {
		if len(p.ImageInfo) == 0 {
			continue
		}
		info := p.ImageInfo[0]
		// 优先用缩略图 URL:原图(可达 8MB)下载易触发 Wikimedia 429 限流。
		rawURL := strings.TrimSpace(info.ThumbURL)
		if rawURL == "" {
			rawURL = info.URL
		}
		clean := stripURLQuery(rawURL)
		// 宽度门槛仍按原图宽度判断,不受缩略图宽度影响。
		if !commonsFileAcceptable(p.Title, clean, info.Width, names) {
			continue
		}
		out = append(out, commonsImageResult{Title: p.Title, URL: clean, Width: info.Width})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Width > out[j].Width })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// commonsFileAcceptable 判断 Commons 搜索结果是否可用:文件名(title)须包含景点中文名或英文名
// (大小写不敏感),格式限 jpg/jpeg/png,宽度不低于 scenicCommonsMinWidth。
func commonsFileAcceptable(title, rawURL string, width int, names []string) bool {
	if width < scenicCommonsMinWidth {
		return false
	}
	lowerTitle := strings.ToLower(title)
	matched := false
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if strings.Contains(lowerTitle, strings.ToLower(n)) {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}
	switch strings.ToLower(filepath.Ext(stripURLQuery(rawURL))) {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

// stripURLQuery 去掉 Wikimedia 附加的 ?utm_source=... 查询参数,便于判断扩展名。
func stripURLQuery(raw string) string {
	if idx := strings.IndexByte(raw, '?'); idx >= 0 {
		return raw[:idx]
	}
	return raw
}

// fetchCommonsImageBytes 下载 Commons 图片:命中 429 限流时等待 2 秒重试一次,仍失败则跳过该张。
// 仅用于 Commons 来源,避免影响其它来源的下载链路。
func fetchCommonsImageBytes(ctx context.Context, client *http.Client, rawURL string) ([]byte, string, error) {
	body, ct, err := fetchBytesWithClient(ctx, client, rawURL)
	if err == nil || !isRateLimitedError(err) {
		return body, ct, err
	}
	log.Printf("    Commons 图片被限流(429),等待 2 秒后重试一次: %s", rawURL)
	time.Sleep(2 * time.Second)
	body, ct, retryErr := fetchBytesWithClient(ctx, client, rawURL)
	if retryErr != nil {
		log.Printf("    Commons 图片重试仍失败,跳过该张 %s: %v", rawURL, retryErr)
	}
	return body, ct, retryErr
}

// uploadImageCandidates 下载候选图片并上传 OSS,返回 OSS URL 列表与成功张数。
// keyPrefix 形如 "scenic/12" 或 "food/3";单张失败只记日志并跳过;
// 达到 galleryCap 张即停止。Commons 来源走专用客户端(可经 COMMONS_PROXY)并带 429 退避。
func uploadImageCandidates(ctx context.Context, client *http.Client, signer *pkg.OssSigner, keyPrefix string, candidates []ScenicImageCandidate, galleryCap int) ([]string, int) {
	var out []string
	for i, c := range candidates {
		if len(out) >= galleryCap {
			break
		}
		var (
			body []byte
			ct   string
			err  error
		)
		// Commons 图走 Commons 客户端(可经 COMMONS_PROXY),其余来源沿用通用下载客户端。
		if c.Source == scenicImageSourceCommons && client != nil {
			body, ct, err = fetchCommonsImageBytes(ctx, client, c.URL)
		} else {
			body, ct, err = fetchBytes(ctx, c.URL)
		}
		if err != nil {
			continue
		}
		ext := extFromContentType(ct, c.URL)
		key := fmt.Sprintf("%s/g%d%s", keyPrefix, i+1, ext)
		ossURL, err := signer.PutObjectBytes(body, key, ct)
		if err != nil {
			continue
		}
		out = append(out, ossURL)
	}
	return out, len(out)
}

// commonsGetJSON 请求 Commons Action API 并解析 JSON。
func commonsGetJSON(ctx context.Context, client *http.Client, apiURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; wenlv-enricher/1.0)")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("Commons 响应解析失败: %w", err)
	}
	return nil
}

// newCommonsHTTPClient 构造 Commons 专用 HTTP 客户端。
// Commons 国内通常需要代理;Go 不使用系统代理,因此这里显式支持 COMMONS_PROXY 与标准环境变量:
// COMMONS_PROXY 非空时优先按它设置(写法与 GOOGLE_MAPS_PROXY 一致,支持 http(s) 与 socks5);
// 为空或地址非法时回退到 http.ProxyFromEnvironment(即 HTTPS_PROXY/HTTP_PROXY)。
func newCommonsHTTPClient() *http.Client {
	transport := &http.Transport{}
	p := strings.TrimSpace(os.Getenv("COMMONS_PROXY"))
	if p == "" {
		transport.Proxy = http.ProxyFromEnvironment
	} else if u, err := url.Parse(p); err != nil {
		log.Printf("Commons 代理地址非法 (%s): %v,回退到 HTTPS_PROXY/HTTP_PROXY", p, err)
		transport.Proxy = http.ProxyFromEnvironment
	} else {
		switch strings.ToLower(u.Scheme) {
		case "socks5", "socks5h":
			if dialer, derr := proxy.SOCKS5("tcp", u.Host, nil, proxy.Direct); derr == nil {
				if ctxDialer, ok := dialer.(proxy.ContextDialer); ok {
					transport.DialContext = ctxDialer.DialContext
				}
			}
		default:
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Timeout: 20 * time.Second, Transport: transport}
}
