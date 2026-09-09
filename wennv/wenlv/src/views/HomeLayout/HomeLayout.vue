<script setup lang="ts">
/**
 * HomeLayout.vue — 登录后内部页面布局
 * 顶栏：左 品牌 | 中 快捷导航(首页/探索/美食/路线/收藏/我的) | 右 天气+语言+我的下拉
 * 导航加「收藏」「我的」：收藏为独立页，个人资料与打卡照片在「我的」内完成
 */
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useWeatherStore } from '@/stores/weather'
import { useUserStore } from '@/stores/user'
import WeatherTrigger from '@/components/WeatherTrigger.vue'
import WeatherPanel from '@/components/WeatherPanel.vue'
import LanguageSwitch from '@/components/LanguageSwitch.vue'

const route = useRoute()
const router = useRouter()
const langStore = useLanguageStore()
const weatherStore = useWeatherStore()
const userStore = useUserStore()
const { open } = storeToRefs(weatherStore)
const { profile } = storeToRefs(userStore)

function goBack() {
  router.push('/')
}

const quickNavs = [
  { route: '/home/index', key: 'bottomNav.home' },
  { route: '/home/explore', key: 'bottomNav.explore' },
  { route: '/home/food', key: 'bottomNav.food' },
  { route: '/home/routes', key: 'bottomNav.routes' },
  { route: '/home/favorites', key: 'bottomNav.favorites' },
  { route: '/home/profile', key: 'bottomNav.mine' },
] as const

const isQuickActive = (target: string) => route.path === target || route.path.startsWith(target + '/')

const triggerRef = ref<InstanceType<typeof WeatherTrigger> | null>(null)
const panelRef = ref<InstanceType<typeof WeatherPanel> | null>(null)

/* ── 我的下拉 ── */
const mineOpen = ref(false)
const mineWrapRef = ref<HTMLDivElement | null>(null)
function toggleMine() { mineOpen.value = !mineOpen.value }
function closeMine() { mineOpen.value = false }
function goMine(path: '/home/profile' | '/home/favorites') {
  mineOpen.value = false
  router.push(path)
}
function onDocClick(e: MouseEvent) {
  if (!mineWrapRef.value) return
  if (!mineWrapRef.value.contains(e.target as Node)) mineOpen.value = false
}
onMounted(() => document.addEventListener('click', onDocClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <div class="internal-layout">
    <header class="internal-topbar">
      <div class="internal-topbar__left">
        <button class="internal-topbar__back" :title="langStore.t('common.back')" @click="goBack">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </button>
        <div class="internal-topbar__brand" @click="router.push('/home/index')">
          <svg class="internal-topbar__brand-mark" width="24" height="24" viewBox="0 0 26 26" fill="none">
            <rect x="6" y="6" width="14" height="14" transform="rotate(45 13 13)" stroke="currentColor" stroke-width="1.2" />
            <rect x="10.5" y="10.5" width="5" height="5" transform="rotate(45 13 13)" fill="currentColor" />
          </svg>
          <div class="internal-topbar__brand-text">
            <span class="internal-topbar__title">蜀韵·成都</span>
            <span class="internal-topbar__subtitle">Shu·Chengdu</span>
          </div>
        </div>
      </div>

      <nav class="internal-topbar__nav" :aria-label="langStore.t('nav.ariaNav')">
        <RouterLink
          v-for="item in quickNavs"
          :key="item.route"
          :to="item.route"
          class="internal-topbar__nav-link"
          :class="{ 'internal-topbar__nav-link--active': isQuickActive(item.route) }"
        >
          {{ langStore.t(item.key) }}
        </RouterLink>
      </nav>

      <div class="internal-topbar__actions">
        <WeatherTrigger ref="triggerRef" />
        <WeatherPanel v-if="open" ref="panelRef" :anchor="triggerRef" />
        <LanguageSwitch />
        <span class="internal-topbar__divider" aria-hidden="true" />

        <!-- 我的下拉 -->
        <div ref="mineWrapRef" class="internal-topbar__mine-wrap">
          <button
            class="internal-topbar__profile"
            :class="{ 'internal-topbar__profile--open': mineOpen, 'router-link-active': route.path.startsWith('/home/profile') || route.path.startsWith('/home/favorites') }"
            aria-haspopup="menu"
            :aria-expanded="mineOpen"
            @click.stop="toggleMine"
          >
            <span v-if="profile.avatar" class="internal-topbar__avatar"><img :src="profile.avatar" alt="avatar" /></span>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
              <circle cx="12" cy="8" r="3.5" />
              <path d="M5 20c1.6-3.6 4.1-5 7-5s5.4 1.4 7 5" />
            </svg>
            {{ langStore.t('bottomNav.mine') }}
            <svg class="internal-topbar__chev" :class="{ 'internal-topbar__chev--open': mineOpen }" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M6 9l6 6 6-6" /></svg>
          </button>

          <div v-if="mineOpen" class="internal-topbar__mine-menu" role="menu">
            <button role="menuitem" class="internal-topbar__mine-item" :class="{ 'internal-topbar__mine-item--active': route.path === '/home/profile' }" @click="goMine('/home/profile')">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><circle cx="12" cy="8" r="3.5"/><path d="M5 20c1.6-3.6 4.1-5 7-5s5.4 1.4 7 5"/></svg>
              {{ langStore.t('mineMenu.profile') }}
            </button>
            <button role="menuitem" class="internal-topbar__mine-item" :class="{ 'internal-topbar__mine-item--active': route.path === '/home/favorites' }" @click="goMine('/home/favorites')">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z"/></svg>
              {{ langStore.t('mineMenu.favorites') }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <main class="internal-main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.internal-layout { min-height: 100vh; display: flex; flex-direction: column; padding-bottom: var(--bottom-nav-height); }
.internal-topbar {
  position: sticky; top: 0; z-index: 100; display: flex; align-items: center; justify-content: space-between;
  height: var(--nav-height); padding: 0 var(--space-6); background: var(--color-nav-bg);
  backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px); border-bottom: 1px solid var(--color-border); transition: background var(--transition-base);
}
.internal-topbar::after { content: ''; position: absolute; bottom: -1px; left: 50%; transform: translateX(-50%); width: 60%; height: 1px; background: linear-gradient(90deg, transparent, var(--color-gold), transparent); pointer-events: none; }
.internal-topbar__left { display: flex; align-items: center; gap: var(--space-3); min-width: 0; }
.internal-topbar__back { flex-shrink: 0; width: 36px; height: 36px; border-radius: var(--radius-full); display: flex; align-items: center; justify-content: center; color: var(--color-text-secondary); border: 1px solid var(--color-border); transition: all var(--transition-fast); }
.internal-topbar__back:hover { border-color: var(--color-gold); color: var(--color-gold); box-shadow: 0 0 12px var(--color-gold-glow); }
.internal-topbar__brand { display: flex; align-items: center; gap: var(--space-3); cursor: pointer; user-select: none; }
.internal-topbar__brand-mark { color: var(--color-gold); flex-shrink: 0; }
.internal-topbar__brand-text { display: flex; flex-direction: column; line-height: 1.2; }
.internal-topbar__title { font-family: var(--font-display); font-size: var(--text-lg); font-weight: 700; color: var(--color-gold); letter-spacing: var(--tracking-wide); }
.internal-topbar__subtitle { font-family: var(--font-en-display); font-size: var(--text-xs); color: var(--color-text-muted); letter-spacing: var(--tracking-wider); text-transform: uppercase; }
.internal-topbar__nav { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); display: flex; align-items: center; gap: 4px; white-space: nowrap; }
.internal-topbar__nav-link { position: relative; font-family: var(--font-display); font-size: var(--text-sm); color: var(--color-text-secondary); letter-spacing: var(--tracking-wide); padding: var(--space-2) var(--space-3); border-radius: var(--radius-full); text-decoration: none; transition: color var(--transition-fast), background var(--transition-fast); }
.internal-topbar__nav-link:hover { color: var(--color-gold-dark); background: color-mix(in srgb, var(--color-gold) 8%, transparent); }
.internal-topbar__nav-link--active { color: var(--color-gold-dark); font-weight: 600; }
.internal-topbar__nav-link--active::after { content: ''; position: absolute; left: 50%; bottom: 2px; transform: translateX(-50%); width: 16px; height: 2px; border-radius: var(--radius-full); background: var(--color-gold); }
.internal-topbar__actions { display: flex; align-items: center; gap: var(--space-3); }
.internal-topbar__divider { width: 1px; height: 18px; background: var(--color-border); }

.internal-topbar__mine-wrap { position: relative; }
.internal-topbar__profile {
  display: inline-flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm);
  padding: var(--space-2) var(--space-3); border-radius: var(--radius-full); color: var(--color-text-secondary);
  border: 1px solid transparent; letter-spacing: var(--tracking-wide); transition: all var(--transition-fast);
}
.internal-topbar__profile:hover, .internal-topbar__profile--open, .internal-topbar__profile.router-link-active { color: var(--color-gold-dark); border-color: var(--color-gold); background: color-mix(in srgb, var(--color-gold) 8%, transparent); }
.internal-topbar__avatar { width: 22px; height: 22px; border-radius: var(--radius-full); overflow: hidden; border: 1px solid color-mix(in srgb, var(--color-gold) 40%, transparent); flex-shrink: 0; }
.internal-topbar__avatar img { width: 100%; height: 100%; object-fit: cover; display: block; }
.internal-topbar__chev { transition: transform var(--transition-fast); opacity: 0.7; }
.internal-topbar__chev--open { transform: rotate(180deg); }

.internal-topbar__mine-menu {
  position: absolute; right: 0; top: calc(100% + 10px); min-width: 168px; padding: var(--space-2);
  background: var(--color-surface-elevated); border: 1px solid var(--color-border); border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg), 0 8px 24px var(--color-gold-glow); display: flex; flex-direction: column; gap: 2px; z-index: 20;
}
.internal-topbar__mine-item {
  display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-3);
  border-radius: var(--radius-md); font-size: var(--text-sm); color: var(--color-text-secondary); text-align: left; transition: all var(--transition-fast);
}
.internal-topbar__mine-item:hover { background: color-mix(in srgb, var(--color-gold) 10%, transparent); color: var(--color-gold-dark); }
.internal-topbar__mine-item--active { background: color-mix(in srgb, var(--color-gold) 12%, transparent); color: var(--color-gold-dark); font-weight: 600; }

.internal-main { flex: 1; padding: var(--space-6); padding-bottom: calc(var(--space-6) + var(--bottom-nav-height)); }

@media (max-width: 1100px) { .internal-topbar__nav { display: none; } }
@media (max-width: 768px) {
  .internal-topbar { padding: 0 var(--space-4); }
  .internal-topbar__subtitle { display: none; }
  .internal-main { padding: var(--space-4); padding-bottom: calc(var(--space-4) + var(--bottom-nav-height)); }
}
</style>
