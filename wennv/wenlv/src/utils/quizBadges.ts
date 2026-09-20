/**
 * quizBadges.ts — 蜀文化知识闯关徽章(localStorage 本地存档)
 * 答对的题目给"徽章":存答对题目的景点 id 集合 + 累计答对总数;
 * 数据按答题记录保存在本机浏览器,不绑定账号(与护照印章走后端不同)。
 * 文化段位按累计答对数定:0-1 青铜 / 2-3 白银 / 4-5 黄金。
 */

/** localStorage 存储键(与主站 localStorage 命名约定对齐) */
export const QUIZ_BADGES_KEY = 'tripstar.quiz_badges'

/** 徽章存档:答对题目的景点 id 集合 + 累计答对总数 */
export interface QuizBadgeStore {
  /** 答对过题目的景点 id 集合(徽章标识) */
  spots: number[]
  /** 累计答对题目总数(段位依据) */
  correctTotal: number
}

/** 文化段位:key 为 locales quiz.rank* 的键名,门槛为累计答对数 */
export interface QuizRank {
  key: 'quiz.rankBronze' | 'quiz.rankSilver' | 'quiz.rankGold'
  min: number
  max: number
}

/** 段位划分:0-1 青铜 / 2-3 白银 / 4-5 黄金 */
export const QUIZ_RANKS: QuizRank[] = [
  { key: 'quiz.rankBronze', min: 0, max: 1 },
  { key: 'quiz.rankSilver', min: 2, max: 3 },
  { key: 'quiz.rankGold', min: 4, max: 5 },
]

/** 读取徽章存档(损坏/缺失返回空档案) */
export function loadQuizBadges(): QuizBadgeStore {
  try {
    const raw = localStorage.getItem(QUIZ_BADGES_KEY)
    if (!raw) return { spots: [], correctTotal: 0 }
    const parsed = JSON.parse(raw) as Partial<QuizBadgeStore>
    return {
      spots: Array.isArray(parsed.spots) ? parsed.spots.filter((n) => typeof n === 'number') : [],
      correctTotal: typeof parsed.correctTotal === 'number' ? parsed.correctTotal : 0,
    }
  } catch {
    return { spots: [], correctTotal: 0 }
  }
}

/** 写入徽章存档 */
function saveQuizBadges(store: QuizBadgeStore) {
  localStorage.setItem(QUIZ_BADGES_KEY, JSON.stringify(store))
}

/** 答题结算:按本轮答对题数累计答对总数,答对过则记下景点 id(幂等) */
export function recordQuizResult(spotId: number, correctCount: number): QuizBadgeStore {
  const store = loadQuizBadges()
  const spots = new Set(store.spots)
  if (correctCount > 0) spots.add(spotId)
  const next: QuizBadgeStore = {
    spots: [...spots],
    correctTotal: store.correctTotal + Math.max(0, correctCount),
  }
  saveQuizBadges(next)
  return next
}

/** 按累计答对数取文化段位 */
export function quizRankOf(correctTotal: number): QuizRank {
  for (const rank of QUIZ_RANKS) {
    if (correctTotal >= rank.min && correctTotal <= rank.max) return rank
  }
  // 超出 5 题封顶黄金(单轮最多 5 题,防异常数据)
  return { key: 'quiz.rankGold', min: 4, max: Number.MAX_SAFE_INTEGER }
}
