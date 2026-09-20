/**
 * mask-poster.ts — 川剧脸谱分享海报 Canvas 绘制模块
 * 与熊猫模板(StoryCardModal / PassportModal)共用 1080×1440 版式骨架:
 * 顶部标题大字 / 中部正文(wrapText)或集章网格 / 底部二维码与日期。
 * 配色:米白纸底 + 川剧红 #C0392B / 藏蓝 #2C3E50 / 金 #D4AC0D,
 * 与熊猫模板的暖米色+黑/竹绿条带形成鲜明对比。
 * 所有对称图形均按 [-1, 1] 镜像循环绘制,保证左右严格对称;
 * 绘制异常由调用方 try-catch 回落到熊猫模板。
 */
import QRCode from 'qrcode'

export const POSTER_W = 1080
export const POSTER_H = 1440

const PAD_X = 96
const CONTENT_W = POSTER_W - PAD_X * 2

export const MASK_COLORS = {
  bg: '#FBF5EA',
  ink: '#26201A',
  body: '#4A3F35',
  meta: '#7A6A55',
  red: '#C0392B',
  blue: '#2C3E50',
  gold: '#D4AC0D',
  deepRed: '#8E2A22',
  line: '#E4D5BC',
  qrBorder: '#E3C86B',
} as const

const FONT_STACK =
  '"PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Noto Sans SC", sans-serif'

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

/** 英文/日文手动换行:按字符宽度截断,ASCII 单词尽量保持在同一行 */
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

/* QR 码按 URL 缓存:模板来回切换时避免重复生成(生成耗时大头) */
const qrCache = new Map<string, HTMLCanvasElement>()

const generateQrCanvas = async (url: string): Promise<HTMLCanvasElement | null> => {
  const cached = qrCache.get(url)
  if (cached) return cached
  try {
    const canvas = document.createElement('canvas')
    await QRCode.toCanvas(canvas, url, {
      width: 320,
      margin: 1,
      color: { dark: MASK_COLORS.blue, light: '#FFFFFF' },
    })
    qrCache.set(url, canvas)
    return canvas
  } catch (error) {
    console.error('生成二维码失败:', error)
    return null
  }
}

/**
 * 川剧脸谱(简化风格化):椭圆白底脸型 + 墨线描边 + 内圈金线;
 * 左右对称绘制金色火焰眉 / 蓝色眼窝 / 白眼珠 / 红色颊纹 / 额侧金饰,
 * 居中绘制红色额印 / 鼻 / 笑口。红蓝金三色 + 白底。
 */
export const drawMaskFace = (
  ctx: CanvasRenderingContext2D,
  cx: number,
  cy: number,
  r: number
) => {
  ctx.save()

  // 脸型轮廓
  ctx.beginPath()
  ctx.ellipse(cx, cy, 0.86 * r, r, 0, 0, Math.PI * 2)
  ctx.fillStyle = '#FFFDF8'
  ctx.fill()
  ctx.lineWidth = Math.max(1.5, r * 0.06)
  ctx.strokeStyle = MASK_COLORS.ink
  ctx.stroke()

  // 内圈金线
  ctx.beginPath()
  ctx.ellipse(cx, cy, 0.76 * r, 0.9 * r, 0, 0, Math.PI * 2)
  ctx.lineWidth = Math.max(1, r * 0.028)
  ctx.strokeStyle = MASK_COLORS.gold
  ctx.stroke()

  // 左右对称区(side = -1 / 1 镜像)
  for (const side of [-1, 1] as const) {
    const sx = (dx: number) => cx + side * dx * r

    // 金色火焰眉(横长椭圆,外端上挑)
    ctx.fillStyle = MASK_COLORS.gold
    ctx.save()
    ctx.translate(sx(0.34), cy - 0.44 * r)
    ctx.rotate(side * -0.5)
    ctx.beginPath()
    ctx.ellipse(0, 0, 0.16 * r, 0.055 * r, 0, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()

    // 蓝色眼窝(竖长椭圆,上端外挑)
    ctx.fillStyle = MASK_COLORS.blue
    ctx.save()
    ctx.translate(sx(0.34), cy - 0.06 * r)
    ctx.rotate(side * 0.5)
    ctx.beginPath()
    ctx.ellipse(0, 0, 0.185 * r, 0.26 * r, 0, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()

    // 白眼珠 + 墨瞳
    ctx.fillStyle = '#FFFFFF'
    ctx.beginPath()
    ctx.arc(sx(0.34), cy - 0.02 * r, 0.06 * r, 0, Math.PI * 2)
    ctx.fill()
    ctx.fillStyle = MASK_COLORS.ink
    ctx.beginPath()
    ctx.arc(sx(0.34), cy - 0.02 * r, 0.028 * r, 0, Math.PI * 2)
    ctx.fill()

    // 红色颊纹
    ctx.fillStyle = MASK_COLORS.red
    ctx.save()
    ctx.translate(sx(0.6), cy + 0.3 * r)
    ctx.rotate(side * 0.35)
    ctx.beginPath()
    ctx.ellipse(0, 0, 0.07 * r, 0.14 * r, 0, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()

    // 额侧金饰(额印两翼)
    ctx.fillStyle = MASK_COLORS.gold
    ctx.save()
    ctx.translate(sx(0.2), cy - 0.6 * r)
    ctx.rotate(side * 0.3)
    ctx.beginPath()
    ctx.ellipse(0, 0, 0.05 * r, 0.11 * r, 0, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()
  }

  // 红色额印
  ctx.fillStyle = MASK_COLORS.red
  ctx.beginPath()
  ctx.ellipse(cx, cy - 0.52 * r, 0.075 * r, 0.16 * r, 0, 0, Math.PI * 2)
  ctx.fill()

  // 鼻
  ctx.fillStyle = MASK_COLORS.ink
  ctx.beginPath()
  ctx.ellipse(cx, cy + 0.16 * r, 0.05 * r, 0.035 * r, 0, 0, Math.PI * 2)
  ctx.fill()

  // 笑口
  ctx.strokeStyle = MASK_COLORS.ink
  ctx.lineWidth = Math.max(1.5, r * 0.05)
  ctx.beginPath()
  ctx.arc(cx, cy + 0.34 * r, 0.26 * r, Math.PI * 0.15, Math.PI * 0.85)
  ctx.stroke()

  ctx.restore()
}

/** 脸谱版红章:双环 + 迷你脸谱(护照集章格用) */
const drawMaskStampSeal = (
  ctx: CanvasRenderingContext2D,
  cx: number,
  cy: number,
  r: number
) => {
  ctx.save()
  ctx.strokeStyle = MASK_COLORS.red
  ctx.lineWidth = 3.5
  ctx.beginPath()
  ctx.arc(cx, cy, r, 0, Math.PI * 2)
  ctx.stroke()
  ctx.lineWidth = 1.5
  ctx.beginPath()
  ctx.arc(cx, cy, r - 7, 0, Math.PI * 2)
  ctx.stroke()
  drawMaskFace(ctx, cx, cy, r * 0.58)
  ctx.restore()
}

/** 脸谱版背景:米白纸底 + 顶部深红/底部藏蓝条带 + 右上角半透明大脸谱水印 */
const drawMaskBackground = (ctx: CanvasRenderingContext2D) => {
  ctx.fillStyle = MASK_COLORS.bg
  ctx.fillRect(0, 0, POSTER_W, POSTER_H)
  ctx.fillStyle = MASK_COLORS.deepRed
  ctx.fillRect(0, 0, POSTER_W, 14)
  ctx.fillStyle = MASK_COLORS.blue
  ctx.fillRect(0, POSTER_H - 14, POSTER_W, 14)

  ctx.save()
  ctx.globalAlpha = 0.1
  drawMaskFace(ctx, POSTER_W - 170, 220, 108)
  ctx.restore()
}

/** 底部:白卡二维码 + 平台名 + 小字行 */
const drawFooter = async (
  ctx: CanvasRenderingContext2D,
  options: MaskPosterOptions
) => {
  roundRectPath(ctx, PAD_X, 1208, 164, 164, 16)
  ctx.fillStyle = '#FFFFFF'
  ctx.shadowColor = 'rgba(38, 32, 26, 0.1)'
  ctx.shadowBlur = 18
  ctx.fill()
  ctx.shadowBlur = 0
  ctx.strokeStyle = MASK_COLORS.qrBorder
  ctx.lineWidth = 2
  roundRectPath(ctx, PAD_X, 1208, 164, 164, 16)
  ctx.stroke()

  const qrCanvas = await generateQrCanvas(options.qrUrl)
  if (qrCanvas) {
    ctx.drawImage(qrCanvas, PAD_X + 16, 1224, 132, 132)
  }

  ctx.fillStyle = MASK_COLORS.ink
  ctx.font = `800 32px ${FONT_STACK}`
  ctx.fillText(options.footerBrand, 300, 1258)
  ctx.fillStyle = MASK_COLORS.meta
  ctx.font = `400 21px ${FONT_STACK}`
  options.footerLines.forEach((line, index) => {
    ctx.fillText(fitText(ctx, line, POSTER_W - 300 - PAD_X), 300, 1300 + index * 34)
  })
}

export interface MaskPosterRouteSpot {
  label: string
  name: string
}

/** 故事卡中部内容 */
export interface MaskPosterStoryContent {
  title: string
  body: string
  routeLabel: string
  routeSpots: MaskPosterRouteSpot[]
}

/** 护照集章网格内容 */
export interface MaskPosterStampItem {
  name: string
  got: boolean
}

export interface MaskPosterOptions {
  /** 顶部小脸谱旁的平台名 */
  brand: string
  /** 大字标题(超宽自动缩字号) */
  headline: string
  /** 英文小字(字距排布) */
  headlineEn: string
  /** 元信息行(天数/进度等) */
  metaText: string
  /** 二维码指向的 URL */
  qrUrl: string
  footerBrand: string
  footerLines: string[]
  /** 二选一:故事卡正文 */
  story?: MaskPosterStoryContent
  /** 二选一:护照集章网格 */
  stamps?: { items: MaskPosterStampItem[] }
}

/**
 * 绘制完整脸谱模板海报(1080×1440,按设备像素比放大)。
 * 故事卡传 story,护照传 stamps,返回 PNG dataURL。
 */
export const drawMaskPoster = async (options: MaskPosterOptions): Promise<string> => {
  const canvas = document.createElement('canvas')
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  canvas.width = POSTER_W * dpr
  canvas.height = POSTER_H * dpr
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('Canvas 2D context unavailable')
  ctx.scale(dpr, dpr)
  ctx.textAlign = 'left'
  ctx.textBaseline = 'alphabetic'

  drawMaskBackground(ctx)

  // 顶部:小脸谱 + 平台名
  drawMaskFace(ctx, 116, 92, 22)
  ctx.fillStyle = MASK_COLORS.red
  ctx.font = `500 26px ${FONT_STACK}`
  ctx.fillText(options.brand, 156, 101)

  // 大字标题(超宽自动缩字号适配)
  let headSize = 138
  ctx.font = `900 ${headSize}px ${FONT_STACK}`
  const headWidth = ctx.measureText(options.headline).width
  if (headWidth > CONTENT_W) {
    headSize = Math.max(72, Math.floor((headSize * CONTENT_W) / headWidth))
    ctx.font = `900 ${headSize}px ${FONT_STACK}`
  }
  ctx.fillStyle = MASK_COLORS.ink
  ctx.fillText(options.headline, PAD_X, 250)

  // 英文小字 + 金色饰条
  ctx.fillStyle = MASK_COLORS.red
  ctx.font = `700 38px ${FONT_STACK}`
  drawLetterSpaced(ctx, options.headlineEn, PAD_X + 4, 316, 10)
  roundRectPath(ctx, PAD_X, 346, 92, 8, 4)
  ctx.fillStyle = MASK_COLORS.gold
  ctx.fill()

  // 元信息行
  ctx.fillStyle = MASK_COLORS.meta
  ctx.font = `400 30px ${FONT_STACK}`
  ctx.fillText(fitText(ctx, options.metaText, CONTENT_W), PAD_X, 410)

  const story = options.story
  const stamps = options.stamps
  if (story) {
    // 故事文案:title 大字 + body 分行排版
    ctx.fillStyle = MASK_COLORS.ink
    ctx.font = `800 52px ${FONT_STACK}`
    const titleLines = wrapText(ctx, story.title, CONTENT_W).slice(0, 2)
    let y = 486
    for (const line of titleLines) {
      ctx.fillText(line, PAD_X, y)
      y += 68
    }

    ctx.fillStyle = MASK_COLORS.body
    ctx.font = `400 29px ${FONT_STACK}`
    y += 6
    const bodyLines = wrapText(ctx, story.body, CONTENT_W).slice(0, 4)
    for (const line of bodyLines) {
      ctx.fillText(line, PAD_X, y)
      y += 48
    }

    // 分隔线
    ctx.strokeStyle = MASK_COLORS.line
    ctx.lineWidth = 2
    ctx.setLineDash([10, 8])
    ctx.beginPath()
    ctx.moveTo(PAD_X, 806)
    ctx.lineTo(POSTER_W - PAD_X, 806)
    ctx.stroke()
    ctx.setLineDash([])

    // 站点路线时间线(最多 6 个)
    ctx.fillStyle = MASK_COLORS.gold
    roundRectPath(ctx, PAD_X, 842, 14, 14, 3)
    ctx.fill()
    ctx.fillStyle = MASK_COLORS.red
    ctx.font = `700 30px ${FONT_STACK}`
    ctx.fillText(story.routeLabel, PAD_X + 26, 858)

    story.routeSpots.forEach((spot, index) => {
      const rowY = 926 + index * 44
      ctx.fillStyle = MASK_COLORS.red
      ctx.beginPath()
      ctx.arc(112, rowY - 9, 7, 0, Math.PI * 2)
      ctx.fill()
      if (index < story.routeSpots.length - 1) {
        ctx.strokeStyle = 'rgba(192, 57, 43, 0.35)'
        ctx.lineWidth = 2
        ctx.beginPath()
        ctx.moveTo(112, rowY - 2)
        ctx.lineTo(112, rowY + 28)
        ctx.stroke()
      }
      ctx.fillStyle = MASK_COLORS.red
      ctx.font = `700 22px ${FONT_STACK}`
      ctx.fillText(spot.label, 136, rowY)
      ctx.fillStyle = MASK_COLORS.ink
      ctx.font = `500 28px ${FONT_STACK}`
      ctx.fillText(fitText(ctx, spot.name, POSTER_W - PAD_X - 196), 196, rowY)
    })
  } else if (stamps) {
    // 集章网格:5 列,已集红章 + 名称,未集虚线空格 + 名称浅灰
    const COLS = 5
    const GAP = 20
    const GRID_TOP = 452
    const GRID_BOTTOM = 1150
    const cellW = (CONTENT_W - (COLS - 1) * GAP) / COLS
    const rows = Math.max(1, Math.ceil(stamps.items.length / COLS))
    const rowH = (GRID_BOTTOM - GRID_TOP) / rows

    stamps.items.forEach((item, index) => {
      const col = index % COLS
      const row = Math.floor(index / COLS)
      const x = PAD_X + col * (cellW + GAP)
      const y = GRID_TOP + row * rowH
      const cx = x + cellW / 2
      const sealR = Math.min(cellW * 0.27, rowH * 0.28)
      const cy = y + rowH * 0.38
      const nameY = cy + sealR + 32

      if (item.got) {
        drawMaskStampSeal(ctx, cx, cy, sealR)
        ctx.fillStyle = MASK_COLORS.ink
        ctx.font = `600 20px ${FONT_STACK}`
        ctx.textAlign = 'center'
        ctx.fillText(fitText(ctx, item.name, cellW - 8), cx, nameY)
        ctx.textAlign = 'left'
      } else {
        ctx.strokeStyle = MASK_COLORS.line
        ctx.lineWidth = 2
        ctx.setLineDash([8, 7])
        roundRectPath(ctx, x, y + 8, cellW, rowH - 16, 14)
        ctx.stroke()
        ctx.setLineDash([])
        ctx.fillStyle = 'rgba(122, 106, 85, 0.30)'
        ctx.font = `900 46px ${FONT_STACK}`
        ctx.textAlign = 'center'
        ctx.fillText('？', cx, cy + 16)
        ctx.fillStyle = 'rgba(122, 106, 85, 0.55)'
        ctx.font = `400 18px ${FONT_STACK}`
        ctx.fillText(fitText(ctx, item.name, cellW - 8), cx, nameY)
        ctx.textAlign = 'left'
      }
    })
  }

  // 底部:二维码 + 平台名 + 日期/扫码小字
  await drawFooter(ctx, options)

  // 离屏画布 → PNG
  return canvas.toDataURL('image/png')
}
