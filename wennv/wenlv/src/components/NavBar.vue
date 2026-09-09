<script setup lang="ts">
/**
 * NavBar.vue — 成都文旅导航栏（页眉）
 * 布局：左Logo + 中锚点导航 + 右元素聚合
 *   左：蜀韵·成都 Logo + 名称
 *   中：首页 / 景点 / 文化 / 非遗（滚动到对应区块）
 *   右：天气按钮 / 语言切换 / 登录 / 注册（右对齐聚合）
 */
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useWeatherStore } from '@/stores/weather'
import { storeToRefs } from 'pinia'
import WeatherTrigger from './WeatherTrigger.vue'
import WeatherPanel from './WeatherPanel.vue'
import LanguageSwitch from './LanguageSwitch.vue'

const navItems = [
  { key: 'home', href: '' },
  { key: 'spots', href: '/#explore' },
  { key: 'culture', href: '/#culture-cards' },
  { key: 'heritage', href: '/#culture-heritage' },
] as const


const router = useRouter()
const route = useRoute()
const langStore = useLanguageStore()
const weatherStore = useWeatherStore()
const { open } = storeToRefs(weatherStore)

const isScrolled = ref(false)

/* 天气下拉：触发点 ref + 面板 ref（面板 Teleport 到 body，用暴露的 getElement 判点击范围） */
const triggerRef = ref<InstanceType<typeof WeatherTrigger> | null>(null)
const panelRef = ref<InstanceType<typeof WeatherPanel> | null>(null)

/** 点击外部收起：pointerdown 先于 click 触发 */
function onGlobalPointerDown(e: PointerEvent) {
  if (!open.value) return
  const target = e.target as Node
  const triggerEl = triggerRef.value?.getElement?.() ?? null
  const panelEl = panelRef.value?.getElement?.() ?? null
  if (triggerEl?.contains(target)) return // 点触发点：交给 click toggle
  if (panelEl?.contains(target)) return   // 点面板内部：不收起
  weatherStore.closeDropdown()
}

watch(open, (val) => {
  if (val) window.addEventListener('pointerdown', onGlobalPointerDown)
  else window.removeEventListener('pointerdown', onGlobalPointerDown)
})

function navigate(path: string) {
  if (path.startsWith('/#')) {
    const hash = path.slice(2)
    if (route.path === '/') {
      document.getElementById(hash)?.scrollIntoView({ behavior: 'smooth' })
    } else {
      router.push('/')
      setTimeout(() => document.getElementById(hash)?.scrollIntoView({ behavior: 'smooth' }), 300)
    }
  } else if (path === '/' || path === '') {
    // 首页：已在首页则平滑回到顶部，否则跳转
    if (route.path === '/') {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    } else {
      router.push('/')
    }
  } else {
    router.push(path)
  }
}

function goToAuth(path: string) {
  router.push(path)
}

/* 当前激活的锚点导航（仅首页有效），用于高亮 */
const activeAnchor = ref('')

function updateActiveAnchor() {
  if (route.path !== '/') {
    activeAnchor.value = ''
    return
  }
  const ids = navItems.map((item) => item.href.slice(2)).filter(Boolean)
  let current = ''
  for (const id of ids) {
    const el = document.getElementById(id)
    if (el && el.getBoundingClientRect().top <= 120) current = id
  }
  activeAnchor.value = current
}

onMounted(() => {
  // 首次拉取天气，NavBar 无需展开即可见温度区间
  weatherStore.fetchWeather()

  const onScroll = () => {
    isScrolled.value = window.scrollY > 40
    updateActiveAnchor()
  }
  window.addEventListener('scroll', onScroll, { passive: true })
  updateActiveAnchor()
  watch(route, () => {
    if (route.path === '/') isScrolled.value = window.scrollY > 40
    updateActiveAnchor()
  })
})

onUnmounted(() => {
  window.removeEventListener('pointerdown', onGlobalPointerDown)
})
</script>

<template>
  <header
    class="navbar"
    :class="{ 'navbar--scrolled': isScrolled }"
  >
      <!-- ── 左：Logo + 名称 ── -->
      <div class="navbar__logo" @click="navigate('/')">
        <svg class="navbar__logo-mark" width="26" height="26" viewBox="0 0 26 26" fill="none">
          <rect x="6" y="6" width="14" height="14" transform="rotate(45 13 13)" stroke="currentColor" stroke-width="1.2"/>
          <rect x="10.5" y="10.5" width="5" height="5" transform="rotate(45 13 13)" fill="currentColor"/>
        </svg>
        <div class="navbar__logo-text">
          <span class="navbar__logo-zh">蜀韵·成都</span>
          <span class="navbar__logo-en">Shu·Chengdu</span>
        </div>
      </div>

      <!-- ── 中：锚点导航（填补页眉中部空白） ── -->
      <nav class="navbar__nav" :aria-label="langStore.t('nav.ariaNav')">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="navbar__nav-link"
          :class="{ 'navbar__nav-link--active': activeAnchor === item.href.slice(2) }"
          @click="navigate(item.href)"
        >
          {{ langStore.t(`nav.${item.key}`) }}
        </button>
      </nav>

      <!-- ── 右：元素聚合（右对齐） ── -->
      <div class="navbar__right">
        <!-- 天气：锚点按钮（含温度区间） + 下拉面板 -->
        <WeatherTrigger ref="triggerRef" />
        <WeatherPanel v-if="open" ref="panelRef" :anchor="triggerRef" />

        <!-- 语言切换：分段式 中/EN（共用组件） -->
        <LanguageSwitch />

        <span class="navbar__divider" aria-hidden="true" />

        <!-- 登录 -->
        <button class="navbar__auth-btn navbar__auth-btn--login" @click="goToAuth('/login')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
            <circle cx="12" cy="8" r="3.5" />
            <path d="M5 20c1.6-3.6 4.1-5 7-5s5.4 1.4 7 5" />
          </svg>
          {{ langStore.t('nav.login') }}
        </button>

        <!-- 注册 -->
        <button class="navbar__auth-btn navbar__auth-btn--register" @click="goToAuth('/register')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
            <circle cx="10" cy="8" r="3.5" />
            <path d="M3.5 20c1.6-3.6 4-5 6.5-5 1.7 0 3.2.5 4.4 1.4" />
            <path d="M18.5 8.5v6M15.5 11.5h6" />
          </svg>
          {{ langStore.t('nav.register') }}
        </button>
      </div>

    <!-- 底部分割金线 -->
    <div class="navbar__underline" />
  </header>
</template>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  height: var(--nav-height);
  padding: 0 var(--space-8); /* 替代原 container-wide 的内边距 */
  display: flex;
  align-items: center;
  justify-content: space-between; /* 左Logo右元素：logo贴左，right贴右 */
  background: transparent;
  transition: background var(--transition-base), box-shadow var(--transition-base);
}

@media (max-width: 768px) {
  .navbar {
    padding: 0 var(--space-4);
  }
}

.navbar--scrolled {
  background: var(--color-nav-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  box-shadow: 0 1px 0 var(--color-border);
}

/* ── 左：Logo ── */
.navbar__logo {
  margin-left: 20px;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  cursor: pointer;
  user-select: none;
}

.navbar__logo-mark {
  color: var(--color-gold);
  flex-shrink: 0;
}

.navbar__logo-text {
  display: flex;
  flex-direction: column;
}

.navbar__logo-zh {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
  line-height: 1.2;
}

.navbar__logo-en {
  font-family: var(--font-en-display);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wider);
  text-transform: uppercase;
}

/* ── 中：锚点导航（绝对定位居中，不随两侧宽度偏移） ── */
.navbar__nav {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  align-items: center;
  gap: var(--space-2);
  white-space: nowrap;
}

.navbar__nav-link {
  position: relative;
  font-family: var(--font-display);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  transition: color var(--transition-fast), background var(--transition-fast);
}

.navbar__nav-link:hover {
  color: var(--color-gold-dark);
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}

/* 激活锚点：下方金色短横线指示 */
.navbar__nav-link--active {
  color: var(--color-gold-dark);
  font-weight: 600;
}

.navbar__nav-link--active::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 2px;
  transform: translateX(-50%);
  width: 16px;
  height: 2px;
  border-radius: var(--radius-full);
  background: var(--color-gold);
}

/* ── 右：元素聚合 ── */
.navbar__right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
}

/* 竖分隔线：区分「工具区」与「账号区」 */
.navbar__divider {
  width: 1px;
  height: 18px;
  background: var(--color-border);
}

/* ── 账号按钮：登录 / 注册 ── */
.navbar__auth-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  transition: all var(--transition-fast);
  letter-spacing: var(--tracking-wide);
}

.navbar__auth-btn svg {
  flex-shrink: 0;
}

.navbar__auth-btn--login {
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
  background: transparent;
}

.navbar__auth-btn--login:hover {
  border-color: var(--color-gold);
  color: var(--color-gold-dark);
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}

/* 注册：实心主按钮，页眉中最强的行动点 */
.navbar__auth-btn--register {
  background: var(--color-gold);
  color: var(--color-bg);
  font-weight: 500;
  border: 1px solid transparent;
}

.navbar__auth-btn--register:hover {
  background: var(--color-gold-light);
  box-shadow: 0 0 20px var(--color-gold-glow);
  transform: translateY(-1px);
}

/* 底部分割金线 */
.navbar__underline {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--color-gold), transparent);
  transition: width var(--transition-slow);
}

.navbar--scrolled .navbar__underline {
  width: 60%;
}

/* Mobile */
@media (max-width: 1024px) {
  .navbar__nav {
    display: none;
  }
}

@media (max-width: 768px) {
  .navbar__auth-btn {
    display: none;
  }
  .navbar__right {
    gap: var(--space-2);
  }
  .navbar__divider {
    display: none;
  }
}
</style>
