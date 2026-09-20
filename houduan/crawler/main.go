// crawler 景点数据爬取脚本。
// 数据源:中文维基百科(需代理访问),获取景点介绍文本与主图。
// 产出:
//  1. 介绍文本 + 图片外链写入 MySQL scenic_spots 表(按 name_zh 匹配,存在则更新,不存在则插入)
//  2. 图片文件下载到本地 images/scenic/ 目录(供后续上传 OSS)
//
// 用法(在 houduan 目录下执行):
//
//	go run ./crawler -proxy http://127.0.0.1:7897 -limit 3   # 试跑前 3 个
//	go run ./crawler -proxy http://127.0.0.1:7897            # 全量
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"wenlv-backend/logger"
	"wenlv-backend/model"
	"wenlv-backend/pkg"
)

// ──── 前端 chengdu.ts 中的景点条目 ────

type spot struct {
	ID        string
	DistID    string
	NameZH    string
	NameEN    string
	DescZH    string
	Tags      []string
	Rating    float64
	Lng, Lat  float64
	ImageURL  string // 维基百科图片外链
	LocalPath string // 本地下载路径
	WikiTitle string
	Extract   string // 维基百科简介文本
}

type district struct {
	ID    string
	Name  string
	Name2 string
}

func main() {
	proxy := flag.String("proxy", "", "HTTP 代理地址(维基百科需代理),留空则直连")
	source := flag.String("source", filepath.Join("..", "wennv", "wenlv", "src", "data", "chengdu.ts"), "前端数据文件路径")
	outDir := flag.String("out", filepath.Join("..", "images", "scenic"), "图片下载目录")
	limit := flag.Int("limit", 0, "仅处理前 N 个景点(0 表示全部)")
	delay := flag.Int("delay", 1500, "每次维基百科请求的间隔毫秒数")
	upload := flag.Bool("upload", false, "图片下载后同步上传 OSS,并把数据库图片地址替换为 OSS URL")
	uploadOnly := flag.Bool("upload-only", false, "跳过爬取,仅把本地已下载图片上传 OSS 并更新数据库(无需代理)")
	descOnly := flag.Bool("desc-only", false, "仅回填景点简介:从数据库读取景点,用维基百科简体正文只更新 desc 字段")
	food := flag.Bool("food", false, "爬取成都美食(写入 foods 表),配合 -upload 上传 OSS")
	foodCard := flag.Bool("food-card", false, "爬取美食页六大风味名片配图(写入 food_cards 表),配合 -upload 上传 OSS")
	routes := flag.Bool("routes", false, "爬取精选路线站点简介与配图(写入 routes 表),配合 -upload 上传 OSS")
	only := flag.String("only", "", "美食/美食名片模式下仅处理名称或标识包含该关键字的条目")
	flag.Parse()

	_ = godotenv.Load()

	logger.ConfigureTool("crawler")
	defer logger.Close()

	// OSS 配置:启用上传相关功能时必须齐全
	var signer *pkg.OssSigner
	if *upload || *uploadOnly {
		signer = pkg.NewOssSigner(&pkg.OssConfig{
			Endpoint:  os.Getenv("OSS_ENDPOINT"),
			AccessKey: os.Getenv("OSS_ACCESS_KEY"),
			SecretKey: os.Getenv("OSS_SECRET_KEY"),
			Bucket:    os.Getenv("OSS_BUCKET"),
		})
		if !signer.Configured() {
			logger.Fatalf("已启用 OSS 上传,但 .env 中 OSS_* 配置不完整")
		}
	}

	client := newHTTPClient(*proxy)

	db := mustConnectDB()
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	// ── 仅回填简介模式:景点列表取自数据库,不依赖 chengdu.ts,只更新 desc ──
	if *descOnly {
		runDescOnly(client, db, *delay)
		return
	}

	spots, dists := parseChengduTS(*source)
	if len(spots) == 0 {
		logger.Fatalf("未能从 %s 解析到景点数据", *source)
	}
	if *limit > 0 && *limit < len(spots) {
		spots = spots[:*limit]
	}
	logger.Infof("解析到 %d 个区县 / %d 个景点,开始爬取(输出目录: %s)", len(dists), len(spots), *outDir)

	// 未配置 OSS 时才需要本地目录留档;配置 OSS 后图片直接内存上传,不在本地落盘
	if signer == nil {
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			logger.Fatalf("创建图片目录失败: %v", err)
		}
	}

	// ── 仅上传模式:本地图片 → OSS → 更新数据库,不访问维基百科 ──
	if *uploadOnly {
		uploadLocalToOSS(db, signer, *outDir, spots)
		return
	}

	// ── 精选路线模式:抓取路线站点简介与配图写入 routes 表 ──
	if *routes {
		routeOut := filepath.Join(filepath.Dir(*outDir), "routes") // images/routes/
		runRoutesCrawl(client, db, signer, routeOut, *delay)
		return
	}

	// ── 美食模式:爬取成都经典美食写入 foods 表 ──
	if *food {
		foodOut := filepath.Join(filepath.Dir(*outDir), "food") // images/food/,与景点图片分开存放
		runFoodCrawl(client, db, signer, foodOut, *delay, *only)
		return
	}

	// ── 美食名片模式:爬取六大风味名片配图写入 food_cards 表 ──
	if *foodCard {
		cardOut := filepath.Join(filepath.Dir(*outDir), "food-card") // images/food-card/
		runFoodCardCrawl(client, db, signer, cardOut, *delay, *only)
		return
	}

	okCnt, noWiki, noImg, failCnt := 0, 0, 0, 0
	for i, s := range spots {
		logger.Infof("[%d/%d] %s", i+1, len(spots), s.NameZH)

		// 1. 维基百科搜索候选词条(无结果时尝试去掉后缀重搜)
		titles, err := searchWiki(client, s.NameZH)
		if err != nil {
			logger.Warnf("    搜索失败: %v", err)
			failCnt++
			continue
		}
		if len(titles) == 0 {
			for _, alt := range altNames(s.NameZH) {
				if titles, err = searchWiki(client, alt); err != nil || len(titles) > 0 {
					break
				}
			}
		}
		if len(titles) == 0 {
			logger.Infof("    未找到维基百科词条,保留原有简介")
			noWiki++
		} else {
			// 2. 依次尝试候选,要求摘要内容提及成都/四川/川菜,防止跨地域误配
			matched, fetchErr := "", false
			for _, t := range titles {
				sum, err := fetchSummary(client, t)
				if err != nil {
					logger.Warnf("    摘要获取失败: %v", err)
					fetchErr = true
					break
				}
				if sum.Type == "disambiguation" {
					continue
				}
				// 正文优先取 Action API(带 variant=zh-cn,返回简体);REST summary 实测
				// 忽略 variant 可能返回繁体,仅在 intro 为空时兜底。sum 仍用于主图/消歧义。
				extract, _ := fetchIntro(client, t)
				if extract == "" {
					extract = sum.Extract
				}
				if !strings.Contains(extract, "成都") && !strings.Contains(extract, "四川") &&
					!strings.Contains(extract, "川菜") && !strings.Contains(extract, "川味") {
					logger.Infof("    候选 [%s] 与成都/四川无关,跳过", t)
					continue
				}
				matched = t
				s.WikiTitle = t
				if extract != "" {
					s.Extract = extract
				}
				if sum.OriginalImage != nil && sum.OriginalImage.Source != "" {
					s.ImageURL = sum.OriginalImage.Source
				} else if sum.Thumbnail != nil {
					s.ImageURL = sum.Thumbnail.Source
				}
				break
			}
			if matched == "" && !fetchErr {
				logger.Infof("    所有候选均不匹配,保留原有简介")
				noWiki++
			}
		}

		// 2.5 词条无主图时,从 Wikimedia Commons 图库补充
		if s.ImageURL == "" {
			if img := commonsImage(client, s.NameZH, s.NameEN); img != "" {
				s.ImageURL = img
				logger.Infof("    Commons 图库补图")
			}
		}

		// 3. 图片处理:配置 OSS 时直接内存上传(不落盘),否则下载到本地留档
		if s.ImageURL != "" && !strings.Contains(s.ImageURL, "placeholder") {
			ext := imageExt(s.ImageURL)
			if signer != nil {
				ossURL, err := uploadImageToOSS(client, signer, s.ImageURL, "scenic/"+s.ID+ext)
				if err != nil {
					logger.Warnf("    OSS 上传失败: %v,仅记录外链", err)
					noImg++
				} else {
					s.ImageURL = ossURL
					logger.Infof("    已上传 OSS: %s", ossURL)
				}
			} else {
				local := filepath.Join(*outDir, s.ID+ext)
				if err := downloadImage(client, s.ImageURL, local); err != nil {
					logger.Warnf("    图片下载失败(%s): %v,仅记录外链", s.ImageURL, err)
					noImg++
				} else {
					s.LocalPath = local
					logger.Infof("    图片已保存: %s", local)
				}
			}
		} else {
			noImg++
			logger.Infof("    该词条无主图")
		}

		// 4. 写入数据库
		if err := upsertSpot(db, s, dists); err != nil {
			logger.Errorf("    入库失败: %v", err)
			failCnt++
			continue
		}
		okCnt++
		if s.Extract != "" {
			logger.Infof("    文本: %s...", truncate(s.Extract, 40))
		}
		time.Sleep(time.Duration(*delay) * time.Millisecond)
	}

	logger.Infof("完成: 成功入库 %d / 无词条 %d / 无主图 %d / 失败 %d", okCnt, noWiki, noImg, failCnt)
}

// ──── HTTP ────

func newHTTPClient(proxy string) *http.Client {
	tr := &http.Transport{}
	if proxy != "" {
		pu, err := url.Parse(proxy)
		if err != nil {
			logger.Fatalf("代理地址无效: %v", err)
		}
		tr.Proxy = http.ProxyURL(pu)
	}
	return &http.Client{Timeout: 30 * time.Second, Transport: tr}
}

// httpGetJSON 带 429 重试与指数退避的 GET 请求。
func httpGetJSON(client *http.Client, apiURL string, out any) error {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<uint(attempt)) * 2 * time.Second // 4s/8s/16s
			logger.Warnf("    请求限流/失败,第 %d 次重试(等待 %v): %v", attempt, wait, lastErr)
			time.Sleep(wait)
		}
		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "wenlv-crawler/1.0 (educational project)")
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				var secs int
				if _, err := fmt.Sscanf(ra, "%d", &secs); err == nil && secs > 0 && secs <= 120 {
					time.Sleep(time.Duration(secs) * time.Second)
				}
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP 429")
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		err = json.NewDecoder(resp.Body).Decode(out)
		resp.Body.Close()
		return err
	}
	return lastErr
}

// searchWiki 按名称搜索维基百科词条,返回候选标题(按匹配度排序)。
// 只接受精确/前缀匹配,排除地铁站与消歧义页,宁可不匹配也不错配。
func searchWiki(client *http.Client, name string) ([]string, error) {
	candidates := []string{name, name + " 成都"}
	seen := map[string]bool{}
	var titles []string
	add := func(t string) {
		if t != "" && !seen[t] {
			seen[t] = true
			titles = append(titles, t)
		}
	}
	for _, q := range candidates {
		u := fmt.Sprintf("https://zh.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&srlimit=5&format=json&utf8=1", url.QueryEscape(q))
		var res struct {
			Query struct {
				Search []struct {
					Title string `json:"title"`
				} `json:"search"`
			} `json:"query"`
		}
		if err := httpGetJSON(client, u, &res); err != nil {
			return nil, err
		}
		// 第一轮:精确匹配
		for _, it := range res.Query.Search {
			if it.Title == name {
				add(it.Title)
			}
		}
		// 第二轮:前缀匹配(标题以景点名开头,排除地铁站/消歧义)
		for _, it := range res.Query.Search {
			if strings.HasPrefix(it.Title, name) && !strings.HasSuffix(it.Title, "站") &&
				!strings.Contains(it.Title, "地铁") && !strings.Contains(it.Title, "消歧义") {
				add(it.Title)
			}
		}
		// 第二轮半:标题包含景点名且包含"成都"(如"成都大熊猫繁育研究基地")
		for _, it := range res.Query.Search {
			if strings.Contains(it.Title, name) && strings.Contains(it.Title, "成都") &&
				!strings.HasSuffix(it.Title, "站") {
				add(it.Title)
			}
		}
		// 第三轮:景点名以标题开头(如搜"都江堰水利工程"命中"都江堰"),
		// 此类宽泛候选交由调用方的地域校验兜底
		for _, it := range res.Query.Search {
			if len([]rune(it.Title)) >= 2 && strings.HasPrefix(name, it.Title) && !strings.HasSuffix(it.Title, "站") {
				add(it.Title)
			}
		}
		if len(titles) > 0 {
			break
		}
	}
	return titles, nil
}

type wikiSummary struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Extract   string `json:"extract"`
	Thumbnail *struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
	OriginalImage *struct {
		Source string `json:"source"`
	} `json:"originalimage"`
}

// altNames 生成去掉常见后缀的备选名称(如"黄龙溪古镇"→"黄龙溪")。
func altNames(name string) []string {
	var alts []string
	for _, suf := range []string{"古镇", "古街", "博物馆", "湿地公园", "风景区", "度假区"} {
		if strings.HasSuffix(name, suf) {
			if t := strings.TrimSuffix(name, suf); len([]rune(t)) >= 2 {
				alts = append(alts, t)
			}
		}
	}
	return alts
}

// commonsImage 从 Wikimedia Commons 按名称搜索图片,优先选大尺寸 jpg/png。
func commonsImage(client *http.Client, nameZH, nameEN string) string {
	for _, q := range []string{nameZH, nameEN} {
		if q == "" {
			continue
		}
		u := fmt.Sprintf("https://commons.wikimedia.org/w/api.php?action=query&generator=search&gsrsearch=%s&gsrnamespace=6&gsrlimit=5&prop=imageinfo&iiprop=url%%7Csize&format=json", url.QueryEscape(q))
		var res struct {
			Query struct {
				Pages map[string]struct {
					ImageInfo []struct {
						URL   string `json:"url"`
						Width int    `json:"width"`
					} `json:"imageinfo"`
				} `json:"pages"`
			} `json:"query"`
		}
		if err := httpGetJSON(client, u, &res); err != nil {
			continue
		}
		best, bestW := "", 0
		for _, p := range res.Query.Pages {
			if len(p.ImageInfo) == 0 {
				continue
			}
			info := p.ImageInfo[0]
			// 去掉 Wikimedia 附加的 ?utm_source=... 查询参数后再判断扩展名
			clean := info.URL
			if idx := strings.IndexByte(clean, '?'); idx >= 0 {
				clean = clean[:idx]
			}
			low := strings.ToLower(clean)
			if !strings.HasSuffix(low, ".jpg") && !strings.HasSuffix(low, ".jpeg") && !strings.HasSuffix(low, ".png") {
				continue
			}
			if info.Width < 640 || info.Width <= bestW {
				continue
			}
			best, bestW = clean, info.Width
		}
		if best != "" {
			return best
		}
	}
	return ""
}

// fetchSummary 获取词条摘要结构(含正文提取、主图 URL)。
// 注:实测 REST summary 会忽略 variant 参数(仍返回条目原文,常为繁体),
// 简繁转换只在 Action API(见 fetchIntro)生效,此处带上仅为尝试。
func fetchSummary(client *http.Client, title string) (wikiSummary, error) {
	u := "https://zh.wikipedia.org/api/rest_v1/page/summary/" + url.PathEscape(title) + "?variant=zh-cn"
	var sum wikiSummary
	err := httpGetJSON(client, u, &sum)
	return sum, err
}

// fetchIntro 通过 Action API 获取词条导言纯文本(已做简体转换)。
// REST summary 对部分词条 extract 返回空,用此接口兜底。
func fetchIntro(client *http.Client, title string) (string, error) {
	u := fmt.Sprintf(
		"https://zh.wikipedia.org/w/api.php?action=query&prop=extracts&exintro=1&explaintext=1&format=json&utf8=1&variant=zh-cn&titles=%s",
		url.QueryEscape(title))
	var res struct {
		Query struct {
			Pages map[string]struct {
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := httpGetJSON(client, u, &res); err != nil {
		return "", err
	}
	for _, p := range res.Query.Pages {
		return strings.TrimSpace(p.Extract), nil
	}
	return "", nil
}

// mentionsChengdu 地域校验:正文须提及成都/四川/川菜/川味,防止跨地域误配。
func mentionsChengdu(text string) bool {
	return strings.Contains(text, "成都") || strings.Contains(text, "四川") ||
		strings.Contains(text, "川菜") || strings.Contains(text, "川味")
}

// runDescOnly 仅回填景点简介:景点列表直接取自数据库(不依赖 chengdu.ts),
// 取维基百科简体导言后只更新 desc 字段,绝不触碰 images/gallery_images/
// detail_sections/estimated_fields/address/tel/open_hours/tags/score/lat/lng。
func runDescOnly(client *http.Client, db *gorm.DB, delayMs int) {
	type dbSpot struct {
		ID     uint
		NameZH string
		NameEN string
	}
	var spots []dbSpot
	if err := db.Model(&model.ScenicSpot{}).
		Select("id, name_zh, name_en").Order("id").Find(&spots).Error; err != nil {
		logger.Fatalf("读取景点列表失败: %v", err)
	}
	logger.Infof("从数据库读取到 %d 个景点,开始回填简介(仅更新 desc 字段)", len(spots))

	okCnt, skipCnt, failCnt := 0, 0, 0
	for i, s := range spots {
		// 1. 搜索候选词条(无结果时去掉常见后缀重搜)
		titles, err := searchWiki(client, s.NameZH)
		if err != nil {
			logger.Infof("[%d/%d] %s → 搜索失败: %v", i+1, len(spots), s.NameZH, err)
			failCnt++
			continue
		}
		if len(titles) == 0 {
			for _, alt := range altNames(s.NameZH) {
				if titles, err = searchWiki(client, alt); err != nil || len(titles) > 0 {
					break
				}
			}
		}

		// 2. 依次尝试候选:优先 Action API(带 variant=zh-cn 的简体正文),
		//    REST summary 仅作 extract 为空时的兜底(实测其变体参数无效)
		text, fetchFail, evaluated := "", false, false
		for _, t := range titles {
			extract, err := fetchIntro(client, t)
			if err != nil {
				logger.Warnf("    候选 [%s] 正文获取失败: %v", t, err)
				fetchFail = true
				continue
			}
			if extract == "" {
				sum, err := fetchSummary(client, t)
				if err != nil {
					logger.Warnf("    候选 [%s] 摘要获取失败: %v", t, err)
					fetchFail = true
					continue
				}
				if sum.Type == "disambiguation" {
					continue
				}
				extract = strings.TrimSpace(sum.Extract)
			}
			evaluated = true
			if !mentionsChengdu(extract) {
				logger.Infof("    候选 [%s] 与成都/四川无关,跳过", t)
				continue
			}
			text = extract
			break
		}

		// 3. 只更新 desc 一个字段
		switch {
		case text != "":
			if err := db.Model(&model.ScenicSpot{}).Where("id = ?", s.ID).Update("desc", text).Error; err != nil {
				logger.Errorf("[%d/%d] %s → 入库失败: %v", i+1, len(spots), s.NameZH, err)
				failCnt++
				continue
			}
			logger.Infof("[%d/%d] %s → 已更新(前 40 字: %s)", i+1, len(spots), s.NameZH, truncate(text, 40))
			okCnt++
		case !evaluated && fetchFail:
			logger.Warnf("[%d/%d] %s → 取正文失败，保留原简介", i+1, len(spots), s.NameZH)
			failCnt++
		default:
			logger.Infof("[%d/%d] %s → 未匹配，保留原简介", i+1, len(spots), s.NameZH)
			skipCnt++
		}
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	logger.Infof("完成: 更新 %d / 未匹配 %d / 失败 %d", okCnt, skipCnt, failCnt)
}

func downloadImage(client *http.Client, imgURL, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return nil // 已存在,跳过
	}
	req, err := http.NewRequest(http.MethodGet, imgURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "wenlv-crawler/1.0 (educational project)")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// imageExt 从图片 URL 推断扩展名(去掉查询串),无法识别时按 .jpg 处理。
func imageExt(rawURL string) string {
	ext := strings.ToLower(filepath.Ext(rawURL))
	if idx := strings.IndexByte(ext, '?'); idx >= 0 {
		ext = ext[:idx]
	}
	if ext == "" || len(ext) > 6 {
		ext = ".jpg"
	}
	return ext
}

// pinnedImage 精确指定配图:用于修正自动检索结果不准确的图片。
// Title 为 Commons 文件名;KeyTag 用于生成新的 OSS 对象名,
// 替换后地址变化可绕过浏览器对旧图的 24h 缓存。
type pinnedImage struct {
	Title  string
	KeyTag string
}

// pinnedFoodImages 需人工指正的菜品配图(键为菜品中文名)。
var pinnedFoodImages = map[string]pinnedImage{
	"兔头":  {"File:Chengdu travel 033 (36023201702).jpg", "v2"},
	"韩包子": {"File:Baozi Chengdu.JPG", "v2"},
}

// pinnedFoodCardImages 需人工指正的美食名片配图(键为名片 card_key)。
var pinnedFoodCardImages = map[string]pinnedImage{
	"chuanchuan": {"File:冷锅 串 Cold-pot Skewers Y1 per skewer (1495465364).jpg", "v2"},
}

// keyTag 返回 OSS 对象名的版本后缀(无版本标识时为空串)。
func keyTag(tag string) string {
	if tag == "" {
		return ""
	}
	return "-" + tag
}

// commonsFileURL 按 Commons 文件名精确取原图地址。
func commonsFileURL(client *http.Client, title string) (string, error) {
	u := fmt.Sprintf("https://commons.wikimedia.org/w/api.php?action=query&titles=%s&prop=imageinfo&iiprop=url&format=json",
		url.QueryEscape(title))
	var res struct {
		Query struct {
			Pages map[string]struct {
				ImageInfo []struct {
					URL string `json:"url"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := httpGetJSON(client, u, &res); err != nil {
		return "", err
	}
	for _, p := range res.Query.Pages {
		if len(p.ImageInfo) == 0 || p.ImageInfo[0].URL == "" {
			continue
		}
		clean := p.ImageInfo[0].URL
		if i := strings.IndexByte(clean, '?'); i >= 0 {
			clean = clean[:i]
		}
		return clean, nil
	}
	return "", fmt.Errorf("未找到图片: %s", title)
}

// ossKeyFromURL 从 OSS 图片地址中取出对象 key(非本桶地址返回空)。
func ossKeyFromURL(signer *pkg.OssSigner, rawURL string) string {
	if signer == nil || rawURL == "" {
		return ""
	}
	prefix := fmt.Sprintf("https://%s.%s/", signer.Bucket, signer.Endpoint)
	if !strings.HasPrefix(rawURL, prefix) {
		return ""
	}
	return strings.TrimPrefix(rawURL, prefix)
}

// removeOldOSSImage 替换配图后清理 OSS 上不再使用的旧对象。
func removeOldOSSImage(signer *pkg.OssSigner, oldURL, newKey string) {
	oldKey := ossKeyFromURL(signer, oldURL)
	if oldKey == "" || oldKey == newKey {
		return // 非本桶地址或对象名未变(原地覆盖),无需删除
	}
	if err := signer.DeleteObject(oldKey); err != nil {
		logger.Warnf("    删除 OSS 旧图失败(%s): %v", oldKey, err)
		return
	}
	logger.Infof("    已删除 OSS 旧图: %s", oldKey)
}

// currentFoodImage 读取菜品当前图片地址(用于替换后清理 OSS 旧对象)。
func currentFoodImage(db *gorm.DB, nameZH string) string {
	var f model.Food
	if err := db.Select("images").Where("name_zh = ?", nameZH).First(&f).Error; err != nil {
		return ""
	}
	return f.Images
}

// currentFoodCardImage 读取美食名片当前配图地址。
func currentFoodCardImage(db *gorm.DB, cardKey string) string {
	var c model.FoodCard
	if err := db.Select("image").Where("card_key = ?", cardKey).First(&c).Error; err != nil {
		return ""
	}
	return c.Image
}

// uploadImageToOSS 直接把远程图片上传到 OSS,全程内存操作、不在本地落盘,返回 OSS 地址。
func uploadImageToOSS(client *http.Client, signer *pkg.OssSigner, imageURL, key string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "wenlv-crawler/1.0 (educational project)")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	ct := mime.TypeByExtension(imageExt(imageURL))
	if ct == "" {
		ct = "application/octet-stream"
	}
	return signer.PutObjectBytes(data, key, ct)
}

// uploadLocalToOSS 仅上传模式:遍历本地图片目录,按文件名(景点ID)匹配景点并更新数据库。
func uploadLocalToOSS(db *gorm.DB, signer *pkg.OssSigner, dir string, spots []spot) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Fatalf("读取图片目录失败: %v", err)
	}
	byStem := map[string]string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		byStem[stem] = filepath.Join(dir, e.Name())
	}
	okCnt, failCnt := 0, 0
	for _, s := range spots {
		local, found := byStem[s.ID]
		if !found {
			continue
		}
		key := "scenic/" + filepath.Base(local)
		ossURL, err := signer.PutObject(local, key)
		if err != nil {
			logger.Warnf("[%s] OSS 上传失败: %v", s.NameZH, err)
			failCnt++
			continue
		}
		if err := db.Model(&model.ScenicSpot{}).Where("name_zh = ?", s.NameZH).
			Update("images", ossURL).Error; err != nil {
			logger.Errorf("[%s] 数据库更新失败: %v", s.NameZH, err)
			failCnt++
			continue
		}
		logger.Infof("[%s] 已上传并更新: %s", s.NameZH, ossURL)
		okCnt++
	}
	logger.Infof("上传完成: 成功 %d / 失败 %d", okCnt, failCnt)
}

// ──── 数据库 ────

func mustConnectDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	if host == "" || name == "" {
		logger.Fatalf("缺少 DB_* 环境变量,请检查 .env")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		logger.Fatalf("连接数据库失败: %v", err)
	}
	// 爬虫日志同样落库 log_entries(module=crawler)
	if err := logger.SetDB(db); err != nil {
		logger.Fatalf("日志表 log_entries 初始化失败: %v", err)
	}
	// 与主服务保持一致,确保表存在且字段注释齐全
	if err := db.AutoMigrate(&model.ScenicSpot{}, &model.Food{}, &model.FoodCard{}, &model.Route{}); err != nil {
		logger.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

func upsertSpot(db *gorm.DB, s spot, dists map[string]district) error {
	desc := s.Extract
	if desc == "" {
		desc = s.DescZH // 维基百科无词条时回退前端文案
	}
	images := s.ImageURL

	var existing model.ScenicSpot
	err := db.Where("name_zh = ?", s.NameZH).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		rec := model.ScenicSpot{
			NameZH: s.NameZH,
			NameEN: s.NameEN,
			Desc:   desc,
			Images: images,
			Tags:   strings.Join(s.Tags, ","),
			Score:  s.Rating,
			Lat:    s.Lat,
			Lng:    s.Lng,
		}
		if d, ok := dists[s.DistID]; ok {
			rec.District = d.Name
		}
		return db.Create(&rec).Error
	}
	if err != nil {
		return err
	}
	updates := map[string]any{
		"name_en": s.NameEN,
		"tags":    strings.Join(s.Tags, ","),
		"score":   s.Rating,
		"lat":     s.Lat,
		"lng":     s.Lng,
	}
	// 仅在本次拿到维基百科文本时才覆盖 desc,避免重复运行时用前端文案覆盖已入库的维基文本
	if s.Extract != "" {
		updates["desc"] = s.Extract
	}
	if d, ok := dists[s.DistID]; ok {
		updates["district"] = d.Name
	}
	// 仅当本次拿到的是 OSS 地址,或库中原封面为空时才覆盖 images:
	// 避免不带 -upload 的常规抓取用维基外链覆盖已上传的 OSS 封面。
	if images != "" && (strings.Contains(images, "aliyuncs.com") || existing.Images == "") {
		updates["images"] = images
	}
	return db.Model(&existing).Updates(updates).Error
}

// ──── 美食爬取 ────

// foodSeed 成都经典美食种子数据。
// Wiki 为维基百科搜索词(缺省用 NameZH);ImgTerm 为 Commons 补图搜索词(缺省用中英文名);
// Desc 为无词条时的回退简介。
type foodSeed struct {
	ID      string
	NameZH  string
	NameEN  string
	Wiki    string
	ImgTerm string
	Tags    []string
	Desc    string
}

func foodSeeds() []foodSeed {
	return []foodSeed{
		{"hotpot", "火锅", "Chengdu Hotpot", "四川火锅", "", []string{"火锅", "麻辣"}, "麻辣鲜香的牛油红汤翻滚沸腾，毛肚、鸭肠七上八下，是成都人聚会的首选，也是川渝饮食文化最响亮的名片。"},
		{"chuanchuan", "串串香", "Chuanchuan", "串串香", "", []string{"串串", "市井"}, "竹签串起荤素百味，浸入红汤涮烫，蘸上干碟香油碟， 成都街头的烟火气美食代表。"},
		{"maocai", "冒菜", "Maocai", "冒菜", "", []string{"冒菜", "一个人的火锅"}, "一个人的火锅。荤素食材在滚汤中冒熟，浇上红油蒜泥，配一碗米饭，实惠又过瘾。"},
		{"mapo", "麻婆豆腐", "Mapo Tofu", "", "", []string{"川菜", "麻辣"}, "川菜之魂。豆腐嫩滑、牛肉酥香，麻、辣、烫、香、酥、嫩、鲜、活八字真味，享誉全球。"},
		{"kungpao", "宫保鸡丁", "Kung Pao Chicken", "", "", []string{"川菜", "荔枝味"}, "糊辣荔枝味型代表作，鸡丁滑嫩、花生酥脆，咸甜酸辣平衡，是最早走向世界的川菜。"},
		{"huiguo", "回锅肉", "Twice-Cooked Pork", "", "", []string{"川菜", "家常"}, "川菜第一菜。五花肉先煮后炒，卷成灯盏窝，豆瓣与甜面酱赋予其灵魂，是四川人心中家的味道。"},
		{"fuqi", "夫妻肺片", "Fuqi Feipian", "", "", []string{"凉菜", "老字号"}, "经典川味凉菜。牛肉牛杂切薄片，淋上红油花椒料汁，麻辣鲜香、细嫩化渣，由成都郭朝华夫妇创制得名。"},
		{"dandan", "担担面", "Dan Dan Noodles", "", "", []string{"面食", "小吃"}, "挑担叫卖起家的成都名小吃。面条细薄，臆子酥香，咸鲜微辣，芽菜与花生碎让口感层次丰富。"},
		{"longchaoshou", "龙抄手", "Long Chaoshou", "龙抄手", "", []string{"小吃", "老字号"}, "成都老字号名小吃。皮薄馅嫩、汤浓味美，原汤、红油、海味等多种口味各有拥趸。"},
		{"zhongshuijiao", "钟水饺", "Zhong Dumplings", "", "", []string{"小吃", "红油"}, "始于光绪年间的中华老字号。水饺皮薄馅足，淋特制红油与复制甜酱油，咸甜微辣，回味悠长。"},
		{"tianshuimian", "甜水面", "Sweet Water Noodles", "", "", []string{"面食", "甜辣"}, "粗壮有嚼劲的手擀面，裹上复制甜酱油、红油辣子与芝麻酱，甜中带辣，是成都独一份的味觉记忆。"},
		{"feichangfen", "肥肠粉", "Feichang Rice Noodles", "", "", []string{"粉", "市井"}, "双流名小吃。红薯粉滑爽筋道，肥肠软糯入味，加一节冒节子，配军屯锅盔堪称绝配。"},
		{"sandapao", "三大炮", "San Da Pao", "", "", []string{"小吃", "糯米"}, "锦里庙会的明星小吃。糯米团掷向案板发出三声炮响，裹上黄豆粉淋红糖汁，香甜软糯。"},
		{"tangyouguozi", "糖油果子", "Sugar Rice Balls", "", "", []string{"小吃", "甜食"}, "街头经典甜食。糯米果子在糖油中炸至红亮，外脆内糯，撒白芝麻串成串，是老成都的童年味道。"},
		{"danhonggao", "蛋烘糕", "Dan Hong Gao", "", "", []string{"小吃", "街头"}, "成都人的心头好。面糊在小铜锅中烘成蛋皮小饼，夹肉松奶油或土豆丝，甜咸皆宜。"},
		{"totou", "兔头", "Rabbit Head", "兔头", "", []string{"麻辣", "夜宵"}, "双流老妈兔头名扬天下。兔头经卤煮浸泡，麻辣入骨，啃起来越嚼越香，配上啤酒是地道夜宵。"},
		{"boboji", "钵钵鸡", "Bo Bo Chicken", "", "", []string{"串串", "凉食"}, "瓦罐土陶盛满红油或藤椒汤料，串好的鸡肉藕片浸于其中，麻辣鲜香、随取随吃。"},
		{"juntunkui", "军屯锅盔", "Juntun Guokui", "军屯锅魁", "Guokui", []string{"锅盔", "彭州"}, "彭州军乐镇传统名小吃。千层油酥面饼先煎后烤，层次分明、香酥化渣，夹上凉粉更是人间美味。"},
		{"yeerba", "叶儿粑", "Leaf Cake", "", "", []string{"小吃", "糯米"}, "用良姜叶包裹蒸制的糯米粑，馅分咸甜两派，清香滋润，是川人逢年过节的味觉符号。"},
		{"hanbaozi", "韩包子", "Han Baozi", "", "", []string{"小吃", "老字号"}, "成都老字号包子。皮薄纹匀、馅心细嫩、松泡化渣，南虾包子更是经典中的经典。"},
		{"laitangyuan", "赖汤圆", "Lai Tangyuan", "", "", []string{"小吃", "甜食"}, "创始于1894年的中华老字号。汤圆皮薄滋润、心里甜，黑芝麻馅香甜可口、不腻不沾牙。"},
		{"bingfen", "冰粉", "Bingfen", "", "", []string{"甜食", "消夏"}, "成都夏日限定。晶莹剔透的冰粉配红糖水、花生碎、山楂片，一碗下去暑气全消。"},
		{"gaiwancha", "盖碗茶", "Gaiwan Tea", "盖碗茶", "Gaiwan tea", []string{"茶饮", "慢生活"}, "一茶一坐，半日浮生。成都茶馆里竹椅盖碗、长嘴铜壶掺茶，泡着的是这座城市最地道的慢生活。"},
	}
}

// runFoodCrawl 爬取美食词条文本与图片,写入 foods 表。
// only 非空时仅处理名称包含该关键字的条目(用于单条重爬)。
func runFoodCrawl(client *http.Client, db *gorm.DB, signer *pkg.OssSigner, outDir string, delayMs int, only string) {
	seeds := foodSeeds()
	if only != "" {
		filtered := make([]foodSeed, 0, 1)
		for _, f := range seeds {
			if strings.Contains(f.NameZH, only) {
				filtered = append(filtered, f)
			}
		}
		seeds = filtered
	}
	if signer == nil {
		logger.Infof("开始爬取成都美食(%d 种),图片保存目录: %s", len(seeds), outDir)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			logger.Fatalf("创建图片目录失败: %v", err)
		}
	} else {
		logger.Infof("开始爬取成都美食(%d 种),图片直接上传 OSS", len(seeds))
	}

	okCnt, noImg, failCnt := 0, 0, 0
	for i, f := range seeds {
		logger.Infof("[%d/%d] %s", i+1, len(seeds), f.NameZH)

		// 1. 搜索词条并做地域校验(须提及成都/四川),防止误配其他地域同名食物
		searchName := f.NameZH
		if f.Wiki != "" {
			searchName = f.Wiki
		}
		titles, err := searchWiki(client, searchName)
		if err != nil {
			logger.Warnf("    搜索失败: %v", err)
			failCnt++
			continue
		}
		extract, imgURL := "", ""
		for _, t := range titles {
			sum, err := fetchSummary(client, t)
			if err != nil {
				logger.Warnf("    摘要获取失败: %v", err)
				break
			}
			if sum.Type == "disambiguation" {
				continue
			}
			// 正文优先取 Action API(带 variant=zh-cn,返回简体);REST summary 实测
			// 忽略 variant 可能返回繁体,仅在 intro 为空时兜底。sum 仍用于主图/消歧义。
			cur, _ := fetchIntro(client, t)
			if cur == "" {
				cur = sum.Extract
			}
			if !strings.Contains(cur, "成都") && !strings.Contains(cur, "四川") &&
				!strings.Contains(cur, "川菜") && !strings.Contains(cur, "川味") {
				logger.Infof("    候选 [%s] 与成都/四川无关,跳过", t)
				continue
			}
			extract = cur
			if sum.OriginalImage != nil && sum.OriginalImage.Source != "" {
				imgURL = sum.OriginalImage.Source
			} else if sum.Thumbnail != nil {
				imgURL = sum.Thumbnail.Source
			}
			break
		}

		// 2. 无主图时从 Commons 补图(优先使用指定搜索词)
		if imgURL == "" {
			termZH, termEN := f.NameZH, f.NameEN
			if f.ImgTerm != "" {
				termZH, termEN = f.ImgTerm, f.ImgTerm
			}
			if img := commonsImage(client, termZH, termEN); img != "" {
				imgURL = img
				logger.Infof("    Commons 图库补图")
			}
		}

		// 2.5 人工指定配图优先(用于修正自动检索不准的图片)
		pin, pinned := pinnedFoodImages[f.NameZH]
		if pinned {
			u, err := commonsFileURL(client, pin.Title)
			if err != nil {
				logger.Warnf("    指定配图获取失败: %v", err)
				failCnt++
				continue
			}
			imgURL = u
			logger.Infof("    使用指定配图: %s", pin.Title)
		} else if imgURL == "" {
			noImg++
			logger.Infof("    未获取到图片")
		}

		// 3. 图片处理:配置 OSS 时直接内存上传(不落盘),否则下载到本地留档
		if imgURL != "" {
			ext := imageExt(imgURL)
			key := "food/" + f.ID + keyTag(pin.KeyTag) + ext
			oldImage := currentFoodImage(db, f.NameZH)
			if signer != nil {
				ossURL, err := uploadImageToOSS(client, signer, imgURL, key)
				if err != nil {
					logger.Warnf("    OSS 上传失败: %v", err)
					noImg++
				} else {
					imgURL = ossURL
					logger.Infof("    已上传 OSS: %s", ossURL)
					removeOldOSSImage(signer, oldImage, key)
				}
			} else {
				local := filepath.Join(outDir, f.ID+ext)
				if err := downloadImage(client, imgURL, local); err != nil {
					logger.Warnf("    图片下载失败(%s): %v", imgURL, err)
					noImg++
				} else {
					logger.Infof("    图片已保存: %s", local)
				}
			}
		}

		// 4. 入库
		if err := upsertFood(db, f, extract, imgURL); err != nil {
			logger.Errorf("    入库失败: %v", err)
			failCnt++
			continue
		}
		okCnt++
		if extract != "" {
			logger.Infof("    文本: %s...", truncate(extract, 40))
		}
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	logger.Infof("美食爬取完成: 成功入库 %d / 无图 %d / 失败 %d", okCnt, noImg, failCnt)
}

// upsertFood 按 name_zh 匹配写入 foods 表;已有记录仅在拿到新文本/图片时覆盖。
func upsertFood(db *gorm.DB, f foodSeed, extract, imgURL string) error {
	var existing model.Food
	err := db.Where("name_zh = ?", f.NameZH).First(&existing).Error
	desc := extract
	if desc == "" {
		desc = f.Desc
	}
	tags := strings.Join(f.Tags, ",")
	if err == gorm.ErrRecordNotFound {
		return db.Create(&model.Food{
			NameZH:   f.NameZH,
			NameEN:   f.NameEN,
			District: "成都",
			Tags:     tags,
			Desc:     desc,
			Images:   imgURL,
		}).Error
	}
	if err != nil {
		return err
	}
	updates := map[string]any{
		"name_en": f.NameEN,
		"tags":    tags,
	}
	if extract != "" {
		updates["desc"] = extract
	}
	if imgURL != "" {
		updates["images"] = imgURL
	}
	return db.Model(&existing).Updates(updates).Error
}

// ──── 美食名片配图爬取 ────

// foodCardSeed 美食页六大风味名片种子数据,Query 为 Commons 配图搜索词。
type foodCardSeed struct {
	Key    string
	NameZH string
	NameEN string
	Query  string
	Sort   int
}

func foodCardSeeds() []foodCardSeed {
	return []foodCardSeed{
		{"hotpot", "火锅", "Hot Pot", "hot pot restaurant China", 1},
		{"chuanchuan", "串串香", "Chuan Chuan", "chuanchuan", 2},
		{"cuisine", "川菜", "Sichuan Cuisine", "Sichuan cuisine", 3},
		{"snacks", "名小吃", "Street Snacks", "Chengdu snack", 4},
		{"tea", "盖碗茶", "Gaiwan Tea", "teahouse Chengdu", 5},
		{"nightfood", "夜宵", "Night Food", "night market food", 6},
	}
}

// commonsSearch 从 Wikimedia Commons 搜索图片,按搜索相关度返回候选 URL 列表。
func commonsSearch(client *http.Client, query string, limit int) []string {
	if query == "" {
		return nil
	}
	u := fmt.Sprintf("https://commons.wikimedia.org/w/api.php?action=query&generator=search&gsrsearch=%s&gsrnamespace=6&gsrlimit=%d&prop=imageinfo&iiprop=url%%7Csize&format=json",
		url.QueryEscape(query), limit)
	var res struct {
		Query struct {
			Pages map[string]struct {
				Index     int `json:"index"`
				ImageInfo []struct {
					URL   string `json:"url"`
					Width int    `json:"width"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := httpGetJSON(client, u, &res); err != nil {
		return nil
	}
	type cand struct {
		idx int
		url string
	}
	var list []cand
	for _, p := range res.Query.Pages {
		if len(p.ImageInfo) == 0 {
			continue
		}
		clean := p.ImageInfo[0].URL
		// 去掉 Wikimedia 附加的 ?utm_source=... 查询参数
		if i := strings.IndexByte(clean, '?'); i >= 0 {
			clean = clean[:i]
		}
		low := strings.ToLower(clean)
		if !strings.HasSuffix(low, ".jpg") && !strings.HasSuffix(low, ".jpeg") && !strings.HasSuffix(low, ".png") {
			continue
		}
		if p.ImageInfo[0].Width < 800 {
			continue
		}
		list = append(list, cand{p.Index, clean})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].idx < list[j].idx })
	urls := make([]string, 0, len(list))
	for _, c := range list {
		urls = append(urls, c.url)
	}
	return urls
}

// runFoodCardCrawl 爬取六大风味名片配图,上传 OSS 并写入 food_cards 表。
// 会排除 foods 表已用图片,保证名片配图与下方「成都味道图鉴」不重复。
// only 非空时仅处理标识或名称包含该关键字的名片(用于单张重爬)。
func runFoodCardCrawl(client *http.Client, db *gorm.DB, signer *pkg.OssSigner, outDir string, delayMs int, only string) {
	used := map[string]bool{}
	var existing []model.Food
	if err := db.Select("images").Find(&existing).Error; err != nil {
		logger.Warnf("读取已有美食图片失败(将不排除重复): %v", err)
	}
	for _, f := range existing {
		if f.Images != "" {
			used[f.Images] = true
		}
	}

	seeds := foodCardSeeds()
	if only != "" {
		filtered := make([]foodCardSeed, 0, 1)
		for _, c := range seeds {
			if strings.Contains(c.Key, only) || strings.Contains(c.NameZH, only) {
				filtered = append(filtered, c)
			}
		}
		seeds = filtered
	}
	if signer == nil {
		logger.Infof("开始爬取美食名片配图(%d 张),图片保存目录: %s", len(seeds), outDir)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			logger.Fatalf("创建图片目录失败: %v", err)
		}
	} else {
		logger.Infof("开始爬取美食名片配图(%d 张),图片直接上传 OSS", len(seeds))
	}

	okCnt, failCnt := 0, 0
	for i, c := range seeds {
		logger.Infof("[%d/%d] %s", i+1, len(seeds), c.NameZH)

		// 优先使用人工指定配图,否则按相关度取第一个未被占用的候选图
		pin, pinned := pinnedFoodCardImages[c.Key]
		picked := ""
		if pinned {
			u, err := commonsFileURL(client, pin.Title)
			if err != nil {
				logger.Warnf("    指定配图获取失败: %v", err)
				failCnt++
				continue
			}
			picked = u
			logger.Infof("    使用指定配图: %s", pin.Title)
		} else {
			for _, u := range commonsSearch(client, c.Query, 12) {
				if used[u] {
					continue
				}
				picked = u
				break
			}
		}
		if picked == "" {
			logger.Infof("    未找到可用配图")
			failCnt++
			continue
		}
		used[picked] = true

		ext := imageExt(picked)
		key := "food-card/" + c.Key + keyTag(pin.KeyTag) + ext
		oldImage := currentFoodCardImage(db, c.Key)
		imageURL := picked
		if signer != nil {
			// 直接内存上传 OSS,不在本地落盘
			ossURL, err := uploadImageToOSS(client, signer, picked, key)
			if err != nil {
				logger.Warnf("    OSS 上传失败: %v", err)
				failCnt++
				continue
			}
			imageURL = ossURL
			logger.Infof("    已上传 OSS: %s", ossURL)
			removeOldOSSImage(signer, oldImage, key)
		} else {
			local := filepath.Join(outDir, c.Key+ext)
			if err := downloadImage(client, picked, local); err != nil {
				logger.Warnf("    图片下载失败(%s): %v", picked, err)
				failCnt++
				continue
			}
			logger.Infof("    图片已保存: %s", local)
		}

		if err := upsertFoodCard(db, c, imageURL); err != nil {
			logger.Errorf("    入库失败: %v", err)
			failCnt++
			continue
		}
		okCnt++
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	logger.Infof("美食名片爬取完成: 成功 %d / 失败 %d", okCnt, failCnt)
}

// upsertFoodCard 按 card_key 写入 food_cards 表(存在则更新配图)。
func upsertFoodCard(db *gorm.DB, c foodCardSeed, imageURL string) error {
	var existing model.FoodCard
	err := db.Where("card_key = ?", c.Key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(&model.FoodCard{
			CardKey: c.Key,
			NameZH:  c.NameZH,
			NameEN:  c.NameEN,
			Image:   imageURL,
			Sort:    c.Sort,
		}).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&existing).Updates(map[string]any{
		"name_zh": c.NameZH,
		"name_en": c.NameEN,
		"image":   imageURL,
		"sort":    c.Sort,
	}).Error
}

// ──── chengdu.ts 解析 ────

var (
	reBlockSpots = regexp.MustCompile(`(?s)export const scenicSpots[^=]*=\s*\[(.*?)\n\]`)
	reBlockDists = regexp.MustCompile(`(?s)export const districts[^=]*=\s*\[(.*?)\n\]`)
	reItem       = regexp.MustCompile(`(?s)\{(.*?)\}`)
	reStr        = func(key string) *regexp.Regexp {
		return regexp.MustCompile(regexp.QuoteMeta(key) + `:\s*'((?:[^'\\]|\\.)*)'`)
	}
	reID      = reStr("id")
	reDistID  = reStr("districtId")
	reNameZH  = reStr("nameZh")
	reNameEN  = reStr("nameEn")
	reDescZH  = reStr("descriptionZh")
	reImage   = reStr("imageUrl")
	reTags    = regexp.MustCompile(`(?s)tags:\s*\[(.*?)\]`)
	reTag     = regexp.MustCompile(`'([^']+)'`)
	reRating  = regexp.MustCompile(`rating:\s*([\d.]+)`)
	reLng     = regexp.MustCompile(`lng:\s*([\d.]+)`)
	reLat     = regexp.MustCompile(`lat:\s*([\d.]+)`)
	reDistDes = regexp.MustCompile(`(?s)descriptionZh:\s*'((?:[^'\\]|\\.)*)'`)
)

func parseChengduTS(path string) ([]spot, map[string]district) {
	raw, err := os.ReadFile(path)
	if err != nil {
		logger.Fatalf("读取数据文件失败: %v", err)
	}
	content := string(raw)

	dists := map[string]district{}
	if m := reBlockDists.FindStringSubmatch(content); m != nil {
		for _, blk := range reItem.FindAllStringSubmatch(m[1], -1) {
			id := pick(reID, blk[1])
			if id == "" {
				continue
			}
			d := district{ID: id, Name: pick(reNameZH, blk[1])}
			// 区县条目里 nameZh 即中文名;描述仅用于人工核对
			_ = reDistDes
			dists[id] = d
		}
	}

	var spots []spot
	if m := reBlockSpots.FindStringSubmatch(content); m != nil {
		for _, blk := range reItem.FindAllStringSubmatch(m[1], -1) {
			body := blk[1]
			if pick(reDistID, body) == "" {
				continue // 跳过无 districtId 的条目(路线等)
			}
			s := spot{
				ID:     pick(reID, body),
				DistID: pick(reDistID, body),
				NameZH: unescape(pick(reNameZH, body)),
				NameEN: unescape(pick(reNameEN, body)),
				DescZH: unescape(pick(reDescZH, body)),
			}
			if s.ID == "" || s.NameZH == "" {
				continue
			}
			if tm := reTags.FindStringSubmatch(body); tm != nil {
				for _, t := range reTag.FindAllStringSubmatch(tm[1], -1) {
					s.Tags = append(s.Tags, t[1])
				}
			}
			if rm := reRating.FindStringSubmatch(body); rm != nil {
				fmt.Sscanf(rm[1], "%f", &s.Rating)
			}
			if lm := reLng.FindStringSubmatch(body); lm != nil {
				fmt.Sscanf(lm[1], "%f", &s.Lng)
			}
			if lm := reLat.FindStringSubmatch(body); lm != nil {
				fmt.Sscanf(lm[1], "%f", &s.Lat)
			}
			s.ImageURL = pick(reImage, body)
			spots = append(spots, s)
		}
	}
	return spots, dists
}

func pick(re *regexp.Regexp, s string) string {
	if m := re.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}

// unescape 还原 TS 单引号字符串中的转义
func unescape(s string) string {
	s = strings.ReplaceAll(s, `\'`, `'`)
	s = strings.ReplaceAll(s, `\n`, " ")
	return s
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
