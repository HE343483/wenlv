<script setup lang="ts">
/**
 * FavoritesPage.vue — 收藏独立页
 * 收藏列表来自后端 /favorites 接口（景点 scenic / 美食 food / 路线 route），按类型分区展示；
 * 详情按收藏的数字ID逐个拉取后端数据后渲染，支持取消收藏与跳转各自详情页。
 */
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import { useUserStore, type FavoriteType } from '@/stores/user'
import {
  getFood,
  getScenic,
  listRoutes,
  parseRouteStops,
  type FoodItem,
  type RouteItem,
  type ScenicItem,
} from '@/api/content'
import { districts } from '@/data/chengdu'
import AppIcon from '@/components/AppIcon.vue'
import HomeBanner from '@/components/HomeBanner.vue'
import type { ScenicSpot } from '@/types'
import { pickDesc, pickName } from '@/utils/storyI18n'

const router = useRouter()
const langStore = useLanguageStore()
const userStore = useUserStore()
const { lang } = storeToRefs(langStore)
const { favorites } = storeToRefs(userStore)

/** 收藏ID集合（形如 scenic-12 / food-3 / route-1） */
const favSet = computed(() => new Set(favorites.value))

/** 从收藏ID列表中筛出指定类型的后端数字ID */
function numericIds(type: FavoriteType): number[] {
  const re = new RegExp(`^${type}-(\\d+)$`)
  return favorites.value.flatMap(id => {
    const m = re.exec(id)
    return m ? [Number(m[1])] : []
  })
}

/* ── 已收藏条目详情（登录后同步收藏，再按数字ID拉取） ── */
const loading = ref(true)
const allScenics = ref<ScenicSpot[]>([])
const allFoods = ref<FoodItem[]>([])
const allRoutes = ref<RouteItem[]>([])

/** 后端 ScenicItem → 前端 ScenicSpot 适配（与探索页一致） */
function mapScenic(item: ScenicItem): ScenicSpot {
  const d = districts.find(x => x.nameZh === item.district)
  const desc = item.desc || ''
  return {
    id: `scenic-${item.id}`,
    districtId: d?.id ?? '',
    nameZh: item.name_zh,
    nameEn: pickName(item, langStore.lang),
    shortDescZh: desc.length > 30 ? desc.slice(0, 30) + '…' : desc,
    shortDescEn: desc.length > 60 ? desc.slice(0, 60) + '…' : desc,
    descriptionZh: desc,
    descriptionEn: desc,
    tags: (item.tags || '').split(',').filter(Boolean),
    imageUrl: item.images || '',
    rating: item.score || 0,
    coords: { lng: item.lng || 0, lat: item.lat || 0 },
  }
}

onMounted(async () => {
  try {
    // 先进主框架时已同步过一次，这里再确保一次，避免收藏页先于同步拿到空列表
    await userStore.syncFavorites()

    const scenicIds = numericIds('scenic')
    const foodIds = numericIds('food')
    const routeIds = new Set(numericIds('route'))

    if (scenicIds.length) {
      const res = await Promise.allSettled(scenicIds.map(id => getScenic(id)))
      allScenics.value = res.flatMap(r => (r.status === 'fulfilled' ? [mapScenic(r.value)] : []))
    }
    if (foodIds.length) {
      const res = await Promise.allSettled(foodIds.map(id => getFood(id)))
      allFoods.value = res.flatMap(r => (r.status === 'fulfilled' ? [r.value] : []))
    }
    if (routeIds.size) {
      const res = await listRoutes({ page_size: 100 })
      allRoutes.value = (res.items || []).filter(r => routeIds.has(r.id))
    }
  } catch {
    /* 拉取失败保持空列表 */
  } finally {
    loading.value = false
  }
})

/* ── 展示列表：按收藏状态实时过滤，取消收藏后卡片立即消失 ── */
const scenicItems = computed(() => allScenics.value.filter(s => favSet.value.has(s.id)))
const foodItems = computed(() => allFoods.value.filter(f => favSet.value.has(`food-${f.id}`)))
const routeItems = computed(() => allRoutes.value.filter(r => favSet.value.has(`route-${r.id}`)))

const total = computed(() => scenicItems.value.length + foodItems.value.length + routeItems.value.length)
const isEmpty = computed(() => !loading.value && total.value === 0)

/* 图片加载失败时回退到首字水印占位 */
const imgFailed = reactive(new Set<string>())
function onMediaError(key: string) {
  imgFailed.add(key)
}

function spotName(s: ScenicSpot): string {
  return lang.value === 'zh' ? s.nameZh : s.nameEn
}
function spotDesc(s: ScenicSpot): string {
  return lang.value === 'zh' ? s.shortDescZh : s.shortDescEn
}

/** 路线标题/简介按语言取值（路线字段为 title_zh/description，与 pickName/pickDesc 约定不同） */
function routeTitle(r: RouteItem): string {
  if (lang.value === 'en') return r.title_en || r.title_zh
  if (lang.value === 'ja') return r.title_ja || r.title_zh
  return r.title_zh
}
function routeDesc(r: RouteItem): string {
  const zh = r.description || ''
  if (lang.value === 'zh') return zh
  if (lang.value === 'en') return r.description_en || zh
  return r.description_ja || zh
}
function foodTags(f: FoodItem): string[] {
  return (f.tags || '').split(',').filter(Boolean).slice(0, 3)
}

/** ── 跳转各自详情页 ── */
function goScenicDetail(id: string) {
  router.push({ name: 'scenic-detail', params: { id } })
}
function goFoodDetail(id: number) {
  router.push({ name: 'food-detail', params: { id: String(id) } })
}
/** 路线无独立详情路由，跳路线页并由 ?route= 自动打开该路线详情弹窗 */
function goRouteDetail(routeKey: string) {
  router.push({ name: 'internal-routes', query: { route: routeKey } })
}

async function remove(id: string) {
  try {
    await userStore.removeFavorite(id)
  } catch (err) {
    alert(err instanceof Error ? err.message : '操作失败')
  }
}
</script>

<template>
  <div class="favorites-page">
    <HomeBanner
      :eyebrow="langStore.lang === 'zh' ? '蜀韵·成都' : 'Shu·Chengdu'"
      :title="langStore.t('favorites.title')"
      :subtitle="langStore.t('favorites.subtitle')"
      watermark="藏"
    />

    <div v-if="total" class="favorites-page__count">
      {{ total }} {{ langStore.t('favorites.countLabel') }}
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="favorites-loading">{{ langStore.t('scenic.loading') }}</div>

    <!-- 空状态 -->
    <div v-else-if="isEmpty" class="favorites-empty">
      <div class="favorites-empty__icon" aria-hidden="true">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
          <path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" />
        </svg>
      </div>
      <h2 class="favorites-empty__title">{{ langStore.t('favorites.empty') }}</h2>
      <p class="favorites-empty__hint">{{ langStore.t('favorites.emptyHint') }}</p>
      <button class="favorites-empty__btn" @click="router.push('/home/explore')">{{ langStore.t('favorites.goExplore') }}</button>
    </div>

    <template v-else>
      <!-- ──── 景点 ──── -->
      <section v-if="scenicItems.length" class="fav-section">
        <header class="fav-section__head">
          <h2 class="fav-section__title">{{ langStore.t('favorites.sectionScenic') }}</h2>
          <span class="fav-section__count">{{ scenicItems.length }}</span>
        </header>
        <div class="favorites-grid">
          <article v-for="spot in scenicItems" :key="spot.id" class="fav-card" tabindex="0" @click="goScenicDetail(spot.id)" @keyup.enter="goScenicDetail(spot.id)">
            <div class="fav-card__media">
              <img
                v-if="spot.imageUrl && !imgFailed.has(spot.id)"
                class="fav-card__photo"
                :src="spot.imageUrl"
                :alt="spotName(spot)"
                loading="lazy"
                referrerpolicy="no-referrer"
                @error="onMediaError(spot.id)"
              />
              <span v-else class="fav-card__char">{{ spot.nameZh.charAt(0) }}</span>
              <span class="fav-card__rating"><AppIcon name="star" :size="11" filled /> {{ spot.rating }}</span>
              <button class="fav-card__fav-btn" :title="langStore.t('favorites.unfavorite')" @click.stop="remove(spot.id)">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"><path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" /></svg>
              </button>
            </div>
            <div class="fav-card__body">
              <h3 class="fav-card__title">{{ spotName(spot) }}</h3>
              <p class="fav-card__desc">{{ spotDesc(spot) }}</p>
              <div class="fav-card__tags">
                <span v-for="tag in spot.tags.slice(0, 3)" :key="tag" class="fav-card__tag">{{ tag }}</span>
              </div>
              <div class="fav-card__footer">
                <button class="fav-card__link" @click.stop="goScenicDetail(spot.id)">{{ langStore.t('favorites.viewDetail') }} →</button>
              </div>
            </div>
          </article>
        </div>
      </section>

      <!-- ──── 美食 ──── -->
      <section v-if="foodItems.length" class="fav-section">
        <header class="fav-section__head">
          <h2 class="fav-section__title">{{ langStore.t('favorites.sectionFood') }}</h2>
          <span class="fav-section__count">{{ foodItems.length }}</span>
        </header>
        <div class="favorites-grid">
          <article
            v-for="f in foodItems"
            :key="f.id"
            class="fav-card"
            tabindex="0"
            @click="goFoodDetail(f.id)"
            @keyup.enter="goFoodDetail(f.id)"
          >
            <div class="fav-card__media">
              <img
                v-if="f.images && !imgFailed.has(`food-${f.id}`)"
                class="fav-card__photo"
                :src="f.images"
                :alt="pickName(f, langStore.lang)"
                loading="lazy"
                referrerpolicy="no-referrer"
                @error="onMediaError(`food-${f.id}`)"
              />
              <span v-else class="fav-card__char">{{ f.name_zh.charAt(0) }}</span>
              <button class="fav-card__fav-btn" :title="langStore.t('favorites.unfavorite')" @click.stop="remove(`food-${f.id}`)">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"><path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" /></svg>
              </button>
            </div>
            <div class="fav-card__body">
              <h3 class="fav-card__title">{{ pickName(f, langStore.lang) }}</h3>
              <p class="fav-card__desc">{{ pickDesc(f, langStore.lang) }}</p>
              <div class="fav-card__tags">
                <span v-for="tag in foodTags(f)" :key="tag" class="fav-card__tag">{{ tag }}</span>
              </div>
              <div class="fav-card__footer">
                <button class="fav-card__link" @click.stop="goFoodDetail(f.id)">{{ langStore.t('favorites.viewDetail') }} →</button>
              </div>
            </div>
          </article>
        </div>
      </section>

      <!-- ──── 路线 ──── -->
      <section v-if="routeItems.length" class="fav-section">
        <header class="fav-section__head">
          <h2 class="fav-section__title">{{ langStore.t('favorites.sectionRoute') }}</h2>
          <span class="fav-section__count">{{ routeItems.length }}</span>
        </header>
        <div class="favorites-grid">
          <article
            v-for="r in routeItems"
            :key="r.id"
            class="fav-card"
            tabindex="0"
            @click="goRouteDetail(r.route_key)"
            @keyup.enter="goRouteDetail(r.route_key)"
          >
            <div class="fav-card__media">
              <img
                v-if="r.cover_image && !imgFailed.has(`route-${r.id}`)"
                class="fav-card__photo"
                :src="r.cover_image"
                :alt="routeTitle(r)"
                loading="lazy"
                referrerpolicy="no-referrer"
                @error="onMediaError(`route-${r.id}`)"
              />
              <span v-else class="fav-card__char">{{ r.title_zh.charAt(0) }}</span>
              <span class="fav-card__rating">{{ r.days || 1 }} {{ langStore.t('routes.daysUnit') }}</span>
              <button class="fav-card__fav-btn" :title="langStore.t('favorites.unfavorite')" @click.stop="remove(`route-${r.id}`)">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"><path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" /></svg>
              </button>
            </div>
            <div class="fav-card__body">
              <h3 class="fav-card__title">{{ routeTitle(r) }}</h3>
              <p class="fav-card__desc">{{ routeDesc(r) }}</p>
              <div class="fav-card__tags">
                <span class="fav-card__tag">{{ langStore.t('favorites.stopCount', { n: parseRouteStops(r.stops).length }) }}</span>
              </div>
              <div class="fav-card__footer">
                <button class="fav-card__link" @click.stop="goRouteDetail(r.route_key)">{{ langStore.t('favorites.viewDetail') }} →</button>
              </div>
            </div>
          </article>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.favorites-page {
  max-width: var(--max-width);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}
.favorites-page__count {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}
.favorites-loading {
  padding: var(--space-12) 0;
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}
.favorites-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-12) var(--space-6);
  background: var(--color-surface);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-xl);
  text-align: center;
}
.favorites-empty__icon { color: var(--color-gold); opacity: 0.7; }
.favorites-empty__title { font-family: var(--font-display); font-size: var(--text-xl); color: var(--color-text-primary); }
.favorites-empty__hint { font-size: var(--text-sm); color: var(--color-text-muted); }
.favorites-empty__btn {
  margin-top: var(--space-2);
  padding: var(--space-3) var(--space-6);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  transition: all var(--transition-fast);
}
.favorites-empty__btn:hover { background: var(--color-gold-light); box-shadow: var(--shadow-gold); }

/* 类型分区（景点 / 美食 / 路线） */
.fav-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.fav-section__head {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--color-border);
}
.fav-section__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}
.fav-section__count {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  color: var(--color-gold);
  border: 1px solid color-mix(in srgb, var(--color-gold) 35%, transparent);
  border-radius: var(--radius-full);
  padding: 1px 10px;
}

.favorites-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}
.fav-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  cursor: pointer;
  transition: all var(--transition-base);
  outline: none;
}
.fav-card:hover, .fav-card:focus-visible { border-color: var(--color-gold-dark); transform: translateY(-3px); box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg); }
.fav-card__media {
  position: relative;
  aspect-ratio: 16 / 10;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(ellipse 60% 60% at 50% 30%, color-mix(in srgb, var(--color-gold) 10%, transparent), transparent 70%), var(--color-bg-alt);
}
.fav-card__photo {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.fav-card__char { font-family: var(--font-display); font-size: 72px; font-weight: 900; color: transparent; -webkit-text-stroke: 1px color-mix(in srgb, var(--color-gold) 40%, transparent); user-select: none; }
.fav-card__rating {
  position: absolute; left: var(--space-3); top: var(--space-3);
  font-size: var(--text-xs); color: var(--color-gold);
  background: color-mix(in srgb, var(--color-bg) 70%, transparent);
  backdrop-filter: blur(6px); border: 1px solid color-mix(in srgb, var(--color-gold) 35%, transparent);
  border-radius: var(--radius-full); padding: 2px 10px;
  display: inline-flex; align-items: center; gap: 4px;
}
.fav-card__fav-btn {
  position: absolute; right: var(--space-3); top: var(--space-3);
  width: 34px; height: 34px; display: flex; align-items: center; justify-content: center;
  border-radius: var(--radius-full); color: var(--color-gold);
  background: color-mix(in srgb, var(--color-bg) 70%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-gold) 40%, transparent);
  backdrop-filter: blur(6px); transition: all var(--transition-fast);
}
.fav-card__fav-btn:hover { background: color-mix(in srgb, var(--color-gold) 12%, transparent); }
.fav-card__body { padding: var(--space-5); display: flex; flex-direction: column; gap: var(--space-3); flex: 1; }
.fav-card__title { font-family: var(--font-display); font-size: var(--text-lg); font-weight: 600; color: var(--color-text-primary); }
.fav-card__desc { font-size: var(--text-sm); color: var(--color-text-secondary); line-height: var(--leading-relaxed); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.fav-card__tags { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.fav-card__tag { font-size: var(--text-xs); padding: 2px 10px; border: 1px solid var(--color-border); border-radius: var(--radius-full); color: var(--color-text-muted); }
.fav-card__footer { margin-top: auto; padding-top: var(--space-3); border-top: 1px solid var(--color-border); }
.fav-card__link { font-size: var(--text-sm); color: var(--color-text-muted); transition: color var(--transition-fast); }
.fav-card__link:hover { color: var(--color-gold); }

@media (max-width: 1024px) { .favorites-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 640px) { .favorites-grid { grid-template-columns: 1fr; } }
</style>
