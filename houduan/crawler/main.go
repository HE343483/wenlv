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
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"wenlv-backend/model"
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

// ossCfg OSS 连接配置(读取 .env 中 OSS_* 变量)
type ossCfg struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

func main() {
	proxy := flag.String("proxy", "", "HTTP 代理地址(维基百科需代理),留空则直连")
	source := flag.String("source", filepath.Join("..", "wennv", "wenlv", "src", "data", "chengdu.ts"), "前端数据文件路径")
	outDir := flag.String("out", filepath.Join("..", "images", "scenic"), "图片下载目录")
	limit := flag.Int("limit", 0, "仅处理前 N 个景点(0 表示全部)")
	delay := flag.Int("delay", 1500, "每次维基百科请求的间隔毫秒数")
	upload := flag.Bool("upload", false, "图片下载后同步上传 OSS,并把数据库图片地址替换为 OSS URL")
	uploadOnly := flag.Bool("upload-only", false, "跳过爬取,仅把本地已下载图片上传 OSS 并更新数据库(无需代理)")
	flag.Parse()

	_ = godotenv.Load()

	// OSS 配置:启用上传相关功能时必须齐全
	var oss *ossCfg
	if *upload || *uploadOnly {
		oss = &ossCfg{
			Endpoint:  os.Getenv("OSS_ENDPOINT"),
			AccessKey: os.Getenv("OSS_ACCESS_KEY"),
			SecretKey: os.Getenv("OSS_SECRET_KEY"),
			Bucket:    os.Getenv("OSS_BUCKET"),
		}
		if oss.Endpoint == "" || oss.AccessKey == "" || oss.SecretKey == "" || oss.Bucket == "" {
			log.Fatal("已启用 OSS 上传,但 .env 中 OSS_* 配置不完整")
		}
	}

	client := newHTTPClient(*proxy)

	spots, dists := parseChengduTS(*source)
	if len(spots) == 0 {
		log.Fatalf("未能从 %s 解析到景点数据", *source)
	}
	if *limit > 0 && *limit < len(spots) {
		spots = spots[:*limit]
	}
	log.Printf("解析到 %d 个区县 / %d 个景点,开始爬取(输出目录: %s)", len(dists), len(spots), *outDir)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("创建图片目录失败: %v", err)
	}

	db := mustConnectDB()
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	// ── 仅上传模式:本地图片 → OSS → 更新数据库,不访问维基百科 ──
	if *uploadOnly {
		uploadLocalToOSS(db, oss, *outDir, spots)
		return
	}

	okCnt, noWiki, noImg, failCnt := 0, 0, 0, 0
	for i, s := range spots {
		log.Printf("[%d/%d] %s", i+1, len(spots), s.NameZH)

		// 1. 维基百科搜索候选词条(无结果时尝试去掉后缀重搜)
		titles, err := searchWiki(client, s.NameZH)
		if err != nil {
			log.Printf("    搜索失败: %v", err)
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
			log.Printf("    未找到维基百科词条,保留原有简介")
			noWiki++
		} else {
			// 2. 依次尝试候选,要求摘要内容提及成都/四川,防止跨地域误配
			matched, fetchErr := "", false
			for _, t := range titles {
				sum, err := fetchSummary(client, t)
				if err != nil {
					log.Printf("    摘要获取失败: %v", err)
					fetchErr = true
					break
				}
				if sum.Type == "disambiguation" {
					continue
				}
				if !strings.Contains(sum.Extract, "成都") && !strings.Contains(sum.Extract, "四川") {
					log.Printf("    候选 [%s] 与成都无关,跳过", t)
					continue
				}
				matched = t
				s.WikiTitle = t
				if sum.Extract != "" {
					s.Extract = sum.Extract
				}
				if sum.OriginalImage != nil && sum.OriginalImage.Source != "" {
					s.ImageURL = sum.OriginalImage.Source
				} else if sum.Thumbnail != nil {
					s.ImageURL = sum.Thumbnail.Source
				}
				break
			}
			if matched == "" && !fetchErr {
				log.Printf("    所有候选均不匹配,保留原有简介")
				noWiki++
			}
		}

		// 2.5 词条无主图时,从 Wikimedia Commons 图库补充
		if s.ImageURL == "" {
			if img := commonsImage(client, s.NameZH, s.NameEN); img != "" {
				s.ImageURL = img
				log.Printf("    Commons 图库补图")
			}
		}

		// 3. 下载图片到本地(无图时保留原占位路径);启用 -upload 时同步上传 OSS
		if s.ImageURL != "" && !strings.Contains(s.ImageURL, "placeholder") {
			ext := strings.ToLower(filepath.Ext(s.ImageURL))
			if idx := strings.IndexByte(ext, '?'); idx >= 0 {
				ext = ext[:idx]
			}
			if ext == "" || len(ext) > 6 {
				ext = ".jpg"
			}
			local := filepath.Join(*outDir, s.ID+ext)
			if err := downloadImage(client, s.ImageURL, local); err != nil {
				log.Printf("    图片下载失败(%s): %v,仅记录外链", s.ImageURL, err)
				noImg++
			} else {
				s.LocalPath = local
				log.Printf("    图片已保存: %s", local)
				if oss != nil {
					key := "scenic/" + s.ID + ext
					if ossURL, err := ossUpload(oss, local, key); err != nil {
						log.Printf("    OSS 上传失败: %v", err)
					} else {
						s.ImageURL = ossURL
						log.Printf("    已上传 OSS: %s", ossURL)
					}
				}
			}
		} else {
			noImg++
			log.Printf("    该词条无主图")
		}

		// 4. 写入数据库
		if err := upsertSpot(db, s, dists); err != nil {
			log.Printf("    入库失败: %v", err)
			failCnt++
			continue
		}
		okCnt++
		if s.Extract != "" {
			log.Printf("    文本: %s...", truncate(s.Extract, 40))
		}
		time.Sleep(time.Duration(*delay) * time.Millisecond)
	}

	log.Printf("完成: 成功入库 %d / 无词条 %d / 无主图 %d / 失败 %d", okCnt, noWiki, noImg, failCnt)
}

// ──── HTTP ────

func newHTTPClient(proxy string) *http.Client {
	tr := &http.Transport{}
	if proxy != "" {
		pu, err := url.Parse(proxy)
		if err != nil {
			log.Fatalf("代理地址无效: %v", err)
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
			log.Printf("    请求限流/失败,第 %d 次重试(等待 %v): %v", attempt, wait, lastErr)
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
			low := strings.ToLower(info.URL)
			if !strings.HasSuffix(low, ".jpg") && !strings.HasSuffix(low, ".jpeg") && !strings.HasSuffix(low, ".png") {
				continue
			}
			if info.Width < 640 || info.Width <= bestW {
				continue
			}
			best, bestW = info.URL, info.Width
		}
		if best != "" {
			return best
		}
	}
	return ""
}

// fetchSummary 获取词条摘要结构(含正文提取、主图 URL)。
func fetchSummary(client *http.Client, title string) (wikiSummary, error) {
	u := "https://zh.wikipedia.org/api/rest_v1/page/summary/" + url.PathEscape(title)
	var sum wikiSummary
	err := httpGetJSON(client, u, &sum)
	return sum, err
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

// ossUpload 以 OSS 签名 V1(Header 签名)直传本地文件,返回公开访问 URL。
// OSS 国内端点可直连,不走维基百科代理。
func ossUpload(cfg *ossCfg, localPath, key string) (string, error) {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(localPath)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	date := time.Now().UTC().Format(http.TimeFormat)

	// StringToSign = VERB \n Content-MD5 \n Content-Type \n Date \n /Bucket/Key
	stringToSign := fmt.Sprintf("PUT\n\n%s\n%s\n/%s/%s", contentType, date, cfg.Bucket, key)
	mac := hmac.New(sha1.New, []byte(cfg.SecretKey))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Endpoint 已含域名(如 oss-cn-chengdu.aliyuncs.com),采用虚拟主机风格
	ossURL := fmt.Sprintf("https://%s.%s/%s", cfg.Bucket, cfg.Endpoint, key)
	req, err := http.NewRequest(http.MethodPut, ossURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Date", date)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "OSS "+cfg.AccessKey+":"+signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("OSS 返回 HTTP %d: %s", resp.StatusCode, string(body))
	}
	return ossURL, nil
}

// uploadLocalToOSS 仅上传模式:遍历本地图片目录,按文件名(景点ID)匹配景点并更新数据库。
func uploadLocalToOSS(db *gorm.DB, cfg *ossCfg, dir string, spots []spot) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("读取图片目录失败: %v", err)
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
		ossURL, err := ossUpload(cfg, local, key)
		if err != nil {
			log.Printf("[%s] OSS 上传失败: %v", s.NameZH, err)
			failCnt++
			continue
		}
		if err := db.Model(&model.ScenicSpot{}).Where("name_zh = ?", s.NameZH).
			Update("images", ossURL).Error; err != nil {
			log.Printf("[%s] 数据库更新失败: %v", s.NameZH, err)
			failCnt++
			continue
		}
		log.Printf("[%s] 已上传并更新: %s", s.NameZH, ossURL)
		okCnt++
	}
	log.Printf("上传完成: 成功 %d / 失败 %d", okCnt, failCnt)
}

// ──── 数据库 ────

func mustConnectDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	if host == "" || name == "" {
		log.Fatal("缺少 DB_* 环境变量,请检查 .env")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	// 与主服务保持一致,确保表存在且字段注释齐全
	if err := db.AutoMigrate(&model.ScenicSpot{}); err != nil {
		log.Fatalf("迁移 scenic_spots 失败: %v", err)
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
	if images != "" {
		updates["images"] = images
	}
	return db.Model(&existing).Updates(updates).Error
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
		log.Fatalf("读取数据文件失败: %v", err)
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
