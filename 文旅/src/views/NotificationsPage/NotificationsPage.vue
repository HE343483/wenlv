<template>
  <div class="notif-page">
    <div class="np-header">
      <h2>消息中心</h2>
    </div>
    <div class="np-list">
      <div class="np-item" v-for="n in notifications" :key="n.id" :class="{ unread: !n.read }">
        <span class="np-icon">{{ n.icon }}</span>
        <div class="np-body">
          <div class="np-title">{{ n.title }}</div>
          <div class="np-time">{{ n.time }}</div>
        </div>
        <span class="np-dot" v-if="!n.read"></span>
      </div>
    </div>
  </div>
</template>
<script setup>
const notifications = [
  { id: 1, icon: '🖼️', title: '收藏馆有新照片', time: '5分钟前', read: false },
  { id: 2, icon: '⭐', title: '宽窄巷子景点有新的评价', time: '1小时前', read: false },
  { id: 3, icon: '🗺️', title: '成都地图已更新', time: '3小时前', read: false },
  { id: 4, icon: '📋', title: '你的行程待确认', time: '昨天', read: true },
  { id: 5, icon: '👤', title: '系统通知：账号安全提示', time: '2天前', read: true },
]
</script>
<style scoped>
.notif-page { height: 100%; overflow-y: auto; padding: 0 20px 24px; background: #FCF7F2; }
.np-header h2 { font-family: "Noto Serif SC","Source Han Serif SC","Songti SC",serif; font-size: 1.2rem; font-weight: 700; color: #2D2D3A; letter-spacing: 2px; padding: 24px 0 16px; }
.np-list { display: flex; flex-direction: column; gap: 6px; }
.np-item { display: flex; align-items: center; gap: 12px; padding: 14px 16px; background: #FFF; border-radius: 12px; border: 1px solid rgba(0,0,0,.04); position: relative; }
.np-icon { font-size: 1.3rem; width: 32px; text-align: center; }
.np-body { flex: 1; }
.np-title { font-size: .78rem; color: #2D2D3A; font-weight: 500; }
.np-time { font-size: .58rem; color: #9CA3AF; margin-top: 2px; }
.np-dot { width: 7px; height: 7px; border-radius: 50%; background: #E85D3A; flex-shrink: 0; }
</style>
