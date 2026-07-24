<template>
  <div class="profile-page">
    <!-- 用户头像 -->
    <div class="profile-header">
      <div class="avatar">
        <svg viewBox="0 0 40 40" width="56" height="56">
          <circle cx="20" cy="20" r="19" fill="rgba(232,93,58,.06)" stroke="rgba(232,93,58,.15)" stroke-width="1"/>
          <circle cx="20" cy="16" r="6" fill="rgba(232,93,58,.12)"/>
          <path d="M8 34 C8 26, 14 22, 20 22 C26 22, 32 26, 32 34" fill="none" stroke="rgba(232,93,58,.12)" stroke-width="1.5"/>
        </svg>
      </div>
      <div class="profile-info">
        <div class="profile-name">{{ auth.username || '用户' }}</div>
        <div class="profile-level">蓉城文化探索者 · Lv.3</div>
      </div>
    </div>

    <!-- 统计数据 -->
    <div class="stats-row">
      <div class="stat-item">
        <span class="stat-num">6</span>
        <span class="stat-label">已打卡</span>
      </div>
      <div class="stat-item">
        <span class="stat-num">4</span>
        <span class="stat-label">收藏攻略</span>
      </div>
      <div class="stat-item">
        <span class="stat-num">18</span>
        <span class="stat-label">积分</span>
      </div>
    </div>

    <!-- 菜单 -->
    <div class="menu-list">
      <div class="menu-item" v-for="m in menus" :key="m.label" @click="handleMenu(m.key)">
        <span class="menu-icon">{{ m.icon }}</span>
        <span class="menu-label">{{ m.label }}</span>
        <span class="menu-arrow">›</span>
      </div>
    </div>

    <!-- 退出 -->
    <button class="logout-btn" @click="handleLogout">退出登录</button>

    <div class="profile-footer">蓉城漫游 v1.0.0</div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import { auth, logout } from '../../stores/auth.js'

const router = useRouter()

const menus = [
  { icon: '📄', key: 'orders',     label: '我的订单' },
  { icon: '⭐', key: 'favorites',  label: '我的收藏' },
  { icon: '🕐', key: 'history',    label: '浏览历史' },
  { icon: '🏅', key: 'badges',     label: '我的成就' },
  { icon: '🔧', key: 'settings',   label: '设置' },
  { icon: '💬', key: 'feedback',   label: '意见反馈' },
  { icon: 'ℹ️', key: 'about',      label: '关于我们' }
]

function handleMenu(key) {
  if (key === 'orders') alert('📄 订单列表功能')
  else if (key === 'favorites') alert('⭐ 收藏列表功能')
  else if (key === 'history') alert('🕐 浏览历史功能')
  else if (key === 'badges') alert('🏅 成就系统功能')
  else if (key === 'settings') alert('🔧 设置页面')
  else if (key === 'feedback') alert('💬 意见反馈功能')
  else if (key === 'about') alert('ℹ️ 蓉城漫游 v1.0.0\n成都文旅 · 数字导览平台')
}

function handleLogout() {
  gsap.to('.profile-page', {
    opacity: 0, y: -10, duration: 0.3, ease: 'power2.in',
    onComplete: () => {
      logout()
      router.push('/login')
    }
  })
}
</script>

<style scoped>
.profile-page {
  height: 100%; overflow-y: auto; padding: 20px 16px;
  background: #FCF7F2;
}
.profile-page::-webkit-scrollbar { width: 3px; }
.profile-page::-webkit-scrollbar-thumb { background: rgba(232,93,58,.15); border-radius: 2px; }

/* 头部 */
.profile-header { display: flex; align-items: center; gap: 16px; margin-bottom: 20px; }
.avatar { flex-shrink: 0; }
.profile-name { font-size: 1rem; font-weight: 700; color: #2D2D3A; margin-bottom: 2px; }
.profile-level { font-size: .65rem; color: rgba(232,93,58,.55); letter-spacing: 1px; }

/* 统计 */
.stats-row {
  display: flex; justify-content: space-around;
  padding: 14px 0; margin-bottom: 18px;
  border-radius: 12px; background: #FFF; border: 1px solid rgba(0,0,0,.04);
  box-shadow: 0 1px 3px rgba(0,0,0,.02);
}
.stat-item { display: flex; flex-direction: column; align-items: center; gap: 4px; }
.stat-num { font-size: 1.1rem; font-weight: 700; color: #E85D3A; }
.stat-label { font-size: .6rem; color: #6B7280; }

/* 菜单 */
.menu-list { display: flex; flex-direction: column; gap: 2px; margin-bottom: 20px; }
.menu-item {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 14px; border-radius: 10px;
  cursor: pointer; transition: all .3s;
}
.menu-item:hover { background: rgba(232,93,58,.03); }
.menu-icon { font-size: 1rem; width: 24px; text-align: center; }
.menu-label { flex: 1; font-size: .78rem; color: #2D2D3A; }
.menu-arrow { font-size: 1.1rem; color: #9CA3AF; }

/* 退出 */
.logout-btn {
  width: 100%; padding: 12px;
  border-radius: 10px; border: 1px solid rgba(239,68,68,.1);
  background: rgba(239,68,68,.02); color: #EF4444;
  font-size: .78rem; cursor: pointer; font-family: inherit;
  transition: all .3s; margin-bottom: 20px;
}
.logout-btn:hover { border-color: rgba(239,68,68,.2); background: rgba(239,68,68,.04); }

.profile-footer { text-align: center; font-size: .58rem; color: #9CA3AF; letter-spacing: 2px; }
</style>
