<template>
  <div class="explore-page" ref="pageRef">
    <!-- ===== Hero 区域 ===== -->
    <section class="hero" ref="heroRef">
      <div class="hero-deco" aria-hidden="true">
        <svg viewBox="0 0 140 140" class="hero-svg">
          <circle cx="70" cy="70" r="68" fill="none" stroke="rgba(232,93,58,.04)" stroke-width="1.5" />
          <circle cx="70" cy="70" r="52" fill="none" stroke="rgba(232,93,58,.03)" stroke-width="1" />
          <g v-for="i in 12" :key="i">
            <line
              :x1="70 + 30 * Math.cos(i * Math.PI / 6)"
              :y1="70 + 30 * Math.sin(i * Math.PI / 6)"
              :x2="70 + 50 * Math.cos(i * Math.PI / 6)"
              :y2="70 + 50 * Math.sin(i * Math.PI / 6)"
              stroke="rgba(232,93,58,.04)" stroke-width="2" stroke-linecap="round"
            />
          </g>
        </svg>
      </div>

      <div class="hero-inner">
        <div class="hero-eyebrow" ref="eyebrowRef">CHENGDU · 成都</div>

        <h1 class="hero-title" ref="titleRef">
          <span class="ht-line">探索</span>
          <span class="ht-line hl">成都</span>
        </h1>

        <p class="hero-desc" ref="descRef">
          在蜀地的山水与烟火之间，发现每一处值得停留的风景
        </p>

        <div class="hero-stats" ref="statsRef">
          <div class="hero-stat">
            <span class="hs-num">{{ spots.length }}</span>
            <span class="hs-label">精选景点</span>
          </div>
          <div class="hero-stat">
            <span class="hs-num">4</span>
            <span class="hs-label">主题分类</span>
          </div>
          <div class="hero-stat">
            <span class="hs-num">6</span>
            <span class="hs-label">推荐路线</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ===== 分类筛选 ===== -->
    <div class="filter-bar" ref="filterRef">
      <button
        v-for="cat in categories"
        :key="cat.key"
        class="filter-pill"
        :class="{ active: activeCategory === cat.key }"
        @click="activeCategory = cat.key"
      >
        <span class="fp-emoji">{{ cat.emoji }}</span>
        <span>{{ cat.label }}</span>
      </button>
    </div>

    <!-- ===== 加载状态 ===== -->
    <div class="loading-state" v-if="loading">
      <div class="loading-spinner"></div>
      <span>正在探索成都景点...</span>
    </div>

    <!-- ===== 景点卡片列表 ===== -->
    <div class="spot-list" v-else ref="listRef">
      <transition-group name="card-enter">
        <article
          v-for="(spot, index) in filteredSpots"
          :key="spot.id"
          class="spot-card"
          :data-index="index"
          ref="cardRefs"
        >
          <!-- 左侧视觉区 -->
          <div class="sc-visual" :style="{ background: spot.gradient }">
            <span class="sc-emoji">{{ spot.emoji }}</span>
          </div>

          <!-- 右侧内容区 -->
          <div class="sc-body">
            <div class="sc-top">
              <span class="sc-tag" :style="{ '--tag-clr': spot.tagColor }">{{ spot.category }}</span>
              <span class="sc-dist">{{ spot.distance }}</span>
            </div>

            <h3 class="sc-name">{{ spot.name }}</h3>

            <p class="sc-desc">{{ spot.description }}</p>

            <div class="sc-footer">
              <button
                class="sc-like"
                :class="{ liked: spot.liked }"
                @click="toggleLike(spot)"
                :aria-label="spot.liked ? '取消收藏' : '收藏'"
              >
                <svg viewBox="0 0 24 24" class="like-icon" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
                </svg>
                <span class="like-count">{{ formatCount(spot.likes) }}</span>

                <!-- 朱砂印：点赞后出现 -->
                <span class="sc-seal" v-if="spot.showSeal" ref="sealRefs">❤︎</span>
              </button>
            </div>
          </div>
        </article>
      </transition-group>

      <!-- 空状态 -->
      <div class="empty-state" v-if="filteredSpots.length === 0">
        <span class="empty-emoji">🔍</span>
        <p>没有找到该分类的景点</p>
      </div>
    </div>

    <!-- ===== 底部 ===== -->
    <div class="explore-footer" ref="footerRef">
      <div class="footer-divider"></div>
      <div class="footer-text">蓉城漫游 · 让世界看见成都</div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, inject, nextTick } from 'vue'
import gsap from 'gsap'
import { auth } from '../../stores/auth.js'

// ===== 注入父级提供的登录控制 =====
const showLogin = inject('showLogin', () => {})

// ===== 分类定义 =====
const categories = [
  { key: '全部', emoji: '🌟', label: '全部' },
  { key: '自然', emoji: '🌿', label: '自然' },
  { key: '人文', emoji: '🏛️', label: '人文' },
  { key: '美食', emoji: '🍜', label: '美食' }
]

const activeCategory = ref('全部')

// ===== 景点模拟数据 =====
const ALL_SPOTS = [
  {
    id: 1, name: '大熊猫繁育研究基地', category: '自然',
    emoji: '🐼', gradient: 'linear-gradient(135deg, #10B981, #34D399)',
    tagColor: '#10B981', distance: '成华区 · 30min',
    description: '翠竹掩映中与国宝近距离接触，看呆萌熊猫嬉戏玩耍',
    likes: 2341, liked: false, showSeal: false
  },
  {
    id: 2, name: '宽窄巷子', category: '人文',
    emoji: '🏛️', gradient: 'linear-gradient(135deg, #E85D3A, #F5A623)',
    tagColor: '#E85D3A', distance: '青羊区 · 市中心',
    description: '青砖黛瓦间的老成都记忆，茶香袅袅中品味慢生活',
    likes: 1832, liked: false, showSeal: false
  },
  {
    id: 3, name: '青城山', category: '自然',
    emoji: '🌿', gradient: 'linear-gradient(135deg, #0EA5A0, #34D399)',
    tagColor: '#0EA5A0', distance: '都江堰市 · 60min',
    description: '青城天下幽，漫步山林间感受道家文化的清静无为',
    likes: 1521, liked: false, showSeal: false
  },
  {
    id: 4, name: '武侯祠', category: '人文',
    emoji: '⚔️', gradient: 'linear-gradient(135deg, #C43E1C, #E85D3A)',
    tagColor: '#C43E1C', distance: '武侯区 · 20min',
    description: '三国文化圣地，感受诸葛亮"鞠躬尽瘁"的千古情怀',
    likes: 1456, liked: false, showSeal: false
  },
  {
    id: 5, name: '锦里古街', category: '美食',
    emoji: '🏮', gradient: 'linear-gradient(135deg, #D4A854, #F5A623)',
    tagColor: '#D4A854', distance: '武侯区 · 15min',
    description: '灯火阑珊的古街，串串香与糖画交织的成都味道',
    likes: 1378, liked: false, showSeal: false
  },
  {
    id: 6, name: '杜甫草堂', category: '人文',
    emoji: '📜', gradient: 'linear-gradient(135deg, #6B7280, #9CA3AF)',
    tagColor: '#6B7280', distance: '青羊区 · 25min',
    description: '诗圣杜甫的故居，茅屋与竹林间流淌着千年诗韵',
    likes: 1243, liked: false, showSeal: false
  },
  {
    id: 7, name: '都江堰', category: '自然',
    emoji: '🌊', gradient: 'linear-gradient(135deg, #0EA5A0, #06B6D4)',
    tagColor: '#0EA5A0', distance: '都江堰市 · 50min',
    description: '千年水利工程奇迹，看岷江水在此分流灌溉天府',
    likes: 1187, liked: false, showSeal: false
  },
  {
    id: 8, name: '金沙遗址博物馆', category: '人文',
    emoji: '🌟', gradient: 'linear-gradient(135deg, #D4A854, #E8B84B)',
    tagColor: '#D4A854', distance: '金牛区 · 35min',
    description: '太阳神鸟的故乡，探秘古蜀文明的璀璨辉煌',
    likes: 1098, liked: false, showSeal: false
  },
  {
    id: 9, name: '建设路小吃街', category: '美食',
    emoji: '🍜', gradient: 'linear-gradient(135deg, #E85D3A, #F97316)',
    tagColor: '#E85D3A', distance: '成华区 · 20min',
    description: '成都本地人的美食天堂，从早到晚的味蕾狂欢',
    likes: 987, liked: false, showSeal: false
  }
]

// ===== 状态 =====
const spots = reactive([])
const loading = ref(true)
const pageRef = ref(null)
const heroRef = ref(null)
const eyebrowRef = ref(null)
const titleRef = ref(null)
const descRef = ref(null)
const statsRef = ref(null)
const filterRef = ref(null)
const listRef = ref(null)
const footerRef = ref(null)
const cardRefs = ref([])
const sealRefs = ref([])

// ===== 计算筛选后的景点 =====
const filteredSpots = computed(() => {
  if (activeCategory.value === '全部') return spots
  return spots.filter(s => s.category === activeCategory.value)
})

// ===== 格式化计数 =====
function formatCount(n) {
  if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k'
  return String(n)
}

// ===== 模拟 API 获取数据 =====
async function fetchSpots() {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(ALL_SPOTS.map(s => ({ ...s })))
    }, 500)
  })
}

// ===== 点赞/取消点赞 =====
function toggleLike(spot) {
  // 未登录 -> 弹出登录
  if (!auth.isLoggedIn) {
    showLogin()
    return
  }

  spot.liked = !spot.liked
  spot.likes += spot.liked ? 1 : -1

  if (spot.liked) {
    // 显示朱砂印
    spot.showSeal = true
    // 3秒后隐藏
    setTimeout(() => { spot.showSeal = false }, 2500)
  }
}

// ===== 初始化 =====
onMounted(async () => {
  // 加载数据
  const data = await fetchSpots()
  spots.splice(0, spots.length, ...data)
  loading.value = false

  await nextTick()

  // ===== GSAP 入场动画序列 =====
  const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })

  // Hero 文字渐入
  tl.fromTo(eyebrowRef.value, { y: -8, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5 }, 0)
  tl.fromTo(titleRef.value?.children, { y: 24, opacity: 0 }, { y: 0, opacity: 1, duration: 0.7, stagger: 0.15 }, 0.15)
  tl.fromTo(descRef.value, { y: 12, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5 }, 0.5)
  tl.fromTo(statsRef.value?.children, { y: 10, opacity: 0 }, { y: 0, opacity: 1, duration: 0.4, stagger: 0.08 }, 0.7)
  tl.fromTo(filterRef.value?.children, { y: 10, opacity: 0 }, { y: 0, opacity: 1, duration: 0.35, stagger: 0.05 }, 0.95)

  // 卡片入场 (stagger)
  if (cardRefs.value?.length) {
    tl.fromTo(
      cardRefs.value,
      { y: 30, opacity: 0 },
      { y: 0, opacity: 1, duration: 0.5, stagger: 0.06, ease: 'power3.out' },
      1.1
    )
  }

  // 底部
  tl.fromTo(footerRef.value, { opacity: 0 }, { opacity: 1, duration: 0.4 }, 1.6)
})
</script>

<style scoped>
/* ================================
   页面容器
   ================================ */
.explore-page {
  height: 100%;
  overflow-y: auto;
  padding: 0 16px 24px;
  background: #FCF7F2;
}

.explore-page::-webkit-scrollbar { width: 3px; }
.explore-page::-webkit-scrollbar-thumb { background: rgba(232, 93, 58, .12); border-radius: 2px; }

/* ================================
   Hero 区域
   ================================ */
.hero {
  position: relative;
  padding: 32px 0 20px;
  overflow: hidden;
}

.hero-deco {
  position: absolute;
  right: -10px;
  top: -20px;
  width: 120px;
  height: 120px;
  pointer-events: none;
  opacity: 0.6;
}

.hero-svg { width: 100%; height: 100%; }

.hero-inner {
  position: relative;
  z-index: 2;
}

.hero-eyebrow {
  font-size: .55rem;
  letter-spacing: 4px;
  color: rgba(232, 93, 58, .4);
  margin-bottom: 8px;
  font-weight: 500;
}

/* 超大衬线标题 —— 设计风险：极端比例反差 */
.hero-title {
  font-family: "Noto Serif SC", "Source Han Serif SC", "Songti SC", serif;
  font-size: 2.8rem;
  font-weight: 900;
  line-height: 1.08;
  color: #2D2D3A;
  margin-bottom: 10px;
  letter-spacing: 4px;
}

.hero-title .ht-line {
  display: block;
}

.hero-title .hl {
  background: linear-gradient(135deg, #E85D3A, #D4A854);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-desc {
  font-size: .7rem;
  color: #9CA3AF;
  line-height: 1.6;
  max-width: 280px;
  letter-spacing: .5px;
  margin-bottom: 18px;
}

.hero-stats {
  display: flex;
  gap: 24px;
}

.hero-stat {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.hs-num {
  font-size: 1.1rem;
  font-weight: 700;
  color: #2D2D3A;
  line-height: 1;
  font-feature-settings: "tnum";
}

.hs-label {
  font-size: .52rem;
  color: #9CA3AF;
  letter-spacing: 1px;
}

/* ================================
   筛选栏
   ================================ */
.filter-bar {
  display: flex;
  gap: 6px;
  margin-bottom: 16px;
  padding-bottom: 2px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.filter-bar::-webkit-scrollbar { display: none; }

.filter-pill {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid rgba(0, 0, 0, .05);
  background: rgba(255, 255, 255, .6);
  color: #6B7280;
  font-size: .68rem;
  font-family: inherit;
  cursor: pointer;
  white-space: nowrap;
  transition: all .3s ease;
}

.filter-pill:hover {
  background: rgba(255, 255, 255, .9);
  border-color: rgba(232, 93, 58, .12);
  color: #2D2D3A;
}

.filter-pill.active {
  background: linear-gradient(135deg, rgba(232, 93, 58, .12), rgba(212, 168, 84, .12));
  border-color: rgba(232, 93, 58, .2);
  color: #E85D3A;
  font-weight: 600;
}

.fp-emoji { font-size: .85rem; line-height: 1; }

/* ================================
   加载状态
   ================================ */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 40px 0;
  color: #9CA3AF;
  font-size: .72rem;
}

.loading-spinner {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid rgba(232, 93, 58, .1);
  border-top-color: #E85D3A;
  animation: spin .7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ================================
   景点卡片
   ================================ */
.spot-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.spot-card {
  display: flex;
  gap: 0;
  border-radius: 14px;
  background: #FFFFFF;
  border: 1px solid rgba(0, 0, 0, .04);
  overflow: hidden;
  transition: transform .35s cubic-bezier(0.16, 1, 0.3, 1),
              box-shadow .35s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, .02);
}

.spot-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 28px rgba(0, 0, 0, .04);
}

/* 左侧视觉色块 */
.sc-visual {
  width: 84px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.sc-emoji {
  font-size: 2rem;
  filter: drop-shadow(0 2px 6px rgba(0, 0, 0, .08));
}

/* 右侧内容 */
.sc-body {
  flex: 1;
  padding: 13px 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.sc-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.sc-tag {
  padding: 1px 8px;
  border-radius: 4px;
  font-size: .52rem;
  font-weight: 500;
  color: var(--tag-clr, #E85D3A);
  background: color-mix(in srgb, var(--tag-clr, #E85D3A) 10%, transparent);
  letter-spacing: 1px;
}

.sc-dist {
  font-size: .52rem;
  color: #9CA3AF;
}

.sc-name {
  font-family: "Noto Serif SC", "Source Han Serif SC", "Songti SC", serif;
  font-size: .88rem;
  font-weight: 700;
  color: #2D2D3A;
  line-height: 1.3;
  letter-spacing: 1px;
  margin: 0;
}

.sc-desc {
  font-size: .64rem;
  color: #6B7280;
  line-height: 1.5;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.sc-footer {
  display: flex;
  align-items: center;
  margin-top: auto;
  padding-top: 4px;
}

/* === 点赞按钮 === */
.sc-like {
  position: relative;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 6px;
  border: none;
  background: rgba(232, 93, 58, .04);
  color: #9CA3AF;
  font-size: .62rem;
  cursor: pointer;
  font-family: inherit;
  transition: all .3s ease;
}

.sc-like:hover {
  background: rgba(232, 93, 58, .08);
  color: #E85D3A;
}

.sc-like.liked {
  background: rgba(232, 93, 58, .1);
  color: #E85D3A;
}

.like-icon {
  transition: transform .3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.sc-like:hover .like-icon {
  transform: scale(1.2);
}

.sc-like.liked .like-icon {
  fill: #E85D3A;
}

.like-count {
  font-feature-settings: "tnum";
}

/* === 朱砂印 (点赞后出现) === */
.sc-seal {
  position: absolute;
  top: -28px;
  right: -4px;
  width: 22px;
  height: 22px;
  background: #E85D3A;
  color: #fff;
  font-size: .55rem;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  box-shadow: 0 2px 8px rgba(232, 93, 58, .35);
  animation: seal-stamp .5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
  pointer-events: none;
}

@keyframes seal-stamp {
  0% {
    transform: scale(0) rotate(-15deg);
    opacity: 0;
  }
  40% {
    transform: scale(1.35) rotate(5deg);
    opacity: 1;
  }
  70% {
    transform: scale(0.9) rotate(-2deg);
  }
  100% {
    transform: scale(1) rotate(0deg);
    opacity: 1;
  }
}

/* ================================
   空状态
   ================================ */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 40px 0;
  color: #9CA3AF;
  font-size: .72rem;
}

.empty-emoji { font-size: 2rem; }

/* ================================
   底部
   ================================ */
.explore-footer {
  text-align: center;
  padding: 28px 0 8px;
}

.footer-divider {
  width: 30px;
  height: 1px;
  background: rgba(0, 0, 0, .06);
  margin: 0 auto 12px;
}

.footer-text {
  font-size: .55rem;
  color: rgba(0, 0, 0, .15);
  letter-spacing: 4px;
}

/* ================================
   卡片 enter 过渡 (配合 transition-group)
   ================================ */
.card-enter-enter-active {
  transition: all .4s ease;
}

.card-enter-enter-from {
  opacity: 0;
  transform: translateY(16px);
}

/* ================================
   响应式折衷：极小屏
   ================================ */
@media (max-width: 360px) {
  .hero-title { font-size: 2.2rem; }
  .sc-visual { width: 68px; }
  .sc-emoji { font-size: 1.6rem; }
}

/* ================================
   prefers-reduced-motion
   ================================ */
@media (prefers-reduced-motion: reduce) {
  .sc-seal {
    animation: none;
    transform: scale(1);
    opacity: 1;
  }
  .sc-like:hover .like-icon { transform: none; }
  .spot-card:hover { transform: none; }
}
</style>
