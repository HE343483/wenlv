<template>
  <router-view v-slot="{ Component, route }">
    <transition :name="route.meta.transition || 'page-fade'" mode="out-in">
      <component :is="Component" :key="route.path" />
    </transition>
  </router-view>
</template>

<style>
/* 全局基础样式 — 成都文旅 · 亮色主题 */
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: 'Noto Sans SC', 'PingFang SC', sans-serif;
  background: #FCF7F2;
  color: #2D2D3A;
  overflow-x: hidden;
  -webkit-font-smoothing: antialiased;
}
::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-thumb { background: rgba(232,93,58,.25); border-radius: 2px; }

/* 路由过渡动画 */
.page-fade-enter-active, .page-fade-leave-active {
  transition: opacity 0.6s ease, transform 0.6s ease;
}
.page-fade-enter-from { opacity: 0; transform: translateY(20px); }
.page-fade-leave-to { opacity: 0; transform: translateY(-20px); }

.page-slide-enter-active, .page-slide-leave-active {
  transition: all 0.5s cubic-bezier(0.16,1,0.3,1);
}
.page-slide-enter-from { opacity: 0; transform: translateX(30px); }
.page-slide-leave-to { opacity: 0; transform: translateX(-30px); }
</style>

<script setup>
// 根组件 — 无额外逻辑
</script>
