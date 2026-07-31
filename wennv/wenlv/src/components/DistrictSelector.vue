<script setup lang="ts">
/**
 * DistrictSelector.vue — 成都行政区选择器
 * 横向滚动式圆形选择器，每个区以圆形容器呈现
 * 蜀锦金 + 代表色光圈 标识选中状态
 */
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import { districts } from '@/data/chengdu'

const langStore = useLanguageStore()
const { lang } = storeToRefs(langStore)
const emit = defineEmits<{ select: [districtId: string] }>()

const selected = ref('all')

function select(id: string) {
  selected.value = id
  emit('select', id)
}
</script>

<template>
  <section class="district-selector">
    <!-- 标题 -->
    <div class="district-selector__header">
      <h2 class="section-title">{{ langStore.t('district.title') }}</h2>
      <p class="section-subtitle">{{ langStore.t('district.subtitle') }}</p>
    </div>

    <!-- 选择器轮播 -->
    <div class="district-selector__carousel">
      <div class="district-selector__track">
        <!-- 全部选项 -->
        <button
          class="district-pill"
          :class="{ 'district-pill--active': selected === 'all' }"
          @click="select('all')"
        >
          <div class="district-pill__ring" style="--ring-color: #C9A96E">
            <span class="district-pill__icon">蜀</span>
          </div>
          <span class="district-pill__name">{{ langStore.t('district.allDistricts') }}</span>
        </button>

        <!-- 各区 -->
        <button
          v-for="d in districts"
          :key="d.id"
          class="district-pill"
          :class="{ 'district-pill--active': selected === d.id }"
          @click="select(d.id)"
        >
          <div
            class="district-pill__ring"
            :style="{
              '--ring-color': d.color,
              '--ring-glow': d.color + '33',
            }"
          >
            <span class="district-pill__icon">{{ d.nameZh.charAt(0) }}</span>
          </div>
          <span class="district-pill__name">{{ lang === 'zh' ? d.nameZh : d.nameEn }}</span>
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.district-selector {
  padding: var(--space-16) 0 var(--space-8);
}

.district-selector__header {
  text-align: center;
  margin-bottom: var(--space-10);
}

.district-selector__carousel {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  padding-bottom: var(--space-4);
}

.district-selector__carousel::-webkit-scrollbar {
  display: none;
}

.district-selector__track {
  display: flex;
  gap: var(--space-5);
  padding: 0 var(--space-4);
  justify-content: center;
  flex-wrap: wrap;
}

/* 单个区圆形容器 */
.district-pill {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  cursor: pointer;
  transition: transform var(--transition-fast);
  padding: var(--space-2);
  border: none;
  background: none;
  color: inherit;
  min-width: 72px;
}

.district-pill:hover {
  transform: translateY(-2px);
}

.district-pill__ring {
  width: 60px;
  height: 60px;
  border-radius: var(--radius-full);
  border: 2px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-base);
  position: relative;
}

.district-pill__ring::before {
  content: '';
  position: absolute;
  inset: -3px;
  border-radius: inherit;
  border: 1.5px solid transparent;
  transition: all var(--transition-base);
}

.district-pill--active .district-pill__ring {
  border-color: var(--ring-color, var(--color-gold));
  box-shadow: 0 0 24px var(--ring-glow, var(--color-gold-glow));
  background: rgba(201, 169, 110, 0.06);
}

.district-pill__icon {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-text-muted);
  transition: color var(--transition-fast);
}

.district-pill--active .district-pill__icon {
  color: var(--ring-color, var(--color-gold));
}

.district-pill__name {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
  transition: color var(--transition-fast);
}

.district-pill--active .district-pill__name {
  color: var(--ring-color);
}

@media (max-width: 768px) {
  .district-selector__track {
    justify-content: flex-start;
    flex-wrap: nowrap;
  }
  .district-pill__ring {
    width: 52px;
    height: 52px;
  }
}
</style>
