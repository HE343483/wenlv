// 文旅热点服务:定时抓取四川省文化和旅游厅"行业动态"新闻,
// 提取标题/摘要/发布时间入库 hot_topics,并用 LLM 翻译英/日文标题摘要。
// 说明:成都市文广旅局官网等本地源存在 WAF(412),暂不可抓,主源采用省厅栏目。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// 爬虫相关常量。
const (
	// hotTopicListURL 省文旅厅"行业动态"列表页。
	hotTopicListURL = "https://wlt.sc.gov.cn/scwlt/hydt/government_open.shtml"
	// hotTopicSourceZH / hotTopicSourceEN 文章来源展示名。
	hotTopicSourceZH = "四川文旅厅"
	hotTopicSourceEN = "Sichuan Culture & Tourism"
	// hotTopicInterval 抓取周期(6 小时)。
	hotTopicInterval = 6 * time.Hour
	// hotTopicMaxListPages 单轮最多抓取的列表分页数(每页约 10 条,10 页 ≈ 100 条安全网)。
	hotTopicMaxListPages = 10
	// hotTopicMaxNewPerRun 单轮最多抓取入库的新文章数,防止首轮请求过多。
	hotTopicMaxNewPerRun = 100
	// hotTopicMaxTranslatePerRun 单轮最多执行 LLM 翻译的文章数,控制调用成本。
	hotTopicMaxTranslatePerRun = 30
	// hotTopicFetchDelay 详情页请求间隔,防止触发限流。
	hotTopicFetchDelay = 1500 * time.Millisecond
	// hotTopicCacheTTL 热点列表 Redis 缓存有效期(24 小时)。
	hotTopicCacheTTL = 24 * time.Hour
)

// hotTopicUA 官网对无 UA 请求可能拒绝,统一模拟浏览器。
const hotTopicUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// chengduKeywords 标题命中任意关键词则标记为成都相关热点。
var chengduKeywords = []string{
	"成都", "蓉城", "天府", "锦江", "锦官城", "熊猫", "都江堰", "青城山",
	"三星堆", "金沙", "宽窄巷子", "锦里", "武侯祠", "杜甫草堂", "春熙路",
	"太古里", "大运", "洛带", "黄龙溪", "青羊", "金牛", "成华", "龙泉驿",
	"双流", "温江", "郫都", "新津", "简阳", "都江堰", "彭州", "崇州", "邛崃",
	"成渝", "蓉",
}

// 列表页条目解析:href 为 /scwlt/hydt/年/月/日/哈希.shtml,title 属性为完整标题。
var (
	hotTopicListRe = regexp.MustCompile(`<a href="(/scwlt/hydt/\d{4}/\d{1,2}/\d{1,2}/[^"]+\.shtml)"[^>]*title='([^']+)'`)
	// 详情页发布时间:发布时间：2026-09-18
	hotTopicPubDateRe = regexp.MustCompile(`发布时间[:：]\s*(\d{4}-\d{1,2}-\d{1,2})`)
	// 详情页正文段落
	hotTopicParaRe = regexp.MustCompile(`<p[^>]*>([\s\S]*?)</p>`)
	// HTML 标签
	hotTopicTagRe = regexp.MustCompile(`<[^>]+>`)
)

// htmlEntities 常见 HTML 实体还原。
var htmlEntities = strings.NewReplacer(
	"&ensp;", " ", "&emsp;", " ", "&nbsp;", " ",
	"&lt;", "<", "&gt;", ">", "&amp;", "&", "&quot;", "\"", "&#39;", "'",
)

// HotTopicService 文旅热点抓取与查询。
type HotTopicService struct {
	repo *repository.HotTopicRepo
	llm  *TripLLM
	http *http.Client
	rdb  *redis.Client
}

// NewHotTopicService 构造热点服务(禁用系统代理,官网为国内直连;redis 为空时不启用缓存)。
func NewHotTopicService(repo *repository.HotTopicRepo, llm *TripLLM, rdb *redis.Client) *HotTopicService {
	return &HotTopicService{
		repo: repo,
		llm:  llm,
		http: &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 20 * time.Second},
		rdb:  rdb,
	}
}

// Start 启动后台定时抓取:启动 10 秒后先跑一轮,之后每 6 小时一轮。
func (s *HotTopicService) Start() {
	go func() {
		time.Sleep(10 * time.Second)
		s.CrawlOnce()
		for range time.Tick(hotTopicInterval) {
			s.CrawlOnce()
		}
	}()
}

// HotTopicItem 前端展示结构。
type HotTopicItem struct {
	ID          uint      `json:"id"`
	TitleZH     string    `json:"title_zh"`
	TitleEN     string    `json:"title_en"`
	TitleJA     string    `json:"title_ja"`
	SummaryZH   string    `json:"summary_zh"`
	SummaryEN   string    `json:"summary_en"`
	SummaryJA   string    `json:"summary_ja"`
	SourceZH    string    `json:"source_zh"`
	SourceEN    string    `json:"source_en"`
	URL         string    `json:"url"`
	Hot         bool      `json:"hot"`
	PublishedAt time.Time `json:"published_at"`
}

// hotTopicCacheKey 热点列表缓存键(按分页参数区分)。
func hotTopicCacheKey(page, pageSize int) string {
	return fmt.Sprintf("wenlv:hot_topics:list:%d:p%d", pageSize, page)
}

// HotTopicPage 分页结果。
type HotTopicPage struct {
	Items    []HotTopicItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ListPage 分页返回热点列表:优先读 Redis 缓存(TTL 24 小时),未命中回源 MySQL 并写缓存。
func (s *HotTopicService) ListPage(page, pageSize int) (*HotTopicPage, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 6
	}
	ctx := context.Background()
	key := hotTopicCacheKey(page, pageSize)
	// 读缓存
	if s.rdb != nil {
		if cached, err := s.rdb.Get(ctx, key).Result(); err == nil {
			var p HotTopicPage
			if err := json.Unmarshal([]byte(cached), &p); err == nil && len(p.Items) > 0 {
				return &p, nil
			}
		} else if err != redis.Nil {
			log.Printf("[热点缓存] 读取失败(回源数据库): %v", err)
		}
	}
	p, err := s.listPageFromDB(page, pageSize)
	if err != nil {
		return nil, err
	}
	// 写缓存(Redis 故障不影响正常返回)
	if s.rdb != nil {
		if data, err := json.Marshal(p); err == nil {
			if err := s.rdb.Set(ctx, key, data, hotTopicCacheTTL).Err(); err != nil {
				log.Printf("[热点缓存] 写入失败: %v", err)
			}
		}
	}
	return p, nil
}

// listPageFromDB 回源 MySQL 分页查询。
func (s *HotTopicService) listPageFromDB(page, pageSize int) (*HotTopicPage, error) {
	items, total, err := s.repo.ListPage(page, pageSize)
	if err != nil {
		return nil, err
	}
	out := make([]HotTopicItem, 0, len(items))
	for _, it := range items {
		out = append(out, HotTopicItem{
			ID: it.ID, TitleZH: it.TitleZH, TitleEN: it.TitleEN, TitleJA: it.TitleJA,
			SummaryZH: it.SummaryZH, SummaryEN: it.SummaryEN, SummaryJA: it.SummaryJA,
			SourceZH: it.SourceZH, SourceEN: it.SourceEN, URL: it.URL, Hot: it.Hot,
			PublishedAt: it.PublishedAt,
		})
	}
	return &HotTopicPage{Items: out, Total: total, Page: page, PageSize: pageSize}, nil
}

// invalidateCache 抓取到新文章后清除热点列表缓存,避免 24 小时内看不到新内容。
func (s *HotTopicService) invalidateCache() {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	keys, err := s.rdb.Keys(ctx, "wenlv:hot_topics:list:*").Result()
	if err != nil {
		log.Printf("[热点缓存] 查找缓存键失败: %v", err)
		return
	}
	if len(keys) > 0 {
		if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
			log.Printf("[热点缓存] 清除失败: %v", err)
		}
	}
}

// hotTopicEntry 列表页解析出的单条条目。
type hotTopicEntry struct {
	path  string // 详情页路径(相对)
	url   string // 详情页绝对链接
	title string // 标题(中文)
	date  time.Time // 从 URL 提取的发布日期
}

// CrawlOnce 抓取一轮:遍历列表页所有分页 → 过滤已入库 → 抓详情补摘要/时间 → 翻译 → 落库。
func (s *HotTopicService) CrawlOnce() {
	log.Printf("[热点抓取] 开始抓取 %s", hotTopicListURL)
	entries, err := s.parseListPages(hotTopicListURL)
	if err != nil {
		log.Printf("[热点抓取] 列表页抓取失败: %v", err)
		return
	}
	fetched := 0
	for _, e := range entries {
		if fetched >= hotTopicMaxNewPerRun {
			break
		}
		exists, err := s.repo.ExistsByURL(e.url)
		if err != nil {
			log.Printf("[热点抓取] 查重失败 %s: %v", e.url, err)
			continue
		}
		if exists {
			continue
		}
		// 每轮只对最新的一批文章做 LLM 翻译,控制调用成本
		translate := fetched < hotTopicMaxTranslatePerRun
		if err := s.crawlArticle(e, translate); err != nil {
			log.Printf("[热点抓取] 文章抓取失败 %s: %v", e.url, err)
			continue
		}
		fetched++
		time.Sleep(hotTopicFetchDelay)
	}
	log.Printf("[热点抓取] 本轮完成,列表 %d 条,新增 %d 条", len(entries), fetched)
	// 有新文章入库时清除列表缓存,让前端立即看到最新内容
	if fetched > 0 {
		s.invalidateCache()
	}
}

// parseListPages 依次抓取列表页第 1 页及后续分页(government_open_{n}.shtml),
// 直到分页不存在(404)或无条目为止,合并后按 URL 日期倒序返回。
func (s *HotTopicService) parseListPages(firstPageURL string) ([]hotTopicEntry, error) {
	entries := make([]hotTopicEntry, 0, 64)
	for page := 1; page <= hotTopicMaxListPages; page++ {
		url := firstPageURL
		if page > 1 {
			url = strings.Replace(firstPageURL, ".shtml", fmt.Sprintf("_%d.shtml", page), 1)
		}
		html, err := s.fetch(url)
		if err != nil {
			if page == 1 {
				return nil, err // 首页失败直接报错
			}
			log.Printf("[热点抓取] 第 %d 页抓取结束: %v", page, err)
			break // 后续页 404/超时视为抓取完毕
		}
		pageEntries := s.parseList(html)
		if len(pageEntries) == 0 {
			break
		}
		entries = append(entries, pageEntries...)
		time.Sleep(hotTopicFetchDelay) // 列表页请求间隔
	}
	// 倒序:最新优先
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].date.After(entries[i].date) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return entries, nil
}

// parseList 解析单个列表页 HTML。
func (s *HotTopicService) parseList(html string) []hotTopicEntry {
	matches := hotTopicListRe.FindAllStringSubmatch(html, -1)
	entries := make([]hotTopicEntry, 0, len(matches))
	for _, m := range matches {
		path, title := m[1], htmlEntities.Replace(m[2])
		entry := hotTopicEntry{
			path:  path,
			url:   "https://wlt.sc.gov.cn" + path,
			title: strings.TrimSpace(title),
			date:  dateFromPath(path),
		}
		if entry.title == "" || entry.date.IsZero() {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

// crawlArticle 抓取单篇文章详情,组装并落库;translate=false 时跳过 LLM 翻译。
func (s *HotTopicService) crawlArticle(e hotTopicEntry, translate bool) error {
	topic := &model.HotTopic{
		TitleZH:     e.title,
		SourceZH:    hotTopicSourceZH,
		SourceEN:    hotTopicSourceEN,
		URL:         e.url,
		Hot:         matchChengdu(e.title),
		PublishedAt: e.date,
	}
	// 详情页:提取发布时间与正文摘要
	if html, err := s.fetch(e.url); err == nil {
		if m := hotTopicPubDateRe.FindStringSubmatch(html); m != nil {
			if t, err := time.ParseInLocation("2006-1-2", m[1], time.Local); err == nil {
				topic.PublishedAt = t
			}
		}
		topic.SummaryZH = extractSummary(html)
	}
	// LLM 翻译英/日文(未配置或失败时留空,前端回退中文)
	if translate {
		s.translateTopic(topic)
	}
	if err := s.repo.Upsert(topic); err != nil {
		return err
	}
	log.Printf("[热点抓取] 已入库: %s", topic.TitleZH)
	return nil
}

// fetch 请求页面并返回 UTF-8 文本。
func (s *HotTopicService) fetch(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", hotTopicUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 上限 2MB
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// extractSummary 从详情页提取正文摘要:取第一个 30~200 字的段落。
func extractSummary(html string) string {
	for _, m := range hotTopicParaRe.FindAllStringSubmatch(html, -1) {
		text := strings.TrimSpace(htmlEntities.Replace(hotTopicTagRe.ReplaceAllString(m[1], "")))
		text = strings.Join(strings.Fields(text), " ")
		runeLen := len([]rune(text))
		if runeLen >= 30 && runeLen <= 200 {
			return text
		}
	}
	return ""
}

// matchChengdu 判断标题是否命中成都相关关键词。
func matchChengdu(title string) bool {
	for _, kw := range chengduKeywords {
		if strings.Contains(title, kw) {
			return true
		}
	}
	return false
}

// dateFromPath 从 /scwlt/hydt/2026/9/18/xxx.shtml 提取日期。
func dateFromPath(path string) time.Time {
	parts := strings.Split(path, "/")
	if len(parts) < 7 {
		return time.Time{}
	}
	y, err1 := strconv.Atoi(parts[3])
	m, err2 := strconv.Atoi(parts[4])
	d, err3 := strconv.Atoi(parts[5])
	if err1 != nil || err2 != nil || err3 != nil {
		return time.Time{}
	}
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
}

// translateTopic 用 LLM 将标题与摘要翻译为英/日文,单次调用,失败静默跳过。
func (s *HotTopicService) translateTopic(topic *model.HotTopic) {
	if s.llm == nil || !s.llm.Available() || topic.TitleZH == "" {
		return
	}
	prompt := fmt.Sprintf(
		"将下面的中文文旅新闻标题和摘要分别翻译为英文和日文,只返回 JSON,格式:"+
			`{"title_en":"...","summary_en":"...","title_ja":"...","summary_ja":"..."}。`+"\n标题:%s\n摘要:%s",
		topic.TitleZH, topic.SummaryZH,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resp, err := s.llm.Chat(ctx, []llmMessage{
		{Role: "system", Content: "你是专业的新闻翻译,输出严格的 JSON,不要包含代码块标记。"},
		{Role: "user", Content: prompt},
	}, 0.2, 1024)
	if err != nil {
		log.Printf("[热点抓取] LLM 翻译失败(保留中文): %v", err)
		return
	}
	if strings.TrimSpace(resp) == "" {
		log.Printf("[热点抓取] LLM 翻译返回为空,跳过翻译(保留中文)")
		return
	}
	var tr struct {
		TitleEN   string `json:"title_en"`
		SummaryEN string `json:"summary_en"`
		TitleJA   string `json:"title_ja"`
		SummaryJA string `json:"summary_ja"`
	}
	// 兼容模型输出带代码块/多余文本的情况
	if err := json.Unmarshal([]byte(extractJSONObjectText(resp, resp)), &tr); err != nil {
		log.Printf("[热点抓取] LLM 翻译结果解析失败: %v", err)
		return
	}
	topic.TitleEN = strings.TrimSpace(tr.TitleEN)
	topic.SummaryEN = strings.TrimSpace(tr.SummaryEN)
	topic.TitleJA = strings.TrimSpace(tr.TitleJA)
	topic.SummaryJA = strings.TrimSpace(tr.SummaryJA)
}
