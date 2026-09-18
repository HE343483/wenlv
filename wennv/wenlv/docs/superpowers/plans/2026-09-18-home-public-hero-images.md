# 主站公开页 Hero/轮播配图 + 后端代理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为公开首页 Hero 与景点轮播接入本地高清成都风景图，并将开发环境 `/api` 代理指向 `http://ra9a3c94.natappfree.cc`。

**Architecture:** 图片以静态资源落在 `public/images/home/`，由 Vite 原样托管；`HomeView` 只负责赋值 `imageUrl` 与 Hero 背景；`Carousel` 在有图时渲染 `<img>`，无图时保留渐变占位。后端地址只改 Vite 代理与 env，主站 `request.ts` 仍走 `/api`，避免浏览器跨域。

**Tech Stack:** Vue 3 + Vite 8 + TypeScript；Vitest + `@vue/test-utils`（Carousel 单测）；纯 CSS（scoped）。

**Spec:** `docs/superpowers/specs/2026-09-18-home-public-hero-images-design.md`

## Global Constraints

- 范围仅公开首页 `/` 的 Hero + 轮播 + 后端代理；不改登录后页、不改 trip 视觉。
- 后端穿透地址 verbatim：`http://ra9a3c94.natappfree.cc`
- 本地图路径前缀 verbatim：`/images/home/`
- 单张图目标约 200–500KB；清晰优先于极致压缩。
- Hero 为全幅背景 + 遮罩，保留现有标题/CTA 结构。
- 未经用户明确要求不要 `git commit`（仓库根在 `e:\wenlv`；本计划勾选的 Commit 步骤默认跳过，改为向用户确认后再提交）。

---

## File Map

| 文件 | 职责 |
|------|------|
| `public/images/home/*.jpg` | Hero + 5 张轮播静态图 |
| `src/components/Carousel.vue` | 有 `imageUrl` 时渲染图片与深色遮罩 |
| `src/components/Carousel.spec.ts` | 断言有/无图两种 DOM |
| `src/views/HomeView/HomeView.vue` | Hero 背景图 + `carouselItems.imageUrl` |
| `vite.config.ts` | `/api` proxy target |
| `.env.example` | 记录 `VITE_API_BASE_URL` |
| `.env`（若不存在则创建） | 本地联调 API 基址 |

---

### Task 1: 下载并放置本地风景图

**Files:**
- Create: `public/images/home/hero-chengdu.jpg`
- Create: `public/images/home/carousel-panda.jpg`
- Create: `public/images/home/carousel-kuanzhai.jpg`
- Create: `public/images/home/carousel-dujiangyan.jpg`
- Create: `public/images/home/carousel-jinli.jpg`
- Create: `public/images/home/carousel-xiling.jpg`
- Create: `public/images/home/SOURCES.md`（记录图源 URL 与许可说明）

**Interfaces:**
- Consumes: 无
- Produces: 浏览器可访问路径 `/images/home/<name>.jpg`（6 个文件必须存在且可读）

- [ ] **Step 1: 创建目录**

```powershell
New-Item -ItemType Directory -Force -Path "e:\wenlv\wennv\wenlv\public\images\home"
```

- [ ] **Step 2: 用 curl 下载高清公开图（Unsplash 直链，w=1920 清晰）**

在 `e:\wenlv\wennv\wenlv` 下执行（若某链失效，改用 Picsum seed 或 Wikimedia 同主题清晰图，并更新 `SOURCES.md`）：

```powershell
cd e:\wenlv\wennv\wenlv\public\images\home

# Hero — 成都城市/天府风景感全景（宽图）
curl.exe -L "https://images.unsplash.com/photo-1599571234909-29ed5d1e6d4c?auto=format&fit=crop&w=1920&q=80" -o hero-chengdu.jpg

# 熊猫基地
curl.exe -L "https://images.unsplash.com/photo-1564349683136-77e08dba1ef7?auto=format&fit=crop&w=1600&q=80" -o carousel-panda.jpg

# 宽窄巷子 / 古街氛围
curl.exe -L "https://images.unsplash.com/photo-1547981609-4b6bfe67ca0b?auto=format&fit=crop&w=1600&q=80" -o carousel-kuanzhai.jpg

# 都江堰 / 青城山水氛围
curl.exe -L "https://images.unsplash.com/photo-1508804185872-d7badad00f7d?auto=format&fit=crop&w=1600&q=80" -o carousel-dujiangyan.jpg

# 锦里 / 夜景古街
curl.exe -L "https://images.unsplash.com/photo-1548919973-5cef591cdbc9?auto=format&fit=crop&w=1600&q=80" -o carousel-jinli.jpg

# 西岭雪山 / 雪山
curl.exe -L "https://images.unsplash.com/photo-1464822759023-fed622ff2c3b?auto=format&fit=crop&w=1600&q=80" -o carousel-xiling.jpg
```

- [ ] **Step 3: 校验文件非空且体积合理**

```powershell
Get-ChildItem "e:\wenlv\wennv\wenlv\public\images\home\*.jpg" | Select-Object Name, Length
```

Expected: 6 个 `.jpg`，每个 `Length` > 50_000（约 50KB）；若某文件过小或下载失败 HTML，删除后换源重下。

- [ ] **Step 4: 写 SOURCES.md**

```markdown
# Image sources

All images downloaded from Unsplash (https://unsplash.com/license) for local hosting.

| File | Intended subject | Source URL |
|------|------------------|------------|
| hero-chengdu.jpg | Chengdu / cityscape hero | (实际下载 URL) |
| carousel-panda.jpg | Giant panda | (实际下载 URL) |
| carousel-kuanzhai.jpg | Historic alley / Chengdu street | (实际下载 URL) |
| carousel-dujiangyan.jpg | Mountains / irrigation landscape | (实际下载 URL) |
| carousel-jinli.jpg | Night street / traditional lane | (实际下载 URL) |
| carousel-xiling.jpg | Snow mountain | (实际下载 URL) |
```

把 Step 2 使用的最终 URL 原样填入。

- [ ] **Step 5: 目视抽查**

用系统看图打开 `hero-chengdu.jpg` 与任一张 carousel，确认不是损坏图/错误主题页；主题与景点「大致相关」即可（公开图库未必都是精确 POI 实拍）。

- [ ] **Step 6: Commit（默认跳过，等用户要求）**

---

### Task 2: Carousel 渲染真实图片（含单测）

**Files:**
- Modify: `src/components/Carousel.vue`
- Create: `src/components/Carousel.spec.ts`
- Test: `src/components/Carousel.spec.ts`

**Interfaces:**
- Consumes: `CarouselItem { id, imageUrl, titleZh, titleEn, subtitleZh?, subtitleEn? }`（已有）
- Produces: 当 `item.imageUrl` 非空字符串时，对应 slide 内存在 `<img class="carousel__image" :src="item.imageUrl">`；为空时不渲染该 `<img>`，仍显示占位层

- [ ] **Step 1: 写失败单测**

创建 `src/components/Carousel.spec.ts`：

```ts
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Carousel from './Carousel.vue'

describe('Carousel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders img when imageUrl is provided', () => {
    const wrapper = mount(Carousel, {
      props: {
        items: [
          {
            id: 'panda',
            imageUrl: '/images/home/carousel-panda.jpg',
            titleZh: '大熊猫繁育研究基地',
            titleEn: 'Giant Panda Base',
          },
        ],
      },
      global: {
        stubs: {},
      },
    })
    const img = wrapper.find('img.carousel__image')
    expect(img.exists()).toBe(true)
    expect(img.attributes('src')).toBe('/images/home/carousel-panda.jpg')
    expect(img.attributes('alt')).toContain('熊猫')
  })

  it('does not render img when imageUrl is empty', () => {
    const wrapper = mount(Carousel, {
      props: {
        items: [
          {
            id: 'empty',
            imageUrl: '',
            titleZh: '占位',
            titleEn: 'Placeholder',
          },
        ],
      },
    })
    expect(wrapper.find('img.carousel__image').exists()).toBe(false)
  })
})
```

若 `Carousel` 依赖 `useLanguageStore` 且需 locales，按项目既有 store 初始化方式补最小 stub（例如在 `beforeEach` 里确保 language store 可创建）；不要改业务逻辑迁就测试。

- [ ] **Step 2: 跑测确认失败**

```powershell
cd e:\wenlv\wennv\wenlv
npm run test:unit -- src/components/Carousel.spec.ts
```

Expected: FAIL（找不到 `img.carousel__image`）

- [ ] **Step 3: 改 Carousel 模板与样式**

在 `carousel__image-placeholder` 内、pattern 之前插入：

```vue
<img
  v-if="item.imageUrl"
  class="carousel__image"
  :src="item.imageUrl"
  :alt="langStore.lang === 'zh' ? item.titleZh : item.titleEn"
  loading="lazy"
  decoding="async"
/>
```

有图时隐藏纯装饰 pattern（避免盖住照片）：

```vue
<div v-if="!item.imageUrl" class="carousel__pattern" />
```

把 `.carousel__overlay` 改为偏暗底部渐变，例如：

```css
.carousel__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    180deg,
    rgba(15, 13, 11, 0.15) 0%,
    rgba(15, 13, 11, 0.35) 45%,
    rgba(15, 13, 11, 0.72) 100%
  );
}

.carousel__image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
}

.carousel__title {
  color: #fffdf8;
  text-shadow: 0 2px 16px rgba(0, 0, 0, 0.45);
}

.carousel__subtitle {
  color: rgba(255, 253, 248, 0.88);
}
```

保留箭头、dots、自动播放逻辑不变。

- [ ] **Step 4: 再跑单测**

```powershell
npm run test:unit -- src/components/Carousel.spec.ts
```

Expected: PASS

- [ ] **Step 5: Commit（默认跳过）**

---

### Task 3: HomeView 接通 Hero 背景与轮播 URL

**Files:**
- Modify: `src/views/HomeView/HomeView.vue`（`carouselItems` 与 `.hero__bg` / `.hero__gradient`）

**Interfaces:**
- Consumes: `/images/home/*.jpg`（Task 1）；`Carousel` 的 `imageUrl` 行为（Task 2）
- Produces: 首页 Hero 使用 `hero-chengdu.jpg`；五条轮播 `imageUrl` 分别指向五张 carousel 图

- [ ] **Step 1: 填充 carouselItems.imageUrl**

将 `carouselItems` 改为：

```ts
const carouselItems: CarouselItem[] = [
  {
    id: 'panda',
    imageUrl: '/images/home/carousel-panda.jpg',
    titleZh: '大熊猫繁育研究基地',
    titleEn: 'Giant Panda Base',
    subtitleZh: '近距离观察国宝大熊猫，感受自然之美',
    subtitleEn: 'See giant pandas up close in their natural habitat',
  },
  {
    id: 'kuanzhai',
    imageUrl: '/images/home/carousel-kuanzhai.jpg',
    titleZh: '宽窄巷子',
    titleEn: 'Kuanzhai Alleys',
    subtitleZh: '漫步清朝古街，品茗听戏，感受成都慢生活',
    subtitleEn: "Stroll Qing Dynasty alleys, sip tea, and feel Chengdu's slow pace",
  },
  {
    id: 'dujiangyan',
    imageUrl: '/images/home/carousel-dujiangyan.jpg',
    titleZh: '都江堰 · 青城山',
    titleEn: 'Dujiangyan & Mt. Qingcheng',
    subtitleZh: '千年水利工程与道教发源地的完美融合',
    subtitleEn: 'Ancient irrigation wonder meets the birthplace of Taoism',
  },
  {
    id: 'jinli',
    imageUrl: '/images/home/carousel-jinli.jpg',
    titleZh: '锦里 · 武侯祠',
    titleEn: 'Jinli & Wuhou Shrine',
    subtitleZh: '三国文化圣地，西蜀最古老的商业街',
    subtitleEn: "Three Kingdoms heritage on western Sichuan's oldest street",
  },
  {
    id: 'xiling',
    imageUrl: '/images/home/carousel-xiling.jpg',
    titleZh: '西岭雪山',
    titleEn: 'Xiling Snow Mountain',
    subtitleZh: '"窗含西岭千秋雪"——诗圣杜甫笔下的雪山胜景',
    subtitleEn: 'The snow-capped peak immortalized by poet Du Fu',
  },
]
```

- [ ] **Step 2: Hero 模板增加背景图层**

在 `hero__bg` 内、`hero__gradient` 之前加入：

```vue
<div
  class="hero__photo"
  role="img"
  aria-label="成都风景"
/>
```

- [ ] **Step 3: Hero 样式**

```css
.hero {
  min-height: 100dvh;
}

.hero__photo {
  position: absolute;
  inset: 0;
  background-image: url('/images/home/hero-chengdu.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  transform: scale(1.02);
}

.hero__gradient {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(
      180deg,
      rgba(15, 13, 11, 0.55) 0%,
      rgba(15, 13, 11, 0.35) 45%,
      var(--color-bg) 100%
    );
}
```

保留 `hero__pattern` 低透明度即可；若与照片冲突过强，将 `.hero__pattern { opacity: 0.25; }`。

确保 `.hero__content` 仍 `z-index: 1`，文字对比度足够（浅色字若不够，给 `.hero__title-zh` / subtitle 加轻微 `text-shadow`，不要改文案结构）。

- [ ] **Step 4: 手动验收**

```powershell
cd e:\wenlv\wennv\wenlv
npm run dev
```

浏览器打开 `/`：
1. Hero 全幅风景清晰，标题与按钮可读
2. 轮播五页均为照片 + 标题，非纯渐变
3. 缩到 ~375px 宽度不破版

- [ ] **Step 5: Commit（默认跳过）**

---

### Task 4: 后端代理改到 natapp

**Files:**
- Modify: `vite.config.ts`（proxy target）
- Modify: `.env.example`
- Create or Modify: `.env`（本地，勿提交密钥以外的敏感信息；本文件仅 API 基址）

**Interfaces:**
- Consumes: 无
- Produces: `npm run dev` 时对 `/api/*` 的请求转发到 `http://ra9a3c94.natappfree.cc`

- [ ] **Step 1: 改 vite proxy**

将 `vite.config.ts` 中：

```ts
target: 'http://localhost:8081',
```

改为：

```ts
target: 'http://ra9a3c94.natappfree.cc',
```

保留 `changeOrigin: true` 与 `ws: true`。注释更新为说明当前为 natapp 穿透地址。

- [ ] **Step 2: 更新 `.env.example`**

```env
# 后端服务地址(留空则使用同源 /api,由 vite 代理转发)
VITE_API_BASE_URL=http://ra9a3c94.natappfree.cc
```

- [ ] **Step 3: 确保本地 `.env` 存在同名变量**

若无 `.env`，从 `.env.example` 复制并保留既有地图 key；至少包含：

```env
VITE_API_BASE_URL=http://ra9a3c94.natappfree.cc
```

- [ ] **Step 4: 验证代理**

重启 `npm run dev` 后：

```powershell
curl.exe -s "http://127.0.0.1:5173/api/photos/wall"
```

Expected: JSON 形如 `{"code":0,"message":"ok","data":{...}}`（与直连穿透一致）。若 5173 端口不同，以终端打印的 Local URL 为准。

- [ ] **Step 5: Commit（默认跳过）**

---

## Spec Coverage Checklist

| Spec 要求 | Task |
|-----------|------|
| `public/images/home/` 6 张图 | Task 1 |
| Hero 全幅 + 遮罩 + 保留 CTA | Task 3 |
| 轮播对应 5 图 + 深色叠字 | Task 2 + 3 |
| 无图兜底占位 | Task 2 |
| Vite 代理 → natapp | Task 4 |
| `.env.example` / API 基址 | Task 4 |
| 不做内部页 / trip / 文化区配图 | 全局约束（无对应改动任务） |
| 验收：清晰、可读、不破版、API 转发 | Task 3 Step 4 + Task 4 Step 4 |

---

## Self-Review Notes

- 无 TBD /「类似 Task N」占位。
- `CarouselItem.imageUrl` 命名与现有类型一致。
- Commit 步骤按用户规则默认跳过，避免未经请求提交。
