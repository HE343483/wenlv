/**
 * 全局 AI 行程生成任务 store
 * 生成过程(WebSocket 订阅)挂在模块级 store 上，
 * 用户切换页面不会中断生成，完成后通过右上角 notification 提醒。
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { notification } from 'ant-design-vue'
import router from '@/router'
import { i18n } from '@/i18n'
import { generateTripPlan } from '@/api/trip'
import type { TripFormData, TripTaskEvent } from '@/types/trip'

export const useTripTaskStore = defineStore('tripTask', () => {
  const generating = ref(false)
  const progress = ref(0)
  const statusText = ref('')
  const planCode = ref('')

  const getStageStatusText = (stage: TripTaskEvent['stage']) => {
    const t = i18n.global.t
    if (stage === 'submitted' || stage === 'initializing') return t('home.loading.initializing')
    if (stage === 'attraction_search') return t('home.loading.searchingAttractions')
    if (stage === 'weather_search') return t('home.loading.queryingWeather')
    if (stage === 'hotel_search') return t('home.loading.recommendingHotels')
    if (stage === 'planning') return t('home.loading.generatingPlan')
    if (stage === 'graph_building') return t('home.loading.generatingPlan')
    if (stage === 'completed') return t('home.loading.done')
    return t('home.loading.initializing')
  }

  const openResultPage = (planId: string) => {
    if (planId) {
      router.push({ path: '/trip/result', query: { plan_id: planId } })
    } else {
      router.push('/trip/result')
    }
  }

  const clearDraft = () => {
    sessionStorage.removeItem('tripPlan')
    sessionStorage.removeItem('graphData')
    sessionStorage.removeItem('planId')
  }

  const start = async (requestData: TripFormData) => {
    if (generating.value) return

    generating.value = true
    progress.value = 5
    statusText.value = i18n.global.t('home.loading.initializing')
    planCode.value = ''
    clearDraft()

    try {
      const response = await generateTripPlan(requestData, {
        onTaskCreated: (task) => {
          planCode.value = task.plan_id || task.task_id
          progress.value = 5
          statusText.value = i18n.global.t('home.loading.initializing')
        },
        onTaskEvent: (event) => {
          if (event.plan_id) planCode.value = event.plan_id
          if (Number.isFinite(event.progress)) {
            progress.value = Math.max(0, Math.min(100, event.progress))
          }
          // 后端 message 为中文固定文案,优先用 stage 对应的 i18n 三语文案,后端 message 仅作兜底
          statusText.value = getStageStatusText(event.stage) || event.message || statusText.value
        },
      })

      progress.value = 100
      statusText.value = i18n.global.t('home.loading.done')

      if (response.success && response.data) {
        const planId = response.plan_id || planCode.value
        sessionStorage.setItem('tripPlan', JSON.stringify(response.data))
        if (response.graph_data) sessionStorage.setItem('graphData', JSON.stringify(response.graph_data))
        if (planId) sessionStorage.setItem('planId', planId)

        // 非模态的角落通知(自动消失)，点击可跳转结果页
        const t = i18n.global.t
        notification.success({
          message: t('home.messages.notifyCompletedTitle'),
          description: t('home.messages.notifyCompletedDesc'),
          duration: 8,
          placement: 'topRight',
          onClick: () => {
            notification.destroy()
            openResultPage(planId)
          },
        })

        // 用户仍停留在行程表单页时保持原有体验：自动进入结果页
        if (router.currentRoute.value.path === '/trip') {
          setTimeout(() => openResultPage(planId), 800)
        }
      } else {
        clearDraft()
        notification.error({
          message: i18n.global.t('home.messages.notifyFailedTitle'),
          description: response.message || i18n.global.t('home.messages.generateFailed'),
          duration: 6,
          placement: 'topRight',
        })
      }
    } catch (error: any) {
      clearDraft()
      notification.error({
        message: i18n.global.t('home.messages.notifyFailedTitle'),
        description: error?.message || i18n.global.t('home.messages.generateRetry'),
        duration: 6,
        placement: 'topRight',
      })
    } finally {
      setTimeout(() => {
        generating.value = false
        progress.value = 0
        statusText.value = ''
        planCode.value = ''
      }, 1000)
    }
  }

  return { generating, progress, statusText, planCode, start }
})
