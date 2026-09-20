/** AI 对话历史会话接口封装（登录态） */
import { get, post, del } from './request'

/** 历史会话条目 */
export interface ChatSessionItem {
  id: number
  title: string
  updated_at: string
}

/** 会话内历史消息（正序） */
export interface ChatSessionMessage {
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

/** 拉取某角色的历史会话列表（最多 30 条，倒序） */
export function listChatSessions(persona_id: string): Promise<ChatSessionItem[]> {
  return get<ChatSessionItem[]>('/chat/sessions', { params: { persona_id } })
}

/** 新建会话，返回会话 ID */
export function createChatSession(persona_id: string, language: string): Promise<{ id: number }> {
  return post<{ id: number }>('/chat/sessions', { persona_id, language })
}

/** 拉取会话历史消息（正序） */
export function listSessionMessages(id: number): Promise<ChatSessionMessage[]> {
  return get<ChatSessionMessage[]>(`/chat/sessions/${id}/messages`)
}

/** 清除某角色的记忆，返回清除条数 */
export function clearPersonaMemories(persona_id: string): Promise<{ cleared: number }> {
  return del<{ cleared: number }>('/chat/memories', { params: { persona_id } })
}
