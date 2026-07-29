<template>
  <div class="landing">
    <!-- ===== 背景粒子 ===== -->
    <canvas ref="canvasRef" class="particle-bg" />

    <!-- ===== 导航栏 ===== -->
    <header class="navbar" ref="navRef">
      <div class="nav-brand">
        <svg class="nav-logo" viewBox="0 0 40 40" width="18" height="22">
          <circle cx="20" cy="20" r="18" fill="none" stroke="#E85D3A" stroke-width="1.5" opacity="0.3"/>
          <circle cx="20" cy="20" r="14" fill="none" stroke="#E85D3A" stroke-width="1" opacity="0.5"/>
          <g v-for="i in 12" :key="i">
            <line :x1="20 + 8 * Math.cos(i * Math.PI / 6)" :y1="20 + 8 * Math.sin(i * Math.PI / 6)"
                  :x2="20 + 14 * Math.cos(i * Math.PI / 6)" :y2="20 + 14 * Math.sin(i * Math.PI / 6)"
                  stroke="#E85D3A" stroke-width="1.5" stroke-linecap="round" opacity="0.5"/>
          </g>
          <circle cx="20" cy="20" r="6" fill="none" stroke="#E85D3A" stroke-width="1.2" opacity="0.4"/>
        </svg>
        <span class="nav-title">蓉城漫游</span>
      </div>
      <div class="nav-links">
        <a href="#features" class="nav-link">亮点</a>
        <a href="#spots" class="nav-link">景点</a>
        <a href="#how" class="nav-link">攻略</a>
      </div>
      <button class="nav-cta" @click="goToLogin">开始探索 →</button>
    </header>

    <!-- ===== Hero ===== -->
    <section class="hero">
      <div class="hero-bg-deco" aria-hidden="true">
        <svg viewBox="0 0 200 200" class="hero-ring">
          <circle cx="100" cy="100" r="96" fill="none" stroke="rgba(232,93,58,.04)" stroke-width="1"/>
          <circle cx="100" cy="100" r="80" fill="none" stroke="rgba(212,168,84,.04)" stroke-width="0.5"/>
          <circle cx="100" cy="100" r="56" fill="none" stroke="rgba(232,93,58,.03)" stroke-width="0.5"/>
          <g v-for="i in 12" :key="i">
            <line :x1="100 + 56 * Math.cos(i * Math.PI / 6)" :y1="100 + 56 * Math.sin(i * Math.PI / 6)"
                  :x2="100 + 76 * Math.cos(i * Math.PI / 6)" :y2="100 + 76 * Math.sin(i * Math.PI / 6)"
                  stroke="rgba(232,93,58,.04)" stroke-width="1" stroke-linecap="round"/>
          </g>
        </svg>
      </div>

      <!-- 浮动文化符号 -->
      <div class="float-symbols">
        <span class="fs-item" ref="fs1">🐼</span>
        <span class="fs-item" ref="fs2">🍵</span>
        <span class="fs-item" ref="fs3">🎋</span>
        <span class="fs-item" ref="fs4">🏮</span>
        <span class="fs-item" ref="fs5">🎭</span>
      </div>

      <div class="hero-content">
        <div class="hero-badge" ref="badgeRef">
          <span class="badge-dot"></span>
          成都文旅 · 数字体验
        </div>

        <h1 class="hero-title" ref="titleRef">
          <span class="ht-line">烟火里的</span>
          <span class="ht-line hl">幸福成都</span>
        </h1>

        <p class="hero-desc" ref="descRef">
          从宽窄巷子的茶香到大熊猫的憨态<br class="br-desk">
          用全新的方式，探索三千年的巴蜀文明
        </p>

        <div class="hero-actions" ref="actionsRef">
          <button class="btn-primary" @click="goToLogin">
            <span>开启蓉城之旅</span>
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
              <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
            </svg>
            <div class="btn-glow"></div>
          </button>
          <button class="btn-ghost" @click="scrollToFeatures">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span>了解更多</span>
          </button>
        </div>

        <div class="hero-stats" ref="statsRef">
          <div class="hs-item"><span class="hs-num">18+</span><span class="hs-label">精选景点</span></div>
          <div class="hs-divider"></div>
          <div class="hs-item"><span class="hs-num">4.9</span><span class="hs-label">用户评分</span></div>
          <div class="hs-divider"></div>
          <div class="hs-item"><span class="hs-num">12k+</span><span class="hs-label">旅行者</span></div>
        </div>
      </div>
    </section>

    <!-- ===== 核心亮点 ===== -->
    <section id="features" class="section-block features" ref="featuresRef">
      <div class="sec-header" ref="featHeaderRef">
        <div class="sec-deco">
          <svg viewBox="0 0 40 40" width="14" height="14"><circle cx="20" cy="20" r="18" fill="none" stroke="#E85D3A" stroke-width="1.5" opacity=".3"/><circle cx="20" cy="20" r="12" fill="none" stroke="#E85D3A" stroke-width="1" opacity=".5"/></svg>
        </div>
        <span class="sec-tag">WHY CHENGDU</span>
        <h2 class="sec-title">为什么是成都</h2>
        <p class="sec-desc">一座来了就不想走的城市，用科技让旅行更简单</p>
      </div>
      <div class="feature-grid">
        <div class="feat-card" v-for="(f, i) in features" :key="f.title"
             :class="{ expanded: expandedFeature === i }"
             @click="toggleFeature(i)" ref="cardRefs">
          <div class="fc-accent" :style="{ background: f.tagClr || '#E85D3A' }"></div>
          <div class="fc-inner">
            <div class="fc-icon-wrap" :style="{ background: f.bg }">
              <span class="fc-icon">{{ f.icon }}</span>
            </div>
            <h3 class="fc-title">{{ f.title }}</h3>
            <p class="fc-desc">{{ expandedFeature === i ? f.detail : f.desc }}</p>
            <span class="fc-expand">{{ expandedFeature === i ? '收起 −' : '了解更多 +' }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 景点轮播 ===== -->
    <section id="spots" class="section-block carousel-section" ref="carouselRef">
      <div class="section-divider">
        <span class="sd-icon">◇</span>
      </div>
      <div class="sec-header" ref="spotHeaderRef">
        <div class="sec-deco">
          <svg viewBox="0 0 40 40" width="14" height="14"><circle cx="20" cy="20" r="18" fill="none" stroke="#E85D3A" stroke-width="1.5" opacity=".3"/><circle cx="20" cy="20" r="12" fill="none" stroke="#E85D3A" stroke-width="1" opacity=".5"/></svg>
        </div>
        <span class="sec-tag">SPOTLIGHT</span>
        <h2 class="sec-title">精选目的地</h2>
        <p class="sec-desc">每一处风景，都值得你为它停留</p>
      </div>

      <div class="carousel-viewport" ref="viewportRef"
           @mouseenter="stopAutoPlay" @mouseleave="startAutoPlay"
           @pointerdown="onCarouselDown" @pointermove="onCarouselMove" @pointerup="onCarouselUp">
        <div class="carousel-track" ref="trackRef" :style="{ transform: `translateX(${carouselOffset}px)` }">
          <div class="carousel-card" v-for="(spot, i) in carouselSpots" :key="spot.id"
               :class="{ active: i === currentSlide }">
            <div class="cc-visual" :style="{ background: spot.gradient }">
              <span class="cc-emoji">{{ spot.emoji }}</span>
              <div class="cc-rating-badge">{{ spot.rating }}</div>
            </div>
            <div class="cc-body">
              <span class="cc-category" :style="{ color: spot.tagClr }">{{ spot.category }}</span>
              <h3 class="cc-name">{{ spot.name }}</h3>
              <p class="cc-desc">{{ spot.desc }}</p>
              <div class="cc-footer">
                <span class="cc-dist">📍 {{ spot.distance }}</span>
                <span class="cc-explore">探索 →</span>
              </div>
            </div>
          </div>
        </div>

        <button class="carousel-btn prev" @click="prevSlide">‹</button>
        <button class="carousel-btn next" @click="nextSlide">›</button>
      </div>

      <div class="carousel-progress">
        <div class="cp-bar" :style="{ width: progressPct + '%' }"></div>
      </div>
      <div class="carousel-dots">
        <button v-for="(_, i) in carouselSpots" :key="i" class="carousel-dot"
                :class="{ active: i === currentSlide }" @click="goToSlide(i)"></button>
      </div>
    </section>

    <!-- ===== 旅行攻略 ===== -->
    <section id="how" class="section-block guide-section" ref="guideRef">
      <div class="section-divider">
        <span class="sd-icon">◇</span>
      </div>
      <div class="sec-header" ref="guideHeaderRef">
        <div class="sec-deco">
          <svg viewBox="0 0 40 40" width="14" height="14"><circle cx="20" cy="20" r="18" fill="none" stroke="#E85D3A" stroke-width="1.5" opacity=".3"/><circle cx="20" cy="20" r="12" fill="none" stroke="#E85D3A" stroke-width="1" opacity=".5"/></svg>
        </div>
        <span class="sec-tag">TRAVEL GUIDE</span>
        <h2 class="sec-title">旅行攻略</h2>
        <p class="sec-desc">三天两夜，解锁最地道的成都玩法</p>
      </div>

      <div class="guide-grid">
        <div class="guide-card" v-for="(g, i) in guides" :key="g.title"
             :class="{ expanded: expandedGuide === i }"
             ref="guideRefs">
          <div class="gc-day" @click="toggleGuide(i)">
            <span class="gc-day-num">DAY {{ i + 1 }}</span>
            <div class="gc-day-line" v-if="i < guides.length - 1 && expandedGuide !== i"></div>
            <svg class="gc-arrow" :class="{ open: expandedGuide === i }" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </div>
          <div class="gc-body">
            <div class="gc-icon">{{ g.icon }}</div>
            <h3 class="gc-title">{{ g.title }}</h3>
            <p class="gc-desc">{{ g.desc }}</p>
            <div class="gc-tags">
              <span v-for="t in g.tags" :key="t">{{ t }}</span>
            </div>
            <!-- 展开详情 -->
            <transition name="gc-expand">
              <div class="gc-detail" v-if="expandedGuide === i">
                <div class="gc-timeline">
                  <div class="gc-tl-item" v-for="step in g.detail" :key="step.time">
                    <div class="gc-tl-dot"></div>
                    <div class="gc-tl-body">
                      <span class="gc-tl-time">{{ step.time }}</span>
                      <span class="gc-tl-place">{{ step.place }}</span>
                      <span class="gc-tl-note">{{ step.note }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== CTA ===== -->
    <section class="cta-section" ref="ctaRef">
      <div class="cta-card">
        <div class="cta-deco"></div>
        <div class="cta-content">
          <h2 class="cta-title">现在就出发</h2>
          <p class="cta-desc">下载蓉城漫游，让每一次旅行都成为故事</p>
          <button class="btn-primary btn-lg" @click="goToLogin">
            <span>免费开始使用</span>
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
              <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
            </svg>
          </button>
        </div>
      </div>

      <div class="cta-footer">
        <div class="cf-brand">蓉城漫游 · CHENGDU CULTURE</div>
        <div class="cf-copy">© 2024 巴蜀文化数字体验平台</div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)
const router = useRouter()

// ===== Refs =====
const canvasRef = ref(null)
const navRef = ref(null)
const badgeRef = ref(null)
const titleRef = ref(null)
const descRef = ref(null)
const actionsRef = ref(null)
const statsRef = ref(null)
const fs1 = ref(null); const fs2 = ref(null); const fs3 = ref(null)
const fs4 = ref(null); const fs5 = ref(null)
const featuresRef = ref(null)
const featHeaderRef = ref(null)
const cardRefs = ref([])
const carouselRef = ref(null)
const spotHeaderRef = ref(null)
const viewportRef = ref(null)
const trackRef = ref(null)
const guideRef = ref(null)
const guideHeaderRef = ref(null)
const guideRefs = ref([])
const ctaRef = ref(null)

// ===== 内容数据 =====
const features = [
  { icon: '🗺️', title: '智能路线规划', desc: 'AI 根据你的时间和兴趣，自动生成最优路线', detail: '基于深度学习的路线推荐引擎，综合交通、天气、人流数据，秒级生成最优行程。支持多人协同编辑，实时调整不走回头路。', bg: 'rgba(232,93,58,.08)', tagClr: '#E85D3A' },
  { icon: '🌐', title: '3D 全景地图', desc: '交互式 3D 地图，景点分布一目了然', detail: '基于 Three.js 构建的高精度 3D 地图，支持 360° 旋转缩放。覆盖成都全域 18+ 景点，实时路况 + 景点热度一目了然。', bg: 'rgba(14,165,160,.08)', tagClr: '#0EA5A0' },
  { icon: '📷', title: 'AR 实景导览', desc: '拿起手机扫一扫，看千年历史在眼前重现', detail: '基于 WebXR 的增强现实引擎，识别景区标志物后自动叠加历史复原影像、文化解说与导航箭头。支持离线缓存，无网可用。', bg: 'rgba(212,168,84,.08)', tagClr: '#D4A854' },
  { icon: '🧠', title: '文化知识图谱', desc: '边逛边学，解锁成都历史与民俗冷知识', detail: '关联 2000+ 文化实体节点的知识图谱，涵盖历史人物、诗词典故、美食民俗。每次签到解锁隐藏故事，深度游必备。', bg: 'rgba(139,126,200,.08)', tagClr: '#8B7EC8' },
]

const carouselSpots = [
  { id: 1, name: '大熊猫繁育研究基地', category: '自然生态', emoji: '🐼', gradient: 'linear-gradient(135deg, #10B981, #34D399)', tagClr: '#10B981', desc: '与国宝近距离接触，在翠竹掩映中看呆萌熊猫嬉戏玩耍', distance: '成华区 · 30min', rating: '⭐ 4.9' },
  { id: 2, name: '宽窄巷子', category: '人文历史', emoji: '🏛️', gradient: 'linear-gradient(135deg, #E85D3A, #F5A623)', tagClr: '#E85D3A', desc: '青砖黛瓦间的老成都记忆，茶香袅袅中品味地道慢生活', distance: '青羊区 · 市中心', rating: '⭐ 4.8' },
  { id: 3, name: '青城山', category: '自然风光', emoji: '🌿', gradient: 'linear-gradient(135deg, #0EA5A0, #34D399)', tagClr: '#0EA5A0', desc: '青城天下幽，漫步山林间感受道家文化的清静无为', distance: '都江堰市 · 60min', rating: '⭐ 4.8' },
  { id: 4, name: '武侯祠', category: '三国文化', emoji: '⚔️', gradient: 'linear-gradient(135deg, #C43E1C, #E85D3A)', tagClr: '#C43E1C', desc: '三国圣地，感受诸葛亮「鞠躬尽瘁」的千古情怀', distance: '武侯区 · 20min', rating: '⭐ 4.7' },
  { id: 5, name: '锦里古街', category: '美食民俗', emoji: '🏮', gradient: 'linear-gradient(135deg, #D4A854, #F5A623)', tagClr: '#D4A854', desc: '灯火阑珊的古街，串串香与糖画交织的成都味道', distance: '武侯区 · 15min', rating: '⭐ 4.6' },
  { id: 6, name: '都江堰', category: '世界遗产', emoji: '🌊', gradient: 'linear-gradient(135deg, #0EA5A0, #06B6D4)', tagClr: '#0EA5A0', desc: '千年水利工程奇迹，看岷江水在此分流灌溉天府', distance: '都江堰市 · 50min', rating: '⭐ 4.9' },
]

const guides = [
  {
    icon: '🏛️', title: '人文成都 · 一日穿越', desc: '从金沙遗址到武侯祠，在千年古都的脉搏中感受历史厚重',
    tags: ['金沙遗址', '武侯祠', '杜甫草堂', '宽窄巷子'],
    detail: [
      { time: '09:00', place: '金沙遗址博物馆', note: '探秘太阳神鸟，感受古蜀文明' },
      { time: '12:00', place: '宽窄巷子', note: '午餐 + 漫步，体验老成都生活' },
      { time: '14:00', place: '武侯祠', note: '三国文化圣地，看诸葛亮传奇一生' },
      { time: '17:00', place: '锦里古街', note: '逛吃逛吃，采购特色伴手礼' },
    ]
  },
  {
    icon: '🌿', title: '自然成都 · 生态之旅', desc: '从大熊猫基地到青城山，在青山绿水中放松身心',
    tags: ['大熊猫基地', '青城山', '都江堰'],
    detail: [
      { time: '08:30', place: '大熊猫繁育基地', note: '看国宝萌宠，最佳观赏时段' },
      { time: '12:30', place: '都江堰景区', note: '午餐后游览古代水利奇迹' },
      { time: '15:30', place: '青城山', note: '后山徒步，感受青城天下幽' },
    ]
  },
  {
    icon: '🍜', title: '美食成都 · 味蕾狂欢', desc: '从锦里到建设路，用舌尖丈量这座美食之都',
    tags: ['锦里', '建设路小吃', '玉林路', '火锅'],
    detail: [
      { time: '10:00', place: '玉林路菜市场', note: '本地人逛的菜市，吃地道早餐' },
      { time: '12:30', place: '建设路小吃街', note: '网红小吃集中营，空胃来战' },
      { time: '15:00', place: '人民公园', note: '鹤鸣茶社喝盖碗茶，体验慢生活' },
      { time: '19:00', place: '火锅店', note: '来成都怎么能不吃一顿火锅！' },
    ]
  },
]

// ===== 交互：亮点展开 =====
const expandedFeature = ref(null)
function toggleFeature(i) {
  expandedFeature.value = expandedFeature.value === i ? null : i
}

// ===== 交互：攻略展开 =====
const expandedGuide = ref(null)
function toggleGuide(i) {
  expandedGuide.value = expandedGuide.value === i ? null : i
}

// ===== 路由 =====
function goToLogin() { router.push('/login') }
function scrollToFeatures() { document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' }) }

// ===== 粒子系统（温暖光点风格）=====
let particleRaf = null
function initParticles() {
  const canvas = canvasRef.value
  if (!canvas) return () => {}
  const ctx = canvas.getContext('2d')
  let w = canvas.width = window.innerWidth
  let h = canvas.height = window.innerHeight
  const COUNT = 50
  const particles = Array.from({ length: COUNT }, () => ({
    x: Math.random() * w, y: Math.random() * h,
    vx: (Math.random() - 0.5) * 0.15, vy: (Math.random() - 0.5) * 0.15,
    r: Math.random() * 2.5 + 1,
    alpha: Math.random() * 0.3 + 0.1,
    pulse: Math.random() * Math.PI * 2
  }))

  function draw(t) {
    ctx.clearRect(0, 0, w, h)
    for (const p of particles) {
      p.x += p.vx; p.y += p.vy
      if (p.x < -20) p.x = w + 20; if (p.x > w + 20) p.x = -20
      if (p.y < -20) p.y = h + 20; if (p.y > h + 20) p.y = -20
      p.pulse += 0.015
      const alpha = p.alpha * (0.6 + 0.4 * Math.sin(p.pulse))
      ctx.beginPath()
      ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2)
      ctx.fillStyle = `rgba(232, 93, 58, ${alpha})`
      ctx.fill()
    }
    particleRaf = requestAnimationFrame(draw)
  }
  particleRaf = requestAnimationFrame(draw)

  function onResize() { w = canvas.width = window.innerWidth; h = canvas.height = window.innerHeight }
  window.addEventListener('resize', onResize)

  return () => {
    window.removeEventListener('resize', onResize)
    if (particleRaf) cancelAnimationFrame(particleRaf)
  }
}

// ===== 轮播 =====
const currentSlide = ref(0)
const slideWidth = 380
const carouselOffset = ref(0)
const progressPct = ref(0)
let autoTimer = null
let dragStartX = 0
let dragOffset = 0
let isDragging = false
let progressTimer = null

function nextSlide() {
  currentSlide.value = (currentSlide.value + 1) % carouselSpots.length
  carouselOffset.value = -(currentSlide.value * slideWidth)
  resetProgress()
}
function prevSlide() {
  currentSlide.value = (currentSlide.value - 1 + carouselSpots.length) % carouselSpots.length
  carouselOffset.value = -(currentSlide.value * slideWidth)
  resetProgress()
}
function goToSlide(i) {
  currentSlide.value = i
  carouselOffset.value = -(i * slideWidth)
  resetProgress()
}
function startAutoPlay() { stopAutoPlay(); autoTimer = setInterval(nextSlide, 3500); startProgress() }
function stopAutoPlay() { if (autoTimer) { clearInterval(autoTimer); autoTimer = null }; stopProgress() }

// 进度条动画
function startProgress() { progressPct.value = 0; stopProgress(); progressTimer = setInterval(() => { progressPct.value = Math.min(100, progressPct.value + 100 / 60) }, 50) }
function stopProgress() { if (progressTimer) { clearInterval(progressTimer); progressTimer = null } }
function resetProgress() { progressPct.value = 0 }

// 拖拽滑动
function onCarouselDown(e) {
  isDragging = true; dragStartX = e.clientX; dragOffset = 0
  stopAutoPlay()
}
function onCarouselMove(e) {
  if (!isDragging) return
  dragOffset = e.clientX - dragStartX
  carouselOffset.value = -(currentSlide.value * slideWidth) + dragOffset
}
function onCarouselUp() {
  if (!isDragging) return
  isDragging = false
  if (Math.abs(dragOffset) > 60) {
    if (dragOffset < 0) nextSlide()
    else prevSlide()
  } else {
    carouselOffset.value = -(currentSlide.value * slideWidth)
  }
  startAutoPlay()
}

// ===== GSAP =====
onMounted(() => {
  const cleanup = initParticles()
  startAutoPlay()

  const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })
  tl.fromTo(navRef.value, { y: -20, opacity: 0 }, { y: 0, opacity: 1, duration: 0.6 }, 0)
  tl.fromTo(badgeRef.value, { y: -10, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5 }, 0.3)
  tl.fromTo(titleRef.value?.children, { y: 25, opacity: 0 }, { y: 0, opacity: 1, duration: 0.7, stagger: 0.15 }, 0.5)
  tl.fromTo(descRef.value, { y: 15, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5 }, 0.9)
  tl.fromTo(actionsRef.value?.children, { y: 15, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5, stagger: 0.1 }, 1.1)
  tl.fromTo(statsRef.value?.children, { y: 12, opacity: 0 }, { y: 0, opacity: 1, duration: 0.4, stagger: 0.08 }, 1.4)

  // 浮动符号动画
  const symbols = [fs1, fs2, fs3, fs4, fs5]
  symbols.forEach((s, i) => {
    if (!s.value) return
    gsap.fromTo(s.value, { opacity: 0, scale: 0, y: 20 }, { opacity: 0.4, scale: 1, y: 0, duration: 1, delay: 0.8 + i * 0.15, ease: 'back.out(2)' })
    gsap.to(s.value, { y: -8 - Math.random() * 10, x: (Math.random() - 0.5) * 8, duration: 2.5 + Math.random() * 2, repeat: -1, yoyo: true, ease: 'sine.inOut', delay: i * 0.2 })
  })

  // Scroll 触发
  const st = []
  function scrollTrigger(elements, trigger) {
    if (!elements?.length) return
    st.push(gsap.fromTo(elements,
      { y: 30, opacity: 0 },
      { y: 0, opacity: 1, duration: 0.5, stagger: 0.1, ease: 'power3.out',
        scrollTrigger: { trigger, start: 'top 85%', toggleActions: 'play none none none' } }
    ))
  }
  // 区块头部渐入
  function scrollHeader(el) {
    if (!el) return
    st.push(gsap.fromTo(el,
      { y: 20, opacity: 0 },
      { y: 0, opacity: 1, duration: 0.5, ease: 'power3.out',
        scrollTrigger: { trigger: el, start: 'top 88%', toggleActions: 'play none none none' } }
    ))
  }
  scrollHeader(featHeaderRef.value)
  scrollHeader(spotHeaderRef.value)
  scrollHeader(guideHeaderRef.value)
  scrollTrigger(cardRefs.value, featuresRef.value)
  scrollTrigger(guideRefs.value, guideRef.value)

  st.push(gsap.fromTo(ctaRef.value,
    { y: 30, opacity: 0 },
    { y: 0, opacity: 1, duration: 0.6, ease: 'power3.out',
      scrollTrigger: { trigger: ctaRef.value, start: 'top 85%', toggleActions: 'play none none none' } }
  ))

  return () => { cleanup(); stopAutoPlay(); ScrollTrigger.getAll().forEach(t => t.kill()) }
})
</script>

<style scoped>
/* ================================
   全局
   ================================ */
.landing {
  position: relative;
  width: 100%;
  min-height: 100vh;
  background: #FCF7F2;
  color: #2D2D3A;
  font-family: 'Inter', 'Noto Sans SC', 'PingFang SC', sans-serif;
  overflow-x: hidden;
}

.particle-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

/* ================================
   导航栏
   ================================ */
.navbar {
  position: fixed;
  top: 0; left: 0; right: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 40px;
  height: 56px;
  background: rgba(252, 247, 242, .85);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(232, 93, 58, .06);
}
.nav-brand { display: flex; align-items: center; gap: 8px; }
.nav-title { font-size: .82rem; font-weight: 700; color: #E85D3A; letter-spacing: 2px; }
.nav-links { display: flex; gap: 28px; }
.nav-link {
  font-size: .68rem; color: #9CA3AF;
  text-decoration: none; letter-spacing: 2px;
  transition: color .3s;
}
.nav-link:hover { color: #E85D3A; }
.nav-cta {
  padding: 6px 20px; border-radius: 8px;
  border: 1px solid rgba(232, 93, 58, .15);
  background: transparent; color: #E85D3A;
  font-size: .68rem; font-family: inherit; cursor: pointer;
  letter-spacing: 1px; transition: all .3s;
}
.nav-cta:hover { background: rgba(232, 93, 58, .06); border-color: rgba(232, 93, 58, .3); }

/* ================================
   Hero
   ================================ */
.hero {
  position: relative; z-index: 1;
  min-height: 100vh;
  display: flex; align-items: center;
  padding: 80px 60px 60px;
  overflow: hidden;
}

/* 背景装饰圆环 */
.hero-bg-deco {
  position: absolute; left: 50%; top: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  width: 100%; max-width: 700px;
  display: flex; justify-content: center;
}
.hero-ring { width: 100%; height: auto; animation: ring-spin 40s linear infinite; }
@keyframes ring-spin { to { transform: rotate(360deg); } }

/* 浮动符号 */
.float-symbols { position: absolute; inset: 0; pointer-events: none; z-index: 1; }
.fs-item {
  position: absolute;
  font-size: 1.6rem;
  opacity: 0;
}
.fs-item:nth-child(1) { top: 12%; left: 8%; }
.fs-item:nth-child(2) { top: 15%; right: 12%; }
.fs-item:nth-child(3) { bottom: 22%; left: 6%; }
.fs-item:nth-child(4) { top: 38%; right: 5%; }
.fs-item:nth-child(5) { bottom: 14%; right: 8%; }

/* 主内容 */
.hero-content { position: relative; z-index: 2; max-width: 660px; margin: 0 auto; text-align: center; }

.hero-badge {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 14px; border-radius: 20px;
  border: 1px solid rgba(232, 93, 58, .08);
  background: rgba(232, 93, 58, .04);
  font-size: .55rem; color: rgba(232, 93, 58, .55);
  letter-spacing: 3px; margin-bottom: 20px;
}
.badge-dot {
  width: 5px; height: 5px; border-radius: 50%;
  background: #E85D3A;
  animation: dot-pulse 2s ease-in-out infinite;
}
@keyframes dot-pulse {
  0%, 100% { opacity: .3; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.3); }
}

.hero-title {
  margin-bottom: 16px;
}
.ht-line {
  display: block;
  font-size: 3.6rem;
  font-weight: 900;
  line-height: 1.12;
  letter-spacing: 4px;
}
.ht-line.hl {
  background: linear-gradient(135deg, #E85D3A, #D4A854);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-desc {
  font-size: .82rem;
  color: #9CA3AF;
  line-height: 1.8;
  letter-spacing: 1px;
  margin-bottom: 32px;
}
.br-desk { display: block; }
@media (max-width: 640px) { .br-desk { display: none; } }

.hero-actions { display: flex; justify-content: center; gap: 14px; margin-bottom: 40px; }

.btn-primary {
  position: relative; display: flex; align-items: center; gap: 8px;
  padding: 12px 28px; border-radius: 10px; border: none;
  background: linear-gradient(135deg, #E85D3A, #F5A623);
  color: #fff; font-size: .82rem; font-weight: 700;
  font-family: inherit; cursor: pointer; letter-spacing: 1px;
  overflow: hidden;
  transition: all .3s cubic-bezier(0.16,1,0.3,1);
  box-shadow: 0 4px 20px rgba(232,93,58,.2);
}
.btn-primary:hover { transform: translateY(-2px); box-shadow: 0 8px 30px rgba(232,93,58,.3); }
.btn-glow {
  position: absolute; inset: 0;
  background: radial-gradient(ellipse at 50% 100%, rgba(255,255,255,.15), transparent 60%);
  pointer-events: none;
}
.btn-ghost {
  display: flex; align-items: center; gap: 6px;
  padding: 12px 20px; border-radius: 10px;
  border: 1px solid rgba(0,0,0,.06);
  background: transparent; color: #9CA3AF;
  font-size: .72rem; font-family: inherit; cursor: pointer;
  letter-spacing: 1px; transition: all .3s;
}
.btn-ghost:hover { border-color: rgba(232,93,58,.15); color: #E85D3A; }

.hero-stats { display: flex; justify-content: center; align-items: center; gap: 24px; }
.hs-item { display: flex; flex-direction: column; gap: 2px; }
.hs-num { font-size: 1.3rem; font-weight: 800; color: #E85D3A; font-feature-settings: "tnum"; }
.hs-label { font-size: .52rem; color: #9CA3AF; letter-spacing: 2px; }
.hs-divider { width: 1px; height: 28px; background: rgba(0,0,0,.06); }

/* ================================
   区块通用 + 分节器
   ================================ */
.section-block {
  position: relative; z-index: 2;
  scroll-margin-top: 60px;
}
.section-divider {
  display: flex; justify-content: center; align-items: center;
  padding: 0 0 32px; gap: 16px;
}
.section-divider::before,
.section-divider::after {
  content: ''; flex: 1; max-width: 80px;
  height: 1px; background: linear-gradient(90deg, transparent, rgba(232,93,58,.1), transparent);
}
.sd-icon { font-size: .5rem; color: rgba(232,93,58,.15); letter-spacing: 0; }

.sec-header { text-align: center; margin-bottom: 40px; }
.sec-deco {
  display: flex; justify-content: center;
  margin-bottom: 14px; opacity: .5;
}
.sec-tag {
  display: inline-block;
  padding: 3px 12px;
  border-radius: 4px;
  background: rgba(232,93,58,.06);
  color: rgba(232,93,58,.5);
  font-size: .5rem; letter-spacing: 3px;
  margin-bottom: 10px;
}
.sec-title {
  font-family: "Noto Serif SC", "Source Han Serif SC", "Songti SC", serif;
  font-size: 1.8rem; font-weight: 800;
  letter-spacing: 2px; margin-bottom: 8px;
}
.sec-desc { font-size: .72rem; color: #9CA3AF; letter-spacing: 1px; }

/* ================================
   核心亮点
   ================================ */
.features {
  position: relative; z-index: 2;
  padding: 20px 40px 80px;
}
.feature-grid {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px;
  max-width: 900px; margin: 0 auto;
}
.feat-card {
  border-radius: 14px;
  background: #FFF;
  border: 1px solid rgba(0,0,0,.04);
  overflow: hidden;
  cursor: pointer;
  transition: all .4s cubic-bezier(0.16,1,0.3,1);
}
.feat-card:hover { transform: translateY(-4px); box-shadow: 0 16px 40px rgba(232,93,58,.06); }
.feat-card.expanded { transform: translateY(-4px); box-shadow: 0 20px 48px rgba(232,93,58,.1); }

.fc-accent {
  height: 3px;
  width: 100%;
  opacity: .6;
}

.fc-inner {
  padding: 24px 20px 22px;
  text-align: center;
}

.fc-icon-wrap {
  width: 52px; height: 52px;
  border-radius: 14px;
  display: flex; align-items: center; justify-content: center;
  margin: 0 auto 16px;
  transition: transform .4s cubic-bezier(0.34,1.56,0.64,1);
  position: relative;
}
.feat-card:hover .fc-icon-wrap { transform: scale(1.08) translateY(-2px); }

.fc-icon { font-size: 1.5rem; }
.fc-title {
  font-family: "Noto Serif SC", "Source Han Serif SC", serif;
  font-size: .82rem; font-weight: 700;
  margin-bottom: 8px; letter-spacing: 1px;
}
.fc-desc { font-size: .64rem; color: #9CA3AF; line-height: 1.7; transition: all .3s; }
.fc-expand {
  display: inline-block;
  margin-top: 12px;
  font-size: .54rem;
  color: rgba(232,93,58,.3);
  letter-spacing: 2px;
  opacity: 0;
  transition: opacity .3s, color .3s;
}
.feat-card:hover .fc-expand { opacity: 1; color: rgba(232,93,58,.5); }
.feat-card.expanded .fc-expand { opacity: 1; color: #E85D3A; }

/* ================================
   景点轮播
   ================================ */
.carousel-section {
  position: relative; z-index: 2;
  padding: 20px 40px 80px;
}
.carousel-viewport {
  position: relative;
  max-width: 820px;
  margin: 0 auto;
  overflow: hidden;
  padding: 12px 0;
}
.carousel-track {
  display: flex;
  gap: 24px;
  transition: transform .55s cubic-bezier(0.16,1,0.3,1);
  will-change: transform;
  padding: 0 20px;
}
.carousel-card {
  display: flex;
  min-width: 360px;
  height: 210px;
  border-radius: 18px;
  overflow: hidden;
  background: #FFF;
  box-shadow: 0 2px 12px rgba(0,0,0,.03);
  flex-shrink: 0;
  transition: all .5s cubic-bezier(0.16,1,0.3,1);
  opacity: .55;
  transform: scale(.92);
  filter: saturate(.6);
}
.carousel-card.active {
  opacity: 1;
  transform: scale(1);
  filter: saturate(1);
  box-shadow: 0 12px 40px rgba(232,93,58,.06);
}
.carousel-card:not(.active):hover {
  opacity: .7;
  transform: scale(.94);
}
.cc-visual {
  width: 200px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden;
}
.cc-visual::after {
  content: '';
  position: absolute; inset: 0;
  background: linear-gradient(135deg, rgba(255,255,255,.1) 0%, transparent 50%);
  pointer-events: none;
}
.cc-emoji { font-size: 3.6rem; filter: drop-shadow(0 4px 12px rgba(0,0,0,.06)); transition: transform .5s cubic-bezier(0.34,1.56,0.64,1); }
.carousel-card.active:hover .cc-emoji { transform: scale(1.1) rotate(-4deg); }
.cc-rating-badge {
  position: absolute; top: 12px; left: 12px;
  padding: 3px 10px; border-radius: 6px;
  background: rgba(255,255,255,.7);
  backdrop-filter: blur(4px);
  font-size: .55rem; color: #D4A854; font-weight: 600;
}
.cc-body {
  flex: 1; padding: 22px 22px 18px;
  display: flex; flex-direction: column; gap: 6px;
  min-width: 0;
}
.cc-category { font-size: .52rem; font-weight: 600; letter-spacing: 2px; }
.cc-name {
  font-family: "Noto Serif SC", "Source Han Serif SC", serif;
  font-size: 1.05rem; font-weight: 700; letter-spacing: 1px; margin: 0; line-height: 1.3;
}
.cc-desc { font-size: .62rem; color: #9CA3AF; line-height: 1.6; margin: 0; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; line-clamp: 2; overflow: hidden; }
.cc-footer { display: flex; justify-content: space-between; align-items: center; margin-top: auto; padding-top: 6px; }
.cc-dist { font-size: .54rem; color: #9CA3AF; }
.cc-explore { font-size: .56rem; color: #E85D3A; font-weight: 600; letter-spacing: 1px; opacity: 0; transition: opacity .3s; }
.carousel-card.active .cc-explore { opacity: .6; }
.carousel-btn {
  position: absolute; top: 50%; transform: translateY(-50%);
  z-index: 10; width: 36px; height: 36px;
  border-radius: 50%; border: none;
  background: #FFF; color: #9CA3AF; font-size: 1.2rem;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  box-shadow: 0 2px 12px rgba(0,0,0,.06);
  transition: all .3s;
}
.carousel-btn:hover { color: #E85D3A; box-shadow: 0 4px 20px rgba(232,93,58,.1); }
.carousel-btn.prev { left: 4px; }
.carousel-btn.next { right: 4px; }
.carousel-progress {
  max-width: 400px; margin: 16px auto 0;
  height: 2px; background: rgba(0,0,0,.04);
  border-radius: 1px; overflow: hidden;
}
.cp-bar { height: 100%; background: #E85D3A; border-radius: 1px; transition: none; }
.carousel-dots { display: flex; justify-content: center; gap: 6px; margin-top: 12px; }
.carousel-dot {
  width: 5px; height: 5px; border-radius: 50%;
  border: none; background: rgba(0,0,0,.06);
  cursor: pointer; transition: all .4s; padding: 0;
}
.carousel-dot.active { background: #E85D3A; width: 22px; border-radius: 3px; }

/* ================================
   旅行攻略
   ================================ */
.guide-section {
  position: relative; z-index: 2;
  padding: 20px 40px 80px;
}
.guide-grid {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px;
  max-width: 900px; margin: 0 auto;
}
.guide-card {
  display: flex; gap: 16px;
  padding: 20px 20px 16px;
  border-radius: 14px;
  background: #FFF; border: 1px solid rgba(0,0,0,.04);
  transition: all .4s cubic-bezier(0.16,1,0.3,1);
  cursor: pointer;
}
.guide-card:hover { transform: translateY(-3px); box-shadow: 0 12px 32px rgba(232,93,58,.05); }
.guide-card.expanded {
  box-shadow: 0 16px 40px rgba(232,93,58,.08);
  border-color: rgba(232,93,58,.06);
}

.gc-day {
  display: flex; align-items: center; gap: 8px; flex-direction: column;
  cursor: pointer; position: relative; min-width: 32px;
}
.gc-day-num {
  writing-mode: vertical-rl;
  font-size: .52rem; font-weight: 700; color: #E85D3A;
  letter-spacing: 3px; white-space: nowrap;
  padding: 4px 6px;
  border-radius: 6px;
  background: rgba(232,93,58,.04);
}
.gc-day-line { width: 1px; flex: 1; background: rgba(232,93,58,.08); min-height: 20px; }
.gc-arrow {
  color: #D1D5DB;
  transition: transform .4s cubic-bezier(0.34,1.56,0.64,1);
}
.gc-arrow.open { transform: rotate(180deg); color: #E85D3A; }

.gc-body { flex: 1; min-width: 0; }
.gc-icon { font-size: 1.2rem; margin-bottom: 6px; }
.gc-title {
  font-family: "Noto Serif SC", "Source Han Serif SC", serif;
  font-size: .82rem; font-weight: 700; margin-bottom: 4px; letter-spacing: 1px;
}
.gc-desc { font-size: .6rem; color: #9CA3AF; line-height: 1.6; margin-bottom: 8px; }
.gc-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.gc-tags span {
  padding: 2px 8px; border-radius: 4px;
  font-size: .5rem; background: rgba(232,93,58,.04); color: rgba(232,93,58,.55);
  letter-spacing: 1px;
}

/* 展开详情 — 时间线 */
.gc-detail {
  margin-top: 14px; padding-top: 14px;
  border-top: 1px solid rgba(232,93,58,.06);
}
.gc-timeline {
  display: flex; flex-direction: column; gap: 12px;
  padding-left: 4px;
}
.gc-tl-item {
  display: flex; gap: 12px; position: relative;
  align-items: flex-start;
}
.gc-tl-item:not(:last-child)::before {
  content: '';
  position: absolute; left: 3px; top: 12px; bottom: -12px;
  width: 1px; background: rgba(232,93,58,.08);
}
.gc-tl-dot {
  width: 7px; height: 7px; border-radius: 50%;
  background: #E85D3A; flex-shrink: 0;
  margin-top: 3px;
  box-shadow: 0 0 8px rgba(232,93,58,.2);
  position: relative; z-index: 1;
}
.gc-tl-body { display: flex; flex-direction: column; gap: 1px; }
.gc-tl-time {
  font-size: .5rem; font-weight: 700; color: #E85D3A; letter-spacing: 1px;
  background: rgba(232,93,58,.04); padding: 0 6px; border-radius: 3px;
  align-self: flex-start;
}
.gc-tl-place { font-size: .68rem; font-weight: 600; color: #2D2D3A; line-height: 1.4; }
.gc-tl-note { font-size: .55rem; color: #9CA3AF; }

/* Guide expand transition */
.gc-expand-enter-active, .gc-expand-leave-active {
  transition: all .35s ease;
  overflow: hidden;
}
.gc-expand-enter-from, .gc-expand-leave-to {
  opacity: 0; max-height: 0; padding-top: 0; margin-top: 0;
}
.gc-expand-enter-to, .gc-expand-leave-from {
  opacity: 1; max-height: 400px;
}

/* ================================
   CTA
   ================================ */
.cta-section {
  position: relative; z-index: 2;
  padding: 40px 40px 30px;
}
.cta-card {
  position: relative;
  max-width: 760px; margin: 0 auto;
  padding: 48px 40px; border-radius: 20px;
  background: linear-gradient(135deg, rgba(232,93,58,.04), rgba(212,168,84,.04));
  border: 1px solid rgba(232,93,58,.06);
  text-align: center; overflow: hidden;
}
.cta-deco {
  position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 300px; height: 300px;
  background: radial-gradient(circle, rgba(232,93,58,.04) 0%, transparent 70%);
  pointer-events: none;
}
.cta-content { position: relative; z-index: 2; }
.cta-title {
  font-family: "Noto Serif SC", serif;
  font-size: 1.8rem; font-weight: 800;
  letter-spacing: 3px; margin-bottom: 10px;
}
.cta-desc { font-size: .72rem; color: #9CA3AF; margin-bottom: 24px; }
.btn-lg { padding: 14px 36px; font-size: .88rem; margin: 0 auto; }

.cta-footer { text-align: center; padding: 24px 0 0; }
.cf-brand { font-size: .62rem; font-weight: 700; color: rgba(232,93,58,.25); letter-spacing: 4px; margin-bottom: 4px; }
.cf-copy { font-size: .5rem; color: rgba(0,0,0,.12); letter-spacing: 2px; }

/* ================================
   响应式
   ================================ */
@media (max-width: 1024px) {
  .feature-grid { grid-template-columns: repeat(2, 1fr); }
  .guide-grid { grid-template-columns: repeat(2, 1fr); }
  .ht-line { font-size: 2.6rem; }
  .carousel-card { min-width: calc(100vw - 80px); }
  .cc-visual { width: 140px; }
}
@media (max-width: 640px) {
  .feature-grid { grid-template-columns: 1fr; }
  .guide-grid { grid-template-columns: 1fr; }
  .hero { padding: 80px 20px 40px; }
  .ht-line { font-size: 2rem; }
  .nav-links { display: none; }
  .hero-actions { flex-direction: column; align-items: center; }
  .carousel-card { flex-direction: column; height: auto; min-width: calc(100vw - 40px); }
  .cc-visual { width: 100%; height: 120px; }
}

/* ================================
   动画降级
   ================================ */
@media (prefers-reduced-motion: reduce) {
  .hero-ring { animation: none; }
  .badge-dot { animation: none; }
  .fs-item { animation: none !important; }
  .carousel-track { transition: none; }
}
</style>
