# AI 行程落地页 · 巴蜀定位文案改版

**日期：** 2026-09-18  
**范围：** `/trip` 落地页（`LandingView.vue` + trip `NavBar`）可见品牌与主文案；中/英/日 i18n  
**状态：** 口头确认（方向 A · 方案 1 蜀韵智游）；待用户审阅本 spec 后进入实现计划  
**不做：** 改表单字段逻辑、后端、结果页大改版、主站首页 `trip.*` 区块（可另开任务对齐）

---

## 1. 背景与目标

落地页仍沿用模板品牌 **TripStar / TRIPSTAR / 旅途星辰** 与「探索世界的每一种可能」，与主站「蜀韵·成都 / 巴蜀文旅」气质脱节。目标改为 **紧贴成都 / 巴蜀** 的 AI 文旅行程入口文案。

**成功标准：**
1. Hero 与导航不再出现 TripStar / TRIPSTAR / 旅途星辰（落地页范围内）。
2. 中文主文案传达天府慢生活 + AI 排程；英日同一定位。
3. Hero 大标题走 i18n，不再在模板写死英文。
4. 表单步骤标题更口语，字段与校验逻辑不变。

---

## 2. 方案选型（已确认）

| 项 | 选择 |
|----|------|
| 定位 | **A：紧贴成都 / 巴蜀** |
| 品牌方向 | **方案 1：蜀韵智游**（替换 TripStar 露出） |
| 语气 | 文雅、不硬广；产品能力说清楚但不堆砌术语 |

---

## 3. 文案定稿（落地页）

### 3.1 品牌与 Hero

| Key / 位置 | zh-CN | en-US | ja-JP |
|------------|-------|-------|-------|
| `app.title` / `app.brand` / `app.footerBrand` | 蜀韵智游 | Shu Journey | 蜀韻スマート旅 |
| `app.subBrand` | 蜀韵·成都 · AI 文旅行程 | Shu·Chengdu · AI Cultural Trip | 蜀韻·成都 · AI 文化旅 |
| `app.badge` | 巴蜀文旅智能体 | Bashu Travel Agent | 巴蜀文旅エージェント |
| Hero `h1`（现写死 TRIPSTAR → 改为 `t('app.brand')` 或 `t('home.heroTitle')`） | 蜀韵智游 | Shu Journey | 蜀韻スマート旅 |
| `home.titleLine` | 把巴适日子，排成一条天府路 | Plan a Tianfu day that feels right | 巴适な一日を、天府の道に |
| `home.heroBadge` | AI 成都行程规划 | AI Chengdu Itinerary | AI 成都旅程プラン |
| `home.heroDesc` | 告诉我天数、同行与喜好，结合真实游客口碑与地图数据，生成可走、可算、可分享的成都行程 | Tell us your days, companions, and tastes — we weave real traveler tips with map data into a walkable, budget-aware Chengdu plan you can share | 日数・同行・好みを伝えてください。リアルな旅の口コミと地図データから、歩ける・計算できる・共有できる成都の旅程をつくります |

### 3.2 导航与操作

| Key | zh-CN | en-US | ja-JP |
|-----|-------|-------|-------|
| `home.nav.cta` | 开始规划 | Start planning | プランを始める |
| `home.nav.backHome` | 返回首页 | Back to home | ホームに戻る |
| `home.submit` | 生成我的成都行程 | Create my Chengdu trip | 私の成都旅をつくる |
| `home.submitting` | 正在为你排天府日程… | Arranging your Tianfu days… | 天府の日程を組んでいます… |

NavBar 模板中硬编码的 `TripStar` 改为 `{{ t('app.brand') }}`。

### 3.3 表单步骤（口语化，逻辑不变）

| Key | zh-CN | en-US | ja-JP |
|-----|-------|-------|-------|
| `home.step1` | 去哪儿玩 | Where to go | どこへ行く |
| `home.step2` | 怎么走与住 | Getting around & stay | 移動と宿 |
| `home.step3` | 还有什么想说的 | Anything else | そのほかの希望 |
| `home.cityPlaceholder` | 输入城市，例如：中国-成都 | e.g. China-Chengdu | 例：中国-成都 |
| `home.specialNeedsPlaceholder` | 比如：带老人出行、想逛宽窄、对花椒过敏…… | e.g. traveling with elders, Kuanzhai stroll, chili sensitivity… | 例：ご高齢の同行、寛窄巷子、山椒アレルギーなど |

### 3.4 加载与页脚（落地页可见）

| Key | zh-CN | en-US | ja-JP |
|-----|-------|-------|-------|
| `home.loading.workingTogether` | AI 正在为你串起一天的巴适节奏… | AI is weaving a day at Chengdu pace… | AIが巴适な一日のリズムを編んでいます… |
| `home.loading.donePrepare` | 好了，天府等你 | Ready — Tianfu awaits | 準備完了、天府が待っています |
| `app.footerCopy` | 保持年份占位：`{year} 蜀韵·成都` | `{year} Shu·Chengdu` | `{year} 蜀韻·成都` |
| 结果图页脚若含品牌（`result.*.footer` 等） | 将「旅途星辰 TripStar」替换为「蜀韵智游」同一定位 | same | same |

兴趣标签、交通住宿选项 **可不改**（已够清晰）；若实现时顺手把「景点推荐来源」说明改得更短，属可选增强，非必须。

---

## 4. 技术改动面

| 文件 | 改动 |
|------|------|
| `src/trip/i18n/locales/zh.json` | 按上表改 key |
| `src/trip/i18n/locales/en.json` | 同上 |
| `src/trip/i18n/locales/ja.json` | 同上 |
| `src/trip/views/LandingView.vue` | Hero `h1` 改为 i18n；去掉写死 `TRIPSTAR` |
| `src/trip/components/NavBar.vue` | 品牌文案改 `t('app.brand')` |

可选：全仓 trip 目录 grep `TripStar|TRIPSTAR|旅途星辰`，落地页与结果页脚一并替换，避免半新半旧。

---

## 5. 明确不做

- 不改表单数据结构、校验规则、API
- 不改主站 `HomeView` 的 `locales/*/trip` 区块（除非用户追加要求）
- 不做落地页视觉大改（配色/布局保持）

---

## 6. 验收清单

- [ ] `/trip` Hero 显示「蜀韵智游」类品牌，无 TRIPSTAR
- [ ] 导航品牌与 CTA 为新文案；中英日切换正确
- [ ] 副标题为「把巴适日子，排成一条天府路」及对应译
- [ ] 提交按钮为「生成我的成都行程」及对应译
- [ ] 步骤标题为口语三步；表单仍可用
- [ ] trip 落地相关字符串无残留「旅途星辰 / TripStar」（或仅残留在未改的无关配置文案中并在报告注明）

---

## 7. 实现顺序（供后续 plan）

1. 更新 zh / en / ja locale 定稿文案  
2. LandingView + NavBar 去掉硬编码品牌  
3. grep 清理残留 + 目视验收三语  
