# 景点详情页数据补全(高德 POI + LLM 参考值)Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `/scenic/scenic-{id}` 详情页的评价横条、实用信息、图文详情、精彩瞬间、景区位置、周边推荐都有真实数据，而不是前端写死的占位。

**Architecture:** 分三类取数。① 高德 POI 能给的（详细地址、电话、部分开放时间、相册图片）由批处理离线采集、图片落 OSS、字段落 `scenic_spots`；② 门票价格 / 年接待游客 / 建议游玩时长 / 图文详情这类没有权威公开源的，由 LLM 生成参考值并在 `estimated_fields` 中标记，前端显示"参考值"徽标；③ 交通方式、周边推荐、景区位置地图不入库，运行时按坐标查高德（Redis 缓存 24h），前端用高德 JS API 渲染。

**Tech Stack:** Go 1.2x + Gin + GORM + MySQL + Redis(go-redis/v9)；高德 Web 服务 API（place/detail、place/around）与 JS API 2.0；OpenAI 兼容 LLM 接口（已有 `TripLLM`）；阿里云 OSS V1 Header 签名直传；Vue 3 + Vite 8 + TypeScript + Vitest。

**Spec:** 本文件「现状与数据来源」章节即规格说明（没有单独的 spec 文件）。

## Global Constraints

- 分支：在 `D:/桌面/文旅/wenlv` 下新建 `feature/scenic-detail-data`，**不要**直接改 `main`，最终版本确认前不合并。
- 所有新增数据库字段必须有中文 `comment`（Navicat 可读）。
- `/api/scenic`、`/api/scenic/:id` 只能**新增**字段，不得改名或删除已有字段（`id/name_zh/name_en/district/tags/score/lat/lng/desc/images` 原样保留）。
- 门票价格、年接待游客、建议游玩时长、LLM 补的开放时间**必须**写入 `estimated_fields`，前端展示时必须带"参考值"标注，不得当作权威事实展示。
- LLM 只允许在给定参考简介内改写，提示词中明确"不确定的信息不要写"；禁止编造具体数字、价格区间、年份。
- 所有外部图片（高德 / Wikimedia）必须下载后上传 OSS（bucket `wenlv-tdx`，`scenic/` 目录，`Cache-Control: public, max-age=86400`），数据库只存 OSS URL；禁止直接把外部图片外链存库（Wikimedia 国内不可访问）。
- 外部接口失败必须降级：后端返回空数组，前端显示"暂无数据"，不得 500、不得显示 `common.loading` 这类加载文案。
- 周边推荐/交通站点只缓存在 Redis（不落 MySQL），TTL 由 `SCENIC_CACHE_TTL_HOURS` 配置，默认 24 小时；Redis 不可用时回源高德，不影响页面。
- 周边推荐/交通接口响应带 `Cache-Control: public, max-age=300`（浏览器 5 分钟）；空结果响应不加浏览器缓存头，避免把"暂时查不到"缓存下来。
- 三语文案（`zh/en/ja`）必须同步；本次只生成中文详情内容，英文/日文详情内容不在本次范围（与现有 `desc` 行为一致）。
- 不新增前端依赖（`@amap/amap-jsapi-loader` 已在 `package.json`）；后端不新增第三方库。
- 未经用户明确要求不要 `git commit`；计划中的 Commit 步骤先向用户确认。

---

## 现状与数据来源

现状（已核对代码）：

- `scenic_spots` 只有 9 个业务字段：`name_zh/name_en/district/tags/score/lat/lng/desc/images`（`houduan/model/scenic.go`）。
- 详情页除 hero 图/名称/标签/评分/区县/简介外全是前端写死的占位：`detail-scorebar` 的年接待/建议游玩是 `—`，实用信息 4 张卡是 `common.loading`，图文详情 2 组是空图框 + `paraPlaceholder`，精彩瞬间是造 6 个空框，地图是 pin 占位，周边推荐是造 3 张卡。
- 现有爬虫 `houduan/crawler/main.go` 只做维基百科导言 + 单张主图，产出 `desc` 与 `images`（单 URL）。

字段取数策略：

| 页面字段 | 数据来源 | 落库 | 标注 |
|---|---|---|---|
| 详细地址 `address` | 高德 place/detail | 是 | 事实 |
| 咨询电话 `tel` | 高德 place/detail | 是 | 事实 |
| 开放时间 `open_hours` | 高德 `biz_ext`（覆盖率不高）→ LLM 兜底 | 是 | 高德值=事实；LLM 值=参考值 |
| 精彩瞬间 `gallery_images` | 高德 place/detail `photos` → OSS | 是 | 事实 |
| 图文详情 `detail_sections` | 现有 `desc` + LLM 改写分段，配 gallery 图 | 是(JSON) | 参考值 |
| 建议游玩时长 `recommend_hours` | LLM | 是 | 参考值 |
| 门票价格 `ticket_price` | LLM | 是 | 参考值 |
| 年接待游客 `yearly_visitors` | LLM | 是 | 参考值 |
| 交通方式 | 高德 place/around 搜"地铁站/公交站"（1.5km） | 否 | Redis TTL + 浏览器 5min |
| 周边推荐 | 高德 place/around（3km，风景名胜/餐饮/购物） | 否 | Redis TTL + 浏览器 5min |
| 景区位置 | 库内 `lng/lat` + 高德 JS API 渲染 | 否 | 高德 CDN |

## 新增字段设计（`scenic_spots`）

| 字段 | 类型 | 说明 |
|---|---|---|
| `address` | varchar(255) | 详细地址 |
| `tel` | varchar(64) | 咨询电话 |
| `amap_poi_id` | varchar(64)，索引 | 高德 POI ID，用于去重与刷新 |
| `open_hours` | varchar(255) | 开放时间 |
| `ticket_price` | varchar(255) | 门票价格（参考值） |
| `recommend_hours` | varchar(64) | 建议游玩时长（参考值） |
| `yearly_visitors` | varchar(64) | 年接待游客（参考值） |
| `gallery_images` | text | 相册图片 OSS URL，逗号分隔（首张与 `images` 封面可重复） |
| `detail_sections` | text | 图文详情段落 JSON：`[{"title","text","image"}]` |
| `estimated_fields` | varchar(255) | 参考值字段名，逗号分隔，如 `ticket_price,recommend_hours` |
| `data_source` | varchar(32) | `amap` / `amap+llm` / `wiki` / `llm` / `manual` |
| `data_updated_at` | datetime | 本次采集时间 |

## File Map

| 文件 | 职责 |
|---|---|
| `houduan/model/scenic.go` | 新增字段 + `ScenicDetailSection` 结构 |
| `houduan/pkg/oss.go` | 新增服务端直传 `PutObject/PutObjectBytes/Configured` |
| `houduan/crawler/main.go` | 改为复用 `pkg.OssSigner.PutObject`（删本地重复实现） |
| `houduan/service/trip_amap.go` | 新增 `SearchAround`、POI 详情补 `biz_ext/rating/cost/open_hours` |
| `houduan/service/scenic_enrich.go` | 采集/补全编排：POI 匹配、图片上传、LLM 分段与参考值 |
| `houduan/service/trip_amap_test.go` | 高德纯函数单测（开放时间字段提取） |
| `houduan/service/scenic_enrich_test.go` | 采集器纯函数单测（坐标距离、图文分段、参考值字段拼装） |
| `houduan/repository/content_repo.go` | `ScenicRepo` 新增 `ListAll`、`UpdateFields` |
| `houduan/cmd/scenic-enrich/main.go` | 批处理入口（`-limit/-only/-dry-run/-with-llm/-with-images/-force`） |
| `houduan/service/scenic_extra.go` | 周边推荐 / 交通站点查询 + Redis 缓存（TTL 可配） |
| `houduan/handler/scenic_extra_handler.go` | `Around` / `Transport` 两个只读接口 + 浏览器缓存头 |
| `houduan/config/config.go`、`houduan/.env.example` | 新增 `SCENIC_CACHE_TTL_HOURS` 配置项 |
| `houduan/handler/bootstrap.go`、`houduan/router/router.go`、`houduan/main.go` | 装配与路由注册 |
| `wennv/wenlv/src/api/content.ts` | 类型扩展 + `getScenicAround` / `getScenicTransport` |
| `wennv/wenlv/src/utils/scenicDetail.ts`、`scenicDetail.spec.ts` | 展示层纯函数 + 单测 |
| `wennv/wenlv/src/views/ScenicDetail/ScenicDetail.vue` | 各区块改用真实数据 + 地图 + 参考值标注 |
| `wennv/wenlv/src/locales/{zh,en,ja}.ts` | 新增三语文案 |

---

### Task 0: 建立功能分支

**Files:** 无代码改动

- [ ] **Step 1: 创建并切到功能分支**

```powershell
cd "d:\桌面\文旅\wenlv"
git checkout -b feature/scenic-detail-data
git branch --show-current
```

Expected: 输出 `feature/scenic-detail-data`

---

### Task 1: 扩展 `ScenicSpot` 模型与迁移

**Files:**
- Modify: `houduan/model/scenic.go`

**Interfaces:**
- Consumes: 无
- Produces:
  - `model.ScenicSpot` 新增导出字段：`Address, Tel, AmapPOIID, OpenHours, TicketPrice, RecommendHours, YearlyVisitors, GalleryImages, DetailSections, EstimatedFields, DataSource string`、`DataUpdatedAt *time.Time`
  - `model.ScenicDetailSection{Title, Text, Image string}`，JSON tag 为 `title/text/image`

- [ ] **Step 1: 追加字段与段落结构**

把 `houduan/model/scenic.go` 改成（保留原有 9 个字段与注释不动，只在 `Images` 之后插入新字段，并在文件末尾追加结构体）：

```go
	Desc      string    `gorm:"type:text;comment:景点介绍" json:"desc"`
	Images    string    `gorm:"type:text;comment:图片URL列表(逗号分隔)" json:"images"` // 逗号分隔 URL
	// ===== 详情页扩展字段(高德 POI 采集 + LLM 参考值) =====
	Address         string     `gorm:"size:255;comment:详细地址" json:"address"`
	Tel             string     `gorm:"size:64;comment:咨询电话" json:"tel"`
	AmapPOIID       string     `gorm:"size:64;index;comment:高德POI ID(去重与刷新用)" json:"amap_poi_id"`
	OpenHours       string     `gorm:"size:255;comment:开放时间" json:"open_hours"`
	TicketPrice     string     `gorm:"size:255;comment:门票价格(参考值)" json:"ticket_price"`
	RecommendHours  string     `gorm:"size:64;comment:建议游玩时长(参考值)" json:"recommend_hours"`
	YearlyVisitors  string     `gorm:"size:64;comment:年接待游客(参考值)" json:"yearly_visitors"`
	GalleryImages   string     `gorm:"type:text;comment:相册图片URL列表(逗号分隔)" json:"gallery_images"`
	DetailSections  string     `gorm:"type:text;comment:图文详情段落JSON" json:"detail_sections"`
	EstimatedFields string     `gorm:"size:255;comment:参考值字段(逗号分隔,前端加标注)" json:"estimated_fields"`
	DataSource      string     `gorm:"size:32;comment:数据来源(amap/amap+llm/wiki/llm/manual)" json:"data_source"`
	DataUpdatedAt   *time.Time `gorm:"comment:数据采集时间" json:"data_updated_at"`
	CreatedAt       time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

// ScenicDetailSection 图文详情的一段(标题 + 正文 + 配图)。
// 以 JSON 数组序列化后存入 ScenicSpot.DetailSections。
type ScenicDetailSection struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Image string `json:"image,omitempty"`
}
```

- [ ] **Step 2: 确认迁移已覆盖**

`houduan/database/mysql.go` 的 `AutoMigrate` 已经包含 `&model.ScenicSpot{}`，无需修改；`tableComments` 已有 `scenic_spots` 表注释。

- [ ] **Step 3: 编译并启动，确认列已建**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go build ./...
$env:SERVER_PORT="8081"; go run .
```

另开终端查库（密码按 `.env`）：

```powershell
mysql -uroot wenlv -e "SELECT COLUMN_NAME, COLUMN_COMMENT FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='wenlv' AND TABLE_NAME='scenic_spots' AND COLUMN_NAME IN ('address','open_hours','gallery_images','detail_sections','estimated_fields');"
```

Expected: 5 行，`COLUMN_COMMENT` 均为中文说明。

- [ ] **Step 4: 提交（先向用户确认）**

```powershell
git add houduan/model/scenic.go
git commit -m "feat(scenic): 扩展景点表详情字段(地址/开放时间/相册/图文详情/参考值标记)"
```

---

### Task 2: OSS 服务端直传下沉到 `pkg`

**Files:**
- Modify: `houduan/pkg/oss.go`
- Modify: `houduan/crawler/main.go`

**Interfaces:**
- Consumes: 无
- Produces: `(*pkg.OssSigner).Configured() bool`、`(*pkg.OssSigner).PutObject(localPath, key string) (string, error)`、`(*pkg.OssSigner).PutObjectBytes(data []byte, key, contentType string) (string, error)`

- [ ] **Step 1: 在 `pkg/oss.go` 增加直传能力**

在 `ResolveURL` 之后追加（并在 import 中补 `bytes`、`crypto/hmac` 已有、`errors`、`io`、`mime`、`net/http`、`os`、`path/filepath`、`strings`）：

```go
// Configured OSS 配置是否完整(四项必填)。
func (s *OssSigner) Configured() bool {
	return s.Endpoint != "" && s.AccessKey != "" && s.SecretKey != "" && s.Bucket != ""
}

// endpointHost 去掉 Endpoint 可能带的协议前缀,统一按虚拟主机风格拼接。
func (s *OssSigner) endpointHost() string {
	host := s.Endpoint
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	return strings.TrimSuffix(host, "/")
}

// PutObject 以 OSS V1 Header 签名直传本地文件,返回公开访问 URL。
// 与上传接口的前端直传 Policy 不同,这里是服务端批处理(爬虫/采集任务)使用。
func (s *OssSigner) PutObject(localPath, key string) (string, error) {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(localPath)))
	if ct == "" {
		ct = "application/octet-stream"
	}
	return s.PutObjectBytes(data, key, ct)
}

// PutObjectBytes 直传字节内容,contentType 为空时按 key 后缀推断。
func (s *OssSigner) PutObjectBytes(data []byte, key, contentType string) (string, error) {
	if !s.Configured() {
		return "", errors.New("OSS 未配置")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	// StringToSign = VERB \n Content-MD5 \n Content-Type \n Date \n /Bucket/Key
	stringToSign := fmt.Sprintf("PUT\n\n%s\n%s\n/%s/%s", contentType, date, s.Bucket, key)
	mac := hmac.New(sha1.New, []byte(s.SecretKey))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	u := fmt.Sprintf("https://%s.%s/%s", s.Bucket, s.endpointHost(), key)
	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Date", date)
	req.Header.Set("Content-Type", contentType)
	// 浏览器缓存 24h:看过的图片不再重复回源 OSS
	req.Header.Set("Cache-Control", "public, max-age=86400")
	req.Header.Set("Authorization", "OSS "+s.AccessKey+":"+signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("OSS 返回 HTTP %d: %s", resp.StatusCode, string(body))
	}
	return s.ResolveURL(key), nil
}
```

- [ ] **Step 1.5: 归一化 Endpoint（审查后修订，已实施）**
`NewOssSigner` 需对 `cfg.Endpoint` 做一次归一化后再赋值（抽出 `normalizeEndpoint(host string) string`：剥掉 `http://` / `https://` 前缀与尾斜杠），`endpointHost()` 复用它。这样 PUT 请求 URL、`ResolveURL`、`GeneratePolicy` 三处 host 都基于归一化值；否则当 `OSS_ENDPOINT` 带协议前缀时，上传会成功但落库 URL 会拼成 `https://bucket.https://oss-.../key` 这种坏链接。
同时新增 `houduan/pkg/oss_test.go`（package pkg，表驱动、零网络）：覆盖裸域名 / 带 `https://` 前缀 / 带尾斜杠 三种 Endpoint 输入下 `ResolveURL` 与 `GeneratePolicy` 的 Host、以及 `Configured()` 的四种组合。

- [ ] **Step 2: 爬虫改用 `pkg` 实现，删掉本地重复代码**

`houduan/crawler/main.go`：

1. import 中加 `"wenlv-backend/pkg"`；
2. `ossCfg` 结构体与 `ossUpload` 函数整体删除；
3. `main` 中 OSS 构造改为：

```go
	var signer *pkg.OssSigner
	if *upload || *uploadOnly {
		signer = pkg.NewOssSigner(&pkg.OssConfig{
			Endpoint:  os.Getenv("OSS_ENDPOINT"),
			AccessKey: os.Getenv("OSS_ACCESS_KEY"),
			SecretKey: os.Getenv("OSS_SECRET_KEY"),
			Bucket:    os.Getenv("OSS_BUCKET"),
		})
		if !signer.Configured() {
			log.Fatal("已启用 OSS 上传,但 .env 中 OSS_* 配置不完整")
		}
	}
```

4. 两处调用点改为：

```go
					if ossURL, err := signer.PutObject(local, "scenic/"+s.ID+ext); err != nil {
```
```go
				if ossURL, err := signer.PutObject(local, "food/"+f.ID+ext); err != nil {
```
```go
		ossURL, err := signer.PutObject(local, key)
```
（`uploadLocalToOSS` 的签名改为 `func uploadLocalToOSS(db *gorm.DB, signer *pkg.OssSigner, dir string, spots []spot)`，`runFoodCrawl` 的第二个 OSS 参数类型同样改为 `*pkg.OssSigner`。）

- [ ] **Step 3: 编译校验**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go build ./...
```

Expected: 无输出（编译通过）；`grep -n "func ossUpload" crawler/main.go` 应无结果。

- [ ] **Step 4: 提交（先向用户确认）**

```powershell
git add houduan/pkg/oss.go houduan/crawler/main.go
git commit -m "refactor(oss): 服务端直传下沉到 pkg.OssSigner,爬虫复用去重"
```

---

### Task 3: 高德周边搜索 + POI 详情字段扩展

**Files:**
- Modify: `houduan/service/trip_amap.go`
- Create: `houduan/service/trip_amap_test.go`

**Interfaces:**
- Consumes: `AmapService.apiKey()`、`AmapService.getJSON()`、`toStringValue()`、`parseFlexInt()`、`parseAmapLocation()`（均在同文件已有）
- Produces:
  - `type AroundPOI struct{ ID, Name, Type, Address string; Location Location; Distance int; Tel string }`
  - `(*AmapService) SearchAround(ctx context.Context, lng, lat float64, keywords, types string, radius, offset int) []AroundPOI`
  - `POIDetailResult` 新增 `Rating, Cost, OpenHours string` 与 `BizExt map[string]any`
  - `pickOpenHours(bizExt map[string]any) string`

- [ ] **Step 1: 先确认高德真实回包字段名**

临时在 `houduan/service/trip_amap.go` 的 `GetPOIDetail` 里打一行 `fmt.Printf("POI biz_ext 原始回包: %+v\n", p.BizExt)`，用真实景点跑一次：

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go run ./cmd/scenic-enrich -only 春熙路 -dry-run
```

（该命令在 Task 6 完成后存在；在此之前可临时写一个 `go run` 小脚本或在浏览器里直接请求）
Expected: 打印出的 `biz_ext` 键名与下方 `pickOpenHours` 的候选列表一致（`open_time` / `opentime_today` / `opentime2` 至少命中一个）；若键名不同，按实际回包调整候选列表顺序，确认后删掉这行调试输出。

- [ ] **Step 2: 写失败测试（纯函数）**

创建 `houduan/service/trip_amap_test.go`：

```go
package service

import "testing"

func TestPickOpenHours(t *testing.T) {
	cases := []struct {
		name   string
		bizExt map[string]any
		want   string
	}{
		{"优先 open_time", map[string]any{"open_time": "09:00-17:00", "opentime_today": "08:00-18:00"}, "09:00-17:00"},
		{"回退 opentime_today", map[string]any{"opentime_today": "08:00-18:00"}, "08:00-18:00"},
		{"空数组视为无值", map[string]any{"open_time": "[]"}, ""},
		{"全部缺失", map[string]any{"rating": "4.5"}, ""},
		{"nil 入参", nil, ""},
	}
	for _, c := range cases {
		if got := pickOpenHours(c.bizExt); got != c.want {
			t.Errorf("%s: pickOpenHours=%q, want %q", c.name, got, c.want)
		}
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go test ./service/ -run 'TestPickOpenHours' -v
```

Expected: 编译失败 `undefined: pickOpenHours`

- [ ] **Step 4: 实现 `pickOpenHours` 与 `SearchAround`**

在 `houduan/service/trip_amap.go` 末尾追加：

```go
// pickOpenHours 从高德 biz_ext 中提取开放时间。
// 高德不同 POI 回传的键名不一致,按优先级取第一个非空值;全部为空返回空串,
// 由人工或 LLM 参考值兜底(高德该字段覆盖率不高是已知情况)。
func pickOpenHours(bizExt map[string]any) string {
	for _, k := range []string{"open_time", "opentime_today", "opentime2", "opentime_week"} {
		v := strings.TrimSpace(toStringValue(bizExt[k]))
		if v == "" || v == "[]" {
			continue
		}
		return v
	}
	return ""
}

// pickBizExt 读取 biz_ext 中的单个字符串字段。
func pickBizExt(bizExt map[string]any, key string) string {
	return strings.TrimSpace(toStringValue(bizExt[key]))
}

// AroundPOI 周边 POI(含直线距离,单位米)。
type AroundPOI struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
	Distance int      `json:"distance"`
	Tel      string   `json:"tel,omitempty"`
}

// SearchAround 以坐标为中心做周边搜索(高德 place/around)。
// keywords 与 types 至少给一个;两者都空时用 types 兜底为风景名胜,
// 避免高德返回参数错误。失败时返回 nil 由调用方降级为空数组。
func (s *AmapService) SearchAround(ctx context.Context, lng, lat float64, keywords, types string, radius, offset int) []AroundPOI {
	if s.apiKey() == "" {
		return nil
	}
	if radius <= 0 || radius > 50000 {
		radius = 3000
	}
	if offset <= 0 || offset > 25 {
		offset = 10
	}
	if keywords == "" && types == "" {
		types = "060000"
	}
	params := url.Values{}
	params.Set("location", fmt.Sprintf("%.6f,%.6f", lng, lat))
	params.Set("keywords", keywords)
	params.Set("types", types)
	params.Set("radius", strconv.Itoa(radius))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("page", "1")
	params.Set("extensions", "base")

	var result struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Pois   []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Address  any    `json:"address"`
			Location string `json:"location"`
			Distance any    `json:"distance"`
			Tel      any    `json:"tel"`
		} `json:"pois"`
	}
	if err := s.getJSON(ctx, "https://restapi.amap.com/v3/place/around", params, &result); err != nil {
		fmt.Printf("高德周边搜索失败: %v\n", err)
		return nil
	}
	if result.Status != "1" {
		fmt.Printf("高德周边搜索返回异常: %s\n", result.Info)
		return nil
	}
	out := make([]AroundPOI, 0, len(result.Pois))
	for _, p := range result.Pois {
		out = append(out, AroundPOI{
			ID:       p.ID,
			Name:     p.Name,
			Type:     p.Type,
			Address:  toStringValue(p.Address),
			Location: parseAmapLocation(p.Location),
			Distance: int(parseFlexInt(toStringValue(p.Distance))),
			Tel:      toStringValue(p.Tel),
		})
	}
	return out
}
```

- [ ] **Step 5: `GetPOIDetail` 补 `biz_ext` 字段**

把 `POIDetailResult` 与 `GetPOIDetail` 改为：

```go
// POIDetailResult POI 详情。
type POIDetailResult struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Address   string         `json:"address"`
	Location  Location       `json:"location"`
	Tel       string         `json:"tel,omitempty"`
	Photos    []string       `json:"photos,omitempty"`
	Rating    string         `json:"rating,omitempty"`
	Cost      string         `json:"cost,omitempty"`
	OpenHours string         `json:"open_hours,omitempty"`
	BizExt    map[string]any `json:"biz_ext,omitempty"`
}
```

内层结构体的 `Photos` 之后加一行：

```go
			BizExt   map[string]any `json:"biz_ext"`
```

`detail` 构造后补：

```go
	detail.BizExt = p.BizExt
	detail.Rating = pickBizExt(p.BizExt, "rating")
	detail.Cost = pickBizExt(p.BizExt, "cost")
	detail.OpenHours = pickOpenHours(p.BizExt)
```

- [ ] **Step 6: 运行测试确认通过**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go test ./service/ -run 'TestPickOpenHours' -v
go build ./...
```

Expected: `--- PASS: TestPickOpenHours`；`go build ./...` 无输出。

- [ ] **Step 7: 提交（先向用户确认）**

```powershell
git add houduan/service/trip_amap.go houduan/service/trip_amap_test.go
git commit -m "feat(amap): 新增周边搜索与 POI 详情 biz_ext(评分/人均/开放时间)"
```

---

### Task 4: `ScenicRepo` 采集所需方法

**Files:**
- Modify: `houduan/repository/content_repo.go`

**Interfaces:**
- Consumes: 已有 `ScenicRepo{db *gorm.DB}`
- Produces: `(*ScenicRepo) ListAll() ([]model.ScenicSpot, error)`、`(*ScenicRepo) UpdateFields(id uint, updates map[string]any) error`

- [ ] **Step 1: 追加两个方法**

在 `ScenicRepo.GetByID` 之后追加：

```go
// ListAll 返回全部景点(采集批处理使用,不分页)。
func (r *ScenicRepo) ListAll() ([]model.ScenicSpot, error) {
	var items []model.ScenicSpot
	if err := r.db.Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateFields 按主键更新指定字段(字段名 → 新值),用于采集结果落库。
func (r *ScenicRepo) UpdateFields(id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&model.ScenicSpot{}).Where("id = ?", id).Updates(updates).Error
}
```

- [ ] **Step 2: 编译**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go build ./...
```

Expected: 无输出。

- [ ] **Step 3: 提交（先向用户确认）**

```powershell
git add houduan/repository/content_repo.go
git commit -m "feat(repo): ScenicRepo 增加全量查询与字段更新方法"
```

---

### Task 5: 采集/补全服务 `service/scenic_enrich.go`

**Files:**
- Create: `houduan/service/scenic_enrich.go`
- Create: `houduan/service/scenic_enrich_test.go`

**Interfaces:**
- Consumes: `repository.ScenicRepo`（Task 4）、`AmapService.SearchAround/GetPOIDetail/SearchPOI`（Task 3）、`TripLLM.ChatWithTimeout`、`extractJSONFromResponse`（已有）、`pkg.OssSigner.PutObjectBytes`（Task 2）、`model.ScenicSpot` / `model.ScenicDetailSection`（Task 1）
- Produces:
  - `type EnrichOptions struct{ WithImages, WithLLM, Force bool }`
  - `type EnrichResult struct{ ScenicID uint; Name string; POIMatched bool; Uploaded int; LLMUsed bool; UpdatedFields []string; Note string }`
  - `func NewScenicEnricher(repo *repository.ScenicRepo, amap *AmapService, llm *TripLLM, signer *pkg.OssSigner) *ScenicEnricher`
  - `func (e *ScenicEnricher) Enrich(ctx context.Context, spot model.ScenicSpot, opts EnrichOptions) (*EnrichResult, error)`
  - `func distanceMeters(lng1, lat1, lng2, lat2 float64) float64`
  - `func buildDetailSections(llmSections []model.ScenicDetailSection, images []string) []model.ScenicDetailSection`
  - `func appendEstimatedField(existing, field string) string`

- [ ] **Step 1: 先补测试**

创建 `houduan/service/scenic_enrich_test.go`：

```go
package service

import (
	"testing"

	"wenlv-backend/model"
)

func TestDistanceMeters(t *testing.T) {
	// 春熙路 → 天府广场,实际约 1.1km
	got := distanceMeters(104.0810, 30.6570, 104.0657, 30.6573)
	if got < 1200 || got > 1800 {
		t.Errorf("distanceMeters=%v, 期望 1200~1800 米", got)
	}
	if distanceMeters(0, 0, 0, 0) != 0 {
		t.Error("同点距离应为 0")
	}
}

func TestBuildDetailSections(t *testing.T) {
	images := []string{"cover.jpg", "g1.jpg", "g2.jpg"}
	llm := []model.ScenicDetailSection{
		{Title: "街区沿革", Text: "甲"},
		{Title: "今日风貌", Text: "乙"},
		{Title: "第三段", Text: "丙"},
	}
	got := buildDetailSections(llm, images)
	if len(got) != 3 {
		t.Fatalf("段落数=%d, want 3", len(got))
	}
	if got[0].Image != "g1.jpg" || got[1].Image != "g2.jpg" {
		t.Errorf("配图应从第 2 张开始配对, got %q/%q", got[0].Image, got[1].Image)
	}
	if got[2].Image != "" {
		t.Errorf("图片不足时段落配图应为空, got %q", got[2].Image)
	}
	if got[0].Title != "街区沿革" || got[0].Text != "甲" {
		t.Errorf("段落内容被改写: %+v", got[0])
	}
}

func TestAppendEstimatedField(t *testing.T) {
	if got := appendEstimatedField("", "ticket_price"); got != "ticket_price" {
		t.Errorf("空串应直接写入, got %q", got)
	}
	if got := appendEstimatedField("ticket_price", "ticket_price"); got != "ticket_price" {
		t.Errorf("重复字段不应重复追加, got %q", got)
	}
	if got := appendEstimatedField("ticket_price", "recommend_hours"); got != "ticket_price,recommend_hours" {
		t.Errorf("追加顺序错误, got %q", got)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go test ./service/ -run 'TestBuildDetailSections|TestAppendEstimatedField|TestDistanceMeters' -v
```

Expected: `undefined: distanceMeters` / `undefined: buildDetailSections` 编译失败。

- [ ] **Step 3: 新建 `houduan/service/scenic_enrich.go`**

```go
// 景点详情数据采集与补全:高德 POI 事实字段 + LLM 参考值 + 图文详情分段。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
)

// EnrichOptions 采集开关。
type EnrichOptions struct {
	WithImages bool // 抓取 POI 相册并上传 OSS
	WithLLM    bool // 调用 LLM 生成图文详情与参考值字段
	Force      bool // 已有值时也覆盖(默认只补空字段)
}

// EnrichResult 单个景点的采集结果,用于批处理日志。
type EnrichResult struct {
	ScenicID      uint
	Name          string
	POIMatched    bool
	Uploaded      int
	LLMUsed       bool
	UpdatedFields []string
	Note          string
}

// ScenicEnricher 景点数据采集器。
type ScenicEnricher struct {
	repo    *repository.ScenicRepo
	amap    *AmapService
	llm     *TripLLM
	signer  *pkg.OssSigner
	gallery int // 每个景点最多收录的相册张数
}

// NewScenicEnricher 构造采集器。signer 未配置时自动跳过图片上传。
func NewScenicEnricher(repo *repository.ScenicRepo, amap *AmapService, llm *TripLLM, signer *pkg.OssSigner) *ScenicEnricher {
	return &ScenicEnricher{repo: repo, amap: amap, llm: llm, signer: signer, gallery: 6}
}

// Enrich 采集单个景点并落库。任一步失败只记录 Note 并继续,不返回错误中断批处理。
func (e *ScenicEnricher) Enrich(ctx context.Context, s model.ScenicSpot, opts EnrichOptions) (*EnrichResult, error) {
	res := &EnrichResult{ScenicID: s.ID, Name: s.NameZH}
	updates := map[string]any{}

	// 1) 匹配高德 POI 并取详情(地址/电话/开放时间)
	poiID, poiName, ok := e.MatchPOI(ctx, s)
	var detail *POIDetailResult
	if ok {
		res.POIMatched = true
		d, err := e.amap.GetPOIDetail(ctx, poiID)
		if err != nil {
			res.Note = "POI 详情获取失败: " + err.Error()
		} else {
			detail = d
			updates["amap_poi_id"] = poiID
			if v := strings.TrimSpace(d.Address); v != "" {
				updates["address"] = v
			}
			if v := strings.TrimSpace(d.Tel); v != "" {
				updates["tel"] = v
			}
			if v := strings.TrimSpace(d.OpenHours); v != "" {
				updates["open_hours"] = v
			}
			res.Note = "匹配 POI: " + poiName
		}
	} else {
		res.Note = "未匹配到高德 POI"
	}

	// 2) 相册图片:下载后上传 OSS,数据库只存 OSS URL
	images := splitComma(s.GalleryImages)
	if opts.WithImages && detail != nil && len(detail.Photos) > 0 && (opts.Force || len(images) == 0) {
		uploaded, err := e.uploadPhotos(ctx, s, detail.Photos)
		if err != nil {
			res.Note += "; 图片上传失败: " + err.Error()
		} else if len(uploaded) > 0 {
			images = uploaded
			res.Uploaded = len(uploaded)
			updates["gallery_images"] = strings.Join(uploaded, ",")
		}
	}
	if cover := strings.TrimSpace(s.Images); cover == "" && len(images) > 0 {
		updates["images"] = images[0]
	}

	// 3) LLM:图文详情 + 参考值字段
	estimated := s.EstimatedFields
	if opts.WithLLM && e.llm != nil && e.llm.Available() {
		sections, err := e.buildLLMSections(ctx, s, images)
		if err != nil {
			res.Note += "; 图文详情生成失败: " + err.Error()
		} else if len(sections) > 0 {
			if raw, err := json.Marshal(sections); err == nil {
				updates["detail_sections"] = string(raw)
				res.LLMUsed = true
			}
		}
		est, err := e.fillEstimated(ctx, s)
		if err != nil {
			res.Note += "; 参考值生成失败: " + err.Error()
		} else {
			filled := 0
			for _, kv := range est {
				if v := strings.TrimSpace(kv.Value); v != "" {
					updates[kv.Field] = v
					estimated = appendEstimatedField(estimated, kv.Field)
					filled++
				}
			}
			if filled > 0 {
				res.LLMUsed = true
			}
		}
	}

	// 4) 非空字段保护:非 Force 模式下不覆盖已有事实值
	if !opts.Force {
		for key := range updates {
			if !isEmptyField(s, key) {
				delete(updates, key)
			}
		}
	}

	updates["estimated_fields"] = estimated
	updates["data_source"] = buildDataSource(res.POIMatched, res.LLMUsed)
	now := time.Now()
	updates["data_updated_at"] = now
	for k := range updates {
		res.UpdatedFields = append(res.UpdatedFields, k)
	}
	if err := e.repo.UpdateFields(s.ID, updates); err != nil {
		return res, err
	}
	return res, nil
}

// estimatedKV LLM 参考值字段。
type estimatedKV struct {
	Field string
	Value string
}

// matchPOI 在候选 POI 中挑最可信的一个:名称完全相同 > 名称互相包含,
// 并在库内坐标有效时要求直线距离 < 5km,避免同名景点跨城市错配。
// 先用 isUsablePOI 过滤掉「地名地址信息」(热点地名/道路名,只有区县级地址、无电话与开放时间)
// 与「交通设施服务」(地铁站/公交站/停车场) —— 实测搜「春熙路」时精确同名规则会选中
// 地名类 POI(地址仅「锦江区」),而数据更全的「春熙路步行街」只算"包含"而落选。
// 过滤后若无候选则返回 not-ok,不做回退:宁可不补数据,也不写区县级垃圾地址。
func (e *ScenicEnricher) MatchPOI(ctx context.Context, s model.ScenicSpot) (string, string, bool) {
	candidates := e.amap.SearchPOI(ctx, s.NameZH, "成都", true)
	if len(candidates) == 0 {
		candidates = e.amap.SearchPOI(ctx, s.NameZH, "", false)
	}
	best, bestName, bestScore := "", "", -1
	for _, c := range candidates {
		if !isUsablePOI(c.Type) {
			continue
		}
		score := 0
		if c.Name == s.NameZH {
			score = 3
		} else if strings.Contains(c.Name, s.NameZH) || strings.Contains(s.NameZH, c.Name) {
			score = 2
		}
		if score == 0 {
			continue
		}
		if s.Lat != 0 && s.Lng != 0 && c.Location.Latitude != 0 && c.Location.Longitude != 0 {
			if distanceMeters(s.Lng, s.Lat, c.Location.Longitude, c.Location.Latitude) > 5000 {
				continue
			}
			score++
		}
		if score > bestScore {
			best, bestName, bestScore = c.ID, c.Name, score
		}
	}
	return best, bestName, best != ""
}

// unusablePOITypes 不可用的高德 POI 类型关键词。
// 「地名地址信息」(热点地名/道路名) 只有区县级地址、无电话与开放时间;
// 「交通设施服务」(地铁站/公交站/停车场) 是交通设施而非景点,都不能作为景点数据来源。
var unusablePOITypes = []string{"地名地址信息", "交通设施服务", "地铁站", "公交站", "停车场"}

// isUsablePOI 判断高德返回的 POI type 是否可作为景点数据来源。
func isUsablePOI(poiType string) bool {
	for _, bad := range unusablePOITypes {
		if strings.Contains(poiType, bad) {
			return false
		}
	}
	return true
}

// uploadPhotos 下载 POI 图片并上传 OSS,返回 OSS URL 列表。
// 单张失败跳过,不影响其余图片。
func (e *ScenicEnricher) uploadPhotos(ctx context.Context, s model.ScenicSpot, photos []string) ([]string, error) {
	if e.signer == nil || !e.signer.Configured() {
		return nil, fmt.Errorf("OSS 未配置")
	}
	var out []string
	for i, u := range photos {
		if len(out) >= e.gallery {
			break
		}
		body, ct, err := fetchBytes(ctx, u)
		if err != nil {
			continue
		}
		ext := extFromContentType(ct, u)
		key := fmt.Sprintf("scenic/%d/g%d%s", s.ID, i+1, ext)
		ossURL, err := e.signer.PutObjectBytes(body, key, ct)
		if err != nil {
			continue
		}
		out = append(out, ossURL)
	}
	return out, nil
}

// buildLLMSections 调 LLM 把已有简介改写成 2-4 段图文详情,并按顺序配相册图。
func (e *ScenicEnricher) buildLLMSections(ctx context.Context, s model.ScenicSpot, images []string) ([]model.ScenicDetailSection, error) {
	if strings.TrimSpace(s.Desc) == "" {
		return nil, nil
	}
	prompt := fmt.Sprintf(`你是成都文旅内容编辑。请为景点"%s"撰写图文详情,严格只输出 JSON,不要输出解释:
{"sections":[{"title":"小标题(6-12字)","text":"段落(80-150字)"}]}
参考简介:%s
要求:写成 2-4 段;只使用参考简介里能确认的信息;不确定的内容不要写;不要编造数字、价格、年份;不要使用 Markdown 语法;使用简体中文。`,
		s.NameZH, s.Desc)
	reply, err := e.llm.ChatWithTimeout(ctx, 90, []llmMessage{
		{Role: "system", Content: "你是严谨的文旅内容编辑,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.3, 1200)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Sections []model.ScenicDetailSection `json:"sections"`
	}
	if err := json.Unmarshal([]byte(extractJSONFromResponse(reply)), &parsed); err != nil {
		return nil, err
	}
	return buildDetailSections(parsed.Sections, images), nil
}

// fillEstimated 让 LLM 给出无权威公开源的字段参考值,未知填空串。
func (e *ScenicEnricher) fillEstimated(ctx context.Context, s model.ScenicSpot) ([]estimatedKV, error) {
	prompt := fmt.Sprintf(`请根据下列已知信息,给出成都景点"%s"的参考信息。严格只输出 JSON,不要输出解释:
{"ticket_price":"","open_hours":"","recommend_hours":"","yearly_visitors":""}
已知信息:
- 所属区县:%s
- 简介:%s
- 高德开放时间:%s
规则:
1. 这些值会作为"参考值"展示,必须保守,宁缺勿错;
2. ticket_price 写具体票制,如"免费""成人60元/人";不确定写空字符串;
3. open_hours 仅在已知信息里没有开放时间时才填;有则留空字符串;
4. recommend_hours 写"2-3小时"这类区间;yearly_visitors 写"约500万人次",不确定写空字符串;
5. 不要编造精确数字、年份、官方排名。`,
		s.NameZH, s.District, s.Desc, s.OpenHours)
	reply, err := e.llm.ChatWithTimeout(ctx, 60, []llmMessage{
		{Role: "system", Content: "你是谨慎的文旅信息编辑,只输出合法 JSON。"},
		{Role: "user", Content: prompt},
	}, 0.2, 600)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		TicketPrice    string `json:"ticket_price"`
		OpenHours      string `json:"open_hours"`
		RecommendHours string `json:"recommend_hours"`
		YearlyVisitors string `json:"yearly_visitors"`
	}
	if err := json.Unmarshal([]byte(extractJSONFromResponse(reply)), &parsed); err != nil {
		return nil, err
	}
	return []estimatedKV{
		{Field: "ticket_price", Value: parsed.TicketPrice},
		{Field: "open_hours", Value: parsed.OpenHours},
		{Field: "recommend_hours", Value: parsed.RecommendHours},
		{Field: "yearly_visitors", Value: parsed.YearlyVisitors},
	}, nil
}

// buildDetailSections 把 LLM 段落与相册图配对:第 1 张图是封面,段落从第 2 张开始配。
func buildDetailSections(llmSections []model.ScenicDetailSection, images []string) []model.ScenicDetailSection {
	out := make([]model.ScenicDetailSection, 0, len(llmSections))
	for i, sec := range llmSections {
		title := strings.TrimSpace(sec.Title)
		text := strings.TrimSpace(sec.Text)
		if title == "" || text == "" {
			continue
		}
		img := ""
		if idx := i + 1; idx < len(images) {
			img = images[idx]
		}
		out = append(out, model.ScenicDetailSection{Title: title, Text: text, Image: img})
	}
	return out
}

// appendEstimatedField 以逗号分隔追加参考值字段名,已存在则不重复。
func appendEstimatedField(existing, field string) string {
	if field == "" {
		return existing
	}
	for _, f := range splitComma(existing) {
		if f == field {
			return existing
		}
	}
	if strings.TrimSpace(existing) == "" {
		return field
	}
	return existing + "," + field
}

// buildDataSource 拼装数据来源标记。
func buildDataSource(poiMatched, llmUsed bool) string {
	switch {
	case poiMatched && llmUsed:
		return "amap+llm"
	case poiMatched:
		return "amap"
	case llmUsed:
		return "llm"
	default:
		return "wiki"
	}
}

// isEmptyField 判断库中该字段是否为空(用于非 Force 模式跳过覆盖)。
func isEmptyField(s model.ScenicSpot, field string) bool {
	switch field {
	case "address":
		return strings.TrimSpace(s.Address) == ""
	case "tel":
		return strings.TrimSpace(s.Tel) == ""
	case "open_hours":
		return strings.TrimSpace(s.OpenHours) == ""
	case "ticket_price":
		return strings.TrimSpace(s.TicketPrice) == ""
	case "recommend_hours":
		return strings.TrimSpace(s.RecommendHours) == ""
	case "yearly_visitors":
		return strings.TrimSpace(s.YearlyVisitors) == ""
	case "gallery_images":
		return strings.TrimSpace(s.GalleryImages) == ""
	case "detail_sections":
		return strings.TrimSpace(s.DetailSections) == ""
	case "images":
		return strings.TrimSpace(s.Images) == ""
	case "amap_poi_id":
		return strings.TrimSpace(s.AmapPOIID) == ""
	default:
		return true
	}
}

// distanceMeters 用 Haversine 公式计算两点直线距离(米)。
// 经度/lat 顺序与高德一致:先经度后纬度。
func distanceMeters(lng1, lat1, lng2, lat2 float64) float64 {
	if lng1 == lng2 && lat1 == lat2 {
		return 0
	}
	const earthRadius = 6371000.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadius * math.Asin(math.Sqrt(a))
}

// splitComma 拆分逗号分隔串并去空。
func splitComma(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// extFromContentType 按响应类型推断图片后缀,推断不出时回退 URL 后缀,再回退 .jpg。
func extFromContentType(contentType, rawURL string) string {
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	}
	if ext := strings.ToLower(filepath.Ext(rawURL)); ext != "" && len(ext) <= 6 {
		return ext
	}
	return ".jpg"
}
```

`fetchBytes` 放在新文件 `houduan/service/scenic_enrich_http.go`：

```go
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// fetchBytes 下载外部图片,返回内容与 Content-Type。
// 限制 20MB,超时 20s;高德图片 CDN 可能要求 Referer,这里统一带普通 UA。
func fetchBytes(ctx context.Context, rawURL string) ([]byte, string, error) {
	if rawURL == "" {
		return nil, "", fmt.Errorf("空 URL")
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; wenlv-enricher/1.0)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return nil, "", err
	}
	if len(body) == 0 {
		return nil, "", fmt.Errorf("空响应体")
	}
	return body, resp.Header.Get("Content-Type"), nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go test ./service/ -run 'TestBuildDetailSections|TestAppendEstimatedField|TestPickOpenHours|TestDistanceMeters' -v
go build ./...
```

Expected: 4 个测试全部 `PASS`；`go build ./...` 无输出。

- [ ] **Step 5: 提交（先向用户确认）**

```powershell
git add houduan/service/scenic_enrich.go houduan/service/scenic_enrich_http.go houduan/service/scenic_enrich_test.go
git commit -m "feat(scenic): 新增景点详情采集服务(高德POI事实字段+LLM参考值+图文详情分段)"
```

---

### Task 6: 批处理入口 `cmd/scenic-enrich`

**Files:**
- Create: `houduan/cmd/scenic-enrich/main.go`

**Interfaces:**
- Consumes: `service.NewScenicEnricher`、`service.EnrichOptions`（Task 5）、`service.NewTripSettings`/`NewAmapService`/`NewTripLLM`（已有）、`config.Load()`、`database.InitMySQL`/`MustAutoMigrate`、`repository.NewScenicRepo`、`pkg.NewOssSigner`
- Produces: 可执行入口，命令行参数 `-limit -only -dry-run -with-llm -with-images -force`

- [ ] **Step 1: 新建 `houduan/cmd/scenic-enrich/main.go`**

```go
// scenic-enrich 景点详情数据补全批处理。
// 数据源:高德 Web 服务(地址/电话/开放时间/相册图) + LLM(图文详情与参考值)。
// 产出:字段写入 MySQL scenic_spots,相册图上传 OSS。
//
// 用法(在 houduan 目录下执行):
//
//	go run ./cmd/scenic-enrich -limit 3 -dry-run                      # 试跑前 3 个,不落库
//	go run ./cmd/scenic-enrich -only 春熙路 -with-llm -with-images    # 重跑单个景点
//	go run ./cmd/scenic-enrich                                        # 全量
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"wenlv-backend/config"
	"wenlv-backend/database"
	"wenlv-backend/model"
	"wenlv-backend/pkg"
	"wenlv-backend/repository"
	"wenlv-backend/service"
)

func main() {
	limit := flag.Int("limit", 0, "仅处理前 N 个景点(0 表示全部)")
	only := flag.String("only", "", "仅处理名称包含该关键字的景点")
	dryRun := flag.Bool("dry-run", false, "只打印将要写入的内容,不写库、不上传")
	withImages := flag.Bool("with-images", true, "抓取相册图并上传 OSS")
	withLLM := flag.Bool("with-llm", true, "调用 LLM 生成图文详情与参考值")
	force := flag.Bool("force", false, "已有值时也覆盖(默认只补空字段)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	db := mustDB(cfg)
	repo := repository.NewScenicRepo(db)

	// 与 main.go 一致:高德/LLM Key 先取 .env,再被 data/runtime_settings.json 覆盖(设置页热更新过的值)
	tripSettings := service.NewTripSettings(service.TripSettingsOptions{
		Defaults: service.TripRuntimeSettings{
			ViteAmapWebKey: cfg.Trip.AmapWebKey,
			GoogleMapsAPIKey: cfg.Trip.GoogleMapsAPIKey,
			GoogleMapsProxy:  cfg.Trip.GoogleMapsProxy,
			OpenAIAPIKey:     cfg.Trip.LLMAPIKey,
			OpenAIBaseURL:    cfg.Trip.LLMBaseURL,
			OpenAIModel:      cfg.Trip.LLMModel,
		},
		DataDir:            cfg.Trip.DataDir,
		PlannerTimeout:     cfg.Trip.PlannerTimeout,
		LLMTimeout:         cfg.Trip.LLMTimeout,
		EnableUserMemory:   cfg.Trip.EnableUserMemory,
		ImageCacheTTLHours: cfg.Trip.ImageCacheTTLHours,
		ImageCacheMaxMB:    cfg.Trip.ImageCacheMaxMB,
	})
	amap := service.NewAmapService(tripSettings)
	llm := service.NewTripLLM(tripSettings)
	signer := pkg.NewOssSigner(&pkg.OssConfig{
		Endpoint:  os.Getenv("OSS_ENDPOINT"),
		AccessKey: os.Getenv("OSS_ACCESS_KEY"),
		SecretKey: os.Getenv("OSS_SECRET_KEY"),
		Bucket:    os.Getenv("OSS_BUCKET"),
	})
	if !amap.Available() {
		log.Fatal("高德 Web 服务 Key 未配置:请在 .env 的 AMAP_WEB_KEY 或前端设置页配置后重试")
	}
	if !signer.Configured() {
		log.Println("提示:OSS 未配置,将跳过相册图片采集")
	}
	if !llm.Available() {
		log.Println("提示:LLM 未配置,将跳过图文详情与参考值生成")
	}

	spots, err := repo.ListAll()
	if err != nil {
		log.Fatalf("读取景点失败: %v", err)
	}
	if *only != "" {
		filtered := make([]model.ScenicSpot, 0, 1)
		for _, s := range spots {
			if strings.Contains(s.NameZH, *only) {
				filtered = append(filtered, s)
			}
		}
		spots = filtered
	}
	if *limit > 0 && *limit < len(spots) {
		spots = spots[:*limit]
	}
	log.Printf("待处理景点 %d 个(dry-run=%v, with-llm=%v, with-images=%v)", len(spots), *dryRun, *withLLM, *withImages)

	// dry-run 时不落库、不上传:把 enricher 的写库动作短路
	if *dryRun {
		log.Println("dry-run 模式:仅打印匹配结果,不写库")
	}

	enricher := service.NewScenicEnricher(repo, amap, llm, signer)
	ctx := context.Background()
	okCnt, failCnt, skipCnt := 0, 0, 0
	for i, s := range spots {
		log.Printf("[%d/%d] %s", i+1, len(spots), s.NameZH)
		if *dryRun {
			// dry-run 只验证 POI 匹配与字段取值,直接调 AmapService 打印
			printDryRun(ctx, amap, s)
			skipCnt++
			continue
		}
		res, err := enricher.Enrich(ctx, s, service.EnrichOptions{
			WithImages: *withImages,
			WithLLM:    *withLLM,
			Force:      *force,
		})
		if err != nil {
			log.Printf("    落库失败: %v", err)
			failCnt++
			continue
		}
		log.Printf("    POI匹配=%v 上传图片=%d LLM=%v 更新字段=%v 备注=%s",
			res.POIMatched, res.Uploaded, res.LLMUsed, res.UpdatedFields, res.Note)
		okCnt++
		time.Sleep(300 * time.Millisecond) // 控制高德 QPS
	}
	log.Printf("完成: 成功 %d / 失败 %d / 试跑跳过 %d", okCnt, failCnt, skipCnt)
}

// printDryRun 打印高德匹配与字段取值,供人工核对后再正式跑。
func printDryRun(ctx context.Context, amap *service.AmapService, s model.ScenicSpot) {
	pois := amap.SearchPOI(ctx, s.NameZH, "成都", true)
	if len(pois) == 0 {
		log.Printf("    未匹配到高德 POI")
		return
	}
	p := pois[0]
	log.Printf("    候选 POI: %s / %s / %s", p.ID, p.Name, p.Address)
	d, err := amap.GetPOIDetail(ctx, p.ID)
	if err != nil {
		log.Printf("    详情获取失败: %v", err)
		return
	}
	log.Printf("    地址=%s 电话=%s 开放时间=%q 相册=%d 张", d.Address, d.Tel, d.OpenHours, len(d.Photos))
}

// mustDB 连接数据库并确保表结构最新(新增列需要迁移)。
func mustDB(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		log.Fatalf("MySQL 连接失败: %v", err)
	}
	database.MustAutoMigrate(db)
	return db
}
```

- [ ] **Step 2: 试跑 dry-run 核对字段**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go run ./cmd/scenic-enrich -limit 3 -dry-run
```

Expected: 打印 3 个景点的候选 POI、地址、电话、开放时间（`开放时间` 允许为空串）、相册张数 > 0；若有景点打印"未匹配到高德 POI"，记录名称，后续在 Task 10 的处理清单里列出。

- [ ] **Step 3: 正式跑 3 个景点验证落库**

```powershell
go run ./cmd/scenic-enrich -limit 3
```

Expected: 每个景点打印 `POI匹配=true 上传图片=6 LLM=true 更新字段=[...]`；相册 URL 形如 `https://wenlv-tdx.oss-cn-chengdu.aliyuncs.com/scenic/12/g1.jpg`。

查库确认：

```powershell
mysql -uroot wenlv -e "SELECT id,name_zh,address,tel,open_hours,LEFT(ticket_price,20) AS ticket_price,estimated_fields,data_source FROM scenic_spots WHERE id IN (SELECT id FROM scenic_spots ORDER BY id LIMIT 3);"
```

Expected: `address`/`tel` 非空；`ticket_price`/`recommend_hours`/`yearly_visitors` 非空且 `estimated_fields` 含这三个字段名。

- [ ] **Step 4: 提交（先向用户确认）**

```powershell
git add houduan/cmd/scenic-enrich/main.go
git commit -m "feat(scenic): 新增景点详情补全批处理入口(高德POI+LLM参考值,支持dry-run)"
```

---

### Task 7: 周边推荐与交通站点只读接口

**Files:**
- Create: `houduan/service/scenic_extra.go`
- Create: `houduan/service/scenic_extra_test.go`
- Create: `houduan/handler/scenic_extra_handler.go`
- Modify: `houduan/config/config.go`
- Modify: `houduan/.env.example`
- Modify: `houduan/handler/bootstrap.go`
- Modify: `houduan/router/router.go`
- Modify: `houduan/main.go`

**Interfaces:**
- Consumes: `AmapService.SearchAround`（Task 3）、`repository.ScenicRepo.GetByID`、`*redis.Client`（go-redis/v9）
- Produces:
  - `type AroundItem struct{ ID, Name, Type, Address string; Distance int; Lat, Lng float64 }`
  - `type TransitStop struct{ Name, Type string; Distance int }`
  - `func NewScenicExtraService(repo *repository.ScenicRepo, amap *AmapService, rdb *redis.Client, cacheTTL time.Duration) *ScenicExtraService`
  - `func (s *ScenicExtraService) Around(ctx context.Context, id uint, limit int) ([]AroundItem, error)`
  - `func (s *ScenicExtraService) Transport(ctx context.Context, id uint) ([]TransitStop, error)`
  - `config.Trip.ScenicCacheTTLHours int`（来自 `SCENIC_CACHE_TTL_HOURS`，默认 24）
  - `handler.NewScenicExtraHandler(svc *service.ScenicExtraService) *ScenicExtraHandler`，方法 `Around(c *gin.Context)` / `Transport(c *gin.Context)`

- [ ] **Step 1: 新增 `SCENIC_CACHE_TTL_HOURS` 配置项**

`houduan/config/config.go`：`Trip` 结构体在 `ImageCacheMaxMB int` 之后加一行：

```go
		ImageCacheTTLHours int
		ImageCacheMaxMB    int
		// ScenicCacheTTLHours 景点详情页实时数据(周边推荐/交通站点)的 Redis 缓存小时数。
		ScenicCacheTTLHours int
```

`houduan/config/config.go` 的 `Load()` 中，在 `c.Trip.ImageCacheMaxMB = getEnvInt("TRIP_IMAGE_CACHE_MAX_MB", 512)` 之后加：

```go
	c.Trip.ScenicCacheTTLHours = getEnvInt("SCENIC_CACHE_TTL_HOURS", 24)
```

`houduan/.env.example` 在 `TRIP_IMAGE_CACHE_MAX_MB=512` 之后加：

```
# 景点详情页实时数据(周边推荐/交通站点)的 Redis 缓存小时数
SCENIC_CACHE_TTL_HOURS=24
```

- [ ] **Step 2: 新建 `houduan/service/scenic_extra.go`**

```go
// 景点详情页的实时数据:周边推荐与邻近交通站点。
// 不落 MySQL,只缓存在 Redis(cacheTTL 由 SCENIC_CACHE_TTL_HOURS 配置,默认 24h);
// Redis 不可用时直接回源高德,不影响页面。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"

	"wenlv-backend/repository"
)

// AroundItem 周边推荐条目。
type AroundItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Address  string  `json:"address"`
	Distance int     `json:"distance"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

// TransitStop 邻近交通站点。
type TransitStop struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Distance int    `json:"distance"`
}

// ScenicExtraService 景点实时信息查询。
type ScenicExtraService struct {
	repo     *repository.ScenicRepo
	amap     *AmapService
	rdb      *redis.Client
	cacheTTL time.Duration
}

// NewScenicExtraService 构造服务。cacheTTL <= 0 时回退 24 小时。
func NewScenicExtraService(repo *repository.ScenicRepo, amap *AmapService, rdb *redis.Client, cacheTTL time.Duration) *ScenicExtraService {
	if cacheTTL <= 0 {
		cacheTTL = 24 * time.Hour
	}
	return &ScenicExtraService{repo: repo, amap: amap, rdb: rdb, cacheTTL: cacheTTL}
}

// Around 返回景点周边 3km 内的景点/餐饮/购物 POI,按距离升序;失败返回空切片而不是错误。
func (s *ScenicExtraService) Around(ctx context.Context, id uint, limit int) ([]AroundItem, error) {
	if limit <= 0 || limit > 12 {
		limit = 6
	}
	key := fmt.Sprintf("scenic:around:%d:%d", id, limit)
	if cached := s.getCache(ctx, key); cached != nil {
		return cached, nil
	}
	spot, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	out := []AroundItem{}
	if spot.Lng == 0 || spot.Lat == 0 {
		return out, nil
	}
	// 修订(实施后):只取"风景名胜(110000)+餐饮服务(050000)",去掉"购物服务(060000)"——
	// 实测郊区的凤凰湖湿地公园 3km 内会把"龙达五金/兴扬居建材"当周边推荐;同时缓存 key 加 v2 版本号绕过旧缓存。
	pois := s.amap.SearchAround(ctx, spot.Lng, spot.Lat, "", "050000|110000", 3000, 20)
	sort.SliceStable(pois, func(i, j int) bool { return pois[i].Distance < pois[j].Distance })
	for _, p := range pois {
		if p.Name == spot.NameZH {
			continue
		}
		out = append(out, AroundItem{
			ID: p.ID, Name: p.Name, Type: p.Type, Address: p.Address,
			Distance: p.Distance, Lat: p.Location.Latitude, Lng: p.Location.Longitude,
		})
		if len(out) >= limit {
			break
		}
	}
	s.setCache(ctx, key, out)
	return out, nil
}

// Transport 返回景点附近 1.5km 内的地铁站与公交站,按距离升序。
func (s *ScenicExtraService) Transport(ctx context.Context, id uint) ([]TransitStop, error) {
	key := fmt.Sprintf("scenic:transit:%d", id)
	if cached := s.getTransitCache(ctx, key); cached != nil {
		return cached, nil
	}
	spot, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	out := []TransitStop{}
	if spot.Lng == 0 || spot.Lat == 0 {
		return out, nil
	}
	// 用高德 typecode 精确筛选交通设施:150700=地铁站,150500=公交车站。
	// 实测(春熙路坐标):keywords="地铁站|公交站" 会混入停车场/酒店等无关 POI,
	// 而空 keywords + types="150700|150500" 返回的全是公交站/地铁站。
	pois := s.amap.SearchAround(ctx, spot.Lng, spot.Lat, "", "150700|150500", 1500, 10)
	sort.SliceStable(pois, func(i, j int) bool { return pois[i].Distance < pois[j].Distance })
	for _, p := range pois {
		st := TransitStop{Name: p.Name, Type: p.Type, Distance: p.Distance}
		out = append(out, st)
		if len(out) >= 4 {
			break
		}
	}
	s.setTransitCache(ctx, key, out)
	return out, nil
}

// getCache 读取 JSON 缓存,未命中或 Redis 不可用时返回 nil。
func (s *ScenicExtraService) getCache(ctx context.Context, key string) []AroundItem {
	if s.rdb == nil {
		return nil
	}
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil || len(raw) == 0 {
		return nil
	}
	var items []AroundItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

// setCache 写入 JSON 缓存。空结果不写缓存:避免把"高德暂时失败/暂无数据"
// 固化一整个 TTL,宁可下次请求再回源。
func (s *ScenicExtraService) setCache(ctx context.Context, key string, items []AroundItem) {
	if s.rdb == nil || len(items) == 0 {
		return
	}
	if raw, err := json.Marshal(items); err == nil {
		_ = s.rdb.Set(ctx, key, raw, s.cacheTTL).Err()
	}
}

func (s *ScenicExtraService) getTransitCache(ctx context.Context, key string) []TransitStop {
	if s.rdb == nil {
		return nil
	}
	raw, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil || len(raw) == 0 {
		return nil
	}
	var items []TransitStop
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

// setTransitCache 写入交通站点缓存,规则同 setCache(空结果不缓存)。
func (s *ScenicExtraService) setTransitCache(ctx context.Context, key string, items []TransitStop) {
	if s.rdb == nil || len(items) == 0 {
		return
	}
	if raw, err := json.Marshal(items); err == nil {
		_ = s.rdb.Set(ctx, key, raw, s.cacheTTL).Err()
	}
}
```

- [ ] **Step 3: 新建 `houduan/handler/scenic_extra_handler.go`**

```go
package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// scenicClientCacheSeconds 浏览器端缓存秒数:5 分钟内同 URL 不再请求后端,
// 由后端 Redis(TTL 见 SCENIC_CACHE_TTL_HOURS)兜住；用户强刷可立即更新。
const scenicClientCacheSeconds = 300

// ScenicExtraHandler 景点详情页实时数据(周边/交通)接口。
type ScenicExtraHandler struct {
	svc *service.ScenicExtraService
}

// NewScenicExtraHandler 构造处理器。
func NewScenicExtraHandler(svc *service.ScenicExtraService) *ScenicExtraHandler {
	return &ScenicExtraHandler{svc: svc}
}

// Around 周边推荐:GET /api/scenic/:id/around?limit=6
// 查询失败或景点无坐标时返回空数组,前端显示"暂无数据"。
func (h *ScenicExtraHandler) Around(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的景点 ID")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	items, err := h.svc.Around(c.Request.Context(), id, limit)
	if handleNotFound(c, err) {
		return
	}
	if err != nil || len(items) == 0 {
		// 高德不可用或无数据时降级为空列表;空结果不加浏览器缓存头,便于恢复后立刻可见
		pkg.OK(c, []service.AroundItem{})
		return
	}
	setScenicClientCache(c)
	pkg.OK(c, items)
}

// Transport 邻近交通站点:GET /api/scenic/:id/transport
func (h *ScenicExtraHandler) Transport(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		pkg.BadRequest(c, "无效的景点 ID")
		return
	}
	items, err := h.svc.Transport(c.Request.Context(), id)
	if handleNotFound(c, err) {
		return
	}
	if err != nil || len(items) == 0 {
		pkg.OK(c, []service.TransitStop{})
		return
	}
	setScenicClientCache(c)
	pkg.OK(c, items)
}

// setScenicClientCache 设置浏览器缓存头(仅在有数据时调用)。
func setScenicClientCache(c *gin.Context) {
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", scenicClientCacheSeconds))
}
```

- [ ] **Step 4: 装配进 Bootstrap、路由与 main**

`houduan/handler/bootstrap.go` 增加字段：

```go
	HotTopic *HotTopicHandler
	// ScenicExtra 景点详情页实时数据(周边推荐/交通站点)
	ScenicExtra *ScenicExtraHandler
```

`houduan/router/router.go` 在 `/scenic/:id` 之后追加：

```go
	api.GET("/scenic/:id/around", h.ScenicExtra.Around)
	api.GET("/scenic/:id/transport", h.ScenicExtra.Transport)
```

`houduan/main.go` 在 `articleSvc` 之后构造服务并挂到 Bootstrap：

```go
	// 景点详情页实时数据:Redis 缓存时长由 SCENIC_CACHE_TTL_HOURS 配置(默认 24h)
	scenicExtraSvc := service.NewScenicExtraService(
		scenicRepo, tripAmap, rdb,
		time.Duration(cfg.Trip.ScenicCacheTTLHours)*time.Hour,
	)
```
```go
		HotTopic:    handler.NewHotTopicHandler(hotTopicSvc),
		ScenicExtra: handler.NewScenicExtraHandler(scenicExtraSvc),
```

注意：`tripAmap` 与 `rdb` 都在 `scenicExtraSvc` 之前就已构造（`tripAmap` 在第 107 行附近），若顺序不满足，把 `scenicExtraSvc := ...` 移到 `tripAmap` 之后。

- [ ] **Step 5: 接口验证**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go build ./...
$env:SERVER_PORT="8081"; go run .
```

另开终端（本机未安装 `redis-cli`，缓存验证用 Go 测试，见下）：

```powershell
curl.exe -si "http://localhost:8081/api/scenic/3/around?limit=6" | Select-Object -First 12
curl.exe -s  "http://localhost:8081/api/scenic/3/transport"
```

Expected:
- 两个接口都返回 `{"code":0,...}` 且 `data` 是数组；`/around` 的条目带 `name/type/address/distance`；
- 响应头含 `Cache-Control: public, max-age=300`（仅当返回非空数组时）；
- 景点无坐标时 `data` 为 `[]`，且响应头**没有** `Cache-Control`。

缓存与 TTL 由单元测试验证（新增 `houduan/service/scenic_extra_test.go`）：连本机 Redis（`REDIS_ADDR`，默认 `127.0.0.1:6379`），Redis 不可用则 `t.Skip`；断言 `setCache`→`getCache` 往返一致、`TTL` 落在 (0, cacheTTL]、空切片不写键。

```powershell
go test ./service/ -count=1 -run 'TestScenicExtraCache' -v
```

Expected: PASS（Redis 未启动时输出 SKIP，并在报告里注明）。

- [ ] **Step 6: 提交（先向用户确认）**

```powershell
git add houduan/config/config.go houduan/.env.example houduan/service/scenic_extra.go houduan/handler/scenic_extra_handler.go houduan/handler/bootstrap.go houduan/router/router.go houduan/main.go
git commit -m "feat(scenic): 新增周边推荐与交通站点接口(高德实时查询,Redis TTL可配+浏览器5分钟缓存)"
```

---

### Task 8: 前端数据层与展示工具

**Files:**
- Modify: `wennv/wenlv/src/api/content.ts`
- Create: `wennv/wenlv/src/utils/scenicDetail.ts`
- Create: `wennv/wenlv/src/utils/scenicDetail.spec.ts`

**Interfaces:**
- Consumes: `get`（`./request`）
- Produces:
  - `ScenicItem` 新增可选字段 `address/tel/open_hours/ticket_price/recommend_hours/yearly_visitors/gallery_images/detail_sections/estimated_fields/data_source/data_updated_at`
  - `interface ScenicDetailSection{ title: string; text: string; image?: string }`
  - `interface ScenicAroundItem{ id: string; name: string; type?: string; address?: string; distance?: number; lat?: number; lng?: number }`
  - `interface ScenicTransitStop{ name: string; type?: string; distance?: number }`
  - `getScenicAround(id: number, limit?: number): Promise<ScenicAroundItem[]>`
  - `getScenicTransport(id: number): Promise<ScenicTransitStop[]>`
  - `splitList(value?: string): string[]`、`parseSections(value?: string): ScenicDetailSection[]`、`estimatedSet(value?: string): Set<string>`、`displayFact(value: string | undefined, placeholder: string): string`

- [ ] **Step 1: 扩展 `api/content.ts`**

`ScenicItem` 内追加字段，并在文件末尾（`getScenic` 之后）追加新接口：

```ts
export interface ScenicDetailSection {
  title: string
  text: string
  image?: string
}

export interface ScenicAroundItem {
  id: string
  name: string
  type?: string
  address?: string
  distance?: number
  lat?: number
  lng?: number
}

export interface ScenicTransitStop {
  name: string
  type?: string
  distance?: number
}

/** 周边推荐(高德实时查询,后端 Redis 缓存 24h) */
export function getScenicAround(id: number, limit = 6): Promise<ScenicAroundItem[]> {
  return get<ScenicAroundItem[]>(`/scenic/${id}/around`, { params: { limit } })
}

/** 邻近交通站点(地铁站/公交站) */
export function getScenicTransport(id: number): Promise<ScenicTransitStop[]> {
  return get<ScenicTransitStop[]>(`/scenic/${id}/transport`)
}
```

- [ ] **Step 2: 写失败测试**

创建 `wennv/wenlv/src/utils/scenicDetail.spec.ts`：

```ts
import { describe, it, expect } from 'vitest'
import { splitList, parseSections, estimatedSet, displayFact } from './scenicDetail'

describe('scenicDetail 展示工具', () => {
  it('splitList 拆分逗号串并去空', () => {
    expect(splitList('a.jpg, b.jpg ,,')).toEqual(['a.jpg', 'b.jpg'])
    expect(splitList(undefined)).toEqual([])
    expect(splitList('')).toEqual([])
  })

  it('parseSections 对非法 JSON 返回空数组', () => {
    expect(parseSections('not-json')).toEqual([])
    expect(parseSections('{"a":1}')).toEqual([])
    expect(parseSections(undefined)).toEqual([])
  })

  it('parseSections 过滤缺标题或正文的段落', () => {
    const raw = JSON.stringify([
      { title: '街区沿革', text: '甲', image: 'g1.jpg' },
      { title: '缺正文' },
      { title: '缺配图', text: '丙' },
      null,
    ])
    expect(parseSections(raw)).toEqual([
      { title: '街区沿革', text: '甲', image: 'g1.jpg' },
      { title: '缺配图', text: '丙', image: '' },
    ])
  })

  it('estimatedSet 解析参考值字段', () => {
    expect([...estimatedSet('ticket_price,recommend_hours')]).toEqual(['ticket_price', 'recommend_hours'])
    expect([...estimatedSet('')]).toEqual([])
  })

  it('displayFact 空值回退占位符', () => {
    expect(displayFact('   ', '暂无数据')).toBe('暂无数据')
    expect(displayFact(undefined, '暂无数据')).toBe('暂无数据')
    expect(displayFact('09:00-17:00', '暂无数据')).toBe('09:00-17:00')
  })
})
```

- [ ] **Step 3: 运行测试确认失败**

```powershell
cd "d:\桌面\文旅\wenlv\wennv\wenlv"
npx vitest run src/utils/scenicDetail.spec.ts
```

Expected: 失败，提示无法解析 `./scenicDetail`（模块不存在）。

- [ ] **Step 4: 实现 `utils/scenicDetail.ts`**

```ts
/** 景点详情页展示层纯函数:解析后端字段并对空值做统一降级。 */
import type { ScenicDetailSection } from '@/api/content'

/** splitList 把逗号分隔串拆成数组,自动去空。 */
export function splitList(value?: string): string[] {
  return (value || '')
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean)
}

/** parseSections 解析后端 detail_sections JSON;格式异常返回空数组,页面自动隐藏该区块。 */
export function parseSections(value?: string): ScenicDetailSection[] {
  if (!value) return []
  let raw: unknown
  try {
    raw = JSON.parse(value)
  } catch {
    return []
  }
  if (!Array.isArray(raw)) return []
  const out: ScenicDetailSection[] = []
  for (const item of raw as Array<Record<string, unknown>>) {
    if (!item || typeof item.title !== 'string' || typeof item.text !== 'string') continue
    out.push({
      title: item.title,
      text: item.text,
      image: typeof item.image === 'string' ? item.image : '',
    })
  }
  return out
}

/** estimatedSet 参考值字段集合,用于给 LLM 生成的数据加"参考值"标注。 */
export function estimatedSet(value?: string): Set<string> {
  return new Set(splitList(value))
}

/** displayFact 展示值:空值统一显示占位符,避免页面出现加载文案。 */
export function displayFact(value: string | undefined, placeholder: string): string {
  const v = (value || '').trim()
  return v || placeholder
}
```

- [ ] **Step 5: 运行测试确认通过**

```powershell
npx vitest run src/utils/scenicDetail.spec.ts
```

Expected: 5 个用例全部通过。

- [ ] **Step 6: 提交（先向用户确认）**

```powershell
git add wennv/wenlv/src/api/content.ts wennv/wenlv/src/utils/scenicDetail.ts wennv/wenlv/src/utils/scenicDetail.spec.ts
git commit -m "feat(scenic): 前端新增景点详情字段类型、周边/交通接口与展示工具(含单测)"
```

---

### Task 9: 详情页改用真实数据

**Files:**
- Modify: `wennv/wenlv/src/locales/zh.ts`、`en.ts`、`ja.ts`
- Modify: `wennv/wenlv/src/views/ScenicDetail/ScenicDetail.vue`

**Interfaces:**
- Consumes: Task 8 的 `getScenicAround`/`getScenicTransport`/`parseSections`/`estimatedSet`/`splitList`/`displayFact`、Task 1/5 落库的 `scenic_spots` 新字段
- Produces: 页面渲染真实数据；无数据区块自动隐藏或显示"暂无数据"；`estimated_fields` 中的字段显示"参考值"徽标

- [ ] **Step 1: 增加三语文案**

`wennv/wenlv/src/locales/zh.ts` 的 `scenicDetail` 段内追加：

```ts
    estimated: '参考值',
    estimatedTip: '该信息由 AI 依据公开资料整理，仅供参考，请以景区官方公告为准',
    noData: '暂无数据',
    dataSource: '数据来源：高德地图 / 维基百科',
    aroundEmpty: '暂无周边推荐',
    transportEmpty: '暂无交通信息',
```

`en.ts`：

```ts
    estimated: 'Reference',
    estimatedTip: 'AI-generated from public sources for reference only. Please check the official notice.',
    noData: 'No data',
    dataSource: 'Sources: AMap / Wikipedia',
    aroundEmpty: 'No nearby recommendations',
    transportEmpty: 'No transport info',
```

`ja.ts`：

```ts
    estimated: '参考値',
    estimatedTip: 'AIが公開情報をもとに整理した参考値です。最新情報は公式案内をご確認ください。',
    noData: 'データなし',
    dataSource: '出典：高徳地図 / ウィキペディア',
    aroundEmpty: '周辺のおすすめはありません',
    transportEmpty: '交通情報はありません',
```

- [ ] **Step 2: 页面脚本接入数据**

`ScenicDetail.vue` 的 `<script setup>` 中：

```ts
import { getScenic, getScenicAround, getScenicTransport } from '@/api/content'
import type { ScenicItem, ScenicAroundItem, ScenicTransitStop } from '@/api/content'
import { parseSections, estimatedSet, splitList, displayFact } from '@/utils/scenicDetail'
import AMapLoader from '@amap/amap-jsapi-loader'

const sections = computed(() => parseSections(scenic.value?.detail_sections))
const galleryImages = computed(() => splitList(scenic.value?.gallery_images))
const estimated = computed(() => estimatedSet(scenic.value?.estimated_fields))
const noData = computed(() => langStore.t('scenicDetail.noData'))
const isEstimated = (field: string) => estimated.value.has(field)

const around = ref<ScenicAroundItem[]>([])
const transit = ref<ScenicTransitStop[]>([])

const transportText = computed(() => {
  if (!transit.value.length) return ''
  return transit.value
    .slice(0, 2)
    .map((s) => `${s.name} 约${Math.round((s.distance || 0) / 10) / 100}km`)
    .join(' / ')
})

/* 景区位置:高德 JS 地图,Key 缺失或加载失败时回退占位 */
const mapKey = import.meta.env.VITE_AMAP_WEB_JS_KEY ?? ''
const mapReady = ref(false)
let amapInstance: { destroy: () => void } | null = null

async function initMap() {
  const spot = scenic.value
  if (!mapKey || !spot?.lng || !spot?.lat) return
  try {
    const AMap = await AMapLoader.load({ key: mapKey, version: '2.0', plugins: ['AMap.Marker'] })
    const map = new AMap.Map('scenic-map', {
      zoom: 15,
      center: [spot.lng, spot.lat],
      viewMode: '3D',
    })
    new AMap.Marker({ position: [spot.lng, spot.lat], title: displayName.value, map })
    amapInstance = map
    mapReady.value = true
  } catch {
    mapReady.value = false
  }
}

onMounted(async () => {
  if (!numericId.value) return
  try {
    scenic.value = await getScenic(numericId.value)
  } catch {
    /* 加载失败时保持占位展示 */
  }
  await initMap()
  try {
    const [a, t] = await Promise.all([
      getScenicAround(numericId.value, 6),
      getScenicTransport(numericId.value),
    ])
    around.value = a || []
    transit.value = t || []
  } catch {
    around.value = []
    transit.value = []
  }
})

onBeforeUnmount(() => {
  if (amapInstance) amapInstance.destroy()
})
```

并在文件顶部 import 中补 `onBeforeUnmount`。

- [ ] **Step 3: 模板改数据驱动**

逐块替换（保留原有 class 与样式，不改视觉）：

1) 评价横条的"年接待游客 / 建议游玩"：

```html
            <div class="detail-scorebar__fact">
              <span class="detail-scorebar__fact-key">{{ langStore.t('scenicDetail.visits') }}</span>
              <span class="detail-scorebar__fact-value">
                {{ displayFact(scenic?.yearly_visitors, noData) }}
                <em v-if="isEstimated('yearly_visitors')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
              </span>
            </div>
            <div class="detail-scorebar__fact">
              <span class="detail-scorebar__fact-key">{{ langStore.t('scenicDetail.recommendTime') }}</span>
              <span class="detail-scorebar__fact-value">
                {{ displayFact(scenic?.recommend_hours, noData) }}
                <em v-if="isEstimated('recommend_hours')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
              </span>
            </div>
```

2) 实用信息 4 张卡（开放时间 / 门票价格 / 交通方式 / 详细地址）：

```html
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="clock" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.openHours') }}</h3>
            <p class="detail-info__value">
              {{ displayFact(scenic?.open_hours, noData) }}
              <em v-if="isEstimated('open_hours')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="ticket" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.ticket') }}</h3>
            <p class="detail-info__value">
              {{ displayFact(scenic?.ticket_price, noData) }}
              <em v-if="isEstimated('ticket_price')" class="detail-est">{{ langStore.t('scenicDetail.estimated') }}</em>
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="bus" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.transport') }}</h3>
            <p class="detail-info__value">
              {{ transportText || langStore.t('scenicDetail.transportEmpty') }}
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="pin" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.address') }}</h3>
            <p class="detail-info__value">{{ displayFact(scenic?.address, noData) }}</p>
          </div>
```

3) 图文详情：整段替换两个写死的 `detail-content__row`：

```html
        <div v-if="sections.length" class="detail-content">
          <div
            v-for="(sec, i) in sections"
            :key="sec.title"
            class="detail-content__row"
            :class="{ 'detail-content__row--reverse': i % 2 === 1 }"
          >
            <div class="detail-content__figure">
              <img
                v-if="sec.image"
                class="detail-content__img"
                :src="sec.image"
                :alt="sec.title"
                referrerpolicy="no-referrer"
              />
              <div v-else class="detail-content__img detail-content__img--empty">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <path d="M21 15l-5-5L5 21"/>
                </svg>
              </div>
            </div>
            <div class="detail-content__text">
              <h3 class="detail-content__caption">{{ sec.title }}</h3>
              <p class="detail-content__para">{{ sec.text }}</p>
            </div>
          </div>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('scenicDetail.noData') }}</p>
```

4) 精彩瞬间相册：

```html
        <div v-if="galleryImages.length" class="detail-gallery">
          <div v-for="(img, i) in galleryImages" :key="img" class="detail-gallery__item">
            <img class="detail-gallery__photo" :src="img" :alt="`${displayName} ${i + 1}`" referrerpolicy="no-referrer" />
          </div>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('scenicDetail.noData') }}</p>
```

5) 景区位置地图：

```html
        <div class="detail-map">
          <div v-show="mapReady" id="scenic-map" class="detail-map__canvas" />
          <div v-if="!mapReady" class="detail-map__placeholder">
            <span class="detail-map__pin"><AppIcon name="pin" :size="40" /></span>
            <span class="detail-map__hint">{{ langStore.t('scenicDetail.noData') }}</span>
          </div>
        </div>
```

6) 周边推荐：

```html
        <div v-if="around.length" class="detail-around">
          <div v-for="item in around" :key="item.id" class="detail-around__card">
            <div class="detail-around__body">
              <h3 class="detail-around__name">{{ item.name }}</h3>
              <p class="detail-around__desc">
                {{ item.address || item.type || '' }}
                <span v-if="item.distance">· 约{{ (item.distance / 1000).toFixed(1) }}km</span>
              </p>
            </div>
          </div>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('scenicDetail.aroundEmpty') }}</p>
```

7) 页脚数据来源与参考值免责说明（放在 `</main>` 前最后一个 section 内）：

```html
        <p class="detail-source">
          {{ langStore.t('scenicDetail.dataSource') }}
          <template v-if="estimated.size">· {{ langStore.t('scenicDetail.estimatedTip') }}</template>
        </p>
```

- [ ] **Step 4: 补样式（追加到 `<style scoped>`）**

```css
.detail-est {
  display: inline-block;
  margin-left: var(--space-1);
  padding: 1px 6px;
  font-size: var(--text-xs);
  font-style: normal;
  color: var(--color-gold);
  border: 1px solid color-mix(in srgb, var(--color-gold) 45%, transparent);
  border-radius: var(--radius-full);
  vertical-align: middle;
}

.detail-empty {
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.detail-gallery__photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  aspect-ratio: 4 / 3;
}

.detail-map__canvas {
  width: 100%;
  aspect-ratio: 16 / 6;
}

.detail-source {
  margin-top: var(--space-8);
  text-align: center;
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}
```

- [ ] **Step 5: 类型检查与页面验证**

```powershell
cd "d:\桌面\文旅\wenlv\wennv\wenlv"
npm run type-check
```

Expected: 无类型错误。

浏览器打开 `http://localhost:5173/scenic/scenic-3`（`scenic-3` 换成 Task 6 已采过的景点 ID）核对：实用信息有地址/开放时间、门票带"参考值"徽标、图文详情有段落与配图、精彩瞬间有多张图、地图渲染标记、周边推荐有 3km 内 POI。未采集的景点应显示"暂无数据"而不是 loading 文案。

- [ ] **Step 6: 提交（先向用户确认）**

```powershell
git add wennv/wenlv/src/locales/zh.ts wennv/wenlv/src/locales/en.ts wennv/wenlv/src/locales/ja.ts wennv/wenlv/src/views/ScenicDetail/ScenicDetail.vue
git commit -m "feat(scenic): 详情页接入真实数据(实用信息/图文详情/精彩瞬间/高德地图/周边推荐)并标注参考值"
```

---

### Task 10: 全量采集与端到端验收

**Files:** 无代码改动（产出人工核对清单）

- [ ] **Step 1: 后端全量测试与构建**

```powershell
cd "d:\桌面\文旅\wenlv\houduan"
go build ./...
go test ./... 2>&1 | Select-Object -Last 20
```

Expected: 构建通过；`service` 包 4 个测试 PASS（其余包无测试文件属正常）。

- [ ] **Step 2: 全量补数据**

```powershell
go run ./cmd/scenic-enrich 2>&1 | Tee-Object -FilePath enrich.log
```

Expected: 日志末尾打印 `完成: 成功 N / 失败 0`；`unmatched` 备注的景点抄录成清单交给用户决定是否人工补。

- [ ] **Step 3: 抽查数据质量**

```powershell
mysql -uroot wenlv -e "SELECT COUNT(*) AS total, SUM(address<>'') AS has_addr, SUM(gallery_images<>'') AS has_gallery, SUM(detail_sections<>'') AS has_sections, SUM(ticket_price<>'') AS has_ticket FROM scenic_spots;"
mysql -uroot wenlv -e "SELECT name_zh, ticket_price, recommend_hours, yearly_visitors, estimated_fields FROM scenic_spots WHERE estimated_fields<>'' LIMIT 8;"
```

Expected: `has_addr` 覆盖率明显高于 0；`has_gallery`/`has_sections` 覆盖大部分景点；`estimated_fields` 为空的行对应字段也为空（不允许出现"有值但未标注参考值"）。

- [ ] **Step 4: 端到端验收（前后端同时启动）**

```powershell
# 终端 A
cd "d:\桌面\文旅\wenlv\houduan"; $env:SERVER_PORT="8081"; go run .
# 终端 B
cd "d:\桌面\文旅\wenlv\wennv\wenlv"; npm run dev
```

浏览器逐项核对（建议 3 个不同区县的景点）：

1. Explore 卡片 → 详情页：评价横条不再有 `—`；
2. 实用信息 4 卡有值或"暂无数据"，无 `common.loading`；
3. 图文详情有段落 + 配图，段落数与图片数匹配；
4. 精彩瞬间有真实图片（OSS 域名）；
5. 景区位置地图有标记，Key 缺失时回退占位不报错；
6. 周边推荐有 3km 内 POI；
7. 中/EN/日 切换文案正常，参考值徽标三语正确；
8. 断网或清空 `VITE_AMAP_WEB_JS_KEY` 时页面仍可正常浏览（地图回退占位）；
9. 再进一次同一景点详情页（5 分钟内），DevTools Network 里 `/around`、`/transport` 显示 `(disk cache)` 或 `(memory cache)`，即浏览器缓存生效；
10. 改 `.env` 的 `SCENIC_CACHE_TTL_HOURS=1` 重启后端，`redis-cli ttl "scenic:around:3:6"` 对新写入的键返回 ≤3600（老键需 `redis-cli del` 或等过期）。

- [ ] **Step 5: 前端单测回归**

```powershell
cd "d:\桌面\文旅\wenlv\wennv\wenlv"
npx vitest run
npm run type-check
```

Expected: 全部通过。

- [ ] **Step 6: 收尾提交（先向用户确认；`enrich.log` 是草稿日志，不要提交）**

```powershell
Remove-Item -Force enrich.log
git status --short
git add houduan wennv docs
git commit -m "chore(scenic): 完成景点详情数据补全验收(全量采集+端到端核对)"
```

---

## 风险与合规

- **高德数据**：批量调用注意 QPS 与日配额（批处理已加 300ms 间隔），生产环境需确认商用授权与配额；图片版权归原作者，页面已标注来源。
- **LLM 参考值**：一律标注"参考值"并附免责说明；上线前人工抽查 `ticket_price`/`yearly_visitors` 各 10 条，出现明显错误时把该字段从 prompt 中移除。
- **维基百科**：国内需代理，故 `desc` 仍走已有爬虫离线更新，本次不新增维基抓取。
- **小红书/抖音**：本次不引入文字内容（版权与风控风险高），仅保留图片管线作为后续可选项。
- **三语**：详情正文暂为中文，与现有 `desc` 行为一致；后续可按 `desc_en/desc_ja` 扩展。

## 缓存策略与运维

三级缓存，各管一段，互不替代：

| 层级 | 缓存对象 | 有效期 | 失效/更新方式 |
|---|---|---|---|
| MySQL | 地址/开放时间/相册/图文详情/参考值 | 永久 | 重跑 `go run ./cmd/scenic-enrich -only <景点名>` |
| Redis | 周边推荐（`scenic:around:{id}:{limit}`）、交通站点（`scenic:transit:{id}`） | `SCENIC_CACHE_TTL_HOURS`，默认 24h | 到点自动过期；需立即刷新用 `redis-cli del <key>` |
| 浏览器 | 上述两个接口响应；OSS 图片 | 接口 300s；图片 86400s | 强刷（Ctrl+F5）或等过期 |

运维要点：

- 改 `SCENIC_CACHE_TTL_HOURS` 只影响**之后新写入**的键，不会改动 Redis 里已有键的 TTL；要立即生效先 `redis-cli del "scenic:around:*"`。
- 空结果不写 Redis、也不加浏览器缓存头：避免高德临时失败被固化一整个 TTL。
- 浏览器缓存生效的前提是请求 URL 稳定。当前 `src/api/request.ts` 未加时间戳等 cache-busting 参数，且开发环境走 Vite 同源 `/api` 代理，所以缓存有效；如果后续给请求统一加时间戳参数，浏览器这层缓存会直接失效（Redis 那层不受影响）。
- 需要"每次进页面都拿最新"时，把 `scenicClientCacheSeconds`（`handler/scenic_extra_handler.go`）改成 0 或直接删掉 `setScenicClientCache` 调用。
