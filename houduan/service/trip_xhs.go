package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"wenlv-backend/model"
)

// ============ 小红书服务 ============
// 移植自 TripStar 的 xhs_service:原生签名直连 edith.xiaohongshu.com,
// 失败时降级到网页 SSR 抓取;景点图片走"关键词磁盘缓存 + 后端代理"方案。

const (
	xhsBaseURL           = "https://edith.xiaohongshu.com"
	xhsSearchAPI         = "/api/sns/web/v1/search/notes"
	xhsFeedAPI           = "/api/sns/web/v1/feed"
	xhsImageCacheTTL     = 24 * time.Hour
	xhsImageMaxBytes     = 10 * 1024 * 1024
	xhsUserAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"
	xhsBrowserUA         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
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

	photoMu    sync.Mutex
	photoCache map[string]photoCacheEntry

	imgMu sync.Mutex
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

func (s *XHSService) cookie() (string, error) {
	cookie := NormalizeXHSCookie(s.settings.Snapshot().XHSCookie)
	if cookie == "" {
		return "", &XHSNotConfiguredError{Msg: "小红书 Cookie 未配置,请先在前端设置页完成配置"}
	}
	return cookie, nil
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
		"authority":           "edith.xiaohongshu.com",
		"accept":              "application/json, text/plain, */*",
		"accept-language":     "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"cache-control":       "no-cache",
		"content-type":        "application/json;charset=UTF-8",
		"origin":              "https://www.xiaohongshu.com",
		"pragma":              "no-cache",
		"referer":             "https://www.xiaohongshu.com/",
		"sec-ch-ua":           `"Not A(Brand";v="99", "Microsoft Edge";v="121", "Chromium";v="121"`,
		"sec-ch-ua-mobile":    "?0",
		"sec-ch-ua-platform":  `"Windows"`,
		"sec-fetch-dest":      "empty",
		"sec-fetch-mode":      "cors",
		"sec-fetch-site":      "same-site",
		"user-agent":          xhsUserAgent,
		"x-b3-traceid":        randomHex(16),
		"x-mns":               "unload",
		"x-s":                 xs,
		"x-s-common":          xsCommon,
		"x-t":                 fmt.Sprintf("%d", xt),
		"x-xray-traceid":      traceID,
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
		if code == "300011" || strings.Contains(msg, "异常") {
			return nil, &XHSCookieExpiredError{Msg: fmt.Sprintf("小红书 Cookie 已被风控拦截 (code=%s): %s。请更换 Cookie 后重试。", code, msg)}
		}
		return nil, &XHSFetchError{Msg: fmt.Sprintf("小红书接口失败 (code=%s): %s", code, msg)}
	}
	return res, nil
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
		"keyword":    keyword,
		"page":       1,
		"page_size":  20,
		"search_id":  randomHex(21),
		"sort":       "general",
		"note_type":  0,
		"ext_flags":  []any{},
		"filters": []map[string]any{
			{"tags": []string{sort}, "type": "sort_type"},
			{"tags": []string{"不限"}, "type": "filter_note_type"},
			{"tags": []string{"不限"}, "type": "filter_note_time"},
			{"tags": []string{"不限"}, "type": "filter_note_range"},
			{"tags": []string{"不限"}, "type": "filter_pos_distance"},
		},
		"geo":            "",
		"image_formats":  []string{"jpg", "webp", "avif"},
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
	cookie, err := s.cookie()
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf("%s %s 旅游 景点攻略", city, keywords)

	resJSON, err := s.searchNotes(ctx, cookie, query, 0)
	if err != nil {
		return "", err
	}
	items := extractItems(resJSON)
	if len(items) > 4 {
		items = items[:4]
	}

	var combined strings.Builder
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
		fmt.Fprintf(&combined, "\n笔记%d:\n标题: %s\n正文内容: %s\n", i+1, title, desc)
	}

	if combined.Len() == 0 {
		return fmt.Sprintf("未在小红书检索到关于 %s %s 的内容。", city, keywords), nil
	}

	return s.extractAttractions(ctx, city, combined.String(), language), nil
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
func (s *XHSService) extractAttractions(ctx context.Context, city, content, language string) string {
	lang := model.NormalizeLang(language)
	translation := ""
	if name, ok := langNames[lang]; ok && lang != "zh" {
		translation = fmt.Sprintf(`
**极其重要的翻译要求:**
目标语言为 %s。你必须将提取结果中的 "name", "reason", "reservation_tips" 字段的内容翻译为 %s。
- "name" 字段使用目标语言 %s 的景点名称(例如中文"故宫博物院" → English "The Palace Museum")。
- "reason" 和 "reservation_tips" 也必须翻译为 %s。
- "duration" 和 "reservation_required" 保持原始数值/布尔值不变。
- **注意**: "name_zh" 必须始终保持简体中文名称,"name_en" 必须始终保持英文名称,不受目标语言影响!
- 严格保持 JSON schema 格式不变!
`, name, name, name, name)
	}

	prompt := fmt.Sprintf(`请从以下真实的素人小红书打卡游记中,提纯出真实存在的【游玩景点】。
要求返回严格的 JSON 数组格式(哪怕只提取到了1个),切勿返回除了JSON以外的任何冗余 markdown 文字!
%s
数组中每个对象必须包含以下字段:
"name": 景点官方名称(用于前端展示,按目标语言填写;若目标语言为中文则与 name_zh 相同)
"name_zh": 景点的中文简体名称(必须是简体中文,例如 "故宫博物院"。此字段始终为中文,不受目标语言影响)
"name_en": 景点的英文名称(必须是英文,使用景点在国际上通用的官方英文名。此字段始终为英文,不受目标语言影响)
"reason": 小红书用户的真实评价/避坑指南
"duration": 游玩时长(数字, 分钟)
"reservation_required": 是否需要提前预约(布尔值 true/false)。请根据游记中提到的"需要预约"、"提前预约"、"抢票"、"约满"、"官方预约"等关键词判断,如果游记未提及则默认为 false
"reservation_tips": 预约相关提示(字符串)。如果需要预约,请提取预约渠道、提前天数等具体信息;如果不需要预约则填空字符串

游记杂文内容如下:
%s

JSON 返回示例:
[{"name": "故宫博物院", "name_zh": "故宫博物院", "name_en": "The Palace Museum", "reason": "必去打卡,建议走中轴线。", "duration": 240, "reservation_required": true, "reservation_tips": "需要提前7天在故宫官网或微信小程序预约"},
 {"name": "老君山金顶", "name_zh": "老君山金顶", "name_en": "Laojun Mountain Golden Summit", "reason": "网红打卡点,夜景绝美。", "duration": 180, "reservation_required": false, "reservation_tips": ""}]
`, translation, content)

	reply, err := s.llm.Chat(ctx, UserMessage(prompt), 0.1, 4000)
	if err != nil {
		fmt.Printf(" 大模型提纯小红书数据异常: %v\n", err)
		return "尝试提取小红书结构化数据失败,降级回常规处理。"
	}
	jsonText := extractJSONArray(reply)
	if jsonText == "" {
		return "尝试提取小红书结构化数据失败,降级回常规处理。"
	}
	var extracted []map[string]any
	if err := json.Unmarshal([]byte(jsonText), &extracted); err != nil {
		fmt.Printf("❌ 小红书提纯 JSON 解析失败: %v\n", err)
		return "尝试提取小红书结构化数据失败,降级回常规处理。"
	}

	valid := make([]map[string]any, 0, len(extracted))
	for _, item := range extracted {
		if stringValue(item["name"]) != "" {
			valid = append(valid, item)
		}
	}
	if len(valid) == 0 {
		return fmt.Sprintf("未在小红书检索到关于 %s 的有效景点信息。", city)
	}

	// 并发补齐经纬度(最多 3 个并发)
	locations := make([]*model.Location, len(valid))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)
	for i, item := range valid {
		wg.Add(1)
		go func(idx int, it map[string]any) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			name := stringValue(it["name"])
			locations[idx] = GeocodeUnified(ctx, s.settings, s.amap, name, city,
				firstNonEmpty(stringValue(it["name_zh"]), name),
				firstNonEmpty(stringValue(it["name_en"]), name))
		}(i, item)
	}
	wg.Wait()

	var out strings.Builder
	out.WriteString("这是小红书热门精选游记的提取结果,附带确切坐标(图片由前端单独搜索获取):\n")
	for i, item := range valid {
		if loc := locations[i]; loc != nil {
			item["location"] = map[string]float64{"longitude": loc.Longitude, "latitude": loc.Latitude}
		}
		line, err := json.Marshal(item)
		if err == nil {
			out.Write(line)
			out.WriteString("\n")
		}
	}
	fmt.Println("✅ [XHS] 小红书数据挖掘完毕,已装载进上下文。")
	return out.String()
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
			s.writeImageCache("kw:"+keyword, content, contentType)
		}
	}
	return url
}

func (s *XHSService) fetchPhotoURL(ctx context.Context, keyword string) string {
	cookie, err := s.cookie()
	if err != nil {
		return ""
	}
	// 搜图强制按"最新"排序,避开综合高赞的含文字攻略图
	resJSON, err := s.searchNotes(ctx, cookie, keyword, 1)
	if err != nil {
		fmt.Printf("小红书单图抓取失败 (%s): %v\n", keyword, err)
		return ""
	}
	items := extractItems(resJSON)
	targetID, targetToken := "", ""
	for _, note := range items {
		if modelType, _ := note["model_type"].(string); modelType == "note" {
			targetID = stringValue(note["id"])
			targetToken = stringValue(note["xsec_token"])
			break
		}
	}
	if targetID == "" {
		return ""
	}

	// 方案 A: 原生 API 获取笔记图片
	if detail, derr := s.getNoteDetail(ctx, cookie, targetID, targetToken); derr == nil {
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
								return u
							}
						}
					}
					for _, key := range []string{"url_default", "url_pre", "url"} {
						if u := stringValue(first[key]); u != "" {
							return u
						}
					}
				}
			}
		}
	}

	// 方案 B: SSR 抓取
	if ssr := s.getNoteDetailSSR(ctx, targetID); ssr != nil {
		if imgList, ok := ssr["imageList"].([]any); ok && len(imgList) > 0 {
			if first, ok := imgList[0].(map[string]any); ok {
				for _, key := range []string{"urlDefault", "urlPattern", "url"} {
					if u := stringValue(first[key]); u != "" {
						return u
					}
				}
			}
		}
	}
	return ""
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

func (s *XHSService) imageCachePaths(cacheKey string) (string, string) {
	sum := sha256.Sum256([]byte(cacheKey))
	digest := hex.EncodeToString(sum[:])
	dir := filepath.Join(s.settings.DataDir(), "photo_cache")
	return filepath.Join(dir, digest+".img"), filepath.Join(dir, digest+".meta")
}

func (s *XHSService) readImageCache(cacheKey string) ([]byte, string, bool) {
	imgPath, metaPath := s.imageCachePaths(cacheKey)
	info, err := os.Stat(imgPath)
	if err != nil || time.Since(info.ModTime()) >= xhsImageCacheTTL {
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

func (s *XHSService) writeImageCache(cacheKey string, content []byte, contentType string) {
	imgPath, metaPath := s.imageCachePaths(cacheKey)
	s.imgMu.Lock()
	defer s.imgMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(imgPath), 0o755); err != nil {
		return
	}
	if err := os.WriteFile(imgPath+".tmp", content, 0o644); err != nil {
		return
	}
	if err := os.Rename(imgPath+".tmp", imgPath); err != nil {
		return
	}
	_ = os.WriteFile(metaPath+".tmp", []byte(contentType), 0o644)
	_ = os.Rename(metaPath+".tmp", metaPath)
}

// PhotoURL 返回景点图片直链(前端仅作兜底展示,推荐直接用 PhotoBytes 代理)。
func (s *XHSService) PhotoURL(ctx context.Context, name, city string) string {
	keyword := name + " 风景"
	return s.photoURLFromXHS(ctx, keyword)
}

// PhotoBytes 获取景点图片字节;缓存 miss 时自动重搜新直链并立即下载。
func (s *XHSService) PhotoBytes(ctx context.Context, name, city string) ([]byte, string, error) {
	keyword := name + " 风景"
	if content, contentType, ok := s.readImageCache("kw:" + keyword); ok {
		return content, contentType, nil
	}
	for attempt := 0; attempt < 2; attempt++ {
		rawURL := s.fetchPhotoURL(ctx, keyword)
		if rawURL == "" {
			return nil, "", errors.New("未能获取景点图片")
		}
		if err := validateImageURL(rawURL); err != nil {
			return nil, "", err
		}
		content, contentType, err := s.downloadImage(ctx, rawURL)
		if err == nil {
			s.writeImageCache("kw:"+keyword, content, contentType)
			return content, contentType, nil
		}
		// 直链多为限时签名,失败后重搜一条全新直链再试一次
		fmt.Printf("⚠️  图片下载失败(第%d次,将重取新直链): %v\n", attempt+1, err)
	}
	return nil, "", errors.New("未能获取景点图片")
}

// FetchImageBytes 按直链代理小红书图片(URL 维度缓存,仅供白名单域名)。
func (s *XHSService) FetchImageBytes(ctx context.Context, rawURL string) ([]byte, string, error) {
	if err := validateImageURL(rawURL); err != nil {
		return nil, "", err
	}
	if content, contentType, ok := s.readImageCache("url:" + rawURL); ok {
		return content, contentType, nil
	}
	content, contentType, err := s.downloadImage(ctx, rawURL)
	if err != nil {
		return nil, "", err
	}
	s.writeImageCache("url:"+rawURL, content, contentType)
	return content, contentType, nil
}