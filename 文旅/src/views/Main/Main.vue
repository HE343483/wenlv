<template>
  <div class="app-shell">
    <!-- 左侧导航 -->
    <AppSidebar />

    <!-- 主内容 -->
    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="page-fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <!-- 右上角抽屉触发按钮 -->
    <button class="drawer-trigger" @click="drawerVisible = true" title="用户菜单">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.8">
        <circle cx="12" cy="12" r="10"/>
        <circle cx="12" cy="10" r="3"/>
        <path d="M5 20c0-4 3.5-6 7-6s7 2 7 6"/>
      </svg>
    </button>

    <!-- 右侧滑出弹窗 -->
    <NotificationDrawer :visible="drawerVisible" @close="drawerVisible = false" />

    <!-- AI 悬浮球 -->
    <AiFloatingBall />
  </div>
</template>

<script setup>
import { ref} from 'vue'
import AppSidebar from '../../components/AppSidebar.vue'
import NotificationDrawer from '../../components/NotificationDrawer.vue'
import AiFloatingBall from '../../components/AiFloatingBall.vue'

// ===== 右侧抽屉 =====
const drawerVisible = ref(false)
</script>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: row;
  height: 100vh;
  background: #FCF7F2;
  overflow: hidden;
}

.main-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

/* ===== 抽屉触发按钮 ===== */
.drawer-trigger {
  position: fixed;
  top: 14px;
  right: 20px;
  z-index: 150;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  background: rgba(255, 255, 255, .85);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  color: #6B7280;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, .04);
  transition: all .25s;
}
.drawer-trigger:hover {
  color: #E85D3A;
  background: rgba(255, 255, 255, .95);
  box-shadow: 0 4px 16px rgba(232, 93, 58, .1);
}

/* ===== 路由过渡 ===== */
.page-fade-enter-active,
.page-fade-leave-active { transition: opacity .3s ease, transform .3s ease; }
.page-fade-enter-from { opacity: 0; transform: translateY(8px); }
.page-fade-leave-to { opacity: 0; transform: translateY(-8px); }
</style>
