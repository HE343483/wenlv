<template>
  <a-modal
    :open="open"
    class="story-card-modal"
    :title="`🐼 ${t('storyCard.title')}`"
    :width="560"
    :footer="null"
    :mask-closable="false"
    @update:open="handleOpenChange"
  >
    <a-spin :spinning="generating" :tip="t('storyCard.generating')">
      <div class="story-card-preview">
        <img
          v-if="posterUrl"
          :src="posterUrl"
          :alt="t('storyCard.title')"
          class="story-card-image"
        />
        <div v-else class="story-card-placeholder"></div>
      </div>
      <p class="story-card-tip">{{ t('storyCard.tip') }}</p>
      <div class="story-card-actions">
        <a-space size="middle">
          <a-button :disabled="generating" @click="generate">
            {{ posterUrl ? t('storyCard.regenerate') : t('storyCard.generate') }}
          </a-button>
          <a-button type="primary" :disabled="generating || !posterUrl" @click="savePoster">
            {{ t('storyCard.save') }}
          </a-button>
        </a-space>
      </div>
    </a-spin>
  </a-modal>
</template>

<script lang="ts">
export interface StorySpot {
  name: string
  day: number
}

export interface TripStorySummary {
  city: string
  days: number
  startDate: string
  endDate: string
  spots: StorySpot[]
}
</script>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import QRCode from 'qrcode'
import { generateStoryCard, type StoryCardLanguage } from '@/trip/services/api'

const props = defineProps<{
  open: boolean
  planId: string
  summary: TripStorySummary
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const { t, locale } = useI18n()

const generating = ref(false)
const posterUrl = ref('')

const currentLanguage = computed<StoryCardLanguage>(() => {
  const raw = String(locale.value || 'zh-CN').toLowerCase()
  if (raw.startsWith('zh')) return 'zh'
  if (raw.startsWith('ja')) return 'ja'
  return 'en'
})

interface StoryCopy {
  title: string
  body: string
}

const SPOT_SEPARATORS: Record<StoryCardLanguage, string> = {
  zh: '、',
  en: ' & ',
  ja: '・',
}

const SPOT_FALLBACK: Record<StoryCardLanguage, string> = {
  zh: '古老街巷',
  en: 'old teahouse alleys',
  ja: '古い街並み',
}

// AI 文案失败时的本地模板兜底(fallback:true 或接口异常)
const LOCAL_TEMPLATES: Record<StoryCardLanguage, StoryCopy> = {
  zh: {
    title: '我在成都的{days}日漫游',
    body: '从{spots}到人间烟火，这座城市的安逸藏进了一碗盖碗茶里。',
  },
  en: {
    title: 'My {days}-Day Chengdu Wander',
    body: "From {spots} to everyday street life, the city's ease hides in a bowl of covered tea.",
  },
  ja: {
    title: '成都{days}日間の旅',
    body: '{spots}から市井の賑わいまで、この街ののどかさは一碗の蓋碗茶に息づいています。',
  },
}

const buildLocalStory = (language: StoryCardLanguage): StoryCopy => {
  const template = LOCAL_TEMPLATES[language]
  const spotNames = props.summary.spots.slice(0, 3).map((spot) => spot.name)
  const spotsText = spotNames.join(SPOT_SEPARATORS[language]) || SPOT_FALLBACK[language]
  return {
    title: template.title.replace('{days}', String(props.summary.days || 1)),
    body: template.body.replace('{spots}', spotsText),
  }
}

// ===== Canvas 海报绘制(1080×1440,按设备像素比放大保证清晰度) =====
const POSTER_W = 1080
const POSTER_H = 1440
const PAD_X = 96
const CONTENT_W = POSTER_W - PAD_X * 2

const COLOR_BG = '#F7F1E5'
const COLOR_INK = '#1F1F1F'
const COLOR_BODY = '#453F37'
const COLOR_META = '#6F665A'
const COLOR_GREEN = '#2D6A4F'
const COLOR_LINE = '#D9CFC0'
const FONT_STACK = '"PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Noto Sans SC", sans-serif'

const roundRectPath = (
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  r: number
) => {
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.arcTo(x + w, y, x + w, y + h, r)
  ctx.arcTo(x + w, y + h, x, y + h, r)
  ctx.arcTo(x, y + h, x, y, r)
  ctx.arcTo(x, y, x + w, y, r)
  ctx.closePath()
}

// 英文/日文手动换行:按字符宽度截断,ASCII 单词尽量保持在同一行
const wrapText = (ctx: CanvasRenderingContext2D, text: string, maxWidth: number): string[] => {
  const lines: string[] = []
  for (const paragraph of String(text ?? '').split('\n')) {
    const tokens = paragraph.match(/[A-Za-z0-9'’&\-]+[,.!?;:]*[ ]?|./g) ?? []
    let line = ''
    for (const token of tokens) {
      if (ctx.measureText(line + token).width <= maxWidth) {
        line += token
        continue
      }
      if (line) {
        lines.push(line.replace(/\s+$/, ''))
        line = token.replace(/^\s+/, '')
      } else {
        line = token
      }
      // 单个 token 超宽时按字符硬切
      while (ctx.measureText(line).width > maxWidth && line.length > 1) {
        let cut = line.length - 1
        while (cut > 1 && ctx.measureText(line.slice(0, cut)).width > maxWidth) cut -= 1
        lines.push(line.slice(0, cut))
        line = line.slice(cut)
      }
    }
    if (line) lines.push(line.replace(/\s+$/, ''))
  }
  return lines
}

const fitText = (ctx: CanvasRenderingContext2D, text: string, maxWidth: number): string => {
  const raw = String(text ?? '')
  if (ctx.measureText(raw).width <= maxWidth) return raw
  let output = raw
  while (output.length > 1 && ctx.measureText(`${output}…`).width > maxWidth) {
    output = output.slice(0, -1)
  }
  return `${output}…`
}

const drawLetterSpaced = (
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  spacing: number
) => {
  let cursor = x
  for (const char of text) {
    ctx.fillText(char, cursor, y)
    cursor += ctx.measureText(char).width + spacing
  }
}

const drawBambooDecor = (ctx: CanvasRenderingContext2D) => {
  ctx.save()
  ctx.strokeStyle = 'rgba(45, 106, 79, 0.12)'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.arc(POSTER_W - 120, 190, 132, 0, Math.PI * 2)
  ctx.stroke()

  ctx.fillStyle = 'rgba(45, 106, 79, 0.24)'
  const stems: Array<[number, number, number]> = [
    [POSTER_W - 172, 96, 128],
    [POSTER_W - 128, 68, 156],
    [POSTER_W - 88, 118, 106],
  ]
  for (const [x, top, height] of stems) {
    roundRectPath(ctx, x, top, 9, height, 4)
    ctx.fill()
  }

  ctx.fillStyle = 'rgba(45, 106, 79, 0.2)'
  const leaves: Array<[number, number, number]> = [
    [POSTER_W - 128, 84, -0.6],
    [POSTER_W - 160, 120, 0.5],
    [POSTER_W - 90, 140, -0.45],
  ]
  for (const [x, y, rotate] of leaves) {
    ctx.save()
    ctx.translate(x, y)
    ctx.rotate(rotate)
    ctx.beginPath()
    ctx.ellipse(0, 0, 30, 8, 0, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()
  }
  ctx.restore()
}

const drawPandaFace = (ctx: CanvasRenderingContext2D, cx: number, cy: number, r: number) => {
  ctx.save()
  ctx.fillStyle = COLOR_INK
  const ears: Array<[number, number]> = [
    [-0.72, -0.72],
    [0.72, -0.72],
  ]
  for (const [dx, dy] of ears) {
    ctx.beginPath()
    ctx.arc(cx + dx * r, cy + dy * r, 0.38 * r, 0, Math.PI * 2)
    ctx.fill()
  }

  ctx.beginPath()
  ctx.arc(cx, cy, r, 0, Math.PI * 2)
  ctx.fillStyle = '#FFFFFF'
  ctx.fill()
  ctx.lineWidth = Math.max(2, r * 0.08)
  ctx.strokeStyle = COLOR_INK
  ctx.stroke()

  ctx.fillStyle = COLOR_INK
  const patches: Array<[number, number]> = [
    [-0.4, 0],
    [0.4, 0],
  ]
  patches.forEach(([dx, dy], index) => {
    ctx.save()
    ctx.translate(cx + dx * r, cy + dy * r)
    ctx.rotate(index === 0 ? -0.45 : 0.45)
    ctx.beginPath()
    ctx.ellipse(0, 0, 0.22 * r, 0.3 * r, 0, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()
  })

  ctx.beginPath()
  ctx.ellipse(cx, cy + 0.36 * r, 0.11 * r, 0.08 * r, 0, 0, Math.PI * 2)
  ctx.fill()
  ctx.restore()
}

const generateQrCanvas = async (): Promise<HTMLCanvasElement | null> => {
  try {
    const canvas = document.createElement('canvas')
    await QRCode.toCanvas(canvas, window.location.origin, {
      width: 320,
      margin: 1,
      color: { dark: COLOR_INK, light: '#FFFFFF' },
    })
    return canvas
  } catch (error) {
    console.error('生成二维码失败:', error)
    return null
  }
}

const drawPoster = async (story: StoryCopy) => {
  const canvas = document.createElement('canvas')
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  canvas.width = POSTER_W * dpr
  canvas.height = POSTER_H * dpr
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.scale(dpr, dpr)
  ctx.textAlign = 'left'
  ctx.textBaseline = 'alphabetic'

  // 背景:暖米色 + 顶部黑色 / 底部竹绿色条带(熊猫配色)
  ctx.fillStyle = COLOR_BG
  ctx.fillRect(0, 0, POSTER_W, POSTER_H)
  ctx.fillStyle = COLOR_INK
  ctx.fillRect(0, 0, POSTER_W, 14)
  ctx.fillStyle = COLOR_GREEN
  ctx.fillRect(0, POSTER_H - 14, POSTER_W, 14)

  drawBambooDecor(ctx)
  drawPandaFace(ctx, 116, 92, 22)

  // 顶部:平台名
  ctx.fillStyle = COLOR_GREEN
  ctx.font = `500 26px ${FONT_STACK}`
  ctx.fillText(t('storyCard.brand'), 156, 101)

  // 城市名大字
  ctx.fillStyle = COLOR_INK
  ctx.font = `900 138px ${FONT_STACK}`
  ctx.fillText(t('storyCard.cityName'), PAD_X, 250)
  ctx.fillStyle = COLOR_GREEN
  ctx.font = `700 38px ${FONT_STACK}`
  drawLetterSpaced(ctx, t('storyCard.cityNameEn'), PAD_X + 4, 316, 10)
  roundRectPath(ctx, PAD_X, 346, 92, 8, 4)
  ctx.fill()

  // 天数与日期
  ctx.fillStyle = COLOR_META
  ctx.font = `400 30px ${FONT_STACK}`
  const dateText = [props.summary.startDate, props.summary.endDate].filter(Boolean).join(' — ')
  const metaText = [
    t('storyCard.daysMeta', { count: props.summary.days || 1 }),
    dateText,
  ]
    .filter(Boolean)
    .join(' · ')
  ctx.fillText(fitText(ctx, metaText, CONTENT_W), PAD_X, 410)

  // 故事文案:title 大字 + body 分行排版
  ctx.fillStyle = COLOR_INK
  ctx.font = `800 52px ${FONT_STACK}`
  const titleLines = wrapText(ctx, story.title, CONTENT_W).slice(0, 2)
  let y = 486
  for (const line of titleLines) {
    ctx.fillText(line, PAD_X, y)
    y += 68
  }

  ctx.fillStyle = COLOR_BODY
  ctx.font = `400 29px ${FONT_STACK}`
  y += 6
  const bodyLines = wrapText(ctx, story.body, CONTENT_W).slice(0, 4)
  for (const line of bodyLines) {
    ctx.fillText(line, PAD_X, y)
    y += 48
  }

  // 分隔线
  ctx.strokeStyle = COLOR_LINE
  ctx.lineWidth = 2
  ctx.setLineDash([10, 8])
  ctx.beginPath()
  ctx.moveTo(PAD_X, 806)
  ctx.lineTo(POSTER_W - PAD_X, 806)
  ctx.stroke()
  ctx.setLineDash([])

  // 站点路线时间线(最多 6 个)
  ctx.fillStyle = COLOR_GREEN
  roundRectPath(ctx, PAD_X, 842, 14, 14, 3)
  ctx.fill()
  ctx.font = `700 30px ${FONT_STACK}`
  ctx.fillText(t('storyCard.routeLabel'), PAD_X + 26, 858)

  const routeSpots = props.summary.spots.slice(0, 6)
  routeSpots.forEach((spot, index) => {
    const rowY = 926 + index * 44
    ctx.fillStyle = COLOR_GREEN
    ctx.beginPath()
    ctx.arc(112, rowY - 9, 7, 0, Math.PI * 2)
    ctx.fill()
    if (index < routeSpots.length - 1) {
      ctx.strokeStyle = 'rgba(45, 106, 79, 0.35)'
      ctx.lineWidth = 2
      ctx.beginPath()
      ctx.moveTo(112, rowY - 2)
      ctx.lineTo(112, rowY + 28)
      ctx.stroke()
    }
    ctx.fillStyle = COLOR_GREEN
    ctx.font = `700 22px ${FONT_STACK}`
    ctx.fillText(`D${spot.day || index + 1}`, 136, rowY)
    ctx.fillStyle = '#23201B'
    ctx.font = `500 28px ${FONT_STACK}`
    ctx.fillText(fitText(ctx, spot.name, POSTER_W - PAD_X - 196), 196, rowY)
  })

  // 底部:二维码 + 平台名 + 三语扫码小字
  roundRectPath(ctx, PAD_X, 1208, 164, 164, 16)
  ctx.fillStyle = '#FFFFFF'
  ctx.shadowColor = 'rgba(31, 31, 31, 0.1)'
  ctx.shadowBlur = 18
  ctx.fill()
  ctx.shadowBlur = 0
  ctx.strokeStyle = '#E7DECC'
  ctx.lineWidth = 2
  roundRectPath(ctx, PAD_X, 1208, 164, 164, 16)
  ctx.stroke()

  const qrCanvas = await generateQrCanvas()
  if (qrCanvas) {
    ctx.drawImage(qrCanvas, PAD_X + 16, 1224, 132, 132)
  }

  ctx.fillStyle = COLOR_INK
  ctx.font = `800 32px ${FONT_STACK}`
  ctx.fillText(t('app.title'), 300, 1258)
  ctx.fillStyle = COLOR_META
  ctx.font = `400 21px ${FONT_STACK}`
  const scanLines = [t('storyCard.scanTip'), t('storyCard.scanTipEn'), t('storyCard.scanTipJa')]
  scanLines.forEach((line, index) => {
    ctx.fillText(line, 300, 1300 + index * 34)
  })

  // 离屏画布 → PNG 预览
  posterUrl.value = canvas.toDataURL('image/png')
}

const generate = async () => {
  generating.value = true
  let story: StoryCopy | null = null
  if (props.planId) {
    try {
      const response = await generateStoryCard(props.planId, currentLanguage.value)
      if (response && !response.fallback && response.title && response.body) {
        story = { title: response.title, body: response.body }
      }
    } catch (error) {
      console.error('生成旅行故事卡失败,使用本地模板兜底:', error)
    }
  }
  if (!story) {
    story = buildLocalStory(currentLanguage.value)
    if (props.planId) {
      message.info(t('storyCard.fallbackNotice'))
    }
  }
  try {
    await drawPoster(story)
  } catch (error) {
    console.error('绘制故事卡海报失败:', error)
    message.error(t('api.generateStoryCardFailed'))
  } finally {
    generating.value = false
  }
}

const savePoster = () => {
  if (!posterUrl.value) return
  try {
    const link = document.createElement('a')
    link.href = posterUrl.value
    link.download = `${t('storyCard.filePrefix')}-${props.summary.startDate || 'trip'}.png`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    message.success(t('storyCard.saveSuccess'))
  } catch (error: any) {
    message.error(t('storyCard.saveFailed', { error: error?.message || error }))
  }
}

const handleOpenChange = (value: boolean) => {
  emit('update:open', value)
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      void generate()
    }
  }
)
</script>

<style scoped>
.story-card-preview {
  display: flex;
  justify-content: center;
  min-height: 320px;
}

.story-card-image {
  width: 100%;
  max-height: 62vh;
  object-fit: contain;
  border-radius: 12px;
  box-shadow: 0 12px 32px rgba(31, 31, 31, 0.16);
}

.story-card-placeholder {
  width: 100%;
  max-width: 360px;
  aspect-ratio: 3 / 4;
  border-radius: 12px;
  background: linear-gradient(160deg, #f0ead9 0%, #dfe9e0 100%);
}

.story-card-tip {
  margin: 16px 0 12px;
  color: #8a9a9e;
  font-size: 12px;
  text-align: center;
}

.story-card-actions {
  display: flex;
  justify-content: center;
  padding-bottom: 4px;
}
</style>
