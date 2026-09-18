# Hero 背景水波纹特效

**日期：** 2026-09-18  
**范围：** 仅公开首页（`/`）Hero 背景层  
**状态：** 设计已口头确认，待用户审阅本 spec 后进入实现计划

---

## 1. 背景与目标

公开首页 Hero 已有全幅成都风景图（`/images/home/hero-chengdu.jpg`）。希望在背景层增加类似知乎文「Canvas 水面折射」的水波纹氛围，增强「锦城 / 水面」意象，且不抢标题与 CTA。

**成功标准：**
1. 桌面：鼠标在 Hero 上移动时，背景出现轻微跟手涟漪。
2. 桌面与手机：均有偶尔、缓慢、随机位置的自动轻扰涟漪。
3. 手机：不跟手、不跟触；仅自动涟漪。
4. 波纹只扭曲照片层；渐变遮罩、图案、标题、CTA、「蜀」水印保持清晰不动。
5. `prefers-reduced-motion: reduce` 时关闭动画，回退静态背景图。
6. Hero 滚出视口或页签隐藏时暂停动画，回到可见时恢复。

---

## 2. 方案选型

采用 **方案 1：Canvas 2D 振幅图折射**。

| 方案 | 结论 |
|------|------|
| 1. Canvas 2D 双缓冲振幅 + 像素偏移 | **采用**（对标参考文，可控、够轻） |
| 2. WebGL / 着色器 | 不采用（对本期氛围过重） |
| 3. CSS / SVG 滤镜近似 | 不采用（跟手扩散感弱） |

**交互（用户确认）：**
- 触发：桌面鼠标跟手（A）+ 偶尔自动（C）
- 强度：轻柔克制（A）
- 移动端：仅自动缓慢涟漪（A）

---

## 3. 架构

### 3.1 组件

新增 `src/components/HeroRipple.vue`：

- Props：`src`（默认 `/images/home/hero-chengdu.jpg`）、可选 `alt`
- 内部：一张铺满的 `<canvas>`；加载图片后写入纹理 `ImageData`，每帧更新振幅图并 `putImageData`
- 对外无业务状态依赖；由 `HomeView` 挂载

### 3.2 挂载位置（`HomeView.vue`）

`hero__bg` 内层级（自下而上）：

1. 静态兜底 `hero__photo`（CSS `background-image`，始终存在）
2. `HeroRipple`（成功运行时盖住静态图；失败 / 降级时隐藏或透明）
3. `hero__gradient`
4. `hero__pattern`

其上仍为 `hero__shu` 与 `hero__content`（现有结构不变）。

### 3.3 算法要点

- 双缓冲振幅数组（经典 ripple map）：邻点平均 − 当前值，再衰减
- 根据振幅差对源图采样偏移，写出当前帧像素
- **扰动强度偏低**：鼠标注入能量与自动扰动半径/能量均取轻柔档，避免标题区观感发糊
- **分辨率**：按容器尺寸绘制；可用 `devicePixelRatio` 上限（如 1.5）与可选降采样（如逻辑宽最长边约 960–1280）平衡清晰与 CPU
- 鼠标：`pointermove` 映射到 canvas 坐标调用 `disturb(x, y)`；节流到约每帧一次
- 自动：定时器（约 2.5–4s 随机间隔）在安全边距内随机落点轻扰
- 桌面才绑定 pointer 跟手；触摸设备不绑定跟手监听

---

## 4. 性能与无障碍

| 项 | 行为 |
|----|------|
| `IntersectionObserver` | Hero 与视口无交集 → 停 `requestAnimationFrame` / 自动定时器 |
| `visibilitychange` | `document.hidden` → 暂停 |
| `prefers-reduced-motion` | 不启动 Canvas 动画；只显示静态 `hero__photo` |
| 图片加载失败 | 保持静态背景，不抛未处理错误 |
| Resize | `ResizeObserver` 重建缓冲（防抖），避免拉伸模糊 |

Canvas 对辅助技术：`aria-hidden="true"`（装饰层）；语义仍由外层 `role="img"` / 文案承担。

---

## 5. 明确不做

- 不扭曲标题、CTA、导航
- 不引入 jquery.ripples 等第三方库
- 不在登录页 / 内部页复用（除非后续单独需求）
- 不做点击爆炸式大涟漪、不做强对比高振幅

---

## 6. 验收清单

- [ ] 桌面移入 Hero：背景有轻柔跟手涟漪，文字清晰
- [ ] 静置数秒：偶发自动涟漪
- [ ] 手机：可见自动涟漪，滑动页面不产生跟手扰动
- [ ] 开启系统「减少动态效果」：无波纹，静态图正常
- [ ] 滚离 Hero / 切后台：CPU 占用明显下降（动画停）
- [ ] 渐变遮罩与底部衔接色仍正常

---

## 7. 实现顺序（供后续 plan）

1. 新增 `HeroRipple.vue`（算法 + 生命周期 + 降级）
2. `HomeView` 接入并调整 `hero__photo` / canvas 叠层样式
3. 手动按验收清单验证桌面与窄屏
