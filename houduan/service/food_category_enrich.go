// 美食大类(川菜/名小吃/夜宵)详情数据采集:
// 文本以中文维基百科条目全文为唯一素材 → LLM 只做压缩改写(不得引入素材外事实);
// 图片按精确度优先级聚合:该类下已采集菜品的 OSS 图 → Wikimedia Commons 关键词补图(下载后上传 OSS)。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
)

const (
	// foodCategoryGalleryWant 类别图集目标张数上限。
	foodCategoryGalleryWant = 6
	// foodCategoryDishNameMax 注入 LLM 提示词的"已收录菜品"名单上限。
	foodCategoryDishNameMax = 8
	// foodCategoryWikiTextCap 注入 LLM 的维基正文截断长度(字符),避免超长拖慢/超上下文。
	foodCategoryWikiTextCap = 6000
	// foodCategoryDataSource 类别页数据来源标记(固定)。
	foodCategoryDataSource = "wiki+llm+oss"
	// foodCategoryEstimatedFields 类别页参考值字段(固定)。
	foodCategoryEstimatedFields = "intro,sections"
	// foodCategoryWikiEndpoint 中文维基 Action API。
	foodCategoryWikiEndpoint = "https://zh.wikipedia.org/w/api.php"
	// foodCategoryLLMTimeout LLM 改写超时(秒)。
	foodCategoryLLMTimeout = 120
)

// CategorySeed 一个美食大类的内容采集种子:类别键、名称、维基候选标题、菜品标签关键词与 Commons 关键词。
type CategorySeed struct {
	Key             string
	NameZH          string
	NameEN          string
	WikiTitle       string
	WikiTitles      []string
	DishTags        []string
	CommonsKeywords []string
}

// CategorySeeds 返回三个美食大类的采集种子。
// WikiTitles 逐条尝试,命中第一个存在且正文非空的条目;
// DishTags 与既有分类页一致(含匹配),用于筛出该类别下的菜品图;
// CommonsKeywords 既作 Commons 搜索词,也作文件名准入关键词。
func CategorySeeds() []CategorySeed {
	return []CategorySeed{
		{
			Key:             "cuisine",
			NameZH:          "川菜",
			NameEN:          "Sichuan Cuisine",
			WikiTitle:       "川菜",
			WikiTitles:      []string{"川菜", "川菜系"},
			DishTags:        []string{"川菜"},
			CommonsKeywords: []string{"川菜", "Sichuan cuisine"},
		},
		{
			Key:             "snacks",
			NameZH:          "名小吃",
			NameEN:          "Street Snacks",
			WikiTitle:       "成都小吃",
			WikiTitles:      []string{"成都小吃", "四川小吃"},
			DishTags:        []string{"小吃"},
			CommonsKeywords: []string{"成都小吃", "Chengdu snacks", "四川小吃"},
		},
		{
			Key:             "nightfood",
			NameZH:          "夜宵",
			NameEN:          "Night Food",
			WikiTitle:       "夜宵",
			WikiTitles:      []string{"夜宵", "消夜", "成都夜市"},
			DishTags:        []string{"夜宵", "串串", "凉菜"},
			CommonsKeywords: []string{"成都夜市", "Chengdu night market", "烧烤"},
		},
	}
}

// FoodCategoryEnrichOptions 采集开关。
type FoodCategoryEnrichOptions struct {
	WithImages bool
	WithLLM    bool
	Force      bool
	// DryRun 只抓维基素材并统计图片张数:不调 LLM、不上传 OSS、不写库。
	DryRun bool
}

// FoodCategoryEnrichResult 单个类别的采集结果。
type FoodCategoryEnrichResult struct {
	Key               string
	WikiTitle         string
	WikiTextLen       int
	DishImageCount    int
	CommonsImageCount int
	ImageCount        int
	LLMUsed           bool
	UpdatedFields     []string
	Note              string
}

// FoodCategoryEnricher 美食大类详情采集器。
type FoodCategoryEnricher struct {
	catRepo  *repository.FoodCategoryRepo
	foodRepo *repository.FoodRepo
	llm      *TripLLM
	signer   *pkg.OssSigner
}

// NewFoodCategoryEnricher 构造美食大类采集器。
func NewFoodCategoryEnricher(catRepo *repository.FoodCategoryRepo, foodRepo *repository.FoodRepo, llm *TripLLM, signer *pkg.OssSigner) *FoodCategoryEnricher {
	return &FoodCategoryEnricher{catRepo: catRepo, foodRepo: foodRepo, llm: llm, signer: signer}
}

// foodMatchesKeywords 标签包含任一关键词即命中(空关键词忽略)。
func foodMatchesKeywords(f model.Food, keywords []string) bool {
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(f.Tags, kw) {
			return true
		}
	}
	return false
}

// dishesForCategory 按标签关键词(包含匹配)筛出该类别下的菜品,保持原顺序。
func dishesForCategory(foods []model.Food, keywords []string) []model.Food {
	out := make([]model.Food, 0)
	for _, f := range foods {
		if foodMatchesKeywords(f, keywords) {
			out = append(out, f)
		}
	}
	return out
}

// collectCategoryImages 收集该类别菜品的图,采用轮询(round-robin)避免单个菜品垄断图集:
// 按菜品顺序,每轮从每道菜各取 1 张,直到取满 limit 或所有菜品的图都被取完
// (3 道菜时结果为 菜A图1、菜B图1、菜C图1、菜A图2……)。每道菜内先 gallery_images 再 images;
// URL 全局去重、跳过占位封面。
func collectCategoryImages(foods []model.Food, keywords []string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	// 先整理每道菜的有效图列表(占位封面剔除,顺序:gallery_images 在前、images 在后)
	perDish := make([][]string, 0)
	for _, f := range dishesForCategory(foods, keywords) {
		urls := splitComma(f.GalleryImages)
		urls = append(urls, splitComma(f.Images)...)
		valid := make([]string, 0, len(urls))
		for _, u := range urls {
			if isPlaceholderCover(u) {
				continue
			}
			valid = append(valid, u)
		}
		if len(valid) > 0 {
			perDish = append(perDish, valid)
		}
	}
	out := make([]string, 0, limit)
	seen := make(map[string]bool, limit)
	for round := 0; len(out) < limit; round++ {
		tookAny := false
		for _, valid := range perDish {
			if len(out) >= limit {
				break
			}
			if round >= len(valid) {
				continue // 这道菜的图已取完
			}
			tookAny = true
			u := valid[round]
			if seen[u] {
				continue
			}
			seen[u] = true
			out = append(out, u)
		}
		if !tookAny {
			break // 所有菜品的图都已取完
		}
	}
	return out
}

// collectCategoryDishNames 返回该类别下已收录菜品的 name_zh 名单(去空、去重、最多 max 个,保持原顺序),
// 供 LLM 提示词的"代表食物"段落引用,仅可提及菜名本身。
func collectCategoryDishNames(foods []model.Food, keywords []string, max int) []string {
	if max <= 0 {
		return nil
	}
	out := make([]string, 0, max)
	seen := make(map[string]bool, max)
	for _, f := range dishesForCategory(foods, keywords) {
		name := strings.TrimSpace(f.NameZH)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
		if len(out) >= max {
			break
		}
	}
	return out
}

// buildCategorySections 段落与配图配对:第 i 段配第 i 张;图片不足时复用最后一张,保证每段都有图。
// 缺 title 或 text 的段落直接丢弃。
func buildCategorySections(llmSections []model.CategorySection, images []string) []model.CategorySection {
	out := make([]model.CategorySection, 0, len(llmSections))
	for i, sec := range llmSections {
		title := strings.TrimSpace(sec.Title)
		text := strings.TrimSpace(sec.Text)
		if title == "" || text == "" {
			continue
		}
		img := ""
		if len(images) > 0 {
			if i < len(images) {
				img = images[i]
			} else {
				img = images[len(images)-1]
			}
		}
		out = append(out, model.CategorySection{Title: title, Text: text, Image: img})
	}
	return out
}

// newWikiHTTPClient 构造中文维基专用 HTTP 客户端:代理配置与 Commons 一致(复用 COMMONS_PROXY)。
// 维基与 Commons 同样在国内通常需要代理;Go 不使用系统代理,故显式支持。
func newWikiHTTPClient() *http.Client {
	return newCommonsHTTPClient()
}

// fetchWikiArticle 依次尝试候选标题,返回第一个存在且正文非空的条目(标题、正文全文、条目链接)。
// 不带 exintro(取全文);redirects=1 自动跟随重定向;variant=zh-cn 保证简体。
func fetchWikiArticle(ctx context.Context, client *http.Client, titles []string) (string, string, string, error) {
	if client == nil {
		return "", "", "", fmt.Errorf("维基客户端未初始化")
	}
	var lastErr error
	for _, t := range titles {
		title := strings.TrimSpace(t)
		if title == "" {
			continue
		}
		text, resolved, err := fetchWikiExtract(ctx, client, title)
		if err != nil {
			lastErr = err
			logger.Warnf("维基条目 %s 抓取失败: %v", title, err)
			continue
		}
		if strings.TrimSpace(text) == "" {
			lastErr = fmt.Errorf("条目 %s 正文为空", title)
			logger.Warnf("维基条目 %s 正文为空,尝试下一个候选", title)
			continue
		}
		if strings.TrimSpace(resolved) == "" {
			resolved = title
		}
		return resolved, text, "https://zh.wikipedia.org/wiki/" + url.PathEscape(resolved), nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("没有可用的维基候选标题")
	}
	return "", "", "", lastErr
}

// fetchWikiExtract 请求单个维基条目的 explaintext 全文。
func fetchWikiExtract(ctx context.Context, client *http.Client, title string) (string, string, error) {
	apiURL := fmt.Sprintf(
		"%s?action=query&prop=extracts&explaintext=1&variant=zh-cn&redirects=1&format=json&utf8=1&titles=%s",
		foodCategoryWikiEndpoint, url.QueryEscape(title))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; wenlv-enricher/1.0)")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var res struct {
		Query struct {
			Pages map[string]struct {
				Title   string `json:"title"`
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", "", fmt.Errorf("维基响应解析失败: %w", err)
	}
	for _, p := range res.Query.Pages {
		return p.Extract, p.Title, nil
	}
	return "", "", fmt.Errorf("维基响应缺少 pages")
}

// searchCommonsCandidates 按类别关键词搜索 Commons 补图,文件名必须包含关键词之一(规则同景点侧)。
func (e *FoodCategoryEnricher) searchCommonsCandidates(ctx context.Context, seed CategorySeed, want int) []ScenicImageCandidate {
	if want <= 0 {
		return nil
	}
	names := make([]string, 0, len(seed.CommonsKeywords))
	for _, kw := range seed.CommonsKeywords {
		if v := strings.TrimSpace(kw); v != "" {
			names = append(names, v)
		}
	}
	if len(names) == 0 {
		names = []string{seed.NameZH}
	}
	client := newCommonsHTTPClient()
	out := make([]ScenicImageCandidate, 0, want)
	seen := make(map[string]bool, want)
	requests := 0
	for _, kw := range seed.CommonsKeywords {
		if len(out) >= want {
			break
		}
		if requests > 0 {
			time.Sleep(scenicCommonsRequestInterval)
		}
		requests++
		results, err := searchCommonsImagesFor(ctx, client, kw, names, want-len(out))
		if err != nil {
			logger.Warnf("Commons 搜索失败 (%s): %v", kw, err)
			continue
		}
		for _, r := range results {
			if len(out) >= want {
				break
			}
			if seen[r.URL] {
				continue
			}
			seen[r.URL] = true
			out = append(out, ScenicImageCandidate{URL: r.URL, Source: scenicImageSourceCommons})
		}
	}
	return out
}

// categoryRewrite LLM 改写结果的 JSON 结构。
type categoryRewrite struct {
	Intro    string                  `json:"intro"`
	Sections []model.CategorySection `json:"sections"`
}

// rewriteCategory 让 LLM 仅依据维基素材与本站已收录菜品名压缩改写类别介绍(不得引入素材之外的事实)。
// dishNames 为该类别下已收录菜品的 name_zh 名单(可为空);非空时注入提示词,仅允许提及菜名本身,
// 让"代表食物/去哪里吃"这类段落有本地依据,同时严格禁止编造店名/价格/排名/年份/称号。
func (e *FoodCategoryEnricher) rewriteCategory(ctx context.Context, seed CategorySeed, wikiTitle, material string, dishNames []string) (*categoryRewrite, error) {
	text := strings.TrimSpace(material)
	if text == "" {
		return nil, fmt.Errorf("维基素材为空")
	}
	if r := []rune(text); len(r) > foodCategoryWikiTextCap {
		text = string(r[:foodCategoryWikiTextCap])
	}
	dishLine := ""
	if len(dishNames) > 0 {
		dishLine = fmt.Sprintf("本站已收录的该类菜品(可用于\"代表食物\"段落,只可提及菜名与已给信息):%s\n", strings.Join(dishNames, "、"))
	}
	prompt := fmt.Sprintf(`你是成都文旅内容编辑。下面是中文维基百科条目"%s"的正文素材,请据此为"成都%s"改写一份类别介绍。
严格规则:
1. 仅可依据下方素材与已收录菜品信息,不得引入素材之外的事实、数字、年份、排名、人物、品牌、店名;
2. 只做压缩与改写,不要整段照抄原文;
3. 严格只输出 JSON,不要输出解释、不要使用 Markdown:
{"intro":"类别概述(150-250字)","sections":[{"title":"小标题(4-10字)","text":"段落(80-150字)"}]}
4. sections 写 3-4 段,建议结构:起源与特点 / 味型与技法(或品种与做法) / 代表菜品 / 去哪里吃;
5. 使用简体中文;素材里没有的信息宁可不写。
%s素材正文:
%s`, wikiTitle, seed.NameZH, dishLine, text)
	reply, err := e.llm.ChatWithEffort(ctx, foodCategoryLLMTimeout, []llmMessage{
		{Role: "system", Content: "你是严谨的文旅内容编辑,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.3, 2000, llmThinkingLevel())
	if err != nil {
		return nil, err
	}
	var parsed categoryRewrite
	if err := json.Unmarshal([]byte(extractJSONFromResponse(reply)), &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

// isEmptyCategoryField 判断库中该字段是否为空(用于非 Force 模式跳过覆盖)。
func isEmptyCategoryField(fc model.FoodCategory, field string) bool {
	switch field {
	case "name_zh":
		return strings.TrimSpace(fc.NameZH) == ""
	case "name_en":
		return strings.TrimSpace(fc.NameEN) == ""
	case "intro":
		return strings.TrimSpace(fc.Intro) == ""
	case "sections":
		return strings.TrimSpace(fc.Sections) == ""
	case "gallery_images":
		return strings.TrimSpace(fc.GalleryImages) == ""
	case "source_url":
		return strings.TrimSpace(fc.SourceURL) == ""
	default:
		return true
	}
}

// appendCategoryNote 追加备注(以"; "分隔),不覆盖前面的失败原因。
func appendCategoryNote(existing, add string) string {
	add = strings.TrimSpace(add)
	if add == "" {
		return existing
	}
	if strings.TrimSpace(existing) == "" {
		return add
	}
	return existing + "; " + add
}

// Enrich 采集单个美食大类并落库。维基素材取不到时不写库(避免编造);
// LLM 失败时保留既有文本;任一步失败只记 Note 并返回,不中断批处理。
func (e *FoodCategoryEnricher) Enrich(ctx context.Context, seed CategorySeed, opts FoodCategoryEnrichOptions) (*FoodCategoryEnrichResult, error) {
	res := &FoodCategoryEnrichResult{Key: seed.Key}

	// 1) 维基素材(全文,逐候选标题尝试)
	title, text, pageURL, err := fetchWikiArticle(ctx, newWikiHTTPClient(), seed.WikiTitles)
	if err != nil {
		res.Note = "未取到维基素材,跳过(不写库、不编造): " + err.Error()
		return res, nil
	}
	res.WikiTitle = title
	res.WikiTextLen = len([]rune(text))

	// 2) 该类下已采集菜品的 OSS 图(已是 OSS URL,不重复上传)
	foods, err := e.foodRepo.ListAll()
	if err != nil {
		res.Note = "读取菜品失败: " + err.Error()
		return res, nil
	}
	dishImages := collectCategoryImages(foods, seed.DishTags, foodCategoryGalleryWant)
	res.DishImageCount = len(dishImages)
	dishNames := collectCategoryDishNames(foods, seed.DishTags, foodCategoryDishNameMax)

	if opts.DryRun {
		res.ImageCount = len(dishImages)
		res.Note = "试跑:未调 LLM、未上传 OSS、未写库"
		return res, nil
	}

	// 3) 图片不足 6 张时用 Commons 补(仅新图上传 OSS,前缀 foodcat/<key>)
	images := append([]string(nil), dishImages...)
	if opts.WithImages && e.signer != nil && e.signer.Configured() && len(images) < foodCategoryGalleryWant {
		want := foodCategoryGalleryWant - len(images)
		candidates := e.searchCommonsCandidates(ctx, seed, want)
		if len(candidates) > 0 {
			urls, n := uploadImageCandidates(ctx, newCommonsHTTPClient(), e.signer, "foodcat/"+seed.Key, candidates, want)
			if n > 0 {
				images = append(images, urls...)
				res.CommonsImageCount = n
				logger.Infof("Commons 补图上传 %d 张", n)
			}
		} else {
			logger.Warnf("Commons 未找到可用补图,保持 %d 张", len(images))
		}
	}
	res.ImageCount = len(images)

	// 4) LLM 依据素材压缩改写(intro + sections)
	updates := map[string]any{}
	if opts.WithLLM && e.llm != nil && e.llm.Available() {
		out, lerr := e.rewriteCategory(ctx, seed, title, text, dishNames)
		if lerr != nil {
			res.Note = appendCategoryNote(res.Note, "LLM 改写失败,保留既有文本: "+lerr.Error())
		} else if out != nil {
			if intro := strings.TrimSpace(out.Intro); intro != "" {
				updates["intro"] = intro
			}
			if sections := buildCategorySections(out.Sections, images); len(sections) > 0 {
				if raw, merr := json.Marshal(sections); merr == nil {
					updates["sections"] = string(raw)
				}
			}
			if len(updates) > 0 {
				res.LLMUsed = true
			}
		}
	}

	// 5) 事实字段:类别名(固定)、类别图集(OSS)、素材来源、来源标记
	updates["name_zh"] = seed.NameZH
	updates["name_en"] = seed.NameEN
	if len(images) > 0 {
		updates["gallery_images"] = strings.Join(images, ",")
	}
	updates["source_url"] = pageURL
	updates["estimated_fields"] = foodCategoryEstimatedFields
	updates["data_source"] = foodCategoryDataSource

	// 6) 非 Force 模式保护已有非空值
	existing, err := e.catRepo.GetByKey(seed.Key)
	if err != nil {
		res.Note = appendCategoryNote(res.Note, "读取类别失败: "+err.Error())
		return res, nil
	}
	if existing == nil {
		// 首次采集:建一条空壳记录,保证有主键可落库
		existing = &model.FoodCategory{Key: seed.Key, NameZH: seed.NameZH, NameEN: seed.NameEN}
		if cerr := e.catRepo.Create(existing); cerr != nil {
			return res, fmt.Errorf("创建类别记录失败: %w", cerr)
		}
	}
	if !opts.Force {
		for key := range updates {
			if !isEmptyCategoryField(*existing, key) {
				delete(updates, key)
			}
		}
	}
	if !hasBusinessUpdate(updates) {
		res.Note = appendCategoryNote(res.Note, "无字段需要更新")
		return res, nil
	}
	updates["data_updated_at"] = time.Now()
	for k := range updates {
		res.UpdatedFields = append(res.UpdatedFields, k)
	}
	if err := e.catRepo.UpdateFields(existing.ID, updates); err != nil {
		return res, fmt.Errorf("落库失败: %w", err)
	}
	return res, nil
}
