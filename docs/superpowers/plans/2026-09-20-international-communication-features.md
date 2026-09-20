# 国际传播三大创新功能（巴蜀故事引擎 + 外国人点菜神器 + 旅行故事卡片出海）Implementation Plan

**Goal:** 面向"数智文旅 + 国际传播"赛道，补齐项目"只会吃喝玩乐、不会讲中国故事"的短板：

1. **巴蜀故事引擎**——景点/美食/路线的介绍由中文扩展为多语种（EN/JA）"故事版"内容 + **文化注解（Culture Note）**，向外国游客解释"为什么"，前端按语言切换展示。
2. **外国人点菜神器**——每道美食生成 AI 意译英文名（破解"夫妻肺片=Husband and Wife Lung Slices"式直译陷阱）+ 食材/过敏原标注 + 辣度可视化 + 可直接指给服务员看的"点菜大卡"模式。
3. **旅行故事卡片出海**——行程结束后一键生成多语种旅行故事海报（路线图 + AI 生成文案 + 平台二维码），用户可保存分享到海外社交平台，形成"游客即传播者"的闭环。

**Architecture:** 完全沿用项目已验证的三层模式：
- **数据层**：`scenic_spots` / `foods` / `routes` 新增多语种字段（全部带中文 comment），GORM 自动迁移；
- **生成层**：新增批处理命令 `cmd/story-enrich`，复用 `service.TripLLM.ChatWithEffort` + `extractJSONFromResponse`，素材只取库内已有中文 `desc`（LLM 只做故事化翻译改写，禁止编造事实），落库并登记 `estimated_fields`；
- **API 层**：model 新字段经既有 `/api/scenic`、`/api/food`、`/api/routes` 自动带出（只增不改不删）；故事卡片文案新增 `POST /api/trip/story-card`；
- **展示层**：前端按 `language store` 当前 locale 选取 `desc_zh/desc_en/desc_ja`，非中文时显示 Culture Note 高亮卡；FoodDetail 加点菜卡；ResultView 加故事海报（Canvas 绘制 + 二维码）。

**Tech Stack:** Go + Gin + GORM + MySQL；OpenAI 兼容 LLM（`TripLLM`）；Vue 3 + Vite + TS + Pinia（language store 已有 zh/en/ja）；Canvas 2D 海报绘制；`qrcode` npm 包生成二维码（本功能唯一新增依赖，需用户确认）。

**Global Constraints（沿既有约定）:**
- 所有新增数据库字段必须有中文 `comment`（Navicat 可读）。
- 既有 API 只能**新增**字段，不得改名或删除已有字段。
- LLM 生成的字段必须写入 `estimated_fields`；LLM 只允许基于库内已有中文简介改写，"不确定就写空"，禁止编造数字、年份、排名、店名。
- 运行时配置（LLM Key）走 `.env` → `data/runtime_settings.json` 覆盖链，与 `cmd/food-enrich` 完全一致。
- 三语文案（zh/en/ja）同步更新到 `src/locales/*.ts` 与 `src/trip/i18n/locales/*.json`。
- 不改 AI 行程核心链路（plan/ws/xhs/amap），新功能全部独立文件/独立路由，不影响现有功能。
- 改动不提交，留在工作区。

---

## Phase 1 后端：多语种故事数据层 + 批量生成器

### Task 1.1 model 新增多语种字段
- [ ] `model/scenic.go`：`DescEN`、`DescJA`（text 故事版介绍）、`CultureNoteEN`、`CultureNoteJA`（text 文化注解）
- [ ] `model/food.go`：同上 4 个字段 + `NameLiteralEN`（varchar 128，直译陷阱名）+ `IngredientsZH`、`IngredientsEN`（varchar 255，食材/过敏原）
- [ ] `model/route.go`：`DescriptionEN`、`DescriptionJA`（text）；`Stops` JSON 内站点 desc 补 `desc_en/desc_ja` 由生成器同步写入
- [ ] `database.MustAutoMigrate` 确认覆盖（应已含全部表，验证即可）

### Task 1.2 `cmd/story-enrich` 批量生成器
- [ ] 新建 `service/story_enrich.go` + `cmd/story-enrich/main.go`（结构照抄 `cmd/food-enrich`：flag 支持 `-only/-limit/-force/-dry-run`，TripSettings→TripLLM 初始化链一致）
- [ ] 每个素材一次 LLM 调用，输出严格 JSON：
  ```json
  {"story_en":"...","story_ja":"...","note_en":"...","note_ja":"...","literal_en":"","ingredients_zh":"","ingredients_en":""}
  ```
  - `story_*`：面向外国游客的故事化介绍（120-200 词/字），基于库内 `desc` 改写，不得新增事实
  - `note_*`：1-3 条文化注解（如火锅=围炉社交哲学、武侯祠=三国忠义），逐条独立成句
  - `literal_en`：菜名字面直译（如"夫妻肺片"→"Husband and Wife Lung Slices"）+ 生成 `ingredients_*`（仅美食）
- [ ] 写库策略：默认只补空字段，`-force` 覆盖；`estimated_fields` 追加生成字段名；`data_updated_at` 刷新
- [ ] 路线 `stops` JSON：解析→按站点名逐个生成→写回 JSON（站点无 `desc` 时跳过）

### Task 1.3 故事卡片文案 API
- [ ] `handler/trip_handler.go` 新增 `POST /api/trip/story-card`：入参 `plan_id` + `language`；读 `trip_plans.PlanJSON` → LLM 生成该语言的 80-120 字旅行故事标题+正文（含城市/天数/亮点站点，禁止编造未到访地点）→ 返回 `{title, body}`
- [ ] `router.go` 注册路由；失败降级为 HTTP 200 + `{fallback:true}`（前端用本地模板文案兜底）

## Phase 2 前端：多语种内容展示 + 点菜卡 + 故事海报

### Task 2.1 多语种内容选择器与故事展示
- [ ] `src/utils/storyI18n.ts`：`pickDesc(item, locale)` 按语言返回 `desc/desc_en/desc_ja`，中文回落原 `desc`
- [ ] `ScenicDetail.vue` / `FoodDetail.vue` / `RoutesPage.vue` 详情：语言非 zh 时显示 `story_*`，并在正文下方渲染 **Culture Note 高亮卡**（🌏 图标 + 三语标题"文化注解"）
- [ ] 三语文案键：`cultureNote.title` 等 → `src/locales/{zh,en,ja}.ts`

### Task 2.2 外国人点菜神器
- [ ] `FoodDetail.vue` 新增"国际点菜卡"区块：中文菜名大字 + `name_en` + `name_literal_en`（⚠ 直译陷阱标签，仅当直译≠意译时显示"这个名字直译会吓到服务员"趣味提示）+ 食材/过敏原（zh/en 双行）+ 辣度辣椒图标（`spice_level` 文本→0-5 映射）
- [ ] `FoodPage.vue` 列表卡片语言非 zh 时副标题显示意译英文名
- [ ] 三语文案键同步

### Task 2.3 旅行故事卡片出海
- [ ] `ResultView.vue` 顶栏新增「生成故事卡片」按钮（三语：生成你的旅行故事卡 / Share Your Story / 旅のストーリーカード）
- [ ] 新建 `src/trip/components/StoryCardModal.vue`：
  - 调 `POST /api/trip/story-card` 取文案（fallback 本地模板）
  - Canvas 绘制 1080×1440 海报：城市名大字 + 天数/日期 + 站点路线 + AI 故事文案 + 平台二维码 + 熊猫配色主题
  - 二维码引入 `qrcode` npm 包（内容=站点首页 URL），无网络时降级为纯文字水印
  - 「保存图片」= `canvas.toDataURL('image/png')` 触发下载
- [ ] 三语文案键 → `src/trip/i18n/locales/*.json`

## Phase 3 运行生成 + 构建验证

- [ ] `go build ./...` 通过；后端重启（SERVER_PORT=8081）
- [ ] `go run ./cmd/story-enrich -limit 2 -dry-run` 核对素材；正式跑全量（scenic → food → routes）
- [ ] 抽查库内 `desc_en/desc_ja/culture_note_*` 质量与 `estimated_fields` 登记情况
- [ ] 前端 `npm run type-check` + `npm run build-only` 通过
- [ ] 手工验收：切换 EN/JA 看景点/美食/路线故事与 Culture Note；点菜卡展示；行程结果页生成故事海报

## 风险与降级

| 风险 | 对策 |
|---|---|
| LLM 生成质量/编造 | prompt 严格限定"仅基于给定素材"，dry-run 先抽查，非 force 只补空 |
| 路线 stops JSON 结构复杂 | 逐站点独立生成、失败跳过不整体失败 |
| Canvas 中文字体跨端差异 | 海报文字用系统字体栈，测试以 Chrome 为准 |
| 无新增依赖约束 | `qrcode` 为唯一新增依赖，开工前与用户确认 |
