<template>
  <aside class="sidebar" :class="isExpanded ? 'expanded' : 'collapsed'">
    <!-- 顶部品牌区 -->
    <div class="sb-header">
      <button class="sb-toggle" @click="isExpanded = !isExpanded" :title="isExpanded ? '收起导航' : '展开导航'">
        <svg v-if="isExpanded" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 5 12 12 19"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="4" y1="6" x2="20" y2="6"/><line x1="4" y1="12" x2="20" y2="12"/><line x1="4" y1="18" x2="20" y2="18"/>
        </svg>
      </button>
      <div class="sb-brand" v-show="isExpanded">
        <svg class="sb-logo" viewBox="0 0 40 40" width="20" height="24">
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
          <div class="sb-title">蓉城漫游</div>
          <div class="sb-sub">CHENGDU</div>
        </div>
      </div>
    </div>

    <!-- 导航菜单 -->
    <nav class="sb-nav">
      <router-link
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="sb-item"
        active-class="sb-item--active"
      >
        <span class="sb-icon" v-html="item.icon"></span>
        <span class="sb-label" v-show="isExpanded">{{ item.label }}</span>
      </router-link>
    </nav>

    <!-- 底部用户 -->
    <div class="sb-footer">
      <div class="sb-item" style="cursor:default;">
        <span class="sb-icon">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
          </svg>
        </span>
        <span class="sb-label" v-show="isExpanded">{{ auth.username || '用户' }}</span>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref } from 'vue'
import { auth } from '../stores/auth.js'

const isExpanded = ref(true)

const navItems = [
  {
    path: '/main/home', label: '首页',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>'
  },
  {
    path: '/main/collection', label: '收藏馆',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>'
  },
  {
    path: '/main/map', label: '成都地图',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>'
  },
  {
    path: '/main/itinerary', label: '行程规划',
    icon: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>'
  }
]
</script>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: rgba(255, 255, 255, .92);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-right: 1px solid rgba(232, 93, 58, .06);
  overflow: hidden;
  z-index: 200;
  flex-shrink: 0;
  transition: width .35s cubic-bezier(0.16, 1, 0.3, 1);
}

.expanded { width: 220px; }
.collapsed { width: 60px; }

/* ===== 头部 ===== */
.sb-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 8px;
  min-height: 52px;
  border-bottom: 1px solid rgba(232, 93, 58, .06);
}

.sb-toggle {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  background: transparent;
  color: #6B7280;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all .25s;
}
.sb-toggle:hover { background: rgba(232, 93, 58, .06); color: #E85D3A; }

.sb-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
  white-space: nowrap;
}
.sb-logo { opacity: .7; flex-shrink: 0; }
.sb-title { font-size: .82rem; font-weight: 700; color: #E85D3A; line-height: 1.2; }
.sb-sub { font-size: .42rem; color: #9CA3AF; letter-spacing: 2px; line-height: 1; }

/* ===== 导航 ===== */
.sb-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 6px;
  overflow-y: auto;
}

.sb-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  color: #6B7280;
  text-decoration: none;
  transition: all .2s ease;
  white-space: nowrap;
  overflow: hidden;
  cursor: pointer;
}
.sb-item:hover {
  background: rgba(232, 93, 58, .06);
  color: #E85D3A;
}
.sb-item--active {
  background: rgba(232, 93, 58, .1) !important;
  color: #E85D3A !important;
  font-weight: 600;
}

.sb-icon {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sb-label {
  font-size: .75rem;
  letter-spacing: 1px;
}

/* ===== 底部 ===== */
.sb-footer {
  padding: 6px;
  border-top: 1px solid rgba(232, 93, 58, .06);
}
.sb-footer .sb-item { border-radius: 10px; }
</style>
