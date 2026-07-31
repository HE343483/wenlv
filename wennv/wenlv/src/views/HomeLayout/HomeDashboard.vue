<script setup lang="ts">
/**
 * HomeDashboard.vue — 内部首页仪表盘
 * 内容: 天气组件 + 区域选择 + 景点网格
 */
import { ref } from 'vue'
import { useLanguageStore } from '@/stores/language'
import WeatherRowCell from '@/components/WeatherRowCell.vue'
import DistrictSelector from '@/components/DistrictSelector.vue'
import ScenicGrid from '@/components/ScenicGrid.vue'

const langStore = useLanguageStore()
const selectedDistrict = ref('all')

function onDistrictSelect(id: string) {
  selectedDistrict.value = id
}
</script>

<template>
  <div class="dashboard">
    <div class="dashboard__weather">
      <WeatherRowCell />
    </div>

    <section class="dashboard__explore">
      <DistrictSelector @select="onDistrictSelect" />
      <ScenicGrid :district-id="selectedDistrict" />
    </section>

    <div class="dashboard__intro">
      <h2 class="section-title">{{ langStore.t('internal.title') }}</h2>
      <p class="section-subtitle">{{ langStore.t('internal.exploreDesc') }}</p>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: var(--max-width);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

.dashboard__weather {
  max-width: 480px;
}

.dashboard__intro {
  text-align: center;
  padding: var(--space-8) 0;
}

@media (max-width: 640px) {
  .dashboard__weather {
    max-width: 100%;
  }
}
</style>
