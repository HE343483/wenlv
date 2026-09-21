<script setup lang="ts">
/**
 * FoodPage.vue — 成都美食
 * 功能：美食名片卡片网格 + 川菜脉络时间线 + 经典名菜/老字号展示
 * 数据来源：本地化 locales + 内联数据，后期可接入后端
 */
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useUserStore } from '@/stores/user'
import { listFoods, listFoodCards } from '@/api/content'
import type { FoodItem } from '@/api/content'
import { pickDesc, pickName } from '@/utils/storyI18n'
import AppIcon from '@/components/AppIcon.vue'
import HomeBanner from '@/components/HomeBanner.vue'

const langStore = useLanguageStore()
const userStore = useUserStore()
const router = useRouter()

/* 跳转美食详情页（key 作为 :id） */
function goFoodDetail(key: string) {
  router.push({ name: 'food-detail', params: { id: key } })
}

/* ── 图鉴菜品收藏(后端 target_type=food) ── */
function isFoodFav(id: number): boolean {
  return userStore.isFavorite(`food-${id}`)
}

function toggleFoodFav(id: number) {
  userStore.toggleFavorite(`food-${id}`).catch((err: unknown) => {
    // 接口失败已回滚本地状态，此处提示用户
    alert(err instanceof Error ? err.message : '操作失败')
  })
}

/* ── 成都味道图鉴(后端接口数据,爬虫自动维护) ── */
const foods = ref<FoodItem[]>([])
const foodsLoading = ref(false)

async function loadFoods() {
  foodsLoading.value = true
  try {
    const res = await listFoods({ page_size: 50 })
    foods.value = res.items
  } finally {
    foodsLoading.value = false
  }
}
onMounted(() => {
  loadFoods()
  loadFoodCards()
})

/* ── 美食名片配图(后端 food_cards 表,爬虫抓取并上传 OSS) ── */
const cardImages = ref<Record<string, string>>({})

async function loadFoodCards() {
  try {
    const cards = await listFoodCards()
    cardImages.value = Object.fromEntries(cards.map(c => [c.card_key, c.image || '']))
  } catch {
    /* 加载失败时名片回退为图标展示 */
  }
}

const foodSectionTitle = computed(() => {
  const map: Record<string, string> = { zh: '成都味道图鉴', en: 'Taste of Chengdu', ja: '成都グルメ図鑑' }
  return map[langStore.lang] || map.zh
})
const foodSectionSub = computed(() => {
  const map: Record<string, string> = {
    zh: '经典川味与街头小吃,点击查看详细介绍',
    en: 'Classic Sichuan flavors & street snacks — tap for details',
    ja: '四川の定番味と街角グルメ、タップで詳細を見る',
  }
  return map[langStore.lang] || map.zh
})
const foodsLoadingText = computed(() => {
  const map: Record<string, string> = { zh: '正在加载美食数据…', en: 'Loading dishes…', ja: 'グルメデータを読み込み中…' }
  return map[langStore.lang] || map.zh
})

/** 简介截断为卡片一行文案 */
function foodBrief(desc: string | undefined): string {
  const d = desc || ''
  return d.length > 42 ? d.slice(0, 42) + '…' : d
}

/** 名片配图:按卡片标识取数据库中的 OSS 图片,取不到时回退为图标展示 */
function cardImage(key: string): string {
  return cardImages.value[key] || ''
}

/* ── 美食名片 ── */
interface FoodCard {
  key: string
  icon: string
  /** 代表菜名(name_zh),点击名片时按此名在已加载菜品中精确匹配数据库 id */
  nameZh: string
}
const foodCards: FoodCard[] = [
  { key: 'hotpot', icon: 'hotpot', nameZh: '火锅' },
  { key: 'chuanchuan', icon: 'chuanchuan', nameZh: '串串香' },
  { key: 'cuisine', icon: 'cuisine', nameZh: '麻婆豆腐' },
  { key: 'snacks', icon: 'dumpling', nameZh: '担担面' },
  { key: 'tea', icon: 'tea', nameZh: '盖碗茶' },
  { key: 'nightfood', icon: 'moon', nameZh: '兔头' },
]

/** 大类卡片：川菜 / 名小吃 / 夜宵 介绍的是"一类",直接进类别页(不解析菜品 id) */
const CATEGORY_CARDS = new Set(['cuisine', 'snacks', 'nightfood'])

/** 名片跳转进行中标志：防止连点造成多次请求/跳转 */
let cardNavigating = false

/**
 * 点击名片：
 * - cuisine/snacks/nightfood 为大类，直接跳 /food/{key} 类别页；
 * - 其余(火锅/串串香/盖碗茶)先确保菜品列表就绪(复用「成都味道图鉴」的 loadFoods)，
 *   再按代表菜名匹配真实菜品 id；确实匹配不到时回退分类页(key)。
 */
async function goCardDetail(card: FoodCard) {
  if (cardNavigating) return
  cardNavigating = true
  try {
    if (CATEGORY_CARDS.has(card.key)) {
      await router.push({ name: 'food-detail', params: { id: card.key } })
      return
    }
    if (foods.value.length === 0) {
      try {
        await loadFoods()
      } catch {
        /* 列表加载异常：走下方兜底逻辑，不向控制台抛错 */
      }
    }
    const found = foods.value.find((f) => f.name_zh === card.nameZh)
    if (!found) {
      console.warn(`[FoodPage] 未匹配到菜品「${card.nameZh}」，回退分类页 /food/${card.key}`)
    }
    await router.push({ name: 'food-detail', params: { id: found ? String(found.id) : card.key } })
  } finally {
    cardNavigating = false
  }
}

/* 统计数值 */
const stats = [
  { key: 'restaurants', value: langStore.lang === 'zh' ? '16万+' : '160K+', label: langStore.t('food.stats.restaurants') },
  { key: 'snacks', value: langStore.lang === 'zh' ? '200+' : '200+', label: langStore.t('food.stats.snacks') },
  { key: 'heritage', value: '89', label: langStore.t('food.stats.heritage') },
  { key: 'streets', value: langStore.lang === 'zh' ? '30+' : '30+', label: langStore.t('food.stats.streets') },
]

/* ── 川菜脉络时间线 ── */
const timeline = [
  {
    eraZh: '古蜀风味',
    eraEn: 'Ancient Shu Tastes',
    periodZh: '约3000年前',
    periodEn: 'c. 3000 years ago',
    descZh: '成都平原物产丰饶，"沃野千里，号为陆海"。从宝墩、三星堆的陶器中，已能窥见古蜀先民"尚滋味"的饮食传统。',
    descEn: 'The fertile Chengdu Plain was called a "land of plenty." Pottery from Sanxingdui hints at ancient Shu\'s love of flavor.',
  },
  {
    eraZh: '汉晋定型',
    eraEn: 'Han & Jin',
    periodZh: '公元前316年—南北朝',
    periodEn: '316 BC – Southern Dynasties',
    descZh: '秦汉以来成都即为西南饮食都会，左思《蜀都赋》"调夫五味，甘甜之和"，川菜"尚辛香、好滋味"的底色初步奠定。',
    descEn: 'Chengdu became the culinary capital of the southwest. Zuo Si\'s "Ode to the Shu Capital" already praised its five-flavor mastery.',
  },
  {
    eraZh: '明清川味',
    eraEn: 'Ming & Qing',
    periodZh: '辣椒传入后',
    periodEn: 'After chili arrived',
    descZh: '明代辣椒经海上丝路传入中国并进入四川，"湖广填四川"带来移民与食材大融合，麻与辣自此成为川菜的味觉灵魂。',
    descEn: 'Chili peppers arrived via the maritime Silk Road. Mass immigration brought ingredients together — numbing and spicy became the soul of Sichuan cuisine.',
  },
  {
    eraZh: '近代菜系',
    eraEn: 'Modern School',
    periodZh: '清末民初',
    periodEn: 'Late Qing – Republic',
    descZh: '上河帮、小河帮、下河帮三大流派格局形成，宫保鸡丁、回锅肉等经典定型；成都小吃遍地开花，奠定"世界美食之都"基石。',
    descEn: 'Three regional schools took shape and classics like Kung Pao Chicken were born, while Chengdu snacks blossomed across the city.',
  },
  {
    eraZh: '今日天府',
    eraEn: 'Tianfu Today',
    periodZh: '当代',
    periodEn: 'Present',
    descZh: '2010年成都入选联合国教科文组织"世界美食之都"。火锅沸腾、川菜出海，麻辣鲜香成为世界认识中国的重要味觉名片。',
    descEn: 'In 2010, Chengdu was named a UNESCO City of Gastronomy. Boiling hotpot and global Sichuan cuisine put Chengdu on the world\'s palate.',
  },
]

/* ── 经典名菜 / 老字号 ── */
const signatureDishes = [
  {
    icon: 'mapo',
    nameZh: '麻婆豆腐',
    nameEn: 'Mapo Tofu',
    tagZh: '川菜之魂',
    tagEn: 'Soul of Sichuan Cuisine',
  },
  {
    icon: 'kungpao',
    nameZh: '宫保鸡丁',
    nameEn: 'Kung Pao Chicken',
    tagZh: '国际名菜',
    tagEn: 'World-famous Dish',
  },
  {
    icon: 'pork',
    nameZh: '回锅肉',
    nameEn: 'Twice-Cooked Pork',
    tagZh: '川菜第一菜',
    tagEn: 'No.1 Sichuan Dish',
  },
  {
    icon: 'beef',
    nameZh: '夫妻肺片',
    nameEn: 'Fuqi Feipian',
    tagZh: '经典凉菜',
    tagEn: 'Classic Cold Dish',
  },
  {
    icon: 'noodles',
    nameZh: '担担面',
    nameEn: 'Dan Dan Noodles',
    tagZh: '成都名小吃',
    tagEn: 'Chengdu Snack Icon',
  },
  {
    icon: 'dumpling',
    nameZh: '钟水饺',
    nameEn: 'Zhong Dumplings',
    tagZh: '中华老字号',
    tagEn: 'China Time-honored Brand',
  },
]

/* ── 美食街区 ── */
const foodStreets = [
  {
    icon: 'spicy',
    nameZh: '锦里',
    nameEn: 'Jinli',
    tagZh: '川味小吃街',
    tagEn: 'Sichuan Snack Street',
  },
  {
    icon: 'fire',
    nameZh: '玉林路',
    nameEn: 'Yulin Road',
    tagZh: '火锅串串天堂',
    tagEn: 'Hotpot Paradise',
  },
  {
    icon: 'skewer',
    nameZh: '建设路',
    nameEn: 'Jianshe Road',
    tagZh: '网红小吃街',
    tagEn: 'Trending Snack Street',
  },
  {
    icon: 'hotpot',
    nameZh: '奎星楼街',
    nameEn: 'Kuixinglou Street',
    tagZh: '苍蝇馆子聚集地',
    tagEn: 'Hidden Gem Eateries',
  },
  {
    icon: 'night',
    nameZh: '九眼桥',
    nameEn: 'Jiuyan Bridge',
    tagZh: '夜宵酒吧街',
    tagEn: 'Night Snack & Bar St.',
  },
  {
    icon: 'hut',
    nameZh: '文殊坊',
    nameEn: 'Wenshu Fang',
    tagZh: '禅意美食街区',
    tagEn: 'Zen Food District',
  },
]
</script>

<template>
  <div class="food-page">
    <!-- ============================================
         HERO
         ============================================ -->
    <HomeBanner
      :eyebrow="langStore.t('food.badge')"
      :title="langStore.t('food.title')"
      :en-title="langStore.t('food.titleEn')"
      :subtitle="langStore.t('food.subtitle')"
      watermark="味"
    />

    <!-- ============================================
         简介 + 数据统计
         ============================================ -->
    <section class="food-intro container">
      <div class="food-intro__body">
        <p class="food-intro__text">{{ langStore.t('food.intro') }}</p>
      </div>

      <div class="food-stats">
        <div v-for="stat in stats" :key="stat.key" class="food-stats__item">
          <span class="food-stats__num">{{ stat.value }}</span>
          <span class="food-stats__label">{{ stat.label }}</span>
        </div>
      </div>
    </section>

    <!-- ============================================
         美食名片 — 六大维度
         ============================================ -->
    <section class="food-cards-section container">
      <header class="food-section-head">
        <h2 class="section-title">{{ langStore.t('food.cardsTitle') }}</h2>
        <p class="section-subtitle">{{ langStore.t('food.cardsSubtitle') }}</p>
      </header>

      <div class="food-cards">
        <article
          v-for="card in foodCards"
          :key="card.key"
          class="food-card"
          tabindex="0"
          @click="goCardDetail(card)"
          @keyup.enter="goCardDetail(card)"
        >
          <div class="food-card__media">
            <!-- 数据库名片配图,取不到时回退为图标 -->
            <img
              v-if="cardImage(card.key)"
              class="food-card__photo"
              :src="cardImage(card.key)"
              :alt="langStore.t(`food.card.${card.key}.name`)"
              loading="lazy"
              referrerpolicy="no-referrer"
            />
            <div v-else class="food-card__icon">
              <AppIcon :name="card.icon" :size="36" />
            </div>
          </div>
          <h3 class="food-card__name">
            {{ langStore.t(`food.card.${card.key}.name`) }}
          </h3>
          <p class="food-card__en-name">
            {{ langStore.t(`food.card.${card.key}.en`) }}
          </p>
          <p class="food-card__desc">
            {{ langStore.t(`food.card.${card.key}.desc`) }}
          </p>
          <div class="food-card__footer">
            <span class="food-card__action">
              {{ langStore.t('food.exploreBtn') }}
              <span class="food-card__arrow">→</span>
            </span>
          </div>
        </article>
      </div>
    </section>

    <!-- ============================================
         成都味道图鉴 — 接口数据照片卡片
         ============================================ -->
    <section class="food-grid-section container">
      <header class="food-section-head">
        <h2 class="section-title">{{ foodSectionTitle }}</h2>
        <p class="section-subtitle">{{ foodSectionSub }}</p>
      </header>

      <div v-if="foodsLoading" class="food-grid__status">{{ foodsLoadingText }}</div>
      <div v-else-if="foods.length === 0" class="food-grid__status">
        {{ langStore.lang === 'zh' ? '暂无美食数据' : 'No dishes yet' }}
      </div>
      <div v-else class="food-grid">
        <article
          v-for="f in foods"
          :key="f.id"
          class="food-grid__card"
          tabindex="0"
          @click="goFoodDetail(String(f.id))"
          @keyup.enter="goFoodDetail(String(f.id))"
        >
          <div class="food-grid__media">
            <img
              v-if="f.images"
              class="food-grid__photo"
              :src="f.images"
              :alt="pickName(f, langStore.lang)"
              loading="lazy"
              referrerpolicy="no-referrer"
              @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
            <span v-if="f.tags" class="food-grid__tag">{{ (f.tags || '').split(',')[0] }}</span>
            <button
              type="button"
              class="food-grid__fav"
              :class="{ 'food-grid__fav--active': isFoodFav(f.id) }"
              :title="isFoodFav(f.id) ? langStore.t('scenic.favorited') : langStore.t('scenic.favorite')"
              :aria-label="isFoodFav(f.id) ? langStore.t('scenic.favorited') : langStore.t('scenic.favorite')"
              @click.stop="toggleFoodFav(f.id)"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round">
                <path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" :fill="isFoodFav(f.id) ? 'currentColor' : 'none'"/>
              </svg>
            </button>
          </div>
          <div class="food-grid__body">
            <h3 class="food-grid__name">
              {{ pickName(f, langStore.lang) }}
            </h3>
            <p v-if="langStore.lang !== 'zh' && f.name_zh" class="food-grid__en">{{ f.name_zh }}</p>
            <p class="food-grid__desc">{{ foodBrief(pickDesc(f, langStore.lang)) }}</p>
          </div>
        </article>
      </div>
    </section>

    <!-- ============================================
         川菜脉络 — 时间线
         ============================================ -->
    <section class="food-timeline-section">
      <div class="container">
        <header class="food-section-head">
          <h2 class="section-title">{{ langStore.t('food.timelineTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('food.timelineSubtitle') }}</p>
        </header>

        <div class="food-timeline">
          <div
            v-for="(era, i) in timeline"
            :key="i"
            class="timeline-era"
            :class="{ 'timeline-era--alt': i % 2 === 1 }"
          >
            <div class="timeline-era__dot" />
            <div class="timeline-era__card">
              <div class="timeline-era__header">
                <h3 class="timeline-era__title">
                  {{ langStore.lang === 'zh' ? era.eraZh : era.eraEn }}
                </h3>
                <span class="timeline-era__period">
                  {{ langStore.lang === 'zh' ? era.periodZh : era.periodEn }}
                </span>
              </div>
              <p class="timeline-era__desc">
                {{ langStore.lang === 'zh' ? era.descZh : era.descEn }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ============================================
         经典名菜 + 美食街区
         ============================================ -->
    <section class="food-dishes-section container">
      <header class="food-section-head">
        <h2 class="section-title">{{ langStore.t('food.dishesTitle') }}</h2>
        <p class="section-subtitle">{{ langStore.t('food.dishesSubtitle') }}</p>
      </header>

      <div class="food-dishes">
        <div
          v-for="item in signatureDishes"
          :key="item.nameZh"
          class="dish-item"
        >
          <span class="dish-item__icon">
            <AppIcon :name="item.icon" :size="28" />
          </span>
          <div class="dish-item__body">
            <h3 class="dish-item__name">
              {{ langStore.lang === 'zh' ? item.nameZh : item.nameEn }}
            </h3>
            <span class="dish-item__tag">
              {{ langStore.lang === 'zh' ? item.tagZh : item.tagEn }}
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- ============================================
         美食街区
         ============================================ -->
    <section class="food-streets-section container">
      <header class="food-section-head">
        <h2 class="section-title">{{ langStore.t('food.streetsTitle') }}</h2>
        <p class="section-subtitle">{{ langStore.t('food.streetsSubtitle') }}</p>
      </header>

      <div class="food-streets">
        <article
          v-for="street in foodStreets"
          :key="street.nameZh"
          class="street-card"
        >
          <span class="street-card__icon" aria-hidden="true">
            <AppIcon :name="street.icon" :size="28" />
          </span>
          <h3 class="street-card__title">
            {{ langStore.lang === 'zh' ? street.nameZh : street.nameEn }}
          </h3>
          <p class="street-card__tag">
            {{ langStore.lang === 'zh' ? street.tagZh : street.tagEn }}
          </p>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ========================================
   区块统一头部
   ======================================== */
.food-section-head {
  text-align: center;
  margin-bottom: var(--space-10);
}

/* ========================================
   简介 + 统计
   ======================================== */
.food-intro {
  padding: var(--space-8) 0 var(--space-12);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-10);
}

.food-intro__body {
  max-width: 720px;
  text-align: center;
}

.food-intro__text {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
  font-weight: 400;
}

.food-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
  width: 100%;
  max-width: 800px;
}

.food-stats__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-5) var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-base);
}

.food-stats__item:hover {
  border-color: var(--color-gold-dark);
}

.food-stats__num {
  font-family: var(--font-en-display);
  font-size: var(--text-3xl);
  font-weight: 700;
  color: var(--color-cinnabar);
  line-height: 1;
}

.food-stats__label {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  text-align: center;
}

/* ========================================
   美食名片
   ======================================== */
.food-cards-section {
  padding: 0 0 var(--space-14);
}

.food-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.food-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-8) var(--space-6);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--space-3);
  transition: all var(--transition-base);
  cursor: pointer;
  outline: none;
  position: relative;
  overflow: hidden;
}

.food-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--color-cinnabar), var(--color-gold), var(--color-cinnabar), transparent);
  opacity: 0;
  transition: opacity var(--transition-base);
}

.food-card:hover,
.food-card:focus-visible {
  border-color: var(--color-gold-dark);
  transform: translateY(-4px);
  box-shadow: 0 12px 32px var(--color-gold-glow), var(--shadow-lg);
}

.food-card:hover::before,
.food-card:focus-visible::before {
  opacity: 1;
}

/* 名片配图:圆形徽章,无图时内部回退为图标 */
.food-card__media {
  width: 92px;
  height: 92px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--space-2);
  border: 2px solid var(--color-gold);
  background: var(--color-bg-alt, #f5f2ec);
  transition: transform var(--transition-base), box-shadow var(--transition-base);
}

.food-card__photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.food-card:hover .food-card__media,
.food-card:focus-visible .food-card__media {
  transform: scale(1.06);
  box-shadow: 0 8px 22px var(--color-gold-glow, rgba(0, 0, 0, 0.15));
}

.food-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-gold);
}

.food-card__name {
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.2;
}

.food-card__en-name {
  font-family: var(--font-en-display);
  font-style: italic;
  font-size: var(--text-sm);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
  font-weight: 400;
  margin-top: calc(-1 * var(--space-1));
}

.food-card__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  margin-top: var(--space-1);
  flex: 1;
}

.food-card__footer {
  margin-top: auto;
  padding-top: var(--space-3);
}

.food-card__action {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  letter-spacing: var(--tracking-wide);
  transition: color var(--transition-fast);
}

.food-card:hover .food-card__action {
  color: var(--color-cinnabar);
}

.food-card__arrow {
  transition: transform var(--transition-fast);
}

.food-card:hover .food-card__arrow {
  transform: translateX(4px);
}

/* ========================================
   时间线
   ======================================== */
.food-timeline-section {
  padding: var(--space-14) 0;
  position: relative;
}

.food-timeline-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--color-border-light), transparent);
}

.food-timeline-section::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--color-border-light), transparent);
}

.food-timeline {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
  max-width: 800px;
  margin: 0 auto;
}

.food-timeline::before {
  content: '';
  position: absolute;
  left: 50%;
  top: 0;
  bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, transparent, var(--color-gold-dark), var(--color-gold), var(--color-gold-dark), transparent);
  opacity: 0.35;
  transform: translateX(-50%);
}

.timeline-era {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: var(--space-8);
}

.timeline-era--alt {
  flex-direction: row-reverse;
  text-align: right;
}

.timeline-era__dot {
  position: absolute;
  left: 50%;
  top: var(--space-5);
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--color-gold);
  border: 4px solid var(--color-bg);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-gold) 40%, transparent), 0 0 16px var(--color-gold-glow);
  transform: translateX(-50%);
  z-index: 2;
  flex-shrink: 0;
}

.timeline-era__card {
  width: calc(50% - var(--space-10));
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  transition: border-color var(--transition-base), transform var(--transition-base);
}

.timeline-era__card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px var(--color-gold-glow);
}

.timeline-era__header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.timeline-era--alt .timeline-era__header {
  flex-direction: row-reverse;
}

.timeline-era__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.2;
}

.timeline-era__period {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  color: var(--color-cinnabar);
  letter-spacing: var(--tracking-wider);
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 40%, transparent);
  border-radius: var(--radius-sm);
  padding: 2px 10px;
  white-space: nowrap;
}

.timeline-era__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

/* ========================================
   经典名菜
   ======================================== */
.food-dishes-section {
  padding: var(--space-14) 0;
}

.food-dishes {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}

.dish-item {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: all var(--transition-base);
}

.dish-item:hover {
  border-color: var(--color-cinnabar);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px var(--color-cinnabar-dim);
}

.dish-item__icon {
  display: flex;
  align-items: center;
  color: var(--color-cinnabar);
  flex-shrink: 0;
}

.dish-item__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.dish-item__name {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.dish-item__tag {
  font-size: var(--text-xs);
  color: var(--color-cinnabar);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   美食街区
   ======================================== */
.food-streets-section {
  padding: 0 0 var(--space-14);
}

.food-streets {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.street-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--space-3);
  transition: all var(--transition-base);
}

.street-card:hover {
  border-color: var(--color-cinnabar);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-cinnabar-dim);
}

.street-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-gold);
}

.street-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.street-card__tag {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  padding: 2px 12px;
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 1024px) {
  .food-cards {
    grid-template-columns: repeat(2, 1fr);
  }
  .food-stats {
    grid-template-columns: repeat(2, 1fr);
  }
  .food-dishes {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .food-timeline::before {
    left: 24px;
  }
  .timeline-era__dot {
    left: 24px;
  }
  .timeline-era__card,
  .timeline-era--alt .timeline-era__card {
    width: calc(100% - 56px);
    margin-left: 56px;
  }
  .timeline-era,
  .timeline-era--alt {
    flex-direction: row;
    text-align: left;
  }
  .timeline-era--alt .timeline-era__header {
    flex-direction: row;
  }
  .food-cards {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .food-stats {
    grid-template-columns: repeat(2, 1fr);
  }
  .food-dishes {
    grid-template-columns: 1fr;
  }
  .food-streets {
    grid-template-columns: 1fr;
  }
}
/* ── 成都味道图鉴(接口数据网格) ── */
.food-grid-section {
  padding: var(--space-10) 0;
}

.food-grid__status {
  text-align: center;
  padding: var(--space-8) 0;
  color: var(--color-text-2);
  font-size: 0.95rem;
}

.food-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: var(--space-5);
}

.food-grid__card {
  background: var(--color-bg, #fff);
  border-radius: var(--radius-lg, 16px);
  overflow: hidden;
  border: 1px solid var(--color-border, rgba(0, 0, 0, 0.06));
  cursor: pointer;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
}

.food-grid__card:hover,
.food-grid__card:focus-visible {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.1);
  outline: none;
}

.food-grid__media {
  position: relative;
  width: 100%;
  aspect-ratio: 4 / 3;
  background: var(--color-bg-alt, #f5f2ec);
  overflow: hidden;
}

.food-grid__photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.35s ease;
}

.food-grid__card:hover .food-grid__photo {
  transform: scale(1.05);
}

.food-grid__tag {
  position: absolute;
  top: 10px;
  left: 10px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 0.75rem;
  color: #fff;
  background: rgba(196, 62, 29, 0.88);
  backdrop-filter: blur(4px);
}

/* 图鉴卡片收藏按钮(与景点卡片一致的右上角心形) */
.food-grid__fav {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 2;
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-full, 999px);
  color: var(--color-text-muted, #8a7f74);
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid var(--color-border, rgba(0, 0, 0, 0.06));
  backdrop-filter: blur(6px);
  cursor: pointer;
  transition: all 0.2s ease;
}

.food-grid__fav:hover {
  color: var(--color-gold, #c8a45c);
  border-color: var(--color-gold, #c8a45c);
}

.food-grid__fav--active {
  color: var(--color-cinnabar, #c43e1d);
  border-color: color-mix(in srgb, var(--color-cinnabar, #c43e1d) 55%, transparent);
}

.food-grid__body {
  padding: 14px 16px 16px;
}

.food-grid__name {
  margin: 0 0 4px;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--color-text, #2b2118);
}

.food-grid__en {
  margin: 0 0 4px;
  font-size: 0.78rem;
  color: var(--color-text-2, #8a7f74);
}

.food-grid__desc {
  margin: 0;
  font-size: 0.85rem;
  line-height: 1.55;
  color: var(--color-text-2, #6f6459);
}
</style>
