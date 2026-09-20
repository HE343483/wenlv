/** 蜀文化知识闯关接口封装(公开接口,无需登录) */
import { get } from './request'

/** 单道闯关题目:三语题干与选项(JSON 数组字符串),正确项由后端判分不回传 */
export interface QuizQuestionItem {
  id: number
  spot_id: number
  question_zh: string
  /** 英/日文题干(LLM 生成,可能为空,展示回落中文) */
  question_en?: string
  question_ja?: string
  /** 三语选项 JSON 数组字符串,如 ["武侯祠","望江楼","青羊宫"] */
  options_zh: string
  options_en?: string
  options_ja?: string
  /** 中文解析(引用库内资料要点) */
  explain_zh: string
  explain_en?: string
  explain_ja?: string
}

/** 答案校验结果:判分在后端完成,answer_idx 为正确项下标 */
export interface QuizCheckResult {
  question_id: number
  correct: boolean
  answer_idx: number
  explain_zh: string
  explain_en?: string
  explain_ja?: string
}

/** 按景点随机抽题:题目不足时后端返回全部,无题返回空数组(前端隐藏答题卡) */
export function listQuizQuestions(spotId: number, count = 5): Promise<QuizQuestionItem[]> {
  return get<QuizQuestionItem[]>('/quiz', { params: { spot_id: spotId, count } })
}

/** 校验答案:answer 为所选选项下标 */
export function checkQuizAnswer(questionId: number, answer: number): Promise<QuizCheckResult> {
  return get<QuizCheckResult>('/quiz/check', {
    params: { question_id: questionId, answer },
  })
}
