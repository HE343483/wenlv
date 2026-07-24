<template>
  <div class="main-layout">
    <!-- ===== 顶部栏 ===== -->
    <header class="top-bar">
      <div class="top-left" @click="$router.push('/main/home')">
        <svg class="brand-icon" viewBox="0 0 40 40" width="22" height="26">
          <circle cx="20" cy="20" r="18" fill="none" stroke="#E85D3A" stroke-width="1.5" opacity="0.3"/>
          <circle cx="20" cy="20" r="14" fill="none" stroke="#E85D3A" stroke-width="1" opacity="0.5"/>
          <g v-for="i in 12" :key="i">
            <line :x1="20 + 8 * Math.cos(i * Math.PI / 6)" :y1="20 + 8 * Math.sin(i * Math.PI / 6)"
                  :x2="20 + 14 * Math.cos(i * Math.PI / 6)" :y2="20 + 14 * Math.sin(i * Math.PI / 6)"
                  stroke="#E85D3A" stroke-width="1.5" stroke-linecap="round" opacity="0.5"/>
          </g>
          <circle cx="20" cy="20" r="6" fill="none" stroke="#E85D3A" stroke-width="1.2" opacity="0.4"/>
        </svg>
        <div>
          <div class="top-title">蓉城漫游</div>
          <div class="top-sub">CHENGDU CULTURE</div>
        </div>
      </div>

      <div class="top-right">
        <!-- 快捷图标 -->
        <button class="top-icon-btn" @click="showWeather = !showWeather" title="天气">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M17.66 6.34l1.41-1.41"/>
          </svg>
        </button>
        <button class="top-icon-btn" @click="openScanner" title="扫一扫">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M3 7V5a2 2 0 0 1 2-2h2M3 17v2a2 2 0 0 0 2 2h2M17 3h2a2 2 0 0 1 2 2v2M21 17v2a2 2 0 0 1-2 2h-2"/><line x1="7" y1="12" x2="17" y2="12"/>
          </svg>
        </button>

        <div class="top-divider"></div>

        <!-- AI助手 -->
        <button class="top-icon-btn ai-btn" @click="openAIChat" title="智能助手">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            <line x1="9" y1="10" x2="15" y2="10"/><line x1="12" y1="7" x2="12" y2="13"/>
          </svg>
        </button>

        <!-- 用户 -->
        <div class="user-chip">
          <span class="user-dot"></span>
          <span class="user-name">{{ auth.username }}</span>
        </div>
      </div>
    </header>

    <!-- ===== 天气浮层 ===== -->
    <transition name="fade-down">
      <div class="weather-bar" v-if="showWeather">
        <span class="w-item">☀️ 成都 32°C 晴</span>
        <span class="w-item">🌤 都江堰 28°C 多云</span>
        <span class="w-item">☁️ 青城山 24°C 阴</span>
      </div>
    </transition>

    <!-- ===== AI 对话浮层 ===== -->
    <transition name="fade-up">
      <div class="ai-chat" v-if="showAIChat">
        <div class="ai-header">
          <span>🤖 蓉城助手</span>
          <button class="ai-close" @click="showAIChat = false">✕</button>
        </div>
        <div class="ai-body">
          <div class="ai-msg ai-reply">您好！我是蓉城文旅助手，请问有什么可以帮助您的？</div>
          <div class="ai-msg ai-user" v-for="m in aiMsgs" :key="m">{{ m }}</div>
        </div>
        <div class="ai-foot">
          <input v-model="aiInput" class="ai-input" placeholder="输入问题..." @keyup.enter="sendAI" />
          <button class="ai-send" @click="sendAI">发送</button>
        </div>
      </div>
    </transition>

    <!-- ===== 主内容区 ===== -->
    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="page-fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <!-- ===== 底部导航栏 ===== -->
    <nav class="bottom-nav">
      <router-link v-for="tab in tabs" :key="tab.path"
        :to="tab.path"
        class="nav-item"
        active-class="nav-item-active"
        :exact="tab.exact"
      >
        <span class="nav-icon" v-html="tab.icon"></span>
        <span class="nav-label">{{ tab.label }}</span>
        <span class="nav-badge" v-if="tab.badge && tab.badge > 0">{{ tab.badge }}</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import { auth, logout } from '../../stores/auth.js'

const router = useRouter()

// ===== 底部标签定义 =====
const tabs = [
  { path: '/main/home',          label: '首页',     exact: true,
    icon: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>' },
  { path: '/main/map',           label: '成都地图',
    icon: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>' },
  { path: '/main/itinerary',     label: '行程',
    icon: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>' },
  { path: '/main/notifications', label: '消息',     badge: 3,
    icon: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>' },
  { path: '/main/profile',       label: '我的',
    icon: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>' }
]

// ===== 天气 =====
const showWeather = ref(false)

// ===== AI 聊天 =====
const showAIChat = ref(false)
const aiInput = ref('')
const aiMsgs = reactive([])

function openScanner() {
  alert('📷 扫码功能已就绪\n可将此功能对接景区AR导览系统')
}

function openAIChat() {
  showAIChat.value = !showAIChat.value
}

function sendAI() {
  const text = aiInput.value.trim()
  if (!text) return
  aiMsgs.push(text)
  aiInput.value = ''
  setTimeout(() => {
    aiMsgs.push('🤖 已收到您的问题，我将为您查询成都景区相关信息。')
  }, 600)
}

// ===== 退出 =====
function handleLogout() {
  gsap.to('.main-layout', {
    opacity: 0, y: -20, duration: 0.4, ease: 'power2.in',
    onComplete: () => { logout(); router.push('/login') }
  })
}

onMounted(() => {
  gsap.fromTo('.bottom-nav', { y: 40, opacity: 0 }, { y: 0, opacity: 1, duration: 0.5, ease: 'power3.out', delay: 0.2 })
})
</script>

<style scoped>
.main-layout {
  display: flex; flex-direction: column; height: 100vh;
  background: #FCF7F2; overflow: hidden;
}

/* ===== 顶部栏 ===== */
.top-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 20px; height: 52px; flex-shrink: 0;
  background: rgba(255,255,255,.85);
  border-bottom: 1px solid rgba(232,93,58,.06);
  backdrop-filter: blur(12px); z-index: 100;
}
.top-left { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.brand-icon { opacity: .7; }
.top-title { font-size: .88rem; font-weight: 700; color: #E85D3A; line-height: 1.2; }
.top-sub { font-size: .48rem; color: #9CA3AF; letter-spacing: 2px; line-height: 1; }
.top-right { display: flex; align-items: center; gap: 4px; }
.top-icon-btn {
  width: 32px; height: 32px; border-radius: 8px;
  border: none; background: transparent;
  color: #6B7280; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all .25s;
}
.top-icon-btn:hover { color: #E85D3A; background: rgba(232,93,58,.06); }
.ai-btn:hover { color: #0EA5A0; background: rgba(14,165,160,.06); }
.top-divider { width: 1px; height: 18px; background: rgba(0,0,0,.06); margin: 0 4px; }
.user-chip {
  display: flex; align-items: center; gap: 5px;
  margin-left: 4px; padding: 3px 10px 3px 6px;
  border-radius: 6px; font-size: .7rem; color: #6B7280;
}
.user-dot { width: 5px; height: 5px; border-radius: 50%; background: #10B981; }
.user-name { max-width: 64px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* ===== 天气浮层 ===== */
.weather-bar {
  display: flex; gap: 16px; padding: 6px 20px;
  background: rgba(255,255,255,.7); border-bottom: 1px solid rgba(0,0,0,.04);
  font-size: .62rem; color: #6B7280; flex-shrink: 0;
}
.w-item { display: flex; align-items: center; gap: 4px; }

/* ===== AI 对话浮层 ===== */
.ai-chat {
  position: fixed; right: 20px; bottom: 70px; z-index: 200;
  width: 320px; height: 400px; border-radius: 14px;
  background: rgba(255,255,255,.96); border: 1px solid rgba(232,93,58,.1);
  backdrop-filter: blur(20px);
  display: flex; flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,.08);
}
.ai-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 16px; border-bottom: 1px solid rgba(232,93,58,.06);
  font-size: .8rem; color: #E85D3A;
}
.ai-close { background: none; border: none; color: #9CA3AF; cursor: pointer; font-size: .85rem; padding: 2px 6px; border-radius: 4px; }
.ai-close:hover { color: #E85D3A; background: rgba(232,93,58,.05); }
.ai-body { flex: 1; overflow-y: auto; padding: 12px 14px; display: flex; flex-direction: column; gap: 8px; }
.ai-msg { max-width: 85%; padding: 8px 12px; border-radius: 10px; font-size: .75rem; line-height: 1.5; }
.ai-reply { background: rgba(232,93,58,.06); color: #4B5563; align-self: flex-start; }
.ai-user { background: rgba(14,165,160,.08); color: #4B5563; align-self: flex-end; }
.ai-foot { display: flex; gap: 6px; padding: 10px 12px; border-top: 1px solid rgba(0,0,0,.04); }
.ai-input {
  flex: 1; padding: 8px 12px; border-radius: 8px;
  border: 1px solid rgba(0,0,0,.06); background: rgba(255,255,255,.8);
  color: #2D2D3A; font-size: .75rem; font-family: inherit; outline: none;
}
.ai-input:focus { border-color: rgba(232,93,58,.25); }
.ai-send {
  padding: 8px 14px; border-radius: 8px;
  border: none; background: linear-gradient(135deg, #E85D3A, #F5A623);
  color: #fff; font-size: .7rem; cursor: pointer; font-family: inherit;
}
.ai-send:hover { opacity: .9; }

/* ===== 主内容 ===== */
.main-content { flex: 1; overflow: hidden; position: relative; }
.page-fade-enter-active, .page-fade-leave-active { transition: opacity .3s ease, transform .3s ease; }
.page-fade-enter-from { opacity: 0; transform: translateY(8px); }
.page-fade-leave-to { opacity: 0; transform: translateY(-8px); }

/* ===== 底部导航 ===== */
.bottom-nav {
  display: flex; align-items: center; justify-content: space-around;
  height: 60px; flex-shrink: 0;
  background: rgba(255,255,255,.9);
  border-top: 1px solid rgba(232,93,58,.06);
  backdrop-filter: blur(12px); z-index: 100;
  padding: 0 4px; padding-bottom: env(safe-area-inset-bottom, 0);
}
.nav-item {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 2px; flex: 1; height: 100%;
  text-decoration: none; color: #9CA3AF;
  transition: color .3s; position: relative;
}
.nav-item:hover { color: #6B7280; }
.nav-item-active { color: #E85D3A !important; }
.nav-icon { display: flex; align-items: center; justify-content: center; height: 24px; }
.nav-label { font-size: .58rem; letter-spacing: .5px; }
.nav-badge {
  position: absolute; top: 2px; right: 50%; margin-right: -22px;
  min-width: 16px; height: 16px; padding: 0 5px;
  border-radius: 8px; background: #E85D3A;
  color: #fff; font-size: .5rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}

/* 过渡动画 */
.fade-down-enter-active, .fade-down-leave-active { transition: all .3s ease; }
.fade-down-enter-from, .fade-down-leave-to { opacity: 0; transform: translateY(-10px); }
.fade-up-enter-active, .fade-up-leave-active { transition: all .3s ease; }
.fade-up-enter-from, .fade-up-leave-to { opacity: 0; transform: translateY(10px); }
</style>
