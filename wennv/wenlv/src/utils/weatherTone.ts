/**
 * WeatherTone — 温度色调计算器（纯函数，无 UI）
 *
 * 依据平均温度 avgTemp = ({{min_temp}} + {{max_temp}}) / 2
 * 进行分段线性插值，输出主色调 hex。
 * 采用柔和的矿物颜料色系，与蜀锦金统一（温暖天气→赭金）。
 *
 *   平均温度区间          主色调        十六进制
 *   {{avg_temp}} ≤ 0      黛蓝          #3C5A72
 *   0 < avg ≤ 10          青蓝          #5A7E96
 *   10 < avg ≤ 25         竹青          #6F8A68
 *   25 < avg ≤ 33         赭金          #B08952
 *   avg > 33              赭红          #A85A42
 */

type RGB = [number, number, number]

const DEEP_BLUE: RGB = [0x3c, 0x5a, 0x72]
const BLUE: RGB = [0x5a, 0x7e, 0x96]
const GREEN: RGB = [0x6f, 0x8a, 0x68]
const ORANGE: RGB = [0xb0, 0x89, 0x52]
const RED: RGB = [0xa8, 0x5a, 0x42]

function blend(a: RGB, b: RGB, t: number): RGB {
  const k = Math.min(1, Math.max(0, t))
  return [
    a[0] + (b[0] - a[0]) * k,
    a[1] + (b[1] - a[1]) * k,
    a[2] + (b[2] - a[2]) * k,
  ]
}

function rgbToHex(rgb: RGB): string {
  return '#' + rgb.map(c => Math.round(c).toString(16).padStart(2, '0')).join('')
}

/**
 * 计算主色调 (分段线性 LERp)
 * 输出示例：avgTemp=5 → 蓝绿过渡色；avgTemp=28 → 橙红过渡色
 */
export function computeTone(avgTemp: number): string {
  let rgb: RGB
  if (avgTemp <= 0) {
    rgb = blend(DEEP_BLUE, BLUE, (avgTemp + 10) / 10)
  } else if (avgTemp <= 10) {
    rgb = blend(BLUE, GREEN, avgTemp / 10)
  } else if (avgTemp <= 25) {
    rgb = blend(GREEN, ORANGE, (avgTemp - 10) / 15)
  } else if (avgTemp <= 33) {
    rgb = blend(ORANGE, RED, (avgTemp - 25) / 8)
  } else {
    rgb = RED
  }
  return rgbToHex(rgb)
}

/**
 * 将色调写入 :root 的 CSS 变量
 *   --temp-tone      主色调 hex
 *   --temp-tone-soft 主色调 12% alpha（面板底色/按钮微光）
 *   --temp-tone-text 由 main.css 经 color-mix 派生（自适应深浅主题）
 * avgTemp 为 null 时移除变量，回退到默认金色
 */
export function applyToneToRoot(avgTemp: number | null): void {
  const root = document.documentElement
  if (avgTemp === null) {
    root.style.removeProperty('--temp-tone')
    root.style.removeProperty('--temp-tone-soft')
    return
  }
  const main = computeTone(avgTemp)
  const r = parseInt(main.slice(1, 3), 16)
  const g = parseInt(main.slice(3, 5), 16)
  const b = parseInt(main.slice(5, 7), 16)
  root.style.setProperty('--temp-tone', main)
  root.style.setProperty('--temp-tone-soft', `rgba(${r}, ${g}, ${b}, 0.12)`)
}
