<script setup lang="ts">
/**
 * QuizCard.vue — 蜀文化知识闯关(可复用答题卡)
 * 挂在景点详情页底部:进入拉取 5 题,逐题作答,选项点击后即时对错高亮
 * 并展示当前语言解析(en/ja 回落中文);完成后按本轮成绩授予文化段位,
 * 答对题目的景点记入 localStorage 徽章(tripstar.quiz_badges)。
 * 景点无题目时整体不渲染,父组件无需感知题目是否存在。
 */
import { computed, onMounted, ref } from 'vue'
import { useLanguageStore } from '@/stores/language'
import {
  listQuizQuestions,
  checkQuizAnswer,
  type QuizQuestionItem,
  type QuizCheckResult,
} from '@/api/quiz'
import { recordQuizResult, quizRankOf } from '@/utils/quizBadges'

const props = defineProps<{ spotId: number }>()

const langStore = useLanguageStore()

/** 每轮抽题数量(与后端默认一致) */
const QUIZ_COUNT = 5

const questions = ref<QuizQuestionItem[]>([])
const loading = ref(true)
/** 当前题号(0 起) */
const idx = ref(0)
/** 本轮已选选项下标(null = 未作答) */
const selected = ref<number | null>(null)
/** 判分结果(后端返回,含正确项与解析) */
const checked = ref<QuizCheckResult | null>(null)
const finished = ref(false)
const correctCount = ref(0)

onMounted(async () => {
  try {
    questions.value = await listQuizQuestions(props.spotId, QUIZ_COUNT)
  } catch {
    questions.value = []
  }
  loading.value = false
})

/** 无题目时整体隐藏(父组件无需感知) */
const hasQuiz = computed(() => questions.value.length > 0)

/** 当前题目 */
const current = computed(() => questions.value[idx.value] ?? null)

/** JSON 数组字符串安全解析(损坏回落空数组) */
function parseOptions(raw: string | undefined): string[] {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr.map((v) => String(v)) : []
  } catch {
    return []
  }
}

/** 按语言取题干/选项/解析:en/ja 缺失回落中文 */
function pickText(zh: string, en?: string, ja?: string): string {
  if (langStore.lang === 'en') return (en || '').trim() || zh
  if (langStore.lang === 'ja') return (ja || '').trim() || zh
  return zh
}

const questionText = computed(() => {
  const q = current.value
  return q ? pickText(q.question_zh, q.question_en, q.question_ja) : ''
})

/** 当前语言下的选项列表(en/ja 逐项回落中文) */
const optionTexts = computed(() => {
  const q = current.value
  if (!q) return [] as string[]
  const zh = parseOptions(q.options_zh)
  const en = parseOptions(q.options_en)
  const ja = parseOptions(q.options_ja)
  return zh.map((text, i) =>
    pickText(text, en[i] || text, ja[i] || text)
  )
})

/** 当前语言下的解析 */
const explainText = computed(() => {
  const r = checked.value
  return r ? pickText(r.explain_zh, r.explain_en, r.explain_ja) : ''
})

/** 选项样式:答后高亮正确项(青)与错选项(朱) */
function optionClass(i: number): string {
  if (!checked.value) return ''
  if (i === checked.value.answer_idx) return 'quiz-option--correct'
  if (i === selected.value) return 'quiz-option--wrong'
  return 'quiz-option--dim'
}

/** 作答:锁定选项 → 后端判分 → 高亮并展示解析 */
async function choose(i: number) {
  const q = current.value
  if (!q || checked.value) return
  selected.value = i
  try {
    checked.value = await checkQuizAnswer(q.id, i)
    if (checked.value.correct) correctCount.value++
  } catch {
    // 判分失败按未作答处理,允许重选
    selected.value = null
  }
}

/** 下一题:末题则结算徽章并出结果卡 */
function next() {
  if (idx.value < questions.value.length - 1) {
    idx.value++
    selected.value = null
    checked.value = null
    return
  }
  finished.value = true
  recordQuizResult(props.spotId, correctCount.value)
}

/** 本轮段位(0-1 青铜 / 2-3 白银 / 4-5 黄金) */
const rankKey = computed(() => quizRankOf(correctCount.value).key)
</script>

<template>
  <div v-if="hasQuiz && !loading" class="quiz-card">
    <div class="quiz-card__head">
      <span class="quiz-card__icon" aria-hidden="true">⛩️</span>
      <div class="quiz-card__titles">
        <h3 class="quiz-card__title">{{ langStore.t('quiz.title') }}</h3>
        <p class="quiz-card__hint">{{ langStore.t('quiz.hint') }}</p>
      </div>
      <span v-if="!finished" class="quiz-card__progress">
        {{ langStore.t('quiz.progress', { current: idx + 1, total: questions.length }) }}
      </span>
    </div>

    <!-- ──── 答题区 ──── -->
    <div v-if="!finished && current" class="quiz-card__body">
      <p class="quiz-card__question">{{ questionText }}</p>
      <ul class="quiz-card__options">
        <li v-for="(opt, i) in optionTexts" :key="i">
          <button
            type="button"
            class="quiz-option"
            :class="[optionClass(i), { 'quiz-option--picked': selected === i && !checked }]"
            :disabled="!!checked"
            @click="choose(i)"
          >
            <span class="quiz-option__key" aria-hidden="true">{{ 'ABC'[i] }}</span>
            <span class="quiz-option__text">{{ opt }}</span>
            <span v-if="checked && i === checked.answer_idx" class="quiz-option__mark" aria-hidden="true">✓</span>
            <span v-else-if="checked && i === selected" class="quiz-option__mark" aria-hidden="true">✕</span>
          </button>
        </li>
      </ul>

      <!-- 判分反馈:对错 + 当前语言解析 -->
      <div v-if="checked" class="quiz-card__feedback" :class="checked.correct ? 'quiz-card__feedback--ok' : 'quiz-card__feedback--no'">
        <p class="quiz-card__verdict">
          {{ checked.correct ? langStore.t('quiz.correct') : langStore.t('quiz.wrong') }}
        </p>
        <p class="quiz-card__explain">
          <span class="quiz-card__explain-label">{{ langStore.t('quiz.explain') }}</span>
          {{ explainText }}
        </p>
        <button type="button" class="quiz-card__next" @click="next">
          {{ langStore.t('quiz.next') }}
        </button>
      </div>
    </div>

    <!-- ──── 结果卡:答对数 + 文化段位 ──── -->
    <div v-else-if="finished" class="quiz-card__result">
      <p class="quiz-card__score">
        {{ langStore.t('quiz.resultScore', { correct: correctCount, total: questions.length }) }}
      </p>
      <p class="quiz-card__rank">
        <span class="quiz-card__rank-label">{{ langStore.t('quiz.rankLabel') }}</span>
        <span class="quiz-card__rank-name">{{ langStore.t(rankKey) }}</span>
      </p>
    </div>
  </div>
</template>

<style scoped>
/* 复用景点详情页卡片基调:表面 + 描边 + 展示字体 */
.quiz-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  box-shadow: 0 4px 16px rgba(46, 58, 61, 0.05);
}

.quiz-card__head {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  margin-bottom: var(--space-5);
}

.quiz-card__icon {
  font-size: var(--text-2xl);
  line-height: 1.2;
}

.quiz-card__titles {
  flex: 1;
  min-width: 0;
}

.quiz-card__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.quiz-card__hint {
  margin: var(--space-1) 0 0;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.quiz-card__progress {
  flex-shrink: 0;
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-gold) 12%, transparent);
  color: var(--color-gold-dark);
  font-size: var(--text-xs);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
}

.quiz-card__question {
  margin: 0 0 var(--space-4);
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  line-height: var(--leading-normal);
}

.quiz-card__options {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.quiz-option {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: var(--space-3) var(--space-4);
  text-align: left;
  background: color-mix(in srgb, var(--color-gold) 4%, var(--color-surface));
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast),
    transform var(--transition-fast);
}

.quiz-option:hover:not(:disabled) {
  border-color: var(--color-gold);
  transform: translateY(-1px);
}

.quiz-option:disabled {
  cursor: default;
}

.quiz-option__key {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  font-size: var(--text-xs);
  font-weight: 700;
  color: var(--color-text-secondary);
}

.quiz-option__text {
  flex: 1;
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  line-height: var(--leading-normal);
}

.quiz-option__mark {
  flex-shrink: 0;
  font-weight: 800;
}

/* 答后高亮:正确项青绿,错选项朱红,其余弱化 */
.quiz-option--correct {
  border-color: var(--color-sage);
  background: var(--color-sage-dim);
}

.quiz-option--correct .quiz-option__key,
.quiz-option--correct .quiz-option__mark {
  color: var(--color-sage);
  border-color: var(--color-sage);
}

.quiz-option--wrong {
  border-color: var(--color-cinnabar);
  background: var(--color-cinnabar-dim);
}

.quiz-option--wrong .quiz-option__key,
.quiz-option--wrong .quiz-option__mark {
  color: var(--color-cinnabar);
  border-color: var(--color-cinnabar);
}

.quiz-option--dim {
  opacity: 0.55;
}

/* ── 反馈区 ── */
.quiz-card__feedback {
  margin-top: var(--space-4);
  padding: var(--space-4);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-gold) 6%, var(--color-surface));
  border: 1px dashed var(--color-border);
}

.quiz-card__feedback--ok {
  border-color: color-mix(in srgb, var(--color-sage) 50%, transparent);
}

.quiz-card__feedback--no {
  border-color: color-mix(in srgb, var(--color-cinnabar) 50%, transparent);
}

.quiz-card__verdict {
  margin: 0 0 var(--space-2);
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  letter-spacing: var(--tracking-wide);
}

.quiz-card__feedback--ok .quiz-card__verdict {
  color: var(--color-sage);
}

.quiz-card__feedback--no .quiz-card__verdict {
  color: var(--color-cinnabar);
}

.quiz-card__explain {
  margin: 0;
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

.quiz-card__explain-label {
  display: inline-block;
  margin-right: var(--space-2);
  padding: 1px var(--space-2);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-gold) 14%, transparent);
  color: var(--color-gold-dark);
  font-size: var(--text-xs);
  font-weight: 600;
}

.quiz-card__next {
  display: block;
  margin: var(--space-4) 0 0 auto;
  padding: var(--space-2) var(--space-6);
  border: none;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--color-gold), var(--color-gold-dark));
  color: var(--color-text-inverse);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.quiz-card__next:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px var(--color-gold-glow);
}

/* ── 结果卡 ── */
.quiz-card__result {
  text-align: center;
  padding: var(--space-5) 0 var(--space-2);
}

.quiz-card__score {
  margin: 0 0 var(--space-3);
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.quiz-card__rank {
  margin: 0;
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: var(--space-3);
}

.quiz-card__rank-label {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.quiz-card__rank-name {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 800;
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wider);
}
</style>
