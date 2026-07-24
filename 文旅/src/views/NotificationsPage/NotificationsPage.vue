<template>
  <div class="notif-page">
    <div class="page-header">
      <h2>消息通知</h2>
      <button class="read-all" @click="markAllRead">全部已读</button>
    </div>

    <div class="tabs">
      <button v-for="t in tabs" :key="t.key" class="tab" :class="{ active: activeTab === t.key }" @click="activeTab = t.key">
        {{ t.label }}
        <span class="tab-badge" v-if="t.count">{{ t.count }}</span>
      </button>
    </div>

    <div class="notif-list">
      <div class="notif-card" v-for="(n, i) in filteredNotifs" :key="i" :class="{ unread: !n.read }" @click="n.read = true">
        <div class="nf-icon">{{ n.icon }}</div>
        <div class="nf-body">
          <div class="nf-title">{{ n.title }}</div>
          <div class="nf-desc">{{ n.desc }}</div>
          <div class="nf-time">{{ n.time }}</div>
        </div>
        <div class="nf-dot" v-if="!n.read"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const activeTab = ref('all')
const tabs = [
  { key: 'all',   label: '全部', count: 0 },
  { key: 'system', label: '系统', count: 1 },
  { key: 'order',  label: '订单', count: 2 },
  { key: 'promo',  label: '活动', count: 0 }
]

const notifs = ref([
  { icon: '🎫', title: '门票预订成功', desc: '武侯祠博物馆 · 成人票 × 2', time: '10分钟前', read: false, type: 'order' },
  { icon: '📢', title: '景区公告：暑期夜游开放', desc: '锦里古街7月15日起开放夜游，营业时间延长至22:00', time: '2小时前', read: false, type: 'system' },
  { icon: '🎉', title: '恭喜获得"蓉城文化达人"勋章', desc: '您已成功打卡3个成都景点，继续探索解锁更多成就', time: '昨天', read: true, type: 'system' },
  { icon: '🏷️', title: '限时优惠：青城山门票7折', desc: '暑期特惠 · 7月31日前预订可享折扣', time: '2天前', read: false, type: 'promo' },
  { icon: '🛎️', title: '行程提醒：明日出发', desc: '蓉城市井文化深度游 · 集合时间 09:00 宽窄巷子', time: '3天前', read: true, type: 'order' },
  { icon: '📝', title: '订单待评价', desc: '您的都江堰之旅已完成，点击评价获取积分', time: '5天前', read: true, type: 'order' },
  { icon: '🎊', title: '金沙遗址线上展馆开放', desc: '线上3D展馆已上线，足不出户欣赏太阳神鸟金饰', time: '1周前', read: true, type: 'promo' }
])

const filteredNotifs = computed(() => {
  if (activeTab.value === 'all') return notifs.value
  return notifs.value.filter(n => n.type === activeTab.value)
})

function markAllRead() {
  notifs.value.forEach(n => n.read = true)
}
</script>

<style scoped>
.notif-page {
  height: 100%; overflow-y: auto; padding: 16px;
  background: #FCF7F2;
}
.notif-page::-webkit-scrollbar { width: 3px; }
.notif-page::-webkit-scrollbar-thumb { background: rgba(232,93,58,.15); border-radius: 2px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.page-header h2 { font-size: 1rem; font-weight: 700; color: #2D2D3A; }
.read-all { background: none; border: none; font-size: .68rem; color: rgba(232,93,58,.55); cursor: pointer; font-family: inherit; }
.read-all:hover { color: #E85D3A; }

/* 标签 */
.tabs { display: flex; gap: 4px; margin-bottom: 14px; }
.tab {
  flex: 1; padding: 6px 10px; border-radius: 8px;
  border: 1px solid rgba(0,0,0,.04); background: #FFF;
  color: #6B7280; font-size: .68rem; cursor: pointer; font-family: inherit;
  transition: all .3s; position: relative;
  box-shadow: 0 1px 2px rgba(0,0,0,.02);
}
.tab.active { border-color: rgba(232,93,58,.2); color: #E85D3A; background: rgba(232,93,58,.03); }
.tab-badge {
  position: absolute; top: -4px; right: -4px;
  min-width: 14px; height: 14px; padding: 0 4px;
  border-radius: 7px; background: #E85D3A;
  color: #fff; font-size: .5rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}

/* 列表 */
.notif-list { display: flex; flex-direction: column; gap: 8px; }
.notif-card {
  display: flex; align-items: flex-start; gap: 10px;
  padding: 12px; border-radius: 10px;
  background: #FFF; border: 1px solid rgba(0,0,0,.03);
  cursor: pointer; transition: all .3s; position: relative;
  box-shadow: 0 1px 3px rgba(0,0,0,.02);
}
.notif-card:hover { border-color: rgba(232,93,58,.08); }
.notif-card.unread { background: rgba(232,93,58,.02); border-color: rgba(232,93,58,.08); }
.nf-icon { font-size: 1.2rem; width: 28px; text-align: center; margin-top: 2px; }
.nf-body { flex: 1; }
.nf-title { font-size: .78rem; font-weight: 600; color: #2D2D3A; margin-bottom: 3px; }
.nf-desc { font-size: .68rem; color: #6B7280; line-height: 1.5; margin-bottom: 4px; }
.nf-time { font-size: .58rem; color: #9CA3AF; }
.nf-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: #E85D3A; flex-shrink: 0; margin-top: 6px;
}
</style>
