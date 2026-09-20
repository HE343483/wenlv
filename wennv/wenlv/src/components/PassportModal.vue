<script setup lang="ts">
/**
 * PassportModal.vue — 「成都旅行护照」证书海报弹窗
 * 自包含 Canvas 绘制（绘制工具与熊猫配色复制自 StoryCardModal，互不依赖）：
 * 1080×1440、devicePixelRatio 放大、暖米底/墨黑/竹绿配色；
 * 版面：三语标题 + 英文小字 + 熊猫脸装饰 + 集章网格（红章/虚线空格）
 *       + 底部生成日期 + 二维码（指向 /passport）+ 平台名。
 * QRCode 失败不阻塞海报生成。
 */
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'
import QRCode from 'qrcode'
import { useLanguageStore } from '@/stores/language'
import type { ScenicItem } from '@/api/content'
import { pickName } from '@/utils/storyI18n'
import { drawMaskPoster } from '@/utils/mask-poster'

const props = defineProps<{
  open: boolean
  scenics: ScenicItem[]
  /** 已集章景点 id 集合（由父组件从后端打卡记录计算后传入） */
  stampIds: Set<number>
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const langStore = useLanguageStore()
/** 模板切换按钮文案走 vue-i18n(src/i18n)的 storyCard.* 键 */
const { t: sharedT } = useI18n()

const generating = ref(false)
const posterUrl = ref('')
/** 海报模板:熊猫(默认) / 川剧脸谱 */
const posterTemplate = ref<'panda' | 'mask'>('panda')
/** 打开弹窗时的盖章快照（保证绘制与预览进度一致） */
const stampSnapshot = ref<Set<number>>(new Set())

const collected = computed(() =>
  props.scenics.filter((s) => stampSnapshot.value.has(s.id)).length
)

const handleOpenChange = (value: boolean) => {
  emit('update:open', value)
}

/* ===== Canvas 海报绘制(1080×1440,按设备像素比放大保证清晰度) ===== */
const POSTER_W = 1080
const POSTER_H = 1440
const PAD_X = 96
const CONTENT_W = POSTER_W - PAD_X * 2

const COLOR_BG = '#F7F1E5'
const COLOR_INK = '#1F1F1F'
const COLOR_META = '#6F665A'
const COLOR_GREEN = '#2D6A4F'
const COLOR_LINE = '#D9CFC0'
const COLOR_STAMP = '#B93A2B'
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

/** 超宽文本截断加省略号 */
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

/** 单枚红章:圆形双环 + 迷你熊猫脸 */
const drawStampSeal = (
  ctx: CanvasRenderingContext2D,
  cx: number,
  cy: number,
  r: number
) => {
  ctx.save()
  ctx.strokeStyle = COLOR_STAMP
  ctx.lineWidth = 3.5
  ctx.beginPath()
  ctx.arc(cx, cy, r, 0, Math.PI * 2)
  ctx.stroke()
  ctx.lineWidth = 1.5
  ctx.beginPath()
  ctx.arc(cx, cy, r - 7, 0, Math.PI * 2)
  ctx.stroke()
  drawPandaFace(ctx, cx, cy, r * 0.58)
  ctx.restore()
}

const generateQrCanvas = async (): Promise<HTMLCanvasElement | null> => {
  try {
    const canvas = document.createElement('canvas')
    await QRCode.toCanvas(canvas, `${window.location.origin}/passport`, {
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

const drawPoster = async () => {
  const now = new Date()
  const dateText = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0'),
  ].join('-')

  // 脸谱模板:委托共享绘制模块,异常时回落到下方熊猫模板(内容数据不变)
  if (posterTemplate.value === 'mask') {
    try {
      const progressText = langStore.t('passport.progress', {
        count: collected.value,
        total: props.scenics.length,
      })
      posterUrl.value = await drawMaskPoster({
        brand: langStore.t('passport.brand'),
        headline: langStore.t('passport.posterTitle'),
        headlineEn: 'CHENGDU TRAVEL PASSPORT',
        metaText: progressText,
        qrUrl: `${window.location.origin}/passport`,
        footerBrand: langStore.t('passport.brand'),
        footerLines: [
          langStore.t('passport.issuedAt', { date: dateText }),
          progressText,
          langStore.t('passport.scanTip'),
        ],
        stamps: {
          items: props.scenics.map((item) => ({
            name: stampSnapshot.value.has(item.id)
              ? item.name_zh
              : pickName(item, langStore.lang),
            got: stampSnapshot.value.has(item.id),
          })),
        },
      })
      return
    } catch (error) {
      console.error('绘制脸谱模板护照失败,回落熊猫模板:', error)
    }
  }

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

  // 顶部:熊猫脸 + 平台名
  drawPandaFace(ctx, 116, 92, 22)
  ctx.fillStyle = COLOR_GREEN
  ctx.font = `500 26px ${FONT_STACK}`
  ctx.fillText(langStore.t('passport.brand'), 156, 101)

  // 标题(按语言)+ 英文小字
  ctx.fillStyle = COLOR_INK
  ctx.font = `900 96px ${FONT_STACK}`
  ctx.fillText(langStore.t('passport.posterTitle'), PAD_X, 248)
  ctx.fillStyle = COLOR_GREEN
  ctx.font = `700 32px ${FONT_STACK}`
  drawLetterSpaced(ctx, 'CHENGDU TRAVEL PASSPORT', PAD_X + 4, 308, 8)
  roundRectPath(ctx, PAD_X, 338, 92, 8, 4)
  ctx.fill()

  // 集章进度
  ctx.fillStyle = COLOR_META
  ctx.font = `400 30px ${FONT_STACK}`
  ctx.fillText(
    fitText(
      ctx,
      langStore.t('passport.progress', { count: collected.value, total: props.scenics.length }),
      CONTENT_W
    ),
    PAD_X,
    404
  )

  // 集章网格:5 列,已集红章 + 中文名,未集虚线空格 + 景点名浅灰
  const grid = props.scenics
  const COLS = 5
  const GAP = 20
  const GRID_TOP = 452
  const GRID_BOTTOM = 1150
  const cellW = (CONTENT_W - (COLS - 1) * GAP) / COLS
  const rows = Math.max(1, Math.ceil(grid.length / COLS))
  const rowH = (GRID_BOTTOM - GRID_TOP) / rows

  grid.forEach((item, index) => {
    const col = index % COLS
    const row = Math.floor(index / COLS)
    const x = PAD_X + col * (cellW + GAP)
    const y = GRID_TOP + row * rowH
    const cx = x + cellW / 2
    const got = stampSnapshot.value.has(item.id)
    const sealR = Math.min(cellW * 0.27, rowH * 0.28)
    const cy = y + rowH * 0.38
    const nameY = cy + sealR + 32

    if (got) {
      drawStampSeal(ctx, cx, cy, sealR)
      ctx.fillStyle = COLOR_INK
      ctx.font = `600 20px ${FONT_STACK}`
      ctx.textAlign = 'center'
      ctx.fillText(fitText(ctx, item.name_zh, cellW - 8), cx, nameY)
      ctx.textAlign = 'left'
    } else {
      ctx.strokeStyle = COLOR_LINE
      ctx.lineWidth = 2
      ctx.setLineDash([8, 7])
      roundRectPath(ctx, x, y + 8, cellW, rowH - 16, 14)
      ctx.stroke()
      ctx.setLineDash([])
      ctx.fillStyle = 'rgba(111, 102, 90, 0.30)'
      ctx.font = `900 46px ${FONT_STACK}`
      ctx.textAlign = 'center'
      ctx.fillText('？', cx, cy + 16)
      ctx.fillStyle = 'rgba(111, 102, 90, 0.55)'
      ctx.font = `400 18px ${FONT_STACK}`
      ctx.fillText(fitText(ctx, pickName(item, langStore.lang), cellW - 8), cx, nameY)
      ctx.textAlign = 'left'
    }
  })

  // 底部:二维码 + 平台名 + 生成日期 + 扫码小字
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
  ctx.fillText(langStore.t('passport.brand'), 300, 1258)
  ctx.fillStyle = COLOR_META
  ctx.font = `400 22px ${FONT_STACK}`
  ctx.fillText(langStore.t('passport.issuedAt', { date: dateText }), 300, 1296)
  ctx.fillText(
    langStore.t('passport.progress', { count: collected.value, total: props.scenics.length }),
    300,
    1328
  )
  ctx.font = `400 20px ${FONT_STACK}`
  ctx.fillText(langStore.t('passport.scanTip'), 300, 1360)

  // 离屏画布 → PNG 预览
  posterUrl.value = canvas.toDataURL('image/png')
}

const generate = async () => {
  generating.value = true
  try {
    await drawPoster()
  } catch (error) {
    console.error('绘制护照证书失败:', error)
    message.error(langStore.t('passport.generateFailed'))
  } finally {
    generating.value = false
  }
}

/** 模板切换:直接重绘,集章数据不变 */
const handleTemplateChange = async () => {
  if (generating.value) return
  generating.value = true
  try {
    await drawPoster()
  } catch (error) {
    console.error('绘制护照证书失败:', error)
    message.error(langStore.t('passport.generateFailed'))
  } finally {
    generating.value = false
  }
}

const downloadPoster = () => {
  if (!posterUrl.value) return
  try {
    const link = document.createElement('a')
    link.href = posterUrl.value
    const now = new Date()
    const dateText = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}`
    link.download = `${langStore.t('passport.posterTitle')}-${dateText}.png`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (error) {
    console.error('下载护照证书失败:', error)
  }
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      stampSnapshot.value = new Set(props.stampIds)
      void generate()
    }
  }
)
</script>

<template>
  <a-modal
    :open="open"
    class="passport-modal"
    :title="langStore.t('passport.title')"
    :width="560"
    :footer="null"
    :mask-closable="false"
    @update:open="handleOpenChange"
  >
    <a-spin :spinning="generating" :tip="langStore.t('passport.generating')">
      <div class="passport-certificate__templates">
        <a-radio-group
          v-model:value="posterTemplate"
          size="small"
          @change="handleTemplateChange"
        >
          <a-radio-button value="panda">{{ sharedT('storyCard.templatePanda') }}</a-radio-button>
          <a-radio-button value="mask">{{ sharedT('storyCard.templateMask') }}</a-radio-button>
        </a-radio-group>
      </div>
      <div class="passport-certificate">
        <img
          v-if="posterUrl"
          :src="posterUrl"
          :alt="langStore.t('passport.posterTitle')"
          class="passport-certificate__image"
        />
        <div v-else class="passport-certificate__placeholder"></div>
      </div>
      <p class="passport-certificate__tip">
        {{ langStore.t('passport.progress', { count: collected, total: scenics.length }) }}
      </p>
      <div class="passport-certificate__actions">
        <a-space size="middle">
          <a-button :disabled="generating" @click="handleOpenChange(false)">
            {{ langStore.t('passport.close') }}
          </a-button>
          <a-button type="primary" :disabled="generating || !posterUrl" @click="downloadPoster">
            {{ langStore.t('passport.download') }}
          </a-button>
        </a-space>
      </div>
    </a-spin>
  </a-modal>
</template>

<style scoped>
.passport-certificate__templates {
  display: flex;
  justify-content: center;
  margin-bottom: 14px;
}

.passport-certificate {
  display: flex;
  justify-content: center;
  min-height: 320px;
}

.passport-certificate__image {
  width: 100%;
  max-height: 62vh;
  object-fit: contain;
  border-radius: 12px;
  box-shadow: 0 12px 32px rgba(31, 31, 31, 0.16);
}

.passport-certificate__placeholder {
  width: 100%;
  max-width: 360px;
  aspect-ratio: 3 / 4;
  border-radius: 12px;
  background: linear-gradient(160deg, #f0ead9 0%, #dfe9e0 100%);
}

.passport-certificate__tip {
  margin: 16px 0 12px;
  color: #8a9a9e;
  font-size: 12px;
  text-align: center;
}

.passport-certificate__actions {
  display: flex;
  justify-content: center;
  padding-bottom: 4px;
}
</style>
