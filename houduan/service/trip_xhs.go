package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ============ 小红书服务 ============
// 移植自 TripStar 的 xhs_service:原生签名直连 edith.xiaohongshu.com,
// 失败时降级到网页 SSR 抓取;景点图片走"关键词磁盘缓存 + 后端代理"方案。

const (
	xhsBaseURL       = "https://edith.xiaohongshu.com"
	xhsSearchAPI     = "/api/sns/web/v1/search/notes"
	xhsFeedAPI       = "/api/sns/web/v1/feed"
	xhsImageMaxBytes = 10 * 1024 * 1024
	xhsUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"
	xhsBrowserUA     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// XHSCookieExpiredError 小红书 Cookie 过期/被风控,需要向前端明确提示。
type XHSCookieExpiredError struct{ Msg string }

func (e *XHSCookieExpiredError) Error() string { return e.Msg }

// XHSNotConfiguredError 小红书 Cookie 未配置;此时行程规划会降级使用地图 POI 兜底。
type XHSNotConfiguredError struct{ Msg string }

func (e *XHSNotConfiguredError) Error() string { return e.Msg }

// XHSFetchError 小红书普通抓取异常。
type XHSFetchError struct{ Msg string }

func (e *XHSFetchError) Error() string { return e.Msg }

// XHSImageProxyError 图片代理抓取失败。
type XHSImageProxyError struct{ Msg string }

func (e *XHSImageProxyError) Error() string { return e.Msg }

// XHSService 小红书服务。
type XHSService struct {
	settings *TripSettings
	signer   *XHSSigner
	llm      *TripLLM
	amap     *AmapService
	client   *http.Client
	cache    *imageDiskCache
	pool     *xhsCookiePool

	photoMu    sync.Mutex
	photoCache map[string]photoCacheEntry
}

type photoCacheEntry struct {
	url       string
	expiresAt time.Time
}

const (
	xhsPhotoCacheTTL     = 6 * time.Hour
	xhsPhotoCacheMissTTL = 10 * time.Minute
)

// NewXHSService 构造小红书服务。
func NewXHSService(settings *TripSettings, llm *TripLLM, amap *AmapService, signer *XHSSigner) *XHSService {
	return &XHSService{
		settings: settings,
		signer:   signer,
		llm:      llm,
		amap:     amap,
		client: &http.Client{
			Timeout:   15 * time.Second,
			Transport: &http.Transport{Proxy: nil},
		},
		cache:      newImageDiskCache(settings.DataDir(), settings.ImageCacheTTL(), settings.ImageCacheMaxBytes()),
		pool:       newXHSCookiePool(),
		photoCache: map[string]photoCacheEntry{},
	}
}

// NormalizeXHSCookie 兼容 Cookie 请求头字符串和浏览器导出的 JSON Cookie 列表。
func NormalizeXHSCookie(cookie string) string {
	normalized := strings.TrimSpace(cookie)
	if normalized == "" {
		return normalized
	}
	if len(normalized) >= 2 {
		first, last := normalized[0], normalized[len(normalized)-1]
		if first == last && (first == '\'' || first == '"') {
			normalized = strings.TrimSpace(normalized[1 : len(normalized)-1])
		}
	}
	if strings.HasPrefix(normalized, "[") && strings.HasSuffix(normalized, "]") {
		var items []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal([]byte(normalized), &items); err == nil {
			pairs := make([]string, 0, len(items))
			for _, it := range items {
				if strings.TrimSpace(it.Name) != "" {
					pairs = append(pairs, it.Name+"="+it.Value)
				}
			}
			if len(pairs) > 0 {
				return strings.Join(pairs, "; ")
			}
		}
	}
	return normalized
}

func transCookies(cookieStr string) map[string]string {
	out := map[string]string{}
	sep := ";"
	if strings.Contains(cookieStr, "; ") {
		sep = "; "
	}
	for _, item := range strings.Split(cookieStr, sep) {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			out[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return out
}

func randomHex(length int) string {
	const chars = "abcdef0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// signedRequest 生成签名并请求小红书接口。
func (s *XHSService) signedRequest(ctx context.Context, cookie, api string, payload map[string]any) (map[string]any, error) {
	cookies := transCookies(cookie)
	a1 := cookies["a1"]

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	xs, xt, xsCommon, err := s.signer.Sign(ctx, api, json.RawMessage(data), a1, "POST")
	if err != nil {
		return nil, err
	}
	traceID, err := s.signer.TraceID(ctx)
	if err != nil {
		// traceid 仅用于链路追踪,失败不影响主流程
		traceID = randomHex(16)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, xhsBaseURL+api, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"authority":          "edith.xiaohongshu.com",
		"accept":             "application/json, text/plain, */*",
		"accept-language":    "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"cache-control":      "no-cache",
		"content-type":       "application/json;charset=UTF-8",
		"origin":             "https://www.xiaohongshu.com",
		"pragma":             "no-cache",
		"referer":            "https://www.xiaohongshu.com/",
		"sec-ch-ua":          `"Not A(Brand";v="99", "Microsoft Edge";v="121", "Chromium";v="121"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         xhsUserAgent,
		"x-b3-traceid":       randomHex(16),
		"x-mns":              "unload",
		"x-s":                xs,
		"x-s-common":         xsCommon,
		"x-t":                fmt.Sprintf("%d", xt),
		"x-xray-traceid":     traceID,
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", cookie)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &XHSFetchError{Msg: fmt.Sprintf("访问小红书失败: %v", err)}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &XHSFetchError{Msg: fmt.Sprintf("小红书响应读取失败: %v", err)}
	}
	var res map[string]any
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, &XHSFetchError{Msg: fmt.Sprintf("小红书响应解析失败: %s", truncateText(string(raw), 200))}
	}
	if success, _ := res["success"].(bool); !success {
		code := fmt.Sprintf("%v", res["code"])
		msg := fmt.Sprintf("%v", res["msg"])
		if xhsCookieExpiredSignal(code, msg) {
			return nil, &XHSCookieExpiredError{Msg: fmt.Sprintf("小红书 Cookie 已失效或被风控 (code=%s): %s。请重新登录后导出并更换 Cookie。", code, msg)}
		}
		return nil, &XHSFetchError{Msg: fmt.Sprintf("小红书接口失败 (code=%s): %s", code, msg)}
	}
	return res, nil
}

// xhsCookieExpiredSignal 判断小红书返回的业务码/文案是否属于 Cookie 失效(需要更换 Cookie)。
//   - 300011 / 文案含"异常":账号被风控拦截
//   - -100 / 文案含"登录已过期":网页端登录态过期(实测返回 code=-100, msg=登录已过期)
func xhsCookieExpiredSignal(code, msg string) bool {
	if code == "300011" || code == "-100" {
		return true
	}
	return strings.Contains(msg, "异常") ||
		strings.Contains(msg, "登录已过期") ||
		strings.Contains(msg, "登录过期")
}

// searchNotes 搜索笔记。
func (s *XHSService) searchNotes(ctx context.Context, cookie, keyword string, sortType int) (map[string]any, error) {
	sortMap := map[int]string{
		0: "general", 1: "time_descending", 2: "popularity_descending",
		3: "comment_descending", 4: "collect_descending",
	}
	sort, ok := sortMap[sortType]
	if !ok {
		sort = "general"
	}
	payload := map[string]any{
		"keyword":   keyword,
		"page":      1,
		"page_size": 20,
		"search_id": randomHex(21),
		"sort":      "general",
		"note_type": 0,
		"ext_flags": []any{},
		"filters": []map[string]any{
			{"tags": []string{sort}, "type": "sort_type"},
			{"tags": []string{"不限"}, "type": "filter_note_type"},
			{"tags": []string{"不限"}, "type": "filter_note_time"},
			{"tags": []string{"不限"}, "type": "filter_note_range"},
			{"tags": []string{"不限"}, "type": "filter_pos_distance"},
		},
		"geo":           "",
		"image_formats": []string{"jpg", "webp", "avif"},
	}
	return s.signedRequest(ctx, cookie, xhsSearchAPI, payload)
}

// getNoteDetail 获取笔记详情。
func (s *XHSService) getNoteDetail(ctx context.Context, cookie, noteID, xsecToken string) (map[string]any, error) {
	payload := map[string]any{
		"source_note_id": noteID,
		"image_formats":  []string{"jpg", "webp", "avif"},
		"extra":          map[string]any{"need_body_topic": "1"},
		"xsec_source":    "pc_search",
		"xsec_token":     xsecToken,
	}
	return s.signedRequest(ctx, cookie, xhsFeedAPI, payload)
}

var ssrStateRe = regexp.MustCompile(`window\.__INITIAL_STATE__=(\{.*?\})</script>`)

// getNoteDetailSSR 网页 SSR 抓取(原生 API 的降级备选)。
func (s *XHSService) getNoteDetailSSR(ctx context.Context, noteID string) map[string]any {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.xiaohongshu.com/explore/"+noteID, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", xhsBrowserUA)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	match := ssrStateRe.FindSubmatch(raw)
	if len(match) < 2 {
		return nil
	}
	var state map[string]any
	if err := json.Unmarshal([]byte(strings.ReplaceAll(string(match[1]), "undefined", "null")), &state); err != nil {
		return nil
	}
	note, _ := state["note"].(map[string]any)
	if note == nil {
		return nil
	}
	detailMap, _ := note["noteDetailMap"].(map[string]any)
	entry, _ := detailMap[noteID].(map[string]any)
	detail, _ := entry["note"].(map[string]any)
	return detail
}

// ============ 景点搜索(小红书 + LLM 提纯) ============

var langNames = map[string]string{
	"en": "English", "ja": "Japanese", "ko": "Korean",
	"fr": "French", "de": "German", "es": "Spanish",
}

// SearchAttractionsText 搜索小红书游记并经 LLM 提纯为结构化景点文本。
func (s *XHSService) SearchAttractionsText(ctx context.Context, city, keywords, language string) (string, error) {
	fmt.Printf("🔍 [XHS] 正在搜索小红书: %s %s\n", city, keywords)
	query := fmt.Sprintf("%s %s 旅游 景点攻略", city, keywords)

	// 搜索 + 逐条取详情整体纳入 Cookie 轮换:某个 Cookie 被风控时换下一个重试
	combined, err := withXHSCookie(s, ctx, func(cookie string) (string, error) {
		resJSON, err := s.searchNotes(ctx, cookie, query, 0)
		if err != nil {
			return "", err
		}
		items := extractItems(resJSON)
		if len(items) > 4 {
			items = items[:4]
		}

		var b strings.Builder
		for i, note := range items {
			if modelType, _ := note["model_type"].(string); modelType != "note" {
				continue
			}
			noteCard, _ := note["note_card"].(map[string]any)
			title := stringValue(noteCard["display_title"])

			desc := ""
			noteID := stringValue(note["id"])
			xsecToken := stringValue(note["xsec_token"])
			if noteID != "" {
				if detail, derr := s.getNoteDetail(ctx, cookie, noteID, xsecToken); derr == nil {
					detailItems := extractItems(detail)
					if len(detailItems) > 0 {
						card, _ := detailItems[0]["note_card"].(map[string]any)
						desc = stringValue(card["desc"])
					}
				}
				if desc == "" {
					if ssr := s.getNoteDetailSSR(ctx, noteID); ssr != nil {
						desc = stringValue(ssr["desc"])
					}
				}
			}
			fmt.Fprintf(&b, "\n笔记%d:\n标题: %s\n正文内容: %s\n", i+1, title, desc)
		}
		return b.String(), nil
	})
	if err != nil {
		return "", err
	}

	if combined == "" {
		return fmt.Sprintf("未在小红书检索到关于 %s %s 的内容。", city, keywords), nil
	}
	return s.extractAttractions(ctx, city, combined, language), nil
}

func extractItems(res map[string]any) []map[string]any {
	data, _ := res["data"].(map[string]any)
	if data == nil {
		return nil
	}
	rawItems, _ := data["items"].([]any)
	out := make([]map[string]any, 0, len(rawItems))
	for _, it := range rawItems {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// extractAttractions 调用 LLM 从游记杂文中提纯景点,并补齐经纬度。
// 提纯逻辑与抖音等其它内容平台共用(见 trip_attraction_extract.go)。
func (s *XHSService) extractAttractions(ctx context.Context, city, content, language string) string {
	return extractAttractionsFromNotes(ctx, s.llm, s.settings, s.amap, city, content, language, "小红书")
}

var jsonArrayRe = regexp.MustCompile(`(?s)\[.*\]`)

func extractJSONArray(text string) string {
	if m := jsonArrayRe.FindString(text); m != "" {
		return m
	}
	return ""
}

// ============ 景点搜图 ============

// photoURLFromXHS 根据关键词从小红书搜索一张首图直链(带内存缓存)。
func (s *XHSService) photoURLFromXHS(ctx context.Context, keyword string) string {
	s.photoMu.Lock()
	if entry, ok := s.photoCache[keyword]; ok && time.Now().Before(entry.expiresAt) {
		s.photoMu.Unlock()
		return entry.url
	}
	s.photoMu.Unlock()

	url := s.fetchPhotoURL(ctx, keyword)

	s.photoMu.Lock()
	ttl := xhsPhotoCacheTTL
	if url == "" {
		ttl = xhsPhotoCacheMissTTL
	}
	s.photoCache[keyword] = photoCacheEntry{url: url, expiresAt: time.Now().Add(ttl)}
	if len(s.photoCache) > 512 {
		for k, v := range s.photoCache {
			if time.Now().After(v.expiresAt) {
				delete(s.photoCache, k)
			}
		}
	}
	s.photoMu.Unlock()

	if url != "" {
		// 搜索结果直链带时效签名,必须立即代取落盘,否则前端稍后访问会 403
		if content, contentType, err := s.downloadImage(ctx, url); err == nil {
			s.cache.Write("kw:"+keyword, content, contentType)
		}
	}
	return url
}

func (s *XHSService) fetchPhotoURL(ctx context.Context, keyword string) string {
	// 搜图属尽力而为:某个 Cookie 被风控时自动换下一个,全部不可用则返回空
	url, err := withXHSCookie(s, ctx, func(cookie string) (string, error) {
		return s.fetchPhotoURLWithCookie(ctx, cookie, keyword)
	})
	// Cookie 未配置属预期状态(由设置页决定),不刷日志
	var notConfigured *XHSNotConfiguredError
	if err != nil && !errors.As(err, &notConfigured) {
		fmt.Printf("小红书单图抓取失败 (%s): %v\n", keyword, err)
	}
	return url
}

func (s *XHSService) fetchPhotoURLWithCookie(ctx context.Context, cookie, keyword string) (string, error) {
	// 搜图强制按"最新"排序,避开综合高赞的含文字攻略图
	resJSON, err := s.searchNotes(ctx, cookie, keyword, 1)
	if err != nil {
		return "", err
	}
	items := extractItems(resJSON)
	// 诊断日志:便于排查某关键词为何搜不到图(返回条数与类型分布)
	diagTypes := map[string]int{}
	for _, note := range items {
		mt, _ := note["model_type"].(string)
		diagTypes[mt]++
	}
	fmt.Printf("🔍 搜图诊断 (%s): items=%d types=%v\n", keyword, len(items), diagTypes)
	// 收集前若干条笔记候选:"最新"流里视频笔记没有图片列表,只看第一条经常空手而归
	type noteCand struct{ id, token string }
	var cands []noteCand
	for _, note := range items {
		if modelType, _ := note["model_type"].(string); modelType == "note" {
			id := stringValue(note["id"])
			if id == "" {
				continue
			}
			cands = append(cands, noteCand{id: id, token: stringValue(note["xsec_token"])})
			if len(cands) >= 8 {
				break
			}
		}
	}
	if len(cands) == 0 {
		return "", nil
	}
	for _, cand := range cands {

		// 方案 A: 原生 API 获取笔记图片
		if detail, derr := s.getNoteDetail(ctx, cookie, cand.id, cand.token); derr == nil {
			detailItems := extractItems(detail)
			if len(detailItems) > 0 {
				card, _ := detailItems[0]["note_card"].(map[string]any)
				if imageList, ok := card["image_list"].([]any); ok && len(imageList) > 0 {
					first, _ := imageList[0].(map[string]any)
					if first != nil {
						if infoList, ok := first["info_list"].([]any); ok && len(infoList) > 0 {
							pick := infoList[0]
							if len(infoList) > 1 {
								pick = infoList[1]
							}
							if m, ok := pick.(map[string]any); ok {
								if u := stringValue(m["url"]); u != "" {
									return u, nil
								}
							}
						}
						for _, key := range []string{"url_default", "url_pre", "url"} {
							if u := stringValue(first[key]); u != "" {
								return u, nil
							}
						}
					}
				}
			}
		}

		// 方案 B: SSR 抓取
		if ssr := s.getNoteDetailSSR(ctx, cand.id); ssr != nil {
			if imgList, ok := ssr["imageList"].([]any); ok && len(imgList) > 0 {
				if first, ok := imgList[0].(map[string]any); ok {
					for _, key := range []string{"urlDefault", "urlPattern", "url"} {
						if u := stringValue(first[key]); u != "" {
							return u, nil
						}
					}
				}
			}
		}
	}
	return "", nil
}

var allowedImageHosts = []string{"xiaohongshu.com", "xhscdn.com"}

func validateImageURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("图片 URL 解析失败: %v", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("不支持的图片 URL 协议: %s", parsed.Scheme)
	}
	host := strings.ToLower(parsed.Hostname())
	for _, suffix := range allowedImageHosts {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return nil
		}
	}
	return fmt.Errorf("仅允许代理小红书图片域名,收到: %s", host)
}

func (s *XHSService) downloadImage(ctx context.Context, rawURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", &XHSImageProxyError{Msg: err.Error()}
	}
	req.Header.Set("User-Agent", xhsBrowserUA)
	req.Header.Set("Referer", "https://www.xiaohongshu.com/")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", &XHSImageProxyError{Msg: fmt.Sprintf("图片下载失败: %v", err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", &XHSImageProxyError{Msg: fmt.Sprintf("图片下载返回 HTTP %d", resp.StatusCode)}
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, xhsImageMaxBytes+1))
	if err != nil {
		return nil, "", &XHSImageProxyError{Msg: fmt.Sprintf("图片读取失败: %v", err)}
	}
	contentType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	if contentType == "" {
		contentType = "image/jpeg"
	}
	if len(content) == 0 {
		return nil, "", &XHSImageProxyError{Msg: "图片内容为空"}
	}
	if len(content) > xhsImageMaxBytes {
		return nil, "", &XHSImageProxyError{Msg: "图片超过大小限制(10MB)"}
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", &XHSImageProxyError{Msg: fmt.Sprintf("响应不是图片 (content-type=%s)", contentType)}
	}
	return content, contentType, nil
}

// PhotoURL 返回景点图片直链(前端仅作兜底展示,推荐直接用 PhotoBytes 代理)。
// 与 PhotoBytes 相同的多级关键词回退:"最新"流下单词条经常搜不到带封面的笔记。
func (s *XHSService) PhotoURL(ctx context.Context, name, city string) string {
	for _, kw := range []string{name + " 风景", name + " 旅游", name, name + " 攻略"} {
		if u := s.photoURLFromXHS(ctx, kw); u != "" {
			return u
		}
	}
	return ""
}

// PhotoBytes 获取景点图片字节;缓存 miss 时自动重搜新直链并立即下载。
func (s *XHSService) PhotoBytes(ctx context.Context, name, city string) ([]byte, string, error) {
	// 多级关键词回退:"最新"流下单个词条经常搜不到带封面图的笔记,依次换词提高命中
	keywords := []string{name + " 风景", name + " 旅游", name, name + " 攻略"}
	kwCacheKey := func(kw string) string { return "kw:" + kw }
	for _, kw := range keywords {
		if content, contentType, ok := s.cache.Read(kwCacheKey(kw)); ok {
			return content, contentType, nil
		}
	}
	for attempt, kw := range keywords {
		rawURL := s.fetchPhotoURL(ctx, kw)
		if rawURL == "" {
			fmt.Printf("⚠️  景点图片搜索无结果(%d/%d): %s\n", attempt+1, len(keywords), kw)
			continue
		}
		if err := validateImageURL(rawURL); err != nil {
			continue
		}
		content, contentType, err := s.downloadImage(ctx, rawURL)
		if err == nil {
			s.cache.Write(kwCacheKey(kw), content, contentType)
			return content, contentType, nil
		}
		// 直链多为限时签名,失败后换下一个关键词重搜全新直链
		fmt.Printf("⚠️  图片下载失败(%s,第%d次,将换词重取): %v\n", kw, attempt+1, err)
	}
	return nil, "", errors.New("未能获取景点图片")
}

// FetchImageBytes 按直链代理小红书图片(URL 维度缓存,仅供白名单域名)。
func (s *XHSService) FetchImageBytes(ctx context.Context, rawURL string) ([]byte, string, error) {
	if err := validateImageURL(rawURL); err != nil {
		return nil, "", err
	}
	if content, contentType, ok := s.cache.Read("url:" + rawURL); ok {
		return content, contentType, nil
	}
	content, contentType, err := s.downloadImage(ctx, rawURL)
	if err != nil {
		return nil, "", err
	}
	s.cache.Write("url:"+rawURL, content, contentType)
	return content, contentType, nil
}
