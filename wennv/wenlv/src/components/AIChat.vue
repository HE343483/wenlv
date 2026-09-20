<template>
  <div
    ref="rootRef"
    class="ai-chat-floating"
    :class="{ dragging: isDragging }"
    :style="{ '--chat-tx': `${offset.x}px`, '--chat-ty': `${offset.y}px` }"
    @pointerdown="onDragStart"
    @click.capture="onClickCapture"
  >
    <div class="container-ai-input">
      <div v-for="index in 15" :key="`chat-area-${index}`" class="area"></div>
      <div class="container-wrap" :class="{ open: chatOpen }">
        <div class="card">
          <div class="background-blur-balls">
            <div class="balls">
              <span class="ball rosa"></span>
              <span class="ball violet"></span>
              <span class="ball green"></span>
              <span class="ball cyan"></span>
            </div>
          </div>
          <div class="content-card" :class="{ clickable: !chatOpen }" @click="openChatPanel">
            <div class="background-blur-card">
              <div class="eyes">
                <span class="eye"></span>
                <span class="eye"></span>
              </div>
              <div class="eyes happy">
                <svg fill="none" viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M8.28386 16.2843C8.9917 15.7665 9.8765 14.731 12 14.731C14.1235 14.731 15.0083 15.7665 15.7161 16.2843C17.8397 17.8376 18.7542 16.4845 18.9014 15.7665C19.4323 13.1777 17.6627 11.1066 17.3088 10.5888C16.3844 9.23666 14.1235 8 12 8C9.87648 8 7.61556 9.23666 6.69122 10.5888C6.33728 11.1066 4.56771 13.1777 5.09858 15.7665C5.24582 16.4845 6.16034 17.8376 8.28386 16.2843Z"
                  ></path>
                </svg>
                <svg fill="none" viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M8.28386 16.2843C8.9917 15.7665 9.8765 14.731 12 14.731C14.1235 14.731 15.0083 15.7665 15.7161 16.2843C17.8397 17.8376 18.7542 16.4845 18.9014 15.7665C19.4323 13.1777 17.6627 11.1066 17.3088 10.5888C16.3844 9.23666 14.1235 8 12 8C9.87648 8 7.61556 9.23666 6.69122 10.5888C6.33728 11.1066 4.56771 13.1777 5.09858 15.7665C5.24582 16.4845 6.16034 17.8376 8.28386 16.2843Z"
                  ></path>
                </svg>
              </div>
            </div>
          </div>
          <div class="container-ai-chat" @click.stop>
            <button type="button" class="chat-close-btn btn-round btn-danger" @click.stop="closeChatPanel">×</button>
            <div class="chat">
              <!-- 历史会话 / 新会话(仅登录用户) -->
              <div v-if="isLoggedIn" class="chat-toolbar">
                <button type="button" class="chat-tool-btn" :disabled="chatLoading" @click="openSessionsDrawer">
                  <svg viewBox="0 0 24 24" width="28" height="28" xmlns="http://www.w3.org/2000/svg">
                    <path
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M3 12a9 9 0 1 0 3-6.7L3 8m0-5v5h5M12 7v5l3 3"
                    ></path>
                  </svg>
                  <span>{{ t('chatHistory.open') }}</span>
                </button>
                <button type="button" class="chat-tool-btn" :disabled="chatLoading" @click="startNewSession">
                  <svg viewBox="0 0 24 24" width="28" height="28" xmlns="http://www.w3.org/2000/svg">
                    <path
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      d="M12 5v14M5 12h14"
                    ></path>
                  </svg>
                  <span>{{ t('chatHistory.newChat') }}</span>
                </button>
              </div>
              <!-- 对话角色切换:普通助手 / 杜甫 / 诸葛亮 -->
              <div class="persona-switch">
                <span class="persona-switch-label">{{ t('personas.switchLabel') }}</span>
                <div class="persona-switch-chips">
                  <button
                    v-for="p in personaOptions"
                    :key="p.id"
                    type="button"
                    class="persona-chip"
                    :class="[{ active: activePersona === p.id }, p.themeClass]"
                    :disabled="chatLoading"
                    @click="switchPersona(p.id)"
                  >
                    <span class="persona-chip-name">{{ t(p.nameKey) }}</span>
                  </button>
                </div>
              </div>
              <!-- 角色记忆开关(仅角色模式显示;未登录禁用并提示) -->
              <div v-if="isPersonaMode" class="persona-memory-row">
                <a-switch v-model:checked="personaMemoryEnabled" :disabled="!isLoggedIn" class="persona-memory-switch" />
                <span class="persona-memory-label">{{ t('personaMemory.toggle') }}</span>
                <a-popconfirm
                  v-if="isLoggedIn"
                  :title="t('personaMemory.clearConfirm')"
                  :ok-text="t('personaMemory.clear')"
                  :cancel-text="t('common.cancel')"
                  @confirm="onClearMemory"
                >
                  <button type="button" class="persona-memory-clear">{{ t('personaMemory.clear') }}</button>
                </a-popconfirm>
                <span v-else class="persona-memory-tip">{{ t('personaMemory.loginTip') }}</span>
              </div>
              <div class="chat-bot">
                <!-- 角色模式头部条 -->
                <div v-if="isPersonaMode" class="persona-header" :class="activePersonaOption.themeClass">
                  <span class="persona-header-name">{{ t(activePersonaOption.nameKey) }}</span>
                  <span class="persona-header-title">{{ t(activePersonaOption.titleKey) }}</span>
                  <span class="persona-header-tip">{{ t('personas.chatModeTip', { name: t(activePersonaOption.nameKey) }) }}</span>
                </div>
                <div class="chat-history" ref="chatMessagesRef">
                  <div v-if="chatHistory.length === 0" class="chat-empty">
                    <p>{{ t('result.chat.welcome') }}</p>
                    <div v-if="!isPersonaMode" class="chat-suggestions">
                      <button
                        v-for="question in quickQuestions"
                        :key="question.labelKey"
                        type="button"
                        class="chat-suggestion"
                        :disabled="chatLoading || !tripPlan"
                        @click="sendQuickQuestion(t(question.questionKey))"
                      >
                        {{ t(question.labelKey) }}
                      </button>
                    </div>
                  </div>
                  <div
                    v-for="(msg, idx) in chatHistory"
                    :key="`chat-${idx}`"
                    class="chat-msg"
                    :class="[msg.role, msg.personaClass]"
                  >
                    {{ msg.content }}
                  </div>
                  <div v-if="chatLoading" class="chat-msg assistant typing">
                    <span class="dot"></span>
                    <span class="dot"></span>
                    <span class="dot"></span>
                  </div>
                </div>
                <textarea
                  v-model="chatInput"
                  :placeholder="chatPlaceholder"
                  name="chat_bot"
                  id="chat_bot"
                  :disabled="chatLoading || !canChat"
                  @keydown.enter.exact.prevent="sendChatMessage"
                ></textarea>
              </div>
              <div class="options">
                <div class="btns-add">
                  <button type="button" disabled>
                    <svg
                      viewBox="0 0 24 24"
                      height="20"
                      width="20"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M7 8v8a5 5 0 1 0 10 0V6.5a3.5 3.5 0 1 0-7 0V15a2 2 0 0 0 4 0V8"
                        stroke-width="2"
                        stroke-linejoin="round"
                        stroke-linecap="round"
                        stroke="currentColor"
                        fill="none"
                      ></path>
                    </svg>
                  </button>
                  <button type="button" disabled>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="20"
                      height="20"
                      viewBox="0 0 24 24"
                    >
                      <path
                        fill="none"
                        stroke="currentColor"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M4 5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v4a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1zm0 10a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v4a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1zm10 0a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v4a1 1 0 0 1-1 1h-4a1 1 0 0 1-1-1zm0-8h6m-3-3v6"
                      ></path>
                    </svg>
                  </button>
                  <button type="button" disabled>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="20"
                      height="20"
                      viewBox="0 0 24 24"
                    >
                      <path
                        fill="currentColor"
                        d="M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10s-4.477 10-10 10m-2.29-2.333A17.9 17.9 0 0 1 8.027 13H4.062a8.01 8.01 0 0 0 5.648 6.667M10.03 13c.151 2.439.848 4.73 1.97 6.752A15.9 15.9 0 0 0 13.97 13zm9.908 0h-3.965a17.9 17.9 0 0 1-1.683 6.667A8.01 8.01 0 0 0 19.938 13M4.062 11h3.965A17.9 17.9 0 0 1 9.71 4.333A8.01 8.01 0 0 0 4.062 11m5.969 0h3.938A15.9 15.9 0 0 0 12 4.248A15.9 15.9 0 0 0 10.03 11m4.259-6.667A17.9 17.9 0 0 1 15.973 11h3.965a8.01 8.01 0 0 0-5.648-6.667"
                      ></path>
                    </svg>
                  </button>
                </div>
                <button
                  type="button"
                  class="btn-submit"
                  :disabled="chatLoading || !chatInput.trim() || !canChat"
                  @click="sendChatMessage"
                >
                  <i>
                    <svg viewBox="0 0 512 512">
                      <path
                        d="M473 39.05a24 24 0 0 0-25.5-5.46L47.47 185h-.08a24 24 0 0 0 1 45.16l.41.13l137.3 58.63a16 16 0 0 0 15.54-3.59L422 80a7.07 7.07 0 0 1 10 10L226.66 310.26a16 16 0 0 0-3.59 15.54l58.65 137.38c.06.2.12.38.19.57c3.2 9.27 11.3 15.81 21.09 16.25h1a24.63 24.63 0 0 0 23-15.46L478.39 64.62A24 24 0 0 0 473 39.05"
                        fill="currentColor"
                      ></path>
                    </svg>
                  </i>
                </button>
              </div>
              <!-- 未登录提示:历史对话不保存 -->
              <div v-if="!isLoggedIn" class="chat-login-tip">{{ t('chatHistory.loginTip') }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- 历史会话抽屉(teleport 到 body,不受面板缩放影响) -->
    <a-drawer
      v-model:open="sessionsDrawerOpen"
      :title="t('chatHistory.open')"
      placement="right"
      :width="330"
      root-class-name="chat-sessions-drawer"
    >
      <a-spin :spinning="sessionsLoading">
        <a-empty v-if="sessionList.length === 0" :description="t('chatHistory.empty')" />
        <div v-else class="session-list">
          <button
            v-for="s in sessionList"
            :key="s.id"
            type="button"
            class="session-item"
            :class="{ active: s.id === currentSessionId }"
            @click="openSession(s)"
          >
            <span class="session-title">{{ s.title }}</span>
            <span class="session-time">{{ formatRelativeTime(s.updated_at) }}</span>
          </button>
        </div>
        <a-button block class="session-new-btn" @click="startNewSession">
          {{ t('chatHistory.newChat') }}
        </a-button>
      </a-spin>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { message as antMessage } from 'ant-design-vue'
import type { ChatMessage, TripPlan } from '@/types/trip'
import { getRuntimeApiBaseUrl } from '@/api/trip'
import { hasToken } from '@/utils/token'
import {
  listChatSessions,
  createChatSession,
  listSessionMessages,
  clearPersonaMemories,
  type ChatSessionItem,
} from '@/api/chatSession'

const props = defineProps<{
  tripPlan: TripPlan | null
}>()

const { t, locale } = useI18n()
const chatOpen = ref(false)
const chatInput = ref('')
const chatHistory = ref<PersonaChatMessage[]>([])
const chatLoading = ref(false)
const chatMessagesRef = ref<HTMLElement | null>(null)

// ===== 对话角色:普通助手 / 历史人物(杜甫/诸葛亮) =====
type PersonaID = 'assistant' | 'du-fu' | 'zhuge-liang'

interface PersonaOption {
  id: PersonaID
  nameKey: string
  titleKey: string
  greetingKey: string
  /** 气泡主题类:杜甫=宣纸色,诸葛亮=青竹色 */
  themeClass: string
}

interface PersonaChatMessage extends ChatMessage {
  /** 该消息所属角色的主题类,用于气泡配色 */
  personaClass?: string
}

const personaOptions: PersonaOption[] = [
  {
    id: 'assistant',
    nameKey: 'personas.assistant.name',
    titleKey: 'personas.assistant.title',
    greetingKey: '',
    themeClass: '',
  },
  {
    id: 'du-fu',
    nameKey: 'personas.duFu.name',
    titleKey: 'personas.duFu.title',
    greetingKey: 'personas.duFu.greeting',
    themeClass: 'persona-dufu',
  },
  {
    id: 'zhuge-liang',
    nameKey: 'personas.zhuge.name',
    titleKey: 'personas.zhuge.title',
    greetingKey: 'personas.zhuge.greeting',
    themeClass: 'persona-zhuge',
  },
]

const activePersona = ref<PersonaID>('assistant')
const activePersonaOption = computed<PersonaOption>(() => {
  const found = personaOptions.find((p) => p.id === activePersona.value)
  return found ?? personaOptions[0]!
})
const isPersonaMode = computed(() => activePersona.value !== 'assistant')

/** 角色模式无需行程计划也可对话;普通助手仍要求先生成行程 */
const canChat = computed(() => (isPersonaMode.value ? true : !!props.tripPlan))

/** 切换角色:清空当前对话并注入所选角色的开场白(开场白同时作为上下文首条消息) */
const switchPersona = (id: PersonaID) => {
  if (activePersona.value === id || chatLoading.value) return
  activePersona.value = id
  // 切角色 = 开启新会话:重置 session_id,并按新 persona_id 重新拉取历史会话列表
  currentSessionId.value = null
  chatHistory.value = []
  injectPersonaGreeting()
  if (isLoggedIn.value) void loadSessions()
  scrollChatToBottom()
}

/** 注入当前角色的开场白(作为上下文首条消息) */
const injectPersonaGreeting = () => {
  const option = personaOptions.find((p) => p.id === activePersona.value)
  if (option?.greetingKey) {
    chatHistory.value.push({
      role: 'assistant',
      content: t(option.greetingKey),
      personaClass: option.themeClass,
    } as PersonaChatMessage)
  }
}

// ===== 历史会话与角色记忆(仅登录用户可用) =====
const isLoggedIn = ref(hasToken())
const currentSessionId = ref<number | null>(null)
const sessionsDrawerOpen = ref(false)
const sessionsLoading = ref(false)
const sessionList = ref<ChatSessionItem[]>([])

/** 角色记忆独立开关:localStorage 持久化,默认关闭 */
const PERSONA_MEMORY_KEY = 'tripstar.persona_memory_enabled'
const personaMemoryEnabled = ref((() => {
  try {
    return localStorage.getItem(PERSONA_MEMORY_KEY) === '1'
  } catch {
    return false
  }
})())
watch(personaMemoryEnabled, (enabled) => {
  try {
    localStorage.setItem(PERSONA_MEMORY_KEY, enabled ? '1' : '0')
  } catch {
    /* 忽略写入失败(如隐私模式) */
  }
})

/** 按当前角色拉取历史会话列表 */
const loadSessions = async () => {
  if (!isLoggedIn.value) return
  sessionsLoading.value = true
  try {
    sessionList.value = await listChatSessions(activePersona.value)
  } catch (err) {
    console.error('Load chat sessions failed:', err)
    sessionList.value = []
  } finally {
    sessionsLoading.value = false
  }
}

const openSessionsDrawer = () => {
  isLoggedIn.value = hasToken()
  if (!isLoggedIn.value) return
  sessionsDrawerOpen.value = true
  void loadSessions()
}

/** 点击历史会话:加载该会话消息进当前对话,关闭抽屉继续聊 */
const openSession = async (session: ChatSessionItem) => {
  if (chatLoading.value) return
  try {
    const msgs = await listSessionMessages(session.id)
    currentSessionId.value = session.id
    const themeClass = activePersonaOption.value.themeClass
    chatHistory.value = msgs.map((m) => ({
      role: m.role,
      content: m.content,
      ...(isPersonaMode.value && m.role === 'assistant' ? { personaClass: themeClass } : {}),
    }))
    sessionsDrawerOpen.value = false
    scrollChatToBottom()
  } catch (err) {
    console.error('Load session messages failed:', err)
    antMessage.error(t('result.chat.networkError'))
  }
}

/** 新会话:清空对话并将 session_id 置空(首次发消息时才懒创建会话) */
const startNewSession = () => {
  currentSessionId.value = null
  chatHistory.value = []
  injectPersonaGreeting()
  sessionsDrawerOpen.value = false
  scrollChatToBottom()
}

/** 登录用户首次发言时创建会话,返回会话 ID(未登录/创建失败返回 null) */
const ensureSession = async (): Promise<number | null> => {
  if (!isLoggedIn.value) return null
  if (currentSessionId.value !== null) return currentSessionId.value
  try {
    const data = await createChatSession(activePersona.value, locale.value)
    currentSessionId.value = typeof data?.id === 'number' ? data.id : null
  } catch (err) {
    console.error('Create chat session failed:', err)
  }
  return currentSessionId.value
}

/** 清除当前角色的记忆(带二次确认) */
const clearingMemory = ref(false)
const onClearMemory = async () => {
  if (clearingMemory.value) return
  clearingMemory.value = true
  try {
    await clearPersonaMemories(activePersona.value)
    antMessage.success(t('personaMemory.cleared'))
  } catch (err) {
    console.error('Clear persona memories failed:', err)
    antMessage.error(t('result.chat.networkError'))
  } finally {
    clearingMemory.value = false
  }
}

/** 相对时间格式化(会话列表用),跟随当前语言 */
const formatRelativeTime = (input: string): string => {
  const ts = Date.parse(input)
  if (!Number.isFinite(ts)) return ''
  const diff = Date.now() - ts
  const rtf = new Intl.RelativeTimeFormat(locale.value, { numeric: 'auto' })
  if (diff < 60_000) return rtf.format(-Math.round(diff / 1000), 'second')
  if (diff < 3_600_000) return rtf.format(-Math.round(diff / 60_000), 'minute')
  if (diff < 86_400_000) return rtf.format(-Math.round(diff / 3_600_000), 'hour')
  if (diff < 30 * 86_400_000) return rtf.format(-Math.round(diff / 86_400_000), 'day')
  return rtf.format(-Math.round(diff / (30 * 86_400_000)), 'month')
}

const quickQuestions = [
  {
    labelKey: 'result.chat.quickPriceLabel',
    questionKey: 'result.chat.quickPriceQuestion',
  },
  {
    labelKey: 'result.chat.quickSuitabilityLabel',
    questionKey: 'result.chat.quickSuitabilityQuestion',
  },
  {
    labelKey: 'result.chat.quickMealLabel',
    questionKey: 'result.chat.quickMealQuestion',
  },
]

const chatPlaceholder = computed(() => {
  if (activePersona.value === 'du-fu') return t('personas.duFu.placeholder')
  if (activePersona.value === 'zhuge-liang') return t('personas.zhuge.placeholder')
  if (!props.tripPlan) return t('result.noTripPlanDesc')
  return t('result.chat.placeholder')
})

const scrollChatToBottom = () => {
  nextTick(() => {
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
    }
  })
}

watch(chatOpen, (open) => {
  if (open) {
    // 打开面板时刷新登录态(登录/登出发生在其他组件)
    isLoggedIn.value = hasToken()
    scrollChatToBottom()
  }
})

const openChatPanel = () => {
  if (!chatOpen.value) {
    chatOpen.value = true
  }
}

const closeChatPanel = () => {
  chatOpen.value = false
}

// ===== 悬浮助手拖拽移动(超过阈值视为拖动,否则仍按点击打开面板) =====
const AI_CHAT_POS_KEY = 'aiChatWidgetPos'
const DRAG_THRESHOLD = 5 // px,超过该位移判定为拖拽

const rootRef = ref<HTMLElement | null>(null)
const isDragging = ref(false)
const offset = ref({ x: 0, y: 0 })

// 恢复上次拖拽保存的位置
try {
  const saved = JSON.parse(localStorage.getItem(AI_CHAT_POS_KEY) || 'null')
  if (saved && Number.isFinite(saved.x) && Number.isFinite(saved.y)) {
    offset.value = { x: saved.x, y: saved.y }
  }
} catch {
  /* 忽略损坏的本地缓存 */
}

let dragStart: { px: number; py: number; ox: number; oy: number; base: DOMRect } | null = null
let dragMoved = false
let suppressClick = false

const onDragStart = (e: PointerEvent) => {
  if (e.button !== 0 || !rootRef.value) return
  // base 为未加位移时的可视矩形,用于拖拽中的出屏边界修正
  const rect = rootRef.value.getBoundingClientRect()
  dragStart = {
    px: e.clientX,
    py: e.clientY,
    ox: offset.value.x,
    oy: offset.value.y,
    base: new DOMRect(rect.left - offset.value.x, rect.top - offset.value.y, rect.width, rect.height),
  }
  dragMoved = false
  window.addEventListener('pointermove', onDragMove)
  window.addEventListener('pointerup', onDragEnd)
}

const onDragMove = (e: PointerEvent) => {
  if (!dragStart) return
  const dx = e.clientX - dragStart.px
  const dy = e.clientY - dragStart.py
  if (!dragMoved && Math.hypot(dx, dy) < DRAG_THRESHOLD) return
  if (!dragMoved) {
    dragMoved = true
    isDragging.value = true
  }
  // 限制组件可视区域始终留在窗口内
  let nx = dragStart.ox + dx
  let ny = dragStart.oy + dy
  nx = Math.min(Math.max(nx, -dragStart.base.left), window.innerWidth - dragStart.base.width - dragStart.base.left)
  ny = Math.min(Math.max(ny, -dragStart.base.top), window.innerHeight - dragStart.base.height - dragStart.base.top)
  offset.value = { x: nx, y: ny }
}

const onDragEnd = () => {
  window.removeEventListener('pointermove', onDragMove)
  window.removeEventListener('pointerup', onDragEnd)
  if (dragMoved) {
    // 拖拽结束后拦截随后的 click,避免误触打开/关闭面板
    suppressClick = true
    setTimeout(() => {
      suppressClick = false
    }, 80)
    localStorage.setItem(AI_CHAT_POS_KEY, JSON.stringify(offset.value))
  }
  dragStart = null
  dragMoved = false
  isDragging.value = false
}

const onClickCapture = (e: MouseEvent) => {
  if (suppressClick) {
    e.stopPropagation()
    e.preventDefault()
  }
}

const sendQuickQuestion = (q: string) => {
  chatInput.value = q
  void sendChatMessage()
}

const sendChatMessage = async () => {
  const text = chatInput.value.trim()
  if (!text || chatLoading.value || !canChat.value) return

  const personaOption = activePersonaOption.value
  chatHistory.value.push({ role: 'user', content: text })
  chatInput.value = ''
  chatLoading.value = true
  scrollChatToBottom()

  // 预先插入空的助手消息,流式增量追加,实现打字机效果
  const assistantMsg = reactive<ChatMessage>({ role: 'assistant', content: '' })
  chatHistory.value.push(assistantMsg)

  try {
    // 登录用户:首次发言时懒创建会话,后续请求携带 session_id;未登录不带(后端不落库)
    const sessionId = await ensureSession()
    if (isPersonaMode.value) {
      // ===== 历史人物角色对话:SSE 流式 POST /api/trip/persona-chat(打字机效果与普通问答一致) =====
      const apiBase = getRuntimeApiBaseUrl()
      const res = await fetch(`${apiBase}/api/trip/persona-chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          persona_id: activePersona.value,
          messages: chatHistory.value.slice(0, -1).map((m) => ({ role: m.role, content: m.content })),
          language: locale.value,
          ...(sessionId !== null ? { session_id: sessionId } : {}),
          memory_enabled: personaMemoryEnabled.value,
        }),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      // 角色气泡配色在流式开始前应用
      ;(assistantMsg as PersonaChatMessage).personaClass = personaOption.themeClass
      await consumeSSEResponse(res, assistantMsg)
    } else {
      // ===== 普通行程问答:SSE 流式输出 =====
      await sendAssistantStream(text, assistantMsg, sessionId)
    }
  } catch (err) {
    console.error('Chat error:', err)
    if (!assistantMsg.content) {
      assistantMsg.content = t('result.chat.networkError')
    } else {
      assistantMsg.content += '\n\n⚠️ ' + t('result.chat.networkError')
    }
  } finally {
    chatLoading.value = false
    scrollChatToBottom()
  }
}

/** 普通行程问答:走 /api/chat/ask/stream 的 SSE 流式管线,增量填充 assistantMsg */
const sendAssistantStream = async (text: string, assistantMsg: ChatMessage, sessionId: number | null) => {
  const apiBase = getRuntimeApiBaseUrl()
  const res = await fetch(`${apiBase}/api/chat/ask/stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message: text,
      trip_plan: props.tripPlan,
      history: chatHistory.value.slice(0, -2),
      language: locale.value,
      ...(sessionId !== null ? { session_id: sessionId } : {}),
    }),
  })
  if (!res.ok || !res.body) {
    throw new Error(`HTTP ${res.status}`)
  }
  await consumeSSEResponse(res, assistantMsg)
}

/** SSE 通用消费器:逐行解析 data: {"delta"|"error"} / [DONE],增量填充 assistantMsg(persona 与普通问答共用) */
const consumeSSEResponse = async (res: Response, assistantMsg: ChatMessage) => {
  const reader = res.body!.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let failed = false

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || '' // 最后一段可能不完整,留待下一批
    for (const rawLine of lines) {
      const line = rawLine.trim()
      if (!line.startsWith('data:')) continue
      const payload = line.slice(5).trim()
      if (payload === '[DONE]') continue
      try {
        const evt = JSON.parse(payload)
        if (evt.delta) {
          assistantMsg.content += evt.delta
          scrollChatToBottom()
        } else if (evt.error) {
          failed = true
          assistantMsg.content += (assistantMsg.content ? '\n\n' : '') + `⚠️ ${evt.error}`
        }
      } catch {
        /* 忽略无法解析的行 */
      }
    }
  }

  // 若模型没有返回任何内容,给出兜底提示
  if (!assistantMsg.content) {
    assistantMsg.content = failed ? t('result.chat.networkError') : t('result.chat.replyFallback')
  }
}
</script>

<style scoped lang="scss">
.ai-chat-floating {
  position: fixed;
  left: 8px;
  bottom: 8px;
  z-index: 1000;
  /* translate 在 scale 之前组合,位移不受缩放影响 */
  transform: translate(var(--chat-tx, 0px), var(--chat-ty, 0px)) scale(0.3);
  cursor: grab;
  touch-action: none;
}

.ai-chat-floating.dragging {
  cursor: grabbing;
  /* 拖拽中禁止选中文字,避免拖动时误选页面内容 */
  user-select: none;
}

.container-ai-input {
  --perspective: 1000px;
  --translateY: 45px;
  position: absolute;
  inset: 0;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  transform-style: preserve-3d;
}

.container-wrap {
  display: flex;
  align-items: center;
  justify-items: center;
  position: absolute;
  left: 0;
  bottom: 0;
  z-index: 9;
  transform-style: preserve-3d;
  cursor: default;
  padding: 4px;
  transition: all 0.3s ease;
}

.container-wrap:hover {
  padding: 0;
}

.container-wrap:active {
  transform: scale(0.95);
}

.container-wrap:after {
  content: "";
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translateX(-50%) translateY(-55%);
  width: 12rem;
  height: 11rem;
  background-color: #e8efed;
  border-radius: 2rem;
  transition: all 0.3s ease;
}

.container-wrap:hover:after {
  transform: translateX(-50%) translateY(-50%);
  height: 10rem;
}

.container-wrap.open .eyes {
  opacity: 0;
}

.container-wrap.open .content-card {
  width: 1260px;
  height: 1100px;
}

.container-wrap.open .background-blur-balls {
  border-radius: 24px;
}

.container-wrap.open .container-ai-chat {
  opacity: 1;
  visibility: visible;
  z-index: 99999;
  pointer-events: auto;
}

.card {
  width: 100%;
  height: 100%;
  /* background-color: #fff; */
  position: relative;
  transform-style: preserve-3d;
  will-change: transform;
  transition: all 0.6s ease;
  border-radius: 3rem;
  display: flex;
  align-items: flex-end;
  transform: translateZ(50px);
  justify-content: flex-start;
}

.card:hover {
  box-shadow:
    0 10px 40px rgba(0, 0, 0, 0.12),
    inset 0 0 10px rgba(255, 255, 255, 0.5);
}

.background-blur-balls {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  z-index: -10;
  border-radius: 3rem;
  transition: all 0.3s ease;
  background-color: rgba(255, 253, 248, 0.92);
  border: 1px solid #e8efed;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  overflow: hidden;
}
.balls {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translateX(-50%) translateY(-50%);
  animation: rotate-background-balls 10s linear infinite;
}

.container-wrap:hover .balls {
  animation-play-state: paused;
}

.background-blur-balls .ball {
  width: 6rem;
  height: 6rem;
  position: absolute;
  border-radius: 50%;
  filter: blur(30px);
}

.background-blur-balls .ball.violet {
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  background-color: #3e7d8a;
}

.background-blur-balls .ball.green {
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  background-color: #8fc4cd;
}

.background-blur-balls .ball.rosa {
  top: 50%;
  left: 0;
  transform: translateY(-50%);
  background-color: #b8453e;
}

.background-blur-balls .ball.cyan {
  top: 50%;
  right: 0;
  transform: translateY(-50%);
  background-color: #5da4b1;
}

.content-card {
  width: 12rem;
  height: 12rem;
  display: flex;
  border-radius: 3rem;
  transition: all 0.3s ease;
  overflow: hidden;
}

.content-card.clickable {
  cursor: pointer;
}

.background-blur-card {
  width: 100%;
  height: 100%;
  backdrop-filter: blur(50px);
}

.eyes {
  position: absolute;
  left: 50%;
  bottom: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  height: 52px;
  gap: 2rem;
  transition: all 0.3s ease;

  & .eye {
    width: 26px;
    height: 52px;
    background-color: #b8453e;
    border-radius: 16px;
    animation: animate-eyes 10s infinite linear;
    transition: all 0.3s ease;
  }
}

.eyes.happy {
  display: none;
  color: #b8453e;
  gap: 0;

  & svg {
    width: 60px;
  }
}

.container-wrap:hover .eyes .eye {
  display: none;
}

.container-wrap:hover .eyes.happy {
  display: flex;
}

.container-ai-chat {
  position: absolute;
  width: 100%;
  height: 100%;
  padding: 36px;
  opacity: 0;
  pointer-events: none;
}

.container-ai-chat .chat-close-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  z-index: 3;
  width: 62px;
  height: 62px;
  border: none;
  border-radius: 50%;
//   background: rgba(0, 0, 0, 0.12);
//   color: rgba(255, 255, 255, 0.92);
  font-size: 50px;
  line-height: 1;
  cursor: pointer;
}

.container-wrap .card .chat {
  display: flex;
  justify-content: space-between;
  flex-direction: column;
  border-radius: 15px;
  width: 100%;
  height: 100%;
  padding: 90px 20px 20px;
  overflow: hidden;
  background-color: #fffdf8;
}

.container-wrap .card .chat .chat-bot {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  transition: all 0.3s ease;
}

.card .chat .chat-bot .chat-history {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  border-radius: 12px;
  padding: 26px 26px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: rgba(93, 164, 177, 0.06);

  &::-webkit-scrollbar {
    width: 12px;
  }

  &::-webkit-scrollbar-thumb {
    background: #d9e2e0;
    border-radius: 5px;
  }
}

.card .chat .chat-bot .chat-empty {
  color: #8a9a9e;
  line-height: 1.5;

  p {
    margin: 0;
    font-size: 48px;
    font-weight: 600;
  }
}

.card .chat .chat-bot .chat-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 12px;
}

.card .chat .chat-bot .chat-suggestion {
  border: none;
  border-radius: 20px;
  padding: 12px 24px;
  font-size: 42px;
  font-weight: 500;
  background-color: #b8453e;
  border-color: #b8453e;
  color: #ffffff;
  opacity: 1;
  filter: alpha(opacity=100);
//   background: linear-gradient(135deg, #ff4141, #9147ff, #3b82f6);
  cursor: pointer;
  transition: all 0.2s ease;
}

.card .chat .chat-bot .chat-suggestion:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

// ===== 对话角色切换(普通助手/杜甫/诸葛亮) =====
.persona-switch {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 4px 14px;
}

.persona-switch-label {
  font-size: 34px;
  font-weight: 600;
  color: #8a9a9e;
  white-space: nowrap;
}

.persona-switch-chips {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.persona-chip {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  border: 2px solid #d9e2e0;
  border-radius: 999px;
  padding: 10px 22px;
  font-size: 34px;
  font-weight: 500;
  background-color: #ffffff;
  color: #2e3a3d;
  cursor: pointer;
  transition: all 0.2s ease;

  .persona-chip-avatar {
    font-size: 38px;
    line-height: 1;
  }

  &:hover:not(:disabled) {
    transform: translateY(-4px);
    border-color: #5da4b1;
  }

  &.active {
    border-color: #b8453e;
    background-color: rgba(184, 69, 62, 0.08);
    color: #b8453e;
  }

  &.persona-zhuge.active {
    border-color: #4a7c59;
    background-color: rgba(74, 124, 89, 0.1);
    color: #3d6a4b;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

// ===== 历史会话 / 新会话工具条(仅登录用户) =====
.chat-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 4px 14px;
}

.chat-tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  border: none;
  border-radius: 999px;
  padding: 8px 20px;
  font-size: 30px;
  font-weight: 500;
  background-color: rgba(93, 164, 177, 0.12);
  color: #3e7d8a;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover:not(:disabled) {
    background-color: rgba(93, 164, 177, 0.24);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

// ===== 角色记忆开关行(仅角色模式) =====
.persona-memory-row {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 0 4px 14px;
}

// 面板按大尺寸设计(scale 0.3),开关同步放大
.persona-memory-row .persona-memory-switch {
  transform: scale(2);
  transform-origin: left center;
  margin-right: 18px;
}

.persona-memory-label {
  font-size: 34px;
  font-weight: 600;
  color: #2e3a3d;
  white-space: nowrap;
}

.persona-memory-clear {
  border: none;
  background: transparent;
  padding: 4px 8px;
  font-size: 28px;
  color: #b8453e;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    opacity: 0.75;
  }
}

.persona-memory-tip {
  margin-left: auto;
  font-size: 28px;
  color: #8a9a9e;
}

// 未登录提示(面板底部低调展示)
.chat-login-tip {
  padding: 10px 4px 0;
  font-size: 28px;
  color: #8a9a9e;
  text-align: center;
}

// 角色模式头部条(头像 + 名字 + 头衔)
.persona-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  border-radius: 16px;
  font-size: 34px;

  .persona-header-avatar {
    font-size: 44px;
    line-height: 1;
  }

  .persona-header-name {
    font-weight: 700;
    color: #b8453e;
  }

  .persona-header-title {
    font-weight: 500;
    color: #8a6a4f;
    background: rgba(184, 69, 62, 0.1);
    border-radius: 999px;
    padding: 2px 16px;
    font-size: 30px;
  }

  .persona-header-tip {
    margin-left: auto;
    font-size: 28px;
    color: #8a9a9e;
  }

  &.persona-dufu {
    background: linear-gradient(135deg, rgba(246, 240, 229, 0.92), rgba(240, 228, 205, 0.6));
  }

  &.persona-zhuge {
    background: linear-gradient(135deg, rgba(230, 239, 228, 0.92), rgba(213, 230, 216, 0.6));

    .persona-header-name {
      color: #3d6a4b;
    }

    .persona-header-title {
      color: #3d6a4b;
      background: rgba(74, 124, 89, 0.14);
    }
  }
}

// 角色消息气泡主题:杜甫=宣纸色(暖米+朱棕),诸葛亮=青竹色(淡绿+墨绿)
.card .chat .chat-bot .chat-msg.persona-dufu {
  background: #f6efdd;
  border-left: 8px solid #b8453e;
}

.card .chat .chat-bot .chat-msg.persona-zhuge {
  background: #e6efe4;
  border-left: 8px solid #4a7c59;
}

.card .chat .chat-bot .chat-msg {
  max-width: 92%;
  font-size: 44px;
  font-weight: 500;
  line-height: 1.6;
  border-radius: 24px;
  padding: 16px 24px;
  color: #2e3a3d;
  background: #f6f0e5;
  white-space: pre-wrap;
  word-break: break-word;
}

.card .chat .chat-bot .chat-msg.user {
  margin-left: auto;
  background-color: #5da4b1;
  border-color: #5da4b1;
  color: #ffffff;
  opacity: 1;
  filter: alpha(opacity=100);

//   background: linear-gradient(135deg, #ff4141, #9147ff);
}

.card .chat .chat-bot .chat-msg.assistant {
  margin-right: auto;
}

.card .chat .chat-bot .chat-msg.typing {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: fit-content;
}

.card .chat .chat-bot .chat-msg.typing .dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
//   background: #9147ff;
  background-color: #5da4b1;
  border-color: #5da4b1;
  color: #ffffff;
  opacity: 1;
  filter: alpha(opacity=100);
  animation: aiChatDotPulse 1.4s infinite ease-in-out both;
}

.card .chat .chat-bot .chat-msg.typing .dot:nth-child(2) {
  animation-delay: 0.16s;
}

.card .chat .chat-bot .chat-msg.typing .dot:nth-child(3) {
  animation-delay: 0.32s;
}

.card .chat .chat-bot textarea {
  background-color: #ffffff;
  border-radius: 16px;
  border: 1px solid #d9e2e0;
  width: 100%;
  min-height: 156px;
  max-height: 178px;
  color: #2e3a3d;
  font-family: sans-serif;
  font-size: 48px;
  font-weight: 500;
  padding: 10px;
  resize: none;
  outline: none;

  &::-webkit-scrollbar {
    width: 6px;
    height: 10px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #d9e2e0;
    border-radius: 5px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #8a9a9e;
    cursor: pointer;
  }

  &::placeholder {
    color: #8a9a9e;
    transition: all 0.3s ease;
  }
  &:focus::placeholder {
    color: #5e6e72;
  }
}

.card .chat .options {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  padding: 20px;

  & button {
    transition: all 0.3s ease;
  }
}

.card .chat .options .btns-add {
  display: flex;
  gap: 16px;

  & button {
    display: flex;
    color: rgba(46, 58, 61, 0.25);
    background-color: transparent;
    border: none;
    cursor: pointer;
    transition: all 0.3s ease;

    &:hover {
      transform: translateY(-10px);
      color: #8a9a9e;
    }
  }
}

.card .chat .options .btn-submit {
  display: flex;
  padding: 15px;
  background-color: #b8453e;
  border-color: #b8453e;
  color: #ffffff;
  opacity: 1;
  filter: alpha(opacity=100);
//   background-image: linear-gradient(to top, #ff4141, #9147ff, #3b82f6);
  border-radius: 10px;
  box-shadow: inset 0 6px 2px -4px rgba(255, 255, 255, 0.5);
  cursor: pointer;
  border: none;
  outline: none;
  opacity: 0.7;
//   transform: translateY(-100%);
  transition: all 0.15s ease;

  & i {
    width: 60px;
    height: 60px;
    padding: 6px;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 10px;
    backdrop-filter: blur(3px);
    color: rgba(255, 255, 255, 0.85);
  }
  & svg {
    transition: all 0.3s ease;
  }
  &:hover {
    opacity: 1;
    & svg {
      color: #ffffff;
      filter: drop-shadow(0 0 5px #ffffff);
    }
  }

  &:focus svg {
    color: #ffffff;
    filter: drop-shadow(0 0 5px #ffffff);
    transform: scale(1.2) rotate(45deg) translateX(-2px) translateY(1px);
  }

  &:active {
    transform: scale(0.92);
  }
}

.card .chat .options .btn-submit:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

@keyframes aiChatDotPulse {
  0%, 80%, 100% {
    transform: scale(0.4);
    opacity: 0.4;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.area:nth-child(15):hover ~ .container-wrap .card,
.area:nth-child(15):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(-15deg) rotateY(15deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(14):hover ~ .container-wrap .card,
.area:nth-child(14):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(-15deg) rotateY(7deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(13):hover ~ .container-wrap .card,
.area:nth-child(13):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(-15deg) rotateY(0)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(12):hover ~ .container-wrap .card,
.area:nth-child(12):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(-15deg) rotateY(-7deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(11):hover ~ .container-wrap .card,
.area:nth-child(11):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(-15deg) rotateY(-15deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(10):hover ~ .container-wrap .card,
.area:nth-child(10):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(0) rotateY(15deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(9):hover ~ .container-wrap .card,
.area:nth-child(9):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(0) rotateY(7deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(8):hover ~ .container-wrap .card,
.area:nth-child(8):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(0) rotateY(0)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(7):hover ~ .container-wrap .card,
.area:nth-child(7):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(0) rotateY(-7deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(6):hover ~ .container-wrap .card,
.area:nth-child(6):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(0) rotateY(-15deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(5):hover ~ .container-wrap .card,
.area:nth-child(5):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(15deg) rotateY(15deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(4):hover ~ .container-wrap .card,
.area:nth-child(4):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(15deg) rotateY(7deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(3):hover ~ .container-wrap .card,
.area:nth-child(3):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(15deg) rotateY(0)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(2):hover ~ .container-wrap .card,
.area:nth-child(2):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(15deg) rotateY(-7deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}
.area:nth-child(1):hover ~ .container-wrap .card,
.area:nth-child(1):hover ~ .container-wrap .eyes .eye {
  transform: perspective(var(--perspective)) rotateX(15deg) rotateY(-15deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(15):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(15):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(-10deg) rotateY(8deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(14):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(14):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(-10deg) rotateY(4deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(13):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(13):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(-10deg) rotateY(0deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(12):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(12):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(-10deg) rotateY(-4deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(11):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(11):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(-10deg) rotateY(-8deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(10):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(10):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(0deg) rotateY(8deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(9):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(9):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(0deg) rotateY(4deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(8):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(8):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(0deg) rotateY(0deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(7):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(7):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(0deg) rotateY(-4deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(6):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(6):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(0deg) rotateY(-8deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(5):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(5):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(10deg) rotateY(8deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(4):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(4):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(10deg) rotateY(4deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(3):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(3):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(10deg) rotateY(0deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(2):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(2):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(10deg) rotateY(-4deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

.area:nth-child(1):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .options
  button,
.area:nth-child(1):hover
  ~ .container-wrap
  .card
  .container-ai-chat
  .chat
  .chat-bot {
  transform: perspective(var(--perspective)) rotateX(10deg) rotateY(-8deg)
    translateZ(var(--translateY)) scale3d(1, 1, 1);
}

@keyframes rotate-background-balls {
  from {
    transform: translateX(-50%) translateY(-50%) rotate(360deg);
  }
  to {
    transform: translateX(-50%) translateY(-50%) rotate(0);
  }
}

@keyframes animate-eyes {
  46% {
    height: 52px;
  }
  48% {
    height: 20px;
  }
  50% {
    height: 52px;
  }
  96% {
    height: 52px;
  }
  98% {
    height: 20px;
  }
  100% {
    height: 52px;
  }
}

@media (max-width: 768px) {
  .ai-chat-floating {
    left: 12px;
    bottom: 12px;
    width: 220px;
    height: 220px;
  }

  .container-wrap.open .content-card {
    width: 300px;
    height: 220px;
  }

  .container-wrap:after {
    width: 6.5rem;
    height: 6rem;
  }

  .container-wrap:hover:after {
    height: 6.5rem;
  }

  .content-card {
    width: 6.5rem;
    height: 6.5rem;
  }
}
</style>

<style lang="scss">
/* 历史会话抽屉:a-drawer teleport 至 body,需全局(非 scoped)样式 */
.chat-sessions-drawer {
  .ant-drawer-body {
    display: flex;
    flex-direction: column;
  }

  .session-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .session-item {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    width: 100%;
    text-align: left;
    border: 1px solid #e5e0d5;
    border-radius: 10px;
    padding: 10px 12px;
    background: #fffdf8;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      border-color: #5da4b1;
    }

    &.active {
      border-color: #b8453e;
      background: rgba(184, 69, 62, 0.06);
    }
  }

  .session-title {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
    font-weight: 600;
    color: #2e3a3d;
  }

  .session-time {
    font-size: 12px;
    color: #8a9a9e;
  }

  .session-new-btn {
    margin-top: 12px;
  }
}
</style>
