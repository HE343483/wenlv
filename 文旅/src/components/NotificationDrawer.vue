<template>
  <transition name="drawer-fade">
    <div class="drawer-overlay" v-if="visible" @click.self="close" @keydown.escape="close">
      <aside class="drawer-panel" :style="{ transform: visible ? 'translateX(0)' : 'translateX(100%)' }">
        <!-- 头部 -->
        <div class="drawer-header">
          <div class="dh-left">
            <svg viewBox="0 0 40 40" class="dh-avatar" width="36" height="36">
              <circle cx="20" cy="20" r="19" fill="rgba(232,93,58,.06)" stroke="rgba(232,93,58,.12)" stroke-width="1"/>
              <circle cx="20" cy="16" r="6" fill="rgba(232,93,58,.12)"/>
              <path d="M8 34 C8 26, 14 22, 20 22 C26 22, 32 26, 32 34" fill="none" stroke="rgba(232,93,58,.12)" stroke-width="1.5"/>
            </svg>
            <div>
              <div class="dh-name">{{ auth.username || '用户' }}</div>
              <div class="dh-level">Lv.3 · 蓉城文化探索者</div>
            </div>
          </div>
          <button class="dh-close" @click="close" aria-label="关闭">✕</button>
        </div>

        <!-- 统计 -->
        <div class="drawer-stats">
          <div class="ds-item">
            <span class="ds-num">6</span>
            <span class="ds-label">已打卡</span>
          </div>
          <div class="ds-item">
            <span class="ds-num">4</span>
            <span class="ds-label">收藏</span>
          </div>
          <div class="ds-item">
            <span class="ds-num">18</span>
            <span class="ds-label">积分</span>
          </div>
        </div>

        <!-- 快捷操作 -->
        <div class="drawer-actions">
          <button class="da-item" @click="goTo('/main/collection')">
            <span class="da-icon">🖼️</span>
            <span>收藏馆</span>
          </button>
          <button class="da-item" @click="goTo('/main/itinerary')">
            <span class="da-icon">📋</span>
            <span>我的行程</span>
          </button>
        </div>

        <!-- 天气 -->
        <div class="drawer-weather">
          <span class="dw-item">☀️ 成都 32°C</span>
          <span class="dw-item">🌤 都江堰 28°C</span>
        </div>

        <!-- 底部 -->
        <div class="drawer-footer">
          <button class="df-logout" @click="handleLogout">退出登录</button>
        </div>
      </aside>
    </div>
  </transition>
</template>

<script setup>
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import { auth, logout } from '../stores/auth.js'

const router = useRouter()

const props = defineProps({
  visible: { type: Boolean, default: false }
})
const emit = defineEmits(['close'])

function close() {
  emit('close')
}

function goTo(path) {
  close()
  router.push(path)
}

function handleLogout() {
  close()
  gsap.to('.app-shell', {
    opacity: 0, y: -20, duration: 0.4, ease: 'power2.in',
    onComplete: () => { logout(); router.push('/login') }
  })
}
</script>

<style scoped>
.drawer-overlay {
  position: fixed;
  inset: 0;
  z-index: 300;
  background: rgba(0, 0, 0, .2);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  justify-content: flex-end;
}

.drawer-panel {
  width: 320px;
  height: 100vh;
  background: rgba(255, 255, 255, .96);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-left: 1px solid rgba(232, 93, 58, .06);
  box-shadow: -8px 0 40px rgba(0, 0, 0, .06);
  display: flex;
  flex-direction: column;
  transition: transform .4s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ===== 头部 ===== */
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 20px 16px;
  border-bottom: 1px solid rgba(0, 0, 0, .04);
}
.dh-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.dh-avatar { flex-shrink: 0; }
.dh-name { font-size: .82rem; font-weight: 700; color: #2D2D3A; line-height: 1.3; }
.dh-level { font-size: .55rem; color: #9CA3AF; }
.dh-close {
  width: 28px; height: 28px;
  border-radius: 8px;
  border: none; background: transparent;
  color: #9CA3AF; cursor: pointer;
  font-size: .82rem;
  display: flex; align-items: center; justify-content: center;
  transition: all .25s;
}
.dh-close:hover { color: #E85D3A; background: rgba(232,93,58,.06); }

/* ===== 统计 ===== */
.drawer-stats {
  display: flex;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(0, 0, 0, .04);
  gap: 0;
}
.ds-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.ds-num {
  font-size: 1rem;
  font-weight: 700;
  color: #2D2D3A;
  font-feature-settings: "tnum";
}
.ds-label { font-size: .52rem; color: #9CA3AF; letter-spacing: 1px; }

/* ===== 快捷操作 ===== */
.drawer-actions {
  flex: 1;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.da-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  border: none;
  background: transparent;
  color: #4B5563;
  font-size: .75rem;
  font-family: inherit;
  cursor: pointer;
  text-align: left;
  transition: all .2s;
  position: relative;
}
.da-item:hover { background: rgba(232,93,58,.06); color: #E85D3A; }
.da-icon { font-size: 1rem; width: 24px; text-align: center; flex-shrink: 0; }
.da-badge {
  position: absolute;
  right: 10px;
  min-width: 18px; height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: #E85D3A;
  color: #fff;
  font-size: .5rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ===== 天气 ===== */
.drawer-weather {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 20px;
  border-top: 1px solid rgba(0, 0, 0, .04);
  border-bottom: 1px solid rgba(0, 0, 0, .04);
}
.dw-item { font-size: .62rem; color: #6B7280; }

/* ===== 底部 ===== */
.drawer-footer {
  padding: 12px 20px 24px;
}
.df-logout {
  width: 100%;
  padding: 10px;
  border-radius: 8px;
  border: 1px solid rgba(239, 68, 68, .1);
  background: transparent;
  color: #EF4444;
  font-size: .7rem;
  font-family: inherit;
  cursor: pointer;
  transition: all .25s;
  letter-spacing: 1px;
}
.df-logout:hover { background: rgba(239,68,68,.06); }

/* ===== 过渡 ===== */
.drawer-fade-enter-active,
.drawer-fade-leave-active { transition: opacity .3s ease; }
.drawer-fade-enter-from,
.drawer-fade-leave-to { opacity: 0; }
</style>
