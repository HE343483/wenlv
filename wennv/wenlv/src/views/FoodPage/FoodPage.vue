<script setup lang="ts">
/**
 * FoodPage.vue — 成都美食
 * 功能：美食名片卡片网格 + 川菜脉络时间线 + 经典名菜/老字号展示
 * 数据来源：本地化 locales + 内联数据，后期可接入后端
 */
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import AppIcon from '@/components/AppIcon.vue'
import HomeBanner from '@/components/HomeBanner.vue'

const langStore = useLanguageStore()
const router = useRouter()

/* 跳转美食详情页（key 作为 :id） */
function goFoodDetail(key: string) {
  router.push({ name: 'food-detail', params: { id: key } })
}

/* ── 美食名片 ── */
const foodCards = [
  { key: 'hotpot', icon: 'hotpot' },
  { key: 'chuanchuan', icon: 'chuanchuan' },
  { key: 'cuisine', icon: 'cuisine' },
  { key: 'snacks', icon: 'dumpling' },
  { key: 'tea', icon: 'tea' },
  { key: 'nightfood', icon: 'moon' },
]

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
          @click="goFoodDetail(card.key)"
          @keyup.enter="goFoodDetail(card.key)"
        >
          <div class="food-card__icon">
            <AppIcon :name="card.icon" :size="36" />
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

.food-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-gold);
  margin-bottom: var(--space-1);
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
</style>
