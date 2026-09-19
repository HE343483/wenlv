package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"wenlv-backend/logger"
)

// ============ 抖音服务 ============
// 与小红的定位一致:搜索真人短视频/图文内容 -> LLM 提纯为结构化景点;景点图片走封面图
// "关键词磁盘缓存 + 后端代理"方案。抖音网页接口强制要求 a_bogus 签名,签名由 DouyinSigner
// 通过 Node 子进程生成(签名 JS 需自行放置到 douyin_sign/douyin.js)。

const (
	douyinBaseURL       = "https://www.douyin.com"
	douyinSearchAPI     = "/aweme/v1/web/general/search/single/"
	douyinImageMaxBytes = 10 * 1024 * 1024
	douyinUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"
	// 关键词→图片直链的内存缓存时长(命中/未命中分别处理)
	douyinPhotoCacheTTL     = 6 * time.Hour
	douyinPhotoCacheMissTTL = 10 * time.Minute
)

// allowedDouyinImageHosts 允许代理的抖音图片 CDN 域名。
var allowedDouyinImageHosts = []string{
	"douyinpic.com",
	"byteimg.com",
	"ibyteimg.com",
	"ibytedtos.com",
	"douyin.com",
}

// DouyinCookieExpiredError 抖音 Cookie 过期/被风控,需要向前端明确提示。
type DouyinCookieExpiredError struct{ Msg string }

func (e *DouyinCookieExpiredError) Error() string { return e.Msg }

// DouyinNotConfiguredError 抖音 Cookie 未配置;此时行程规划会降级使用地图 POI 兜底。
type DouyinNotConfiguredError struct{ Msg string }

func (e *DouyinNotConfiguredError) Error() string { return e.Msg }

// DouyinFetchError 抖音普通抓取异常。
type DouyinFetchError struct{ Msg string }

func (e *DouyinFetchError) Error() string { return e.Msg }

// DouyinImageProxyError 抖音图片代理抓取失败。
type DouyinImageProxyError struct{ Msg string }

func (e *DouyinImageProxyError) Error() string { return e.Msg }

// DouyinService 抖音服务。
type DouyinService struct {
	settings *TripSettings
	signer   *DouyinSigner
	llm      *TripLLM
	amap     *AmapService
	client   *http.Client
	cache    *imageDiskCache

	photoMu    sync.Mutex
	photoCache map[string]photoCacheEntry
}

// NewDouyinService 构造抖音服务。
func NewDouyinService(settings *TripSettings, llm *TripLLM, amap *AmapService, signer *DouyinSigner) *DouyinService {
	return &DouyinService{
		settings: settings,
		signer:   signer,
		llm:      llm,
		amap:     amap,
		client: &http.Client{
			Timeout:   15 * time.Second,
			Transport: &http.Transport{Proxy: nil},
		},
		cache:      newImageDiskCache(settings.DataDir(), settings.ImageCacheTTL(), settings.ImageCacheMaxBytes()),
		photoCache: map[string]photoCacheEntry{},
	}
}

// NormalizeDouyinCookie 兼容 Cookie 请求头字符串和浏览器导出的 JSON Cookie 列表。
func NormalizeDouyinCookie(cookie string) string {
	return NormalizeXHSCookie(cookie)
}

func (s *DouyinService) cookie() (string, error) {
	cookie := NormalizeDouyinCookie(s.settings.Snapshot().DouyinCookie)
	if cookie == "" {
		return "", &DouyinNotConfiguredError{Msg: "抖音 Cookie 未配置,请先在前端设置页完成配置"}
	}
	return cookie, nil
}

// searchParams 构造抖音搜索接口的查询参数(对齐抖音网页端真实请求)。
func douyinSearchParams(keyword string, offset int) url.Values {
	return url.Values{
		"device_platform":    {"webapp"},
		"aid":                {"6383"},
		"channel":            {"channel_pc_web"},
		"search_channel":     {"aweme_general"},
		"enable_history":     {"1"},
		"keyword":            {keyword},
		"search_source":      {"normal_search"},
		"query_correct_type": {"1"},
		"is_filter_search":   {"0"},
		"from_group_id":      {""},
		"offset":             {fmt.Sprintf("%d", offset)},
		"count":              {"10"},
		"pc_client_type":     {"1"},
		"version_code":       {"170400"},
		"version_name":       {"17.4.0"},
		"cookie_enabled":     {"true"},
		"screen_width":       {"1920"},
		"screen_height":      {"1080"},
		"browser_language":   {"zh-CN"},
		"browser_platform":   {"Win32"},
		"browser_name":       {"Chrome"},
		"browser_version":    {douyinWebVersion},
		"browser_online":     {"true"},
		"engine_name":        {"Blink"},
		"engine_version":     {douyinWebVersion},
		"os_name":            {"Windows"},
		"os_version":         {"10"},
		"cpu_core_num":       {"8"},
		"device_memory":      {"8"},
		"platform":           {"PC"},
		"downlink":           {"10"},
		"effective_type":     {"4g"},
		"round_trip_time":    {"50"},
	}
}

const douyinWebVersion = "121.0.0.0"

// signedGet 生成 a_bogus 签名并请求抖音接口。
func (s *DouyinService) signedGet(ctx context.Context, cookie string, api string, params url.Values) (map[string]any, error) {
	// msToken 以 Cookie 为准,同时作为查询参数回传(抖音网页端行为)
	if token := transCookies(cookie)["msToken"]; token != "" {
		params.Set("msToken", token)
	}
	query := params.Encode()
	aBogus, err := s.signer.SignABogus(ctx, api+"?"+query, douyinUserAgent)
	if err != nil {
		return nil, err
	}
	fullURL := douyinBaseURL + api + "?" + query + "&a_bogus=" + url.QueryEscape(aBogus)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, &DouyinFetchError{Msg: fmt.Sprintf("抖音请求构造失败: %v", err)}
	}
	headers := map[string]string{
		"accept":           "application/json, text/plain, */*",
		"accept-language":  "zh-CN,zh;q=0.9,en;q=0.8",
		"cache-control":    "no-cache",
		"pragma":           "no-cache",
		"referer":          "https://www.douyin.com/",
		"sec-fetch-dest":   "empty",
		"sec-fetch-mode":   "cors",
		"sec-fetch-site":   "same-origin",
		"user-agent":       douyinUserAgent,
		"x-requested-with": "XMLHttpRequest",
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", cookie)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &DouyinFetchError{Msg: fmt.Sprintf("访问抖音失败: %v", err)}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &DouyinFetchError{Msg: fmt.Sprintf("抖音响应读取失败: %v", err)}
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 461 {
		return nil, &DouyinCookieExpiredError{Msg: fmt.Sprintf("抖音请求被风控拦截 (HTTP %d)。请更换 Cookie 后重试。", resp.StatusCode)}
	}
	var res map[string]any
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, &DouyinFetchError{Msg: fmt.Sprintf("抖音响应解析失败: %s", truncateText(string(raw), 200))}
	}
	statusCode := flexToString(res["status_code"])
	if statusCode != "" && statusCode != "0" {
		msg := flexToString(res["status_msg"])
		// 8/2154 等为登录态/风控相关错误码
		if statusCode == "8" || statusCode == "2154" || statusCode == "10037" || strings.Contains(msg, "登录") {
			return nil, &DouyinCookieExpiredError{Msg: fmt.Sprintf("抖音 Cookie 已失效或被风控 (status_code=%s): %s。请更换 Cookie 后重试。", statusCode, msg)}
		}
		return nil, &DouyinFetchError{Msg: fmt.Sprintf("抖音接口失败 (status_code=%s): %s", statusCode, msg)}
	}
	return res, nil
}

// searchAwemes 关键词搜索,返回 aweme 列表(仅保留可用的短视频/图文内容)。
func (s *DouyinService) searchAwemes(ctx context.Context, cookie, keyword string, offset int) ([]map[string]any, error) {
	res, err := s.signedGet(ctx, cookie, douyinSearchAPI, douyinSearchParams(keyword, offset))
	if err != nil {
		return nil, err
	}
	rawItems, _ := res["data"].([]any)
	out := make([]map[string]any, 0, len(rawItems))
	for _, it := range rawItems {
		item, ok := it.(map[string]any)
		if !ok {
			continue
		}
		aweme, ok := item["aweme_info"].(map[string]any)
		if !ok || aweme == nil {
			continue
		}
		out = append(out, aweme)
	}
	return out, nil
}

// ============ 景点搜索(抖音 + LLM 提纯) ============

// SearchAttractionsText 搜索抖音真人分享并经 LLM 提纯为结构化景点文本。
func (s *DouyinService) SearchAttractionsText(ctx context.Context, city, keywords, language string) (string, error) {
	logger.Infof("[抖音] 正在搜索抖音: %s %s", city, keywords)
	cookie, err := s.cookie()
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf("%s %s 旅游 景点攻略", city, keywords)

	awemes, err := s.searchAwemes(ctx, cookie, query, 0)
	if err != nil {
		return "", err
	}
	if len(awemes) > 4 {
		awemes = awemes[:4]
	}

	var combined strings.Builder
	for i, aweme := range awemes {
		desc := stringValue(aweme["desc"])
		if strings.TrimSpace(desc) == "" {
			continue
		}
		author, _ := aweme["author"].(map[string]any)
		nickname := stringValue(author["nickname"])
		fmt.Fprintf(&combined, "\n内容%d:\n作者: %s\n正文内容: %s\n", i+1, nickname, desc)
	}

	if combined.Len() == 0 {
		return fmt.Sprintf("未在抖音检索到关于 %s %s 的内容。", city, keywords), nil
	}
	return extractAttractionsFromNotes(ctx, s.llm, s.settings, s.amap, city, combined.String(), language, "抖音"), nil
}

// ============ 景点搜图(封面图) ============

// photoURLFromDouyin 根据关键词从抖音搜索一张封面图直链(带内存缓存)。
func (s *DouyinService) photoURLFromDouyin(ctx context.Context, keyword string) string {
	s.photoMu.Lock()
	if entry, ok := s.photoCache[keyword]; ok && time.Now().Before(entry.expiresAt) {
		s.photoMu.Unlock()
		return entry.url
	}
	s.photoMu.Unlock()

	url := s.fetchPhotoURL(ctx, keyword)

	s.photoMu.Lock()
	ttl := douyinPhotoCacheTTL
	if url == "" {
		ttl = douyinPhotoCacheMissTTL
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
		// 抖音封面直链带时效签名,必须立即代取落盘,否则前端稍后访问会 403
		if content, contentType, err := s.downloadImage(ctx, url); err == nil {
			s.cache.Write("dy:kw:"+keyword, content, contentType)
		}
	}
	return url
}

func (s *DouyinService) fetchPhotoURL(ctx context.Context, keyword string) string {
	cookie, err := s.cookie()
	if err != nil {
		return ""
	}
	awemes, err := s.searchAwemes(ctx, cookie, keyword, 0)
	if err != nil {
		logger.Warnf("抖音单图抓取失败 (%s): %v", keyword, err)
		return ""
	}
	logger.Infof("抖音搜图诊断 (%s): items=%d", keyword, len(awemes))
	for _, aweme := range awemes {
		if u := firstImageURL(aweme); u != "" {
			return u
		}
	}
	return ""
}

// firstImageURL 从 aweme 数据中提取首图:图文帖优先取 images,视频帖取封面。
func firstImageURL(aweme map[string]any) string {
	if images, ok := aweme["images"].([]any); ok && len(images) > 0 {
		if first, ok := images[0].(map[string]any); ok {
			if u := firstURLFromList(first["url_list"]); u != "" {
				return u
			}
		}
	}
	video, _ := aweme["video"].(map[string]any)
	if video == nil {
		return ""
	}
	for _, key := range []string{"origin_cover", "cover", "dynamic_cover"} {
		cover, _ := video[key].(map[string]any)
		if cover == nil {
			continue
		}
		if u := firstURLFromList(cover["url_list"]); u != "" {
			return u
		}
	}
	return ""
}

func firstURLFromList(v any) string {
	list, _ := v.([]any)
	for _, it := range list {
		if u := stringValue(it); u != "" {
			return u
		}
	}
	return ""
}

func validateDouyinImageURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("图片 URL 解析失败: %v", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("不支持的图片 URL 协议: %s", parsed.Scheme)
	}
	host := strings.ToLower(parsed.Hostname())
	for _, suffix := range allowedDouyinImageHosts {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return nil
		}
	}
	return fmt.Errorf("仅允许代理抖音图片域名,收到: %s", host)
}

func (s *DouyinService) downloadImage(ctx context.Context, rawURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", &DouyinImageProxyError{Msg: err.Error()}
	}
	req.Header.Set("User-Agent", douyinUserAgent)
	req.Header.Set("Referer", "https://www.douyin.com/")
	if cookie, cerr := s.cookie(); cerr == nil {
		req.Header.Set("Cookie", cookie)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", &DouyinImageProxyError{Msg: fmt.Sprintf("图片下载失败: %v", err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", &DouyinImageProxyError{Msg: fmt.Sprintf("图片下载返回 HTTP %d", resp.StatusCode)}
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, douyinImageMaxBytes+1))
	if err != nil {
		return nil, "", &DouyinImageProxyError{Msg: fmt.Sprintf("图片读取失败: %v", err)}
	}
	contentType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
	if contentType == "" {
		contentType = "image/jpeg"
	}
	if len(content) == 0 {
		return nil, "", &DouyinImageProxyError{Msg: "图片内容为空"}
	}
	if len(content) > douyinImageMaxBytes {
		return nil, "", &DouyinImageProxyError{Msg: "图片超过大小限制(10MB)"}
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", &DouyinImageProxyError{Msg: fmt.Sprintf("响应不是图片 (content-type=%s)", contentType)}
	}
	return content, contentType, nil
}

// PhotoURL 返回景点图片直链(前端仅作兜底展示,推荐直接用 PhotoBytes 代理)。
func (s *DouyinService) PhotoURL(ctx context.Context, name, city string) string {
	for _, kw := range []string{name + " 风景", name + " 旅游", name, name + " 攻略"} {
		if u := s.photoURLFromDouyin(ctx, kw); u != "" {
			return u
		}
	}
	return ""
}

// PhotoBytes 获取景点图片字节;缓存 miss 时自动重搜新直链并立即下载。
func (s *DouyinService) PhotoBytes(ctx context.Context, name, city string) ([]byte, string, error) {
	keywords := []string{name + " 风景", name + " 旅游", name, name + " 攻略"}
	kwCacheKey := func(kw string) string { return "dy:kw:" + kw }
	for _, kw := range keywords {
		if content, contentType, ok := s.cache.Read(kwCacheKey(kw)); ok {
			return content, contentType, nil
		}
	}
	for attempt, kw := range keywords {
		rawURL := s.fetchPhotoURL(ctx, kw)
		if rawURL == "" {
			logger.Warnf("抖音景点图片搜索无结果(%d/%d): %s", attempt+1, len(keywords), kw)
			continue
		}
		if err := validateDouyinImageURL(rawURL); err != nil {
			continue
		}
		content, contentType, err := s.downloadImage(ctx, rawURL)
		if err == nil {
			s.cache.Write(kwCacheKey(kw), content, contentType)
			return content, contentType, nil
		}
		logger.Warnf("抖音图片下载失败(%s,第%d次,将换词重取): %v", kw, attempt+1, err)
	}
	return nil, "", errors.New("未能获取景点图片")
}

// FetchImageBytes 按直链代理抖音图片(URL 维度缓存,仅供白名单域名)。
func (s *DouyinService) FetchImageBytes(ctx context.Context, rawURL string) ([]byte, string, error) {
	if err := validateDouyinImageURL(rawURL); err != nil {
		return nil, "", err
	}
	if content, contentType, ok := s.cache.Read("dy:url:" + rawURL); ok {
		return content, contentType, nil
	}
	content, contentType, err := s.downloadImage(ctx, rawURL)
	if err != nil {
		return nil, "", err
	}
	s.cache.Write("dy:url:"+rawURL, content, contentType)
	return content, contentType, nil
}

// flexToString 将任意 JSON 标量转成字符串(抖音 status_code 有时是数字有时是字符串)。
func flexToString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", t)
	}
}
