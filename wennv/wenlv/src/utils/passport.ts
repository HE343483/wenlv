/**
 * passport.ts — 数字足迹护照（MySQL 落库版）
 * 集章与后端 check_ins 表绑定：用户在景点详情页点击"打卡集章"并上传
 * 现场照片后才获得印章，数据按账号永久保存，换设备不丢。
 * 本文件只做数据适配：把打卡记录映射为护照印章（id 去重 + 时间戳）。
 */
import { listCheckIns, type CheckInItem } from '@/api/checkin'

/** 单枚印章：景点 ID + 打卡时间戳 + 打卡照片 */
export interface PassportStamp {
  id: number
  ts: number
  photos: string[]
  comment: string
}

/** 拉取我的全部打卡记录并映射为印章列表（按盖章时间升序） */
export async function fetchStamps(): Promise<PassportStamp[]> {
  const items = await listCheckIns()
  return toStamps(items)
}

/** 打卡记录 → 印章列表（去重、升序），供测试与弹窗复用 */
export function toStamps(items: CheckInItem[]): PassportStamp[] {
  const seen = new Set<number>()
  const stamps: PassportStamp[] = []
  for (const it of items) {
    if (seen.has(it.scenic_id)) continue
    seen.add(it.scenic_id)
    stamps.push({
      id: it.scenic_id,
      ts: new Date(it.visited_at || it.created_at).getTime() || 0,
      photos: Array.isArray(it.photos) ? it.photos : [],
      comment: it.comment || '',
    })
  }
  return stamps.sort((a, b) => a.ts - b.ts)
}
