# 国际传播第二期：四项新创新功能方案

> 对应赛道 1.7 数智文旅 + 国际传播。与第一期三大功能（故事引擎/点菜神器/故事卡）构成完整叙事：
> **听得懂 → 玩得转 → 讲得出**。本文档为方案，未写任何代码。

## 功能总览

| # | 功能 | 一句话 | 后端改动 | 前端改动 | 数据生成 |
|---|------|--------|---------|---------|---------|
| 1 | 多语 AI 语音导览 | 一键朗读三语故事正文 | 无 | speech 工具 + 2 个详情页按钮 | 无 |
| 2 | 数字足迹护照 | 浏览景点集章,集满生成护照证书海报 | 无(纯前端 localStorage) | 盖章逻辑 + 护照页 + 证书 Canvas | 无 |
| 3 | 入境游实用工具箱 | 老外第一次来中国的三语生存指南 | 无 | 新页面 + 首页入口 | 无(纯 i18n 内容) |
| 4 | 四川话一分钟 | 方言彩蛋:巴适/安逸/雄起 | 无 | 详情页方言区块 | 无(纯 i18n 内容) |

四项均**不新增后端接口、不新增依赖**（第 2 项复用现有 qrcode 依赖），全部前端实现。

---

## 1. 多语 AI 语音导览（TTS）

**原理**：浏览器原生 `SpeechSynthesis` API，零 API 费用、零后端。Edge/Chrome 桌面与移动端均内置中/英/日语音。

### 实现
- 新建 `src/utils/speech.ts`：
  - `speak(text, lang)` / `stopSpeaking()` / `isSpeaking`（响应式）
  - 语言映射：zh→`zh-CN`、en→`en-US`、ja→`ja-JP`；`speechSynthesis.getVoices()` 按 `lang.startsWith` 匹配最优 voice，找不到该语言语音时按钮置灰并提示"当前浏览器无该语言语音"
  - 长文本按句子拆分（。！？.!?）依次入队朗读，避免部分浏览器截断长 Utterance
  - 语速 0.95，更接近导览语感
- `ScenicDetail.vue` / `FoodDetail.vue` 故事正文标题旁加 🔊 朗读/⏹ 停止按钮：
  - 朗读内容 = `pickDesc(item, lang)` 当前语言正文
  - 切换语言时先 `stopSpeaking()`；离开页面（onUnmounted）停止
- i18n 三语键：`voice.play / voice.stop / voice.unsupported`（主站 `src/locales`）

### 验收
日/英文界面点朗读，能听到对应语言语音；中文界面正常；切换语言朗读停止。

---

## 2. 数字足迹护照（游戏化集章）

**原理**：浏览景点详情即"盖章"，集章数据存 localStorage（免登录可用、免后端）；集满进度生成一张"成都旅行护照"证书海报，**复用故事卡的 Canvas 绘制模式与 qrcode 依赖**。

### 实现
- 盖章逻辑：新建 `src/utils/passport.ts`
  - localStorage 键 `wenlv.passport.stamps`：`[{ id, name_zh, name_en, name_ja, ts }]`（去重）
  - `addStamp(item)` / `getStamps()` / `stampCount()`
  - `ScenicDetail.vue` 数据加载成功后自动 `addStamp`（无感打卡，不弹窗打扰；首页护照入口显示小红点提醒新章）
- 护照页 `src/views/PassportPage.vue`（路由 `/passport`，主站导航加入口 + HomeDashboard 卡片入口）：
  - 已集章网格（23 格，未集灰色剪影）、进度条 `x/23`
  - 「生成护照证书」按钮（任意进度均可生成，进度不同证书不同——激励继续集）
- 证书 `src/components/PassportModal.vue`：纯 Canvas 1080×1440（参考 StoryCardModal 的绘制工具函数 roundRectPath/wrapText/fitText/drawLetterSpaced，可抽到 `src/utils/posterCanvas.ts` 共用）
  - 设计：熊猫配色同故事卡（暖米底/墨黑/竹绿）、标题「成都旅行护照 CHENGDU TRAVEL PASSPORT」、印章网格（已集章画红色印章样式 + 景点名，未集章画虚线空格）、底部生成日期 + 二维码
  - **二维码指向优化**：编码 `${origin}/passport`（带路径），比首页直达；部署公网后扫码即打开护照页
  - 下载 PNG 按钮（toDataURL）
- i18n 三语键：`passport.*`（title/progress/generate/download/stampTip 等）

### 验收
浏览 3 个景点 → 护照页显示 3 枚章 → 生成证书含 3 枚红章 + 可下载；刷新页面章不丢。

---

## 3. 入境游实用工具箱（First-time in Chengdu）

**定位**：给国际游客的"落地生存指南"，静态内容页，体现产品认真服务外国人的态度。与点菜神器同一条叙事线。

### 实现
- 新建 `src/views/TravelGuidePage.vue`（路由 `/guide`，导航"入境指南"入口）
- 内容全部走 i18n 三语（`guide.*` 键组），六大板块卡片：
  1. **支付** 💳：支付宝/微信绑定境外银行卡步骤、外卡限额提示、现金场景
  2. **交通** 🚄：12306 英文版/App 购票、地铁扫码、网约车（如 DiDi 英文界面）
  3. **住宿** 🏨：选择"涉外接待"资质酒店、登记证件说明
  4. **通讯** 📱：eSIM/漫游、常用 App（地图/翻译/支付）清单
  5. **应急** 🚨：110/120/119、12308 外交部求助热线、丢失护照处理
  6. **礼仪小贴士** 🙏：筷子礼仪、排队、餐厅砍价文化等 3-4 条
- 每条为"图标 + 标题 + 一两句说明"，信息密度低、可扫读；顶部一句话定位文案："Designed for first-time visitors to China"

### 验收
三语切换内容完整；移动端排版正常。

---

## 4. 四川话一分钟（方言彩蛋）

**定位**：跨文化传播里方言最有感染力，作为景点/美食详情页的文化彩蛋。

### 实现
- 内容放 i18n（`dialect.*`），静态数据 6 个词：
  巴适（bā shì）、安逸（ān yì）、雄起（xióng qǐ）、摆龙门阵、要得、扎起
  每词含：`word`（中文）、`pinyin`、`meaning`（英/日解释）、`scene`（使用场景一句话）
- 展示位置：`ScenicDetail.vue` 与 `FoodDetail.vue` 的 Culture Note 卡下方，新区块「🌶 四川话一分钟」
  - **所有语言显示**（中文用户看拼音注释也有趣；外语用户获得最地道的体验）
  - 卡片样式与 Culture Note 琥珀色系区分（用辣椒红渐变）
- 不做方言音频（zh-CN TTS 读四川话不标准，避免负面体验）

### 验收
三语界面均显示方言卡；与 Culture Note 卡视觉区分、不显拥挤。

---

## 实施顺序与工作量

1. **Phase A（半天内）**：功能 1 语音导览 → 功能 4 方言卡（都是详情页小改动，一起做）
2. **Phase B**：功能 3 工具箱（纯内容页）
3. **Phase C**：功能 2 护照（工作量最大：Canvas 抽公共工具 + 新页面 + 盖章逻辑）
4. **Phase D**：全量验收（type-check + build + 浏览器三语实测）+ 演示流程串联

## 演示叙事（给评委）

> 「外国游客打开网站听到英语讲武侯祠的故事（语音导览）→ 点菜不用怕（点菜神器）→ 落地生存有指南（工具箱）→ 学一句'巴适得很'（方言彩蛋）→ 越逛越想集齐护照（游戏化）→ 生成熊猫护照证书分享给朋友，扫码回访（传播闭环）」

## 已知限制说明

- TTS 依赖浏览器内置语音：主流 Chrome/Edge/Safari 均可用，个别系统缺日语语音时按钮置灰并提示
- 护照集章存 localStorage：换设备/清缓存会丢（MVP 取舍；后续可挂到用户账号，表结构与接口预留到三期）
- 二维码在 localhost 开发环境扫码无效（手机访问不了 localhost），演示需部署公网或用局域网 IP
