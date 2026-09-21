<script setup lang="ts">
/**
 * NavBar.vue — 公开顶栏
 * 左：Logo + 天气图标 | 中：首页/探索/美食/路线/收藏 | 右：语言 + 登录或我的
 * AI 行程(/trip)不再单独占导航位，由「路线」页导流进入，/trip 下「路线」保持高亮
 */
import { ref, watch, onMounted, onUnmounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useWeatherStore } from '@/stores/weather'
import { useUserStore } from '@/stores/user'
import { hasToken, getRefreshToken, clearTokens } from '@/utils/token'
import { logout as apiLogout } from '@/api/auth'
import { storeToRefs } from 'pinia'
import WeatherTrigger from './WeatherTrigger.vue'
import LanguageSwitch from './LanguageSwitch.vue'

const navItems = [
  { key: 'home', path: '/' },
  { key: 'explore', path: '/home/explore' },
  { key: 'food', path: '/home/food' },
  { key: 'routes', path: '/home/routes' },
  { key: 'favorites', path: '/home/favorites' },
  { key: 'passport', path: '/passport' },
  { key: 'guide', path: '/guide' },
] as const

const router = useRouter()
const route = useRoute()
const langStore = useLanguageStore()
const weatherStore = useWeatherStore()
const userStore = useUserStore()
const { open } = storeToRefs(weatherStore)
const { profile } = storeToRefs(userStore)

const isScrolled = ref(false)
const loggedIn = ref(hasToken())
const mineOpen = ref(false)
const mineWrapRef = ref<HTMLDivElement | null>(null)

function goToAuth(path: string) {
  router.push(path)
}

function handleLogout() {
  mineOpen.value = false
  const refresh = getRefreshToken()
  if (refresh) apiLogout(refresh).catch(() => {})
  clearTokens()
  userStore.resetAll()
  loggedIn.value = false
  if (route.path === '/') window.scrollTo({ top: 0, behavior: 'smooth' })
  else router.push('/')
}

function handleMineClick() {
  mineOpen.value = !mineOpen.value
}

function goMinePage() {
  mineOpen.value = false
  router.push('/home/profile')
}

function onDocClick(e: MouseEvent) {
  if (!mineWrapRef.value) return
  if (!mineWrapRef.value.contains(e.target as Node)) mineOpen.value = false
}

const triggerRef = ref<InstanceType<typeof WeatherTrigger> | null>(null)

function onGlobalPointerDown(e: PointerEvent) {
  if (!open.value) return
  const target = e.target as Node
  const triggerEl = triggerRef.value?.getElement?.() ?? null
  if (triggerEl?.contains(target)) return
  weatherStore.closeDropdown()
}

watch(open, (val) => {
  if (val) window.addEventListener('pointerdown', onGlobalPointerDown)
  else window.removeEventListener('pointerdown', onGlobalPointerDown)
})

function navigate(path: string) {
  if (path === '/') {
    // 已登录用户点「首页/品牌」应回到登录态主站首页,而非游客落地页(如 /guide、/passport 等页面上的导航)
    const homePath = loggedIn.value ? '/home/index' : '/'
    if (route.path === homePath) window.scrollTo({ top: 0, behavior: 'smooth' })
    else router.push(homePath)
    return
  }
  router.push(path)
}

function isNavActive(path: string) {
  if (path === '/') return route.path === '/'
  // AI 行程(/trip)归入「路线」导航项：在该模块下路线保持高亮
  if (path === '/home/routes') return route.path.startsWith('/home/routes') || route.path.startsWith('/trip')
  return route.path === path || route.path.startsWith(`${path}/`)
}

onMounted(() => {
  weatherStore.fetchWeather()
  document.addEventListener('click', onDocClick)
  const onScroll = () => {
    isScrolled.value = window.scrollY > 40
  }
  window.addEventListener('scroll', onScroll, { passive: true })
  watch(route, () => {
    if (route.path === '/') isScrolled.value = window.scrollY > 40
    loggedIn.value = hasToken()
    mineOpen.value = false
  })
})

onUnmounted(() => {
  window.removeEventListener('pointerdown', onGlobalPointerDown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick)
})
</script>

<template>
  <header
    class="navbar"
    :class="{ 'navbar--scrolled': isScrolled }"
  >
      <div class="navbar__left">
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

        <WeatherTrigger ref="triggerRef" />
      </div>

      <nav class="navbar__nav" :aria-label="langStore.t('nav.ariaNav')">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="navbar__nav-link"
          :class="{ 'navbar__nav-link--active': isNavActive(item.path) }"
          @click="navigate(item.path)"
        >
          {{ langStore.t(`nav.${item.key}`) }}
        </button>
      </nav>

      <div class="navbar__right">
        <LanguageSwitch />

        <span class="navbar__divider" aria-hidden="true" />

        <template v-if="!loggedIn">
          <button class="navbar__auth-btn navbar__auth-btn--login" @click="goToAuth('/login')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
              <circle cx="12" cy="8" r="3.5" />
              <path d="M5 20c1.6-3.6 4.1-5 7-5s5.4 1.4 7 5" />
            </svg>
            {{ langStore.t('nav.login') }}
          </button>
        </template>

        <div v-else ref="mineWrapRef" class="navbar__mine-wrap">
          <button
            type="button"
            class="navbar__auth-btn navbar__auth-btn--mine"
            :class="{
              'navbar__auth-btn--open': mineOpen,
              'navbar__auth-btn--active': route.path.startsWith('/home/profile'),
            }"
            :aria-expanded="mineOpen"
            aria-haspopup="menu"
            @click.stop="handleMineClick"
          >
            <span v-if="profile.avatar" class="navbar__avatar"><img :src="profile.avatar" alt="" /></span>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
              <circle cx="12" cy="8" r="3.5" />
              <path d="M5 20c1.6-3.6 4.1-5 7-5s5.4 1.4 7 5" />
            </svg>
            {{ langStore.t('nav.mine') }}
            <svg class="navbar__chev" :class="{ 'navbar__chev--open': mineOpen }" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M6 9l6 6 6-6" /></svg>
          </button>

          <Transition name="mine-drop">
            <div v-if="mineOpen" class="navbar__mine-menu" role="menu">
              <button type="button" role="menuitem" class="navbar__mine-item" @click="goMinePage">
                {{ langStore.t('nav.mine') }}
              </button>
              <button type="button" role="menuitem" class="navbar__mine-item navbar__mine-item--logout" @click="handleLogout">
                {{ langStore.t('mineMenu.logout') }}
              </button>
            </div>
          </Transition>
        </div>
      </div>

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

/* ── 左：Logo + 天气 ── */
.navbar__left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
  margin-left: 20px;
  z-index: 1;
}

.navbar__logo {
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
  flex-shrink: 0;
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
  justify-content: center;
  gap: var(--space-2);
  /* 预留最长文案（进入主页 / Register / ログアウト）空间，切语言不挤动布局 */
  width: 8.75rem;
  flex-shrink: 0;
  box-sizing: border-box;
  font-size: var(--text-sm);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-full);
  transition: color var(--transition-fast), background var(--transition-fast),
    border-color var(--transition-fast), box-shadow var(--transition-fast),
    transform var(--transition-fast);
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-primary);
  background: color-mix(in srgb, var(--color-surface) 72%, transparent);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1px solid color-mix(in srgb, var(--color-border) 55%, transparent);
  box-shadow: var(--shadow-sm);
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.navbar__auth-btn svg {
  flex-shrink: 0;
}

.navbar__auth-btn--login:hover,
.navbar__auth-btn--mine:hover,
.navbar__auth-btn--open,
.navbar__auth-btn--active {
  background: color-mix(in srgb, var(--color-surface) 88%, transparent);
  border-color: var(--color-gold);
  color: var(--color-gold-dark);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.navbar__mine-wrap {
  position: relative;
}

.navbar__avatar {
  width: 22px;
  height: 22px;
  border-radius: var(--radius-full);
  overflow: hidden;
  flex-shrink: 0;
  border: 1px solid color-mix(in srgb, var(--color-gold) 40%, transparent);
}

.navbar__avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.navbar__chev {
  opacity: 0.7;
  transition: transform var(--transition-fast);
}

.navbar__chev--open {
  transform: rotate(180deg);
}

.navbar__mine-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 10px);
  min-width: 148px;
  padding: var(--space-2);
  background: var(--color-surface-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  display: flex;
  flex-direction: column;
  gap: 2px;
  z-index: 20;
}

.navbar__mine-item {
  width: 100%;
  padding: var(--space-3);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  text-align: left;
  transition: all var(--transition-fast);
}

.navbar__mine-item:hover {
  background: color-mix(in srgb, var(--color-gold) 10%, transparent);
  color: var(--color-gold-dark);
}

.navbar__mine-item--logout {
  color: var(--color-cinnabar);
}

.mine-drop-enter-active,
.mine-drop-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.mine-drop-enter-from,
.mine-drop-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.98);
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
