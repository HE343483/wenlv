<script setup lang="ts">
/**
 * AppIcon.vue — 统一图标组件
 * 风格：底部导航栏同款细线条 SVG（24×24, stroke-width 1.5, currentColor, round caps/joins）
 * 用法：<AppIcon name="tea" size="20" />
 */
defineProps<{
  name: string
  size?: number
  /** 实心渲染（当前仅 star 使用） */
  filled?: boolean
}>()
</script>

<template>
  <svg
    :width="size ?? 20"
    :height="size ?? 20"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.5"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <!-- ──── 文化 ──── -->
    <!-- 蜀锦 / 织布 -->
    <template v-if="name === 'brocade'">
      <path d="M4 6h16v12H4z"/>
      <path d="M4 10h16M4 14h16"/>
      <path d="M10 6v12M14 6v12"/>
    </template>
    <!-- 川剧 / 脸谱 -->
    <template v-else-if="name === 'opera'">
      <path d="M5 12a7 7 0 0 1 14 0v2H5v-2z"/>
      <circle cx="9" cy="10" r="1.5"/>
      <circle cx="15" cy="10" r="1.5"/>
      <path d="M9 14c.5 1 1.5 1.5 3 1.5s2.5-.5 3-1.5"/>
      <path d="M5 14v2c0 2 2 3 7 3s7-1 7-3v-2"/>
    </template>
    <!-- 盖碗茶 -->
    <template v-else-if="name === 'tea'">
      <path d="M5 3v7a7 7 0 0 0 14 0V3"/>
      <path d="M5 3h14"/>
      <path d="M3 3h18"/>
      <path d="M9 17v3c0 1 .5 2 3 2s3-1 3-2v-3"/>
      <path d="M16 11a2 2 0 0 0 4 0v-1a2 2 0 0 0-4 0"/>
    </template>
    <!-- 川菜 / 辣椒 -->
    <template v-else-if="name === 'cuisine' || name === 'spicy'">
      <path d="M12 2C8 6 6 10 6 15a6 6 0 0 0 12 0c0-5-2-9-6-13z"/>
      <path d="M12 9v6"/>
      <path d="M9 12h6"/>
    </template>
    <!-- 古蜀文明 / 文物 -->
    <template v-else-if="name === 'history'">
      <path d="M8 3L4 8v12a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1V8l-4-5H8z"/>
      <path d="M4 8h16"/>
      <path d="M12 3v6"/>
      <path d="M9 12h6"/>
    </template>
    <!-- 蜀绣 / 针线 -->
    <template v-else-if="name === 'embroidery' || name === 'thread'">
      <path d="M4 20L20 4"/>
      <path d="M8 16l4-4"/>
      <path d="M16 8l-4 4"/>
      <circle cx="4" cy="20" r="2"/>
      <circle cx="20" cy="4" r="2"/>
    </template>

    <!-- ──── 美食 ──── -->
    <!-- 火锅 -->
    <template v-else-if="name === 'hotpot'">
      <ellipse cx="12" cy="16" rx="8" ry="4"/>
      <path d="M12 12v4"/>
      <path d="M8 14v2M16 14v2"/>
      <path d="M10 5c-1 2-3 3-3 5h10c0-2-2-3-3-5"/>
      <path d="M12 2v3"/>
    </template>
    <!-- 串串香 -->
    <template v-else-if="name === 'chuanchuan'">
      <path d="M6 20l4-12"/>
      <path d="M10 20l4-12"/>
      <path d="M14 20l4-12"/>
      <circle cx="8" cy="7" r="2"/>
      <circle cx="12" cy="5" r="2"/>
      <circle cx="16" cy="7" r="2"/>
    </template>
    <!-- 小吃 / 饺子 -->
    <template v-else-if="name === 'snacks' || name === 'dumpling'">
      <path d="M5 12c0 4 3 8 7 8s7-4 7-8"/>
      <path d="M5 12c0-3 3-6 7-6s7 3 7 6"/>
      <path d="M8 12h8"/>
      <path d="M12 8v8"/>
    </template>
    <!-- 夜宵 / 月亮 -->
    <template v-else-if="name === 'nightfood' || name === 'moon'">
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
    </template>
    <!-- 麻婆豆腐 -->
    <template v-else-if="name === 'mapo'">
      <rect x="5" y="8" width="14" height="12" rx="2"/>
      <path d="M5 14h14"/>
      <circle cx="9" cy="11" r="1.5"/>
      <circle cx="15" cy="11" r="1.5"/>
      <path d="M12 4l-2 4h4l-2-4z"/>
    </template>
    <!-- 宫保鸡丁 -->
    <template v-else-if="name === 'kungpao'">
      <path d="M6 4c-1 2-2 4-2 7 0 3 2 6 5 8"/>
      <path d="M18 4c1 2 2 4 2 7 0 3-2 6-5 8"/>
      <circle cx="12" cy="6" r="2"/>
      <path d="M12 12v4"/>
    </template>
    <!-- 回锅肉 -->
    <template v-else-if="name === 'pork'">
      <path d="M4 10c0-2 2-4 4-4h8c2 0 4 2 4 4v2c0 2-2 4-4 4H8c-2 0-4-2-4-4v-2z"/>
      <path d="M8 14v4M16 14v4"/>
    </template>
    <!-- 夫妻肺片 -->
    <template v-else-if="name === 'beef'">
      <path d="M4 8c0-2 2-3 4-3h8c2 0 4 1 4 3v4c0 2-2 3-4 3H8c-2 0-4-1-4-3V8z"/>
      <path d="M8 12l2-3 2 2 2-2 2 3"/>
    </template>
    <!-- 面条 -->
    <template v-else-if="name === 'noodles'">
      <path d="M4 5h16"/>
      <path d="M4 9h16"/>
      <path d="M4 13h16"/>
      <path d="M6 17h12"/>
      <path d="M8 5v14"/>
      <path d="M16 5v14"/>
    </template>
    <!-- 火焰 -->
    <template v-else-if="name === 'fire'">
      <path d="M12 2c-2 4-5 7-5 10a5 5 0 0 0 10 0c0-3-3-6-5-10z"/>
      <path d="M12 10c0 2 1 3 1 3"/>
    </template>
    <!-- 小摊 / 夜市 -->
    <template v-else-if="name === 'skewer'">
      <path d="M8 20l3-8"/>
      <path d="M16 20l-3-8"/>
      <circle cx="12" cy="4" r="2"/>
      <path d="M12 6v6"/>
    </template>
    <!-- 街区 / 门楼 -->
    <template v-else-if="name === 'hut'">
      <path d="M3 10l9-7 9 7"/>
      <path d="M5 10v10h14V10"/>
      <path d="M10 20v-6h4v6"/>
    </template>

    <!-- ──── 导航 / 通用 ──── -->
    <!-- 位置图钉 -->
    <template v-else-if="name === 'pin' || name === 'location'">
      <path d="M12 2C8 2 5 5 5 9c0 5 7 13 7 13s7-8 7-13c0-4-3-7-7-7z"/>
      <circle cx="12" cy="9" r="3"/>
    </template>
    <!-- 地图 -->
    <template v-else-if="name === 'map'">
      <path d="M3 7v12a1 1 0 0 0 1 1h4l3-4 3 4h4a1 1 0 0 0 1-1V7a1 1 0 0 0-1-1h-4l-3 4-3-4H4a1 1 0 0 0-1 1z"/>
      <path d="M12 6v12"/>
    </template>
    <!-- 编辑 / 铅笔 -->
    <template v-else-if="name === 'edit' || name === 'pencil'">
      <path d="M17 3a2.83 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/>
    </template>
    <!-- 门票 -->
    <template v-else-if="name === 'ticket'">
      <path d="M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v2z"/>
      <path d="M9 12h6"/>
      <path d="M10 9l2 3 2-3"/>
    </template>
    <!-- 时钟 / 营业时间 -->
    <template v-else-if="name === 'clock'">
      <circle cx="12" cy="12" r="9"/>
      <path d="M12 7v5l3.5 2"/>
    </template>
    <!-- 交通 / 公交 -->
    <template v-else-if="name === 'bus'">
      <rect x="4" y="3" width="16" height="14" rx="3"/>
      <path d="M4 11h16"/>
      <circle cx="8" cy="15" r="1"/>
      <circle cx="16" cy="15" r="1"/>
      <path d="M7 21v-2M17 21v-2M8 17l-1 2M16 17l1 2"/>
    </template>

    <!-- ──── 数据展示 ──── -->
    <!-- 卷轴 / 新闻 -->
    <template v-else-if="name === 'scroll'">
      <path d="M8 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2h-2"/>
      <rect x="8" y="3" width="8" height="6" rx="1"/>
      <path d="M8 12h8M8 15h6"/>
      <path d="M8 3v14"/>
    </template>
    <!-- 熊猫 -->
    <template v-else-if="name === 'panda'">
      <circle cx="8" cy="10" r="5"/>
      <circle cx="16" cy="10" r="5"/>
      <circle cx="5" cy="8" r="2"/>
      <circle cx="19" cy="8" r="2"/>
      <circle cx="8" cy="10" r="2" fill="currentColor"/>
      <circle cx="16" cy="10" r="2" fill="currentColor"/>
      <path d="M8 17c2 1.5 6 1.5 8 0"/>
    </template>
    <!-- 房屋 -->
    <template v-else-if="name === 'house'">
      <path d="M3 9.5L12 3l9 6.5"/>
      <path d="M5 9.5V20a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1V9.5"/>
      <path d="M10 21v-7h4v7"/>
    </template>

    <!-- 夜间 / 城市 -->
    <template v-else-if="name === 'night'">
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
      <path d="M3 18h18"/>
      <path d="M6 21h12"/>
    </template>

    <!-- 星级评分 -->
    <template v-else-if="name === 'star'">
      <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" :fill="filled ? 'currentColor' : 'none'"/>
    </template>

    <!-- 天气 / 晴 -->
    <template v-else-if="name === 'sunny'">
      <circle cx="12" cy="12" r="4.5"/>
      <path d="M12 1.5v3M12 19.5v3M4.2 4.2l2.1 2.1M17.7 17.7l2.1 2.1M1.5 12h3M19.5 12h3M4.2 19.8l2.1-2.1M17.7 6.3l2.1-2.1"/>
    </template>

    <!-- 默认占位 -->
    <template v-else>
      <circle cx="12" cy="12" r="9"/>
      <path d="M12 8v8M8 12h8"/>
    </template>
  </svg>
</template>