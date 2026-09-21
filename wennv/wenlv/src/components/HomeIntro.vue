<script setup lang="ts">
/**
 * HomeIntro.vue — 首页出场动画（宣纸白底 · 真水墨风）
 * 主题元素：宣纸白底（占满全屏）+ 水墨长卷自右向左展开
 *          + 墨分五色山水（turbulence 毛边渗墨：淡墨远山 / 中墨山峦 / 浓墨近山 / 焦墨行舟）
 *          + 留白冷月 / 飞鸟淡墨 / 飞白江水 + 杜甫《绝句》全诗四句竖排逐字墨书 + 朱砂印章
 * 水墨语言：宣纸纤维肌理 + 墨滴入水晕开聚形 + 纸缘暖调微晕 + 落款朱砂
 * 退出条件：水墨长卷展开 → 全诗 28 字逐字墨书全部写完 → 再停留 3s（点击可提前结束）；
 *          且 Hero 大图预加载完成（兜底强制放行）
 * 交互：点击任意处可跳过；由父组件负责 sessionStorage 会话级去重
 */
import { onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 需要预加载的首页主视觉大图，加载完成后才允许退场 */
    heroImage?: string
  }>(),
  { heroImage: '/images/home/hero-chengdu.jpg' }
)

const emit = defineEmits<{ (e: 'done'): void }>()

const visible = ref(true)
const leaving = ref(false)
/**
 * 水墨长卷展开完成时刻（秒）：0.4s 延迟 + 1.9s 展开 ≈ 2.3s
 * 题诗随长卷展开同步书写，卷展完时全诗刚好写完（“展开即整首诗展示完成”）
 */
const SCROLL_OPEN_SEC = 0.4 + 1.9
/** 逐字墨书参数：首字起始 + 步进 + 单字时长，三者与 CSS 中 .ch 动画保持一致 */
const CHAR_START_SEC = 0.5
const CHAR_STEP_SEC = 0.06
const CHAR_DUR_SEC = 0.6
/** 全诗 28 字逐字墨书完成时刻（秒）：首字起始 + 27×步进 + 单字时长 ≈ 2.72s */
const POEM_END_SEC = CHAR_START_SEC + 27 * CHAR_STEP_SEC + CHAR_DUR_SEC
/** 全诗展示完成后再停留 3s 才可退场 */
const AFTER_POEM_HOLD_MS = 5000
/** 进度条填充时长（秒），与“长卷展开 + 全诗完成 + 3s 停留”对齐 */
const barDuration = Math.max(SCROLL_OPEN_SEC, POEM_END_SEC) + AFTER_POEM_HOLD_MS / 1000

let finished = false
const timers: number[] = []
let watchDog: number | undefined

function finish() {
  if (finished) return
  finished = true
  if (watchDog !== undefined) window.clearTimeout(watchDog)
  leaving.value = true
  // 与 CSS 退场过渡时长（0.9s）对齐后再移除节点
  timers.push(
    window.setTimeout(() => {
      visible.value = false
      document.body.style.overflow = ''
      emit('done')
    }, 920)
  )
}

onMounted(() => {
  // 出场期间锁定页面滚动，避免动画未结束时页面可拖动
  document.body.style.overflow = 'hidden'

  // 全诗墨书完成后再停留 3s，之后且图片就绪才可退场；点击 finish 可提前结束
  const MIN_SHOW = Math.round(POEM_END_SEC * 1000 + AFTER_POEM_HOLD_MS)
  let imageReady = false
  let minDone = false

  const tryFinish = () => {
    if (imageReady && minDone) finish()
  }

  // 预加载 Hero 大图：出场动画正好覆盖首页内容加载
  const img = new Image()
  img.onload = () => {
    imageReady = true
    tryFinish()
  }
  img.onerror = () => {
    // 图即使加载失败也不应卡住用户
    imageReady = true
    tryFinish()
  }
  img.src = props.heroImage

  timers.push(
    window.setTimeout(() => {
      minDone = true
      tryFinish()
    }, MIN_SHOW)
  )
  // 兜底：全诗 + 3s 再多等 4s，无论如何强制放行
  watchDog = window.setTimeout(() => finish(), MIN_SHOW + 4000)
})

onBeforeUnmount(() => {
  timers.forEach((t) => window.clearTimeout(t))
  if (watchDog !== undefined) window.clearTimeout(watchDog)
  document.body.style.overflow = ''
})
</script>

<template>
  <div
    v-if="visible"
    class="intro"
    :class="{ 'intro--leaving': leaving }"
    aria-hidden="true"
    @click="finish"
  >
    <!-- 宣纸白底 + 纸面微光 -->
    <div class="intro__wash" />

    <div class="intro__stage">
      <!-- ──── 水墨长卷 ──── -->
      <div class="scroll-band">
        <!-- 卷纸：随左轴向左展开（clip-path 与轴同步） -->
        <div class="scroll-paper">
          <svg
            class="scroll-scene"
            viewBox="0 0 1100 420"
            preserveAspectRatio="xMidYMid slice"
            aria-hidden="true"
          >
            <defs>
              <!-- 墨渗毛边：湍流置换，各层 seed 不同避免重复纹理 -->
              <filter id="ink-far" x="-10%" y="-10%" width="120%" height="120%">
                <feTurbulence type="fractalNoise" baseFrequency="0.010 0.024" numOctaves="3" seed="3" result="n" />
                <feDisplacementMap in="SourceGraphic" in2="n" scale="34" />
              </filter>
              <filter id="ink-mid" x="-10%" y="-10%" width="120%" height="120%">
                <feTurbulence type="fractalNoise" baseFrequency="0.014 0.030" numOctaves="3" seed="11" result="n" />
                <feDisplacementMap in="SourceGraphic" in2="n" scale="26" />
              </filter>
              <filter id="ink-snow" x="-10%" y="-10%" width="120%" height="120%">
                <feTurbulence type="fractalNoise" baseFrequency="0.016 0.034" numOctaves="3" seed="7" result="n" />
                <feDisplacementMap in="SourceGraphic" in2="n" scale="22" />
              </filter>
              <filter id="ink-detail" x="-10%" y="-10%" width="120%" height="120%">
                <feTurbulence type="fractalNoise" baseFrequency="0.03 0.06" numOctaves="2" seed="8" result="n" />
                <feDisplacementMap in="SourceGraphic" in2="n" scale="12" />
              </filter>
              <!-- 留白冷月：纸上淡灰晕圈 -->
              <radialGradient id="moon-glow" cx="50%" cy="50%" r="50%">
                <stop offset="0%" stop-color="#8C8578" stop-opacity="0.10" />
                <stop offset="55%" stop-color="#8C8578" stop-opacity="0.05" />
                <stop offset="100%" stop-color="#8C8578" stop-opacity="0" />
              </radialGradient>
            </defs>

            <!-- 冷月：留白 + 淡墨勾边（千秋雪之夜） -->
            <g class="scene-el scene-moon">
              <circle cx="190" cy="96" r="52" fill="url(#moon-glow)" />
              <circle cx="190" cy="96" r="16" fill="#FDFAF1" stroke="#57503F" stroke-opacity="0.35" stroke-width="1.2" />
            </g>

            <!-- 飞鸟：两笔淡墨 -->
            <g class="scene-el scene-birds" stroke="#2E2820" stroke-width="1.6" stroke-linecap="round" opacity="0.45" fill="none">
              <path d="M768,118 q9,-9 18,0" />
              <path d="M798,131 q8,-8 16,0" />
            </g>

            <!-- 远山：淡墨一抹，如雾 -->
            <g filter="url(#ink-far)">
              <path
                class="scene-el scene-far"
                d="M0,258 C120,190 230,238 350,206 C470,175 590,242 715,212 C835,188 955,234 1100,198 L1100,420 L0,420 Z"
                fill="#3A342A" opacity="0.08"
              />
            </g>

            <!-- 中山：中墨山峦 -->
            <g filter="url(#ink-mid)">
              <path
                class="scene-el scene-mid"
                d="M0,296 C130,238 255,286 395,252 C520,222 640,296 785,264 C905,238 1005,288 1100,256 L1100,420 L0,420 Z"
                fill="#332D23" opacity="0.16"
              />
            </g>

            <!-- 西岭雪山：浓墨勾斫 + 留白雪冠 -->
            <g filter="url(#ink-snow)">
              <path class="scene-el scene-snow" d="M40,332 L152,150 L200,196 L268,118 L300,158 L365,332 Z" fill="#241F17" opacity="0.62" />
              <path class="scene-el scene-snow" d="M268,118 L292,152 L280,158 L268,146 L254,162 L242,148 L232,164 L218,152 L200,196 Z" fill="#FDFAF1" opacity="0.92" />
              <path class="scene-el scene-snow" d="M152,150 L170,180 L158,186 L146,176 L136,188 L128,178 Z" fill="#FDFAF1" opacity="0.85" />
            </g>

            <!-- 近山：浓墨压脚 -->
            <g filter="url(#ink-mid)">
              <path
                class="scene-el scene-near"
                d="M0,352 C95,320 185,342 305,324 C425,302 505,342 625,328 C745,308 865,342 985,322 C1035,315 1065,324 1100,317 L1100,420 L0,420 Z"
                fill="#221D15" opacity="0.42"
              />
            </g>

            <!-- 江水：飞白笔意 -->
            <g class="scene-el scene-water" filter="url(#ink-detail)" stroke="#4A4335" stroke-width="1.6" stroke-linecap="round" fill="none">
              <path d="M110,384 H430" opacity="0.22" />
              <path d="M560,394 H900" opacity="0.14" />
            </g>

            <!-- 东吴行舟：焦墨剪影 -->
            <g class="scene-el scene-boat" filter="url(#ink-detail)">
              <path d="M470,362 q42,14 86,0 l-9,11 h-68 z" fill="#14100B" opacity="0.88" />
              <path d="M512,360 L512,324 Q532,340 537,360 Z" fill="#14100B" opacity="0.7" />
            </g>
          </svg>

          <!-- 卷首题诗：杜甫《绝句》全诗四句竖排，从右向左依次为首句→尾句 -->
          <div class="poem">
            <span class="poem__col">
              <i class="ch" style="--i: 0">两</i><i class="ch" style="--i: 1">个</i><i class="ch" style="--i: 2">黄</i><i class="ch" style="--i: 3">鹂</i><i class="ch" style="--i: 4">鸣</i><i class="ch" style="--i: 5">翠</i><i class="ch" style="--i: 6">柳</i>
            </span>
            <span class="poem__col">
              <i class="ch" style="--i: 7">一</i><i class="ch" style="--i: 8">行</i><i class="ch" style="--i: 9">白</i><i class="ch" style="--i: 10">鹭</i><i class="ch" style="--i: 11">上</i><i class="ch" style="--i: 12">青</i><i class="ch" style="--i: 13">天</i>
            </span>
            <span class="poem__col">
              <i class="ch" style="--i: 14">窗</i><i class="ch" style="--i: 15">含</i><i class="ch" style="--i: 16">西</i><i class="ch" style="--i: 17">岭</i><i class="ch" style="--i: 18">千</i><i class="ch" style="--i: 19">秋</i><i class="ch" style="--i: 20">雪</i>
            </span>
            <span class="poem__col">
              <i class="ch" style="--i: 21">门</i><i class="ch" style="--i: 22">泊</i><i class="ch" style="--i: 23">东</i><i class="ch" style="--i: 24">吴</i><i class="ch" style="--i: 25">万</i><i class="ch" style="--i: 26">里</i><i class="ch" style="--i: 27">船</i>
            </span>
            <span class="poem__sig">杜甫 · 绝句</span>
            <span class="poem__seal">蜀韵</span>
          </div>
        </div>

        <!-- 卷轴：黑檀木轴，右轴固定，左轴随卷展开向左滚动 -->
        <div class="rod rod--left" />
        <div class="rod rod--right" />
      </div>

      <!-- ──── 标题区 ──── -->
      <h1 class="intro__title">蜀韵 · 成都</h1>
      <p class="intro__sub">
        锦绣天府 · 巴蜀文化出海
        <span class="intro__sub-en">SHU · CHENGDU</span>
      </p>

      <!-- 加载进度 -->
      <div class="intro__progress">
        <span
          class="intro__progress-bar"
          :style="{ animationDuration: `${barDuration}s` }"
        />
      </div>
      <p class="intro__skip">轻触任意处跳过</p>
    </div>

    <!-- ──── 水墨语言层 ──── -->
    <!-- 纸缘暖调微晕（替代暗角） -->
    <div class="intro__vignette" />
    <!-- 宣纸纤维肌理 -->
    <svg class="intro__grain" aria-hidden="true">
      <filter id="grain-f">
        <feTurbulence type="fractalNoise" baseFrequency="0.8" numOctaves="2" stitchTiles="stitch" />
        <feColorMatrix type="saturate" values="0" />
      </filter>
      <rect width="100%" height="100%" filter="url(#grain-f)" />
    </svg>
  </div>
</template>

<style scoped>
.intro {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  /* 宣纸白底：与首页米白同族，白底占满 */
  background: #FAF6EC;
  overflow: hidden;
  cursor: pointer;
  transition:
    opacity 0.9s ease,
    transform 0.9s ease,
    background-color 0.9s ease;
}

/* 落幕：宣纸渐变为首页米白底后淡出 */
.intro--leaving {
  background: var(--color-bg);
  opacity: 0;
  transform: scale(1.02);
  pointer-events: none;
}

/* 宣纸底 + 中央纸面微光 + 绢丝纹理 */
.intro__wash {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 56% 46% at 50% 42%, rgba(255, 253, 246, 0.9) 0%, transparent 66%),
    repeating-conic-gradient(
      transparent 0deg 89deg,
      rgba(140, 125, 90, 0.016) 90deg 91deg,
      transparent 91deg 179deg,
      rgba(140, 125, 90, 0.016) 180deg 181deg
    );
  background-size: 100% 100%, 64px 64px;
  animation: intro-wash 1.4s ease both;
}

@keyframes intro-wash {
  from { opacity: 0; }
  to { opacity: 1; }
}

.intro__stage {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--space-8);
  transition: transform 0.9s ease, opacity 0.9s ease;
}

.intro--leaving .intro__stage {
  transform: translateY(-16px);
  opacity: 0.3;
}

/* ============================================================
   水墨长卷
   ============================================================ */
.scroll-band {
  --rod-w: 15px;
  position: relative;
  width: min(1080px, 88vw);
  /* 高度兼顾桌面与矮窗口：矮窗口下不压缩到装不下整首竖排诗 */
  height: clamp(240px, 50vh, 400px);
  margin-bottom: var(--space-10);
}

/* 卷纸：宣纸白，随左轴展开（右→左揭示） */
.scroll-paper {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background:
    linear-gradient(180deg, #FDFAF2 0%, #F6F1E4 100%);
  border-top: 1px solid rgba(90, 78, 55, 0.18);
  border-bottom: 1px solid rgba(90, 78, 55, 0.18);
  box-shadow:
    0 14px 40px rgba(96, 80, 48, 0.14),
    0 3px 10px rgba(96, 80, 48, 0.10);
  clip-path: inset(-40px 0 -40px 100%);
  animation: scroll-unroll 1.9s cubic-bezier(0.3, 0.1, 0.25, 1) 0.4s both;
}

@keyframes scroll-unroll {
  from { clip-path: inset(-40px 0 -40px 100%); }
  to { clip-path: inset(-40px 0 -40px 0); }
}

/* 卷内水墨画 */
.scroll-scene {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

/* 卷尾云气留白：右端墨晕自然消散入纸色雾气，题诗落于留白处（水墨构图惯例） */
.scroll-paper::after {
  content: '';
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
  background:
    radial-gradient(
      ellipse 62% 78% at 87% 48%,
      rgba(250, 246, 236, 0.97) 0%,
      rgba(250, 246, 236, 0.88) 42%,
      rgba(250, 246, 236, 0.45) 68%,
      rgba(250, 246, 236, 0) 100%
    );
}

/* 墨晕聚形：墨滴在水中散开又聚拢的慢镜头 */
.scene-el {
  animation: ink-bloom 1.4s cubic-bezier(0.2, 0.55, 0.25, 1) both;
}

.scene-moon { animation-delay: 1.2s; }
.scene-far { animation-delay: 1.1s; }
.scene-mid { animation-delay: 1.3s; }
.scene-snow { animation-delay: 1.5s; }
.scene-near { animation-delay: 1.7s; }
.scene-birds { animation-delay: 1.8s; }
.scene-water { animation-delay: 1.9s; }
.scene-boat { animation-delay: 2.1s; }

@keyframes ink-bloom {
  from {
    opacity: 0;
    filter: blur(18px);
    transform: translateX(28px);
  }
  to {
    opacity: 1;
    filter: blur(0);
    transform: translateX(0);
  }
}

/* 卷首题诗：竖排，右侧起笔，落于卷尾云气留白处 */
.poem {
  position: absolute;
  top: 50%;
  right: 6.5%;
  transform: translateY(-50%);
  z-index: 2;
  display: flex;
  flex-direction: row-reverse;
  align-items: center;
  gap: clamp(10px, 1.5vw, 18px);
}

.poem__col {
  writing-mode: vertical-rl;
  font-family: var(--font-display);
  font-weight: 700;
  /* 同时参考视宽与视高：矮窗口自动缩小，保证 7 字竖排整列完整不被裁切 */
  font-size: clamp(14px, min(2.2vw, 4.4vh), 24px);
  line-height: 1;
  color: #1A1610;
  letter-spacing: 0.14em;
  font-style: normal;
}

/* 逐字墨书：墨色晕开 → 收笔成形（全诗 28 字，随长卷展开同步书写，--i 步进 0.06s） */
.ch {
  display: inline-block;
  animation: ch-brush 0.6s cubic-bezier(0.22, 0.61, 0.36, 1) both;
  animation-delay: calc(0.5s + var(--i) * 0.06s);
}

@keyframes ch-brush {
  from {
    opacity: 0;
    filter: blur(9px);
    transform: scale(1.3);
  }
  to {
    opacity: 1;
    filter: blur(0);
    transform: scale(1);
  }
}

.poem__sig {
  writing-mode: vertical-rl;
  font-family: var(--font-display);
  font-size: clamp(12px, 1.4vw, 15px);
  color: rgba(38, 31, 22, 0.75);
  letter-spacing: 0.22em;
  animation: sig-fade 0.6s ease 2s both;
}

@keyframes sig-fade {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* 朱砂印章 — 落款盖章效果（白纸上一点红，水墨点睛） */
.poem__seal {
  writing-mode: vertical-rl;
  align-self: flex-end;
  font-family: var(--font-display);
  font-size: clamp(12px, 1.5vw, 17px);
  font-weight: 700;
  letter-spacing: 0.14em;
  color: #C0392B;
  border: 1.5px solid #C0392B;
  border-radius: 3px;
  padding: 0.32em 0.2em;
  background: rgba(192, 57, 43, 0.06);
  transform-origin: center;
  animation: seal-stamp 0.55s cubic-bezier(0.2, 1.4, 0.4, 1) 2.05s both;
}

@keyframes seal-stamp {
  0% {
    opacity: 0;
    transform: scale(2) rotate(7deg);
  }
  62% {
    opacity: 1;
    transform: scale(0.92) rotate(-4deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) rotate(-3deg);
  }
}

/* 卷轴：黑檀木 + 鎏金轴头 */
.rod {
  position: absolute;
  top: -14px;
  height: calc(100% + 28px);
  width: var(--rod-w);
  border-radius: 8px;
  background: linear-gradient(90deg, #2B2418, #171208 45%, #0C0A07);
  box-shadow: 0 3px 14px rgba(60, 48, 26, 0.4);
  z-index: 2;
}

/* 轴头：鎏金圆珠 */
.rod::before,
.rod::after {
  content: '';
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 35%, #8A6F3C, #C9A96E 45%, #4A3A22);
}

.rod::before { top: -6px; }
.rod::after { bottom: -6px; }

.rod--right {
  right: calc(var(--rod-w) / -2);
}

/* 左轴：随卷纸展开向左滚动（与 scroll-unroll 同步） */
.rod--left {
  left: calc(100% - var(--rod-w) / 2);
  animation: rod-roll 1.9s cubic-bezier(0.3, 0.1, 0.25, 1) 0.4s both;
}

@keyframes rod-roll {
  from { left: calc(100% - var(--rod-w) / 2); }
  to { left: calc(var(--rod-w) / -2); }
}

/* ============================================================
   标题区（白纸墨字）
   ============================================================ */
.intro__title {
  font-family: var(--font-display);
  font-size: clamp(1.4rem, 4vw, 2rem);
  font-weight: 700;
  color: #26211A;
  animation: intro-title 1s ease 2.6s both;
}

@keyframes intro-title {
  from {
    opacity: 0;
    letter-spacing: 0.45em;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    letter-spacing: 0.14em;
    transform: translateY(0);
  }
}

.intro__sub {
  margin-top: var(--space-3);
  font-size: var(--text-sm);
  color: rgba(70, 60, 44, 0.62);
  letter-spacing: var(--tracking-widest);
  text-align: center;
  animation: intro-fade-up 0.9s ease 2.8s both;
}

.intro__sub-en {
  display: block;
  margin-top: var(--space-2);
  font-family: var(--font-en-display);
  font-size: var(--text-xs);
  font-style: italic;
  color: #9A7F4E;
  letter-spacing: 0.3em;
}

@keyframes intro-fade-up {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 加载进度 — 墨线填充，末端一点朱砂 */
.intro__progress {
  width: min(260px, 48vw);
  height: 2px;
  margin-top: var(--space-10);
  border-radius: var(--radius-full);
  background: rgba(70, 60, 44, 0.14);
  overflow: hidden;
  animation: intro-fade-up 0.6s ease 0.6s both;
}

.intro__progress-bar {
  display: block;
  width: 100%;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #3A342A, #6B5E45);
  transform-origin: left;
  animation-name: intro-bar;
  animation-timing-function: cubic-bezier(0.3, 0.1, 0.3, 1);
  animation-fill-mode: both;
  animation-delay: 0.5s;
}

@keyframes intro-bar {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}

.intro__skip {
  margin-top: var(--space-5);
  font-size: var(--text-xs);
  color: rgba(70, 60, 44, 0.4);
  letter-spacing: var(--tracking-wide);
  animation: intro-fade-up 0.8s ease 3s both;
}

/* ============================================================
   水墨语言层
   ============================================================ */
/* 纸缘暖调微晕：宣纸陈年边缘的暖色，替代暗角 */
.intro__vignette {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
  background: radial-gradient(
    ellipse 120% 92% at 50% 46%,
    transparent 52%,
    rgba(178, 152, 104, 0.13) 100%
  );
}

/* 宣纸纤维肌理：静态细微噪点（不跳动，纸上无胶片颗粒） */
.intro__grain {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 4;
  opacity: 0.04;
  pointer-events: none;
  mix-blend-mode: multiply;
}

/* 弱动效偏好：跳过装饰性动画，整体快速淡入 */
@media (prefers-reduced-motion: reduce) {
  .intro__wash,
  .scroll-paper,
  .rod--left,
  .scene-el,
  .ch,
  .poem__sig,
  .poem__seal,
  .intro__title,
  .intro__sub,
  .intro__progress,
  .intro__skip {
    animation-duration: 0.3s !important;
    animation-delay: 0s !important;
  }
}

@media (max-width: 480px) {
  .scroll-band {
    --rod-w: 12px;
    height: 240px;
    margin-bottom: var(--space-8);
  }

  .poem {
    right: 5%;
    gap: 8px;
  }

  .poem__col {
    font-size: 15px;
  }

  .poem__seal {
    font-size: 11px;
  }

  .intro__progress {
    margin-top: var(--space-8);
  }
}
</style>
