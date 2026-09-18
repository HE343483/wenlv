/** Line-art geometry for the Chengdu culture scroll. ViewBox 7200×900. */

export const SCROLL_ART = {
  width: 7200,
  height: 900,
  ground: 738,
} as const

export const ERA_WIDTH = 1200

export type Pt = { x: number; y: number }
export type Seg = { d: string; kind?: 'stroke' | 'fill' | 'soft' }
export type House = { x: number; w: number; wall: number; roof: number; door?: boolean }
export type Tower = { x: number; w: number; h: number; cols: number; rows: number; crown?: boolean }
export type Tree = { x: number; h: number; kind: 'pine' | 'willow' | 'cypress' | 'leaf' }

const G = SCROLL_ART.ground
const W = SCROLL_ART.width

function n(v: number) {
  return Math.round(v * 10) / 10
}

export function project(x: number, y: number, t: number, vp: Pt): Pt {
  return { x: n(x + (vp.x - x) * t), y: n(y + (vp.y - y) * t) }
}

export function line(a: Pt, b: Pt) {
  return `M${a.x} ${a.y}L${b.x} ${b.y}`
}

export function polyline(pts: Pt[], close = false) {
  const head = pts[0]
  if (!head) return ''
  return `M${head.x} ${head.y}${pts.slice(1).map((p) => `L${p.x} ${p.y}`).join('')}${close ? 'Z' : ''}`
}

export function ridgePath(baseY: number, amp: number, cycles: number, phase: number, jitter: number) {
  const pts: Pt[] = [{ x: 0, y: baseY + 40 }]
  for (let x = 0; x <= W; x += 30) {
    const t = x / W
    const y =
      baseY -
      Math.abs(Math.sin(t * Math.PI * cycles + phase)) * amp -
      Math.abs(Math.sin(t * Math.PI * cycles * 0.41 + phase * 1.7)) * amp * 0.42 -
      Math.sin(t * 18 + phase) * jitter
    pts.push({ x, y: n(y) })
  }
  pts.push({ x: W, y: baseY + 220 }, { x: 0, y: baseY + 220 })
  return polyline(pts, true)
}

export const farRidges = [
  { d: ridgePath(268, 86, 7.2, 0.2, 6), opacity: 0.35 },
  { d: ridgePath(312, 64, 9.4, 1.4, 5), opacity: 0.5 },
  { d: ridgePath(348, 48, 11.5, 2.1, 4), opacity: 0.72 },
]

export function birdMark(x: number, y: number, s = 1) {
  return `M${x} ${y}q${7 * s} ${-6 * s} ${14 * s} 0q${7 * s} ${-6 * s} ${14 * s} 0`
}

export const farBirds = [
  birdMark(640, 168, 1.1),
  birdMark(1880, 150, 0.9),
  birdMark(3120, 142, 1),
  birdMark(4510, 158, 0.85),
  birdMark(6020, 148, 1),
]

/** 古蜀 · 高台 */
export function terracePaths(): string[] {
  const base: Pt[] = [
    { x: 210, y: G },
    { x: 820, y: G },
    { x: 770, y: G - 38 },
    { x: 260, y: G - 38 },
  ]
  const mid: Pt[] = [
    { x: 260, y: G - 38 },
    { x: 770, y: G - 38 },
    { x: 720, y: G - 78 },
    { x: 310, y: G - 78 },
  ]
  const top: Pt[] = [
    { x: 310, y: G - 78 },
    { x: 720, y: G - 78 },
    { x: 670, y: G - 122 },
    { x: 360, y: G - 122 },
  ]
  const stair: string[] = []
  for (let i = 0; i < 7; i++) {
    const y = G - i * 17.4
    const inset = 8 + i * 6
    stair.push(`M${505 - inset} ${y}H${525 + inset}`)
  }
  return [polyline(base, true), polyline(mid, true), polyline(top, true), ...stair]
}

/** 三星堆神树简化折线 */
export function sacredTreePath() {
  const cx = 515
  const root = G - 122
  return [
    `M${cx} ${root}L${cx + 4} ${root - 318}`,
    `M${cx} ${root - 70}L${cx - 86} ${root - 118}L${cx - 142} ${root - 96}L${cx - 188} ${root - 148}`,
    `M${cx} ${root - 92}L${cx + 78} ${root - 128}L${cx + 138} ${root - 108}L${cx + 186} ${root - 156}`,
    `M${cx} ${root - 168}L${cx - 72} ${root - 214}L${cx - 128} ${root - 198}L${cx - 164} ${root - 248}`,
    `M${cx} ${root - 186}L${cx + 70} ${root - 228}L${cx + 124} ${root - 210}L${cx + 168} ${root - 262}`,
    `M${cx} ${root - 248}L${cx - 48} ${root - 292}L${cx - 78} ${root - 338}`,
    `M${cx} ${root - 262}L${cx + 52} ${root - 308}L${cx + 86} ${root - 352}`,
    `M${cx - 142} ${root - 96}L${cx - 132} ${root - 78}`,
    `M${cx + 138} ${root - 108}L${cx + 150} ${root - 88}`,
    `M${cx - 128} ${root - 198}L${cx - 120} ${root - 178}`,
    `M${cx + 124} ${root - 210}L${cx + 136} ${root - 190}`,
  ].join('')
}

export const sacredOrnaments = [
  { cx: 327, cy: G - 270, r: 5 },
  { cx: 351, cy: G - 370, r: 4.5 },
  { cx: 387, cy: G - 320, r: 4 },
  { cx: 437, cy: G - 460, r: 4.5 },
  { cx: 601, cy: G - 264, r: 5 },
  { cx: 639, cy: G - 348, r: 4 },
  { cx: 683, cy: G - 370, r: 4.5 },
  { cx: 601, cy: G - 474, r: 4 },
]

/** 金沙太阳神鸟简化 */
export function sunBirdPath(cx: number, cy: number, r: number) {
  const inner = r * 0.42
  let d = `M${cx + r} ${cy}A${r} ${r} 0 1 1 ${cx - r} ${cy}A${r} ${r} 0 1 1 ${cx + r} ${cy}`
  d += `M${cx + inner} ${cy}A${inner} ${inner} 0 1 1 ${cx - inner} ${cy}A${inner} ${inner} 0 1 1 ${cx + inner} ${cy}`
  for (let i = 0; i < 4; i++) {
    const a = (i * Math.PI) / 2 - 0.35
    const b = a + 1.05
    const x1 = n(cx + Math.cos(a) * r * 0.72)
    const y1 = n(cy + Math.sin(a) * r * 0.72)
    const x2 = n(cx + Math.cos((a + b) / 2) * r * 1.18)
    const y2 = n(cy + Math.sin((a + b) / 2) * r * 1.18)
    const x3 = n(cx + Math.cos(b) * r * 0.72)
    const y3 = n(cy + Math.sin(b) * r * 0.72)
    d += `M${x1} ${y1}Q${x2} ${y2} ${x3} ${y3}`
  }
  return d
}

export const ritualPoles = [
  { x: 148, h: 168 },
  { x: 186, h: 132 },
  { x: 888, h: 150 },
]

/** 都江堰水系 */
export function riverBank(xs: number, xe: number, yTop: number, yBot: number) {
  return `M${xs} ${yTop}C${xs + 180} ${yTop - 16} ${xs + 360} ${yTop + 22} ${xs + 520} ${yTop + 8}C${xs + 700} ${yTop - 10} ${xs + 880} ${yTop + 18} ${xe} ${yTop + 6}L${xe} ${yBot}C${xe - 160} ${yBot + 8} ${xs + 640} ${yBot - 6} ${xs + 420} ${yBot + 4}C${xs + 240} ${yBot - 10} ${xs + 80} ${yBot + 6} ${xs} ${yBot}Z`
}

export const qinWater = {
  outer: riverBank(1188, 2388, 598, 656),
  inner: riverBank(1688, 2390, 668, 724),
  fishMouth: polyline(
    [
      { x: 1612, y: 662 },
      { x: 1768, y: 628 },
      { x: 1844, y: 662 },
      { x: 1760, y: 698 },
      { x: 1668, y: 692 },
    ],
    true,
  ),
  weir: [
    'M1588 648 L1628 672 L1662 654 L1704 684 L1744 662 L1788 694 L1826 670 L1868 698',
    'M1710 706 L1748 724 L1790 710 L1834 730 L1876 716 L1922 734',
    'M1908 640 L1960 628 L2012 646 L2060 632 L2110 652',
  ],
  bottleMouth: [
    'M2140 560 L2168 640 L2174 722',
    'M2268 548 L2236 638 L2230 722',
    'M2174 676 L2230 676',
    'M2176 702 L2228 702',
  ],
  ripples: Array.from({ length: 11 }, (_, i) => {
    const x = 1240 + i * 96
    const y = 618 + (i % 3) * 28 + (i > 5 ? 36 : 0)
    const w = 54 + (i % 4) * 10
    return `M${x} ${y}q${w / 3} ${6} ${w / 2} 0q${w / 3} ${-6} ${w / 2} 0`
  }),
}

export function pavilion(x: number, ground: number, w = 64) {
  const wall = 36
  const roof = 22
  return [
    `M${x} ${ground}H${x + w}V${ground - wall}H${x}Z`,
    `M${x - 10} ${ground - wall}L${x + w / 2} ${ground - wall - roof}L${x + w + 10} ${ground - wall}`,
    `M${x - 14} ${ground - wall + 5}H${x + w + 14}`,
    `M${x + w / 2 - 8} ${ground}V${ground - wall * 0.7}H${x + w / 2 + 8}V${ground}`,
  ].join('')
}

/** 汉阙 */
export function queTower(x: number, side: 'left' | 'right', scale = 1) {
  const s = scale
  const g = G
  const dir = side === 'left' ? 1 : -1
  const bodyW = 54 * s
  const bodyH = 118 * s
  const roofH = 26 * s
  const subW = 34 * s
  const subH = 72 * s
  const mainX = x
  const subX = x - dir * (bodyW * 0.15 + subW)
  const eave = 14 * s
  const parts: string[] = []

  parts.push(`M${mainX} ${g}H${mainX + bodyW}V${g - 14 * s}H${mainX}Z`)
  parts.push(`M${mainX + 4 * s} ${g - 14 * s}H${mainX + bodyW - 4 * s}V${g - 14 * s - bodyH}H${mainX + 4 * s}Z`)
  parts.push(
    `M${mainX + 4 * s - eave} ${g - 14 * s - bodyH}L${mainX + bodyW / 2} ${g - 14 * s - bodyH - roofH}L${mainX + bodyW - 4 * s + eave} ${g - 14 * s - bodyH}`,
  )
  parts.push(`M${mainX + 4 * s - eave - 4} ${g - 14 * s - bodyH + 6}H${mainX + bodyW - 4 * s + eave + 4}`)
  parts.push(
    `M${mainX + 12 * s} ${g - 28 * s}H${mainX + bodyW - 12 * s}V${g - 52 * s}H${mainX + 12 * s}Z`,
    `M${mainX + 12 * s} ${g - 64 * s}H${mainX + bodyW - 12 * s}V${g - 88 * s}H${mainX + 12 * s}Z`,
    `M${mainX + 12 * s} ${g - 100 * s}H${mainX + bodyW - 12 * s}V${g - 120 * s}H${mainX + 12 * s}Z`,
  )

  parts.push(`M${subX} ${g}H${subX + subW}V${g - 10 * s}H${subX}Z`)
  parts.push(`M${subX + 3 * s} ${g - 10 * s}H${subX + subW - 3 * s}V${g - 10 * s - subH}H${subX + 3 * s}Z`)
  parts.push(
    `M${subX + 3 * s - 10 * s} ${g - 10 * s - subH}L${subX + subW / 2} ${g - 10 * s - subH - 18 * s}L${subX + subW - 3 * s + 10 * s} ${g - 10 * s - subH}`,
  )
  return parts.join('')
}

export function crenellatedWall(x0: number, x1: number, top: number, merlon = 22) {
  let d = `M${x0} ${G}V${top}`
  for (let x = x0; x < x1; x += merlon) {
    const up = (Math.floor((x - x0) / merlon) % 2 === 0)
    d += `H${n(Math.min(x + merlon * 0.55, x1))}V${up ? top - 16 : top}H${n(Math.min(x + merlon, x1))}V${top}`
  }
  d += `H${x1}V${G}`
  return d
}

export function cityGate(cx: number, wallTop: number) {
  const w = 92
  const archH = 78
  return [
    `M${cx - w / 2} ${G}V${wallTop - 8}H${cx + w / 2}V${G}`,
    `M${cx - 28} ${G}V${G - archH}Q${cx} ${G - archH - 28} ${cx + 28} ${G - archH}V${G}`,
    `M${cx - 48} ${wallTop - 8}L${cx} ${wallTop - 40}L${cx + 48} ${wallTop - 8}`,
  ].join('')
}

export function templeHall(x: number, w: number, wall: number, roof: number, platform = 12) {
  const g = G - platform
  return [
    `M${x - 8} ${G}H${x + w + 8}V${G - platform}H${x - 8}Z`,
    `M${x} ${g}H${x + w}V${g - wall}H${x}Z`,
    `M${x - 16} ${g - wall}L${x + w / 2} ${g - wall - roof}L${x + w + 16} ${g - wall}`,
    `M${x - 22} ${g - wall + 7}H${x + w + 22}`,
    `M${x + w * 0.42} ${g}V${g - wall * 0.72}H${x + w * 0.58}V${g}`,
    `M${x + 18} ${g - wall * 0.45}H${x + 36}V${g - wall * 0.62}H${x + 18}Z`,
    `M${x + w - 36} ${g - wall * 0.45}H${x + w - 18}V${g - wall * 0.62}H${x + w - 36}Z`,
  ].join('')
}

export function housePath(h: House) {
  const g = G
  const x = h.x
  const w = h.w
  const wall = h.wall
  const roof = h.roof
  const door = h.door !== false
  const parts = [
    `M${x} ${g}H${x + w}V${g - wall}H${x}Z`,
    `M${x - 9} ${g - wall}L${x + w / 2} ${g - wall - roof}L${x + w + 9} ${g - wall}`,
    `M${x - 13} ${g - wall + 5}H${x + w + 13}`,
  ]
  if (door) {
    parts.push(`M${x + w * 0.38} ${g}V${g - wall * 0.62}H${x + w * 0.62}V${g}`)
  }
  const winY = g - wall * 0.72
  parts.push(`M${x + 10} ${winY}H${x + 22}V${winY + 12}H${x + 10}Z`)
  if (w > 56) parts.push(`M${x + w - 22} ${winY}H${x + w - 10}V${winY + 12}H${x + w - 22}Z`)
  return parts.join('')
}

export const tangHouses: House[] = [
  { x: 3648, w: 78, wall: 50, roof: 26 },
  { x: 3714, w: 64, wall: 44, roof: 22 },
  { x: 3770, w: 88, wall: 58, roof: 30 },
  { x: 3848, w: 70, wall: 48, roof: 24 },
  { x: 3908, w: 96, wall: 62, roof: 32 },
  { x: 3994, w: 58, wall: 42, roof: 20 },
  { x: 4044, w: 80, wall: 54, roof: 28 },
  { x: 4116, w: 72, wall: 46, roof: 24 },
  { x: 4180, w: 90, wall: 60, roof: 30 },
  { x: 4448, w: 84, wall: 52, roof: 26 },
  { x: 4520, w: 68, wall: 46, roof: 22 },
  { x: 4578, w: 92, wall: 58, roof: 30 },
  { x: 4660, w: 74, wall: 48, roof: 24 },
]

export function pagodaPath(cx: number, stories = 7) {
  let w = 86
  let y = G
  const parts: string[] = [`M${cx - 48} ${G}H${cx + 48}V${G - 12}H${cx - 48}Z`]
  y = G - 12
  for (let i = 0; i < stories; i++) {
    const wall = 26 - i * 1.2
    const roof = 13 - i * 0.4
    const over = 15 - i * 0.9
    parts.push(`M${cx - w / 2} ${y}H${cx + w / 2}V${y - wall}H${cx - w / 2}Z`)
    parts.push(`M${cx - w / 2 - over} ${y - wall}L${cx} ${y - wall - roof}L${cx + w / 2 + over} ${y - wall}`)
    parts.push(`M${cx - w / 2 - over - 3} ${y - wall + 4}H${cx + w / 2 + over + 3}`)
    y = y - wall - roof * 0.42
    w *= 0.86
  }
  parts.push(`M${cx} ${y}L${cx} ${y - 28}M${cx - 8} ${y - 10}L${cx} ${y - 28}L${cx + 8} ${y - 10}`)
  return parts.join('')
}

export function stall(x: number, w: number) {
  return [
    `M${x} ${G}H${x + w}V${G - 28}H${x}Z`,
    `M${x - 8} ${G - 28}L${x + w / 2} ${G - 46}L${x + w + 8} ${G - 28}`,
    `M${x + 8} ${G - 8}H${x + w - 8}`,
  ].join('')
}

export const marketStalls = [3688, 3760, 3840, 3928, 4488, 4566, 4644].map((x, i) => stall(x, 42 + (i % 3) * 6))

/** 明清一点透视街巷 */
export const ALLEY_VP: Pt = { x: 5436, y: 488 }
const ALLEY_L = 5012
const ALLEY_R = 5864
const ALLEY_EAVE = 348
export const alleyTs = [0, 0.08, 0.16, 0.24, 0.32, 0.4, 0.47, 0.54, 0.6, 0.66, 0.71, 0.76]

export const alleyLines: string[] = (() => {
  const vp = ALLEY_VP
  const d: string[] = []
  d.push(line({ x: ALLEY_L, y: G }, project(ALLEY_L, G, 0.82, vp)))
  d.push(line({ x: ALLEY_R, y: G }, project(ALLEY_R, G, 0.82, vp)))
  d.push(line({ x: ALLEY_L, y: ALLEY_EAVE }, project(ALLEY_L, ALLEY_EAVE, 0.82, vp)))
  d.push(line({ x: ALLEY_R, y: ALLEY_EAVE }, project(ALLEY_R, ALLEY_EAVE, 0.82, vp)))
  d.push(line({ x: ALLEY_L - 18, y: ALLEY_EAVE - 36 }, project(ALLEY_L - 18, ALLEY_EAVE - 36, 0.78, vp)))
  d.push(line({ x: ALLEY_R + 18, y: ALLEY_EAVE - 36 }, project(ALLEY_R + 18, ALLEY_EAVE - 36, 0.78, vp)))

  for (const t of alleyTs) {
    const lb = project(ALLEY_L, G, t, vp)
    const lt = project(ALLEY_L, ALLEY_EAVE, t, vp)
    const rb = project(ALLEY_R, G, t, vp)
    const rt = project(ALLEY_R, ALLEY_EAVE, t, vp)
    d.push(line(lb, lt), line(rb, rt))
    const roofL = project(ALLEY_L - 18, ALLEY_EAVE - 36, t, vp)
    const roofR = project(ALLEY_R + 18, ALLEY_EAVE - 36, t, vp)
    d.push(line(lt, roofL), line(rt, roofR))
  }

  for (let i = 0; i < alleyTs.length - 1; i++) {
    const t0 = alleyTs[i] ?? 0
    const t1 = alleyTs[i + 1] ?? 0
    const l0 = project(ALLEY_L, G, t0, vp)
    const l1 = project(ALLEY_L, G, t1, vp)
    const r0 = project(ALLEY_R, G, t0, vp)
    const r1 = project(ALLEY_R, G, t1, vp)
    const doorH = 0.55
    const dl0 = project(ALLEY_L, G - (G - ALLEY_EAVE) * doorH, t0, vp)
    const dl1 = project(ALLEY_L, G - (G - ALLEY_EAVE) * doorH, t1, vp)
    const dr0 = project(ALLEY_R, G - (G - ALLEY_EAVE) * doorH, t0, vp)
    const dr1 = project(ALLEY_R, G - (G - ALLEY_EAVE) * doorH, t1, vp)
    const inset = 0.28
    const lDoor0 = {
      x: n(l0.x + (l1.x - l0.x) * inset),
      y: n(l0.y + (l1.y - l0.y) * inset),
    }
    const lDoor1 = {
      x: n(l0.x + (l1.x - l0.x) * (1 - inset)),
      y: n(l0.y + (l1.y - l0.y) * (1 - inset)),
    }
    const lDoorT0 = {
      x: n(dl0.x + (dl1.x - dl0.x) * inset),
      y: n(dl0.y + (dl1.y - dl0.y) * inset),
    }
    const lDoorT1 = {
      x: n(dl0.x + (dl1.x - dl0.x) * (1 - inset)),
      y: n(dl0.y + (dl1.y - dl0.y) * (1 - inset)),
    }
    d.push(polyline([lDoor0, lDoorT0, lDoorT1, lDoor1]))
    const rDoor0 = {
      x: n(r0.x + (r1.x - r0.x) * inset),
      y: n(r0.y + (r1.y - r0.y) * inset),
    }
    const rDoor1 = {
      x: n(r0.x + (r1.x - r0.x) * (1 - inset)),
      y: n(r0.y + (r1.y - r0.y) * (1 - inset)),
    }
    const rDoorT0 = {
      x: n(dr0.x + (dr1.x - dr0.x) * inset),
      y: n(dr0.y + (dr1.y - dr0.y) * inset),
    }
    const rDoorT1 = {
      x: n(dr0.x + (dr1.x - dr0.x) * (1 - inset)),
      y: n(dr0.y + (dr1.y - dr0.y) * (1 - inset)),
    }
    d.push(polyline([rDoor0, rDoorT0, rDoorT1, rDoor1]))
  }

  for (let k = 1; k <= 6; k++) {
    const gx = ALLEY_L + ((ALLEY_R - ALLEY_L) * k) / 7
    d.push(line({ x: gx, y: G }, project(gx, G, 0.8, vp)))
  }

  return d
})()

export const alleyLanterns = alleyTs.slice(0, 8).map((t, i) => {
  const side = i % 2 === 0 ? ALLEY_L + 22 : ALLEY_R - 22
  const p = project(side, ALLEY_EAVE + 28, t, ALLEY_VP)
  const s = Math.max(3.2, 9 * (1 - t * 0.85))
  return { cx: p.x, cy: p.y, rx: s * 0.55, ry: s }
})

export function stoneBridge(x: number, w: number) {
  const mid = x + w / 2
  return [
    `M${x} ${G}Q${mid} ${G - 78} ${x + w} ${G}`,
    `M${x + 18} ${G}Q${mid} ${G - 54} ${x + w - 18} ${G}`,
    `M${x} ${G - 8}H${x + 12}M${x + w - 12} ${G - 8}H${x + w}`,
  ].join('')
}

export function windowGrid(x: number, y: number, w: number, h: number, cols: number, rows: number) {
  const inset = Math.max(4, w * 0.12)
  const innerW = w - inset * 2
  const innerH = h - inset * 2
  const cw = innerW / cols
  const ch = innerH / rows
  const pad = Math.min(2.6, cw * 0.22)
  let d = ''
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      const x0 = n(x + inset + c * cw + pad)
      const y0 = n(y + inset + r * ch + pad)
      const x1 = n(x + inset + (c + 1) * cw - pad)
      const y1 = n(y + inset + (r + 1) * ch - pad)
      d += `M${x0} ${y0}H${x1}V${y1}H${x0}Z`
    }
  }
  return d
}

export function towerPath(t: Tower) {
  const x = t.x
  const y = G - t.h
  const body = `M${x} ${G}H${x + t.w}V${y}H${x}Z`
  const windows = windowGrid(x, y, t.w, t.h, t.cols, t.rows)
  const crown = t.crown
    ? `M${x - 6} ${y}L${x + t.w / 2} ${y - 28}L${x + t.w + 6} ${y}M${x + t.w / 2} ${y - 28}V${y - 48}`
    : `M${x - 3} ${y}H${x + t.w + 3}`
  return { body, windows, crown }
}

export const republicanBlocks: House[] = [
  { x: 6024, w: 70, wall: 96, roof: 18 },
  { x: 6098, w: 58, wall: 84, roof: 16 },
  { x: 6160, w: 80, wall: 110, roof: 20 },
]

export function brickFacade(h: House) {
  const y = G - h.wall
  const arches: string[] = []
  const count = Math.max(3, Math.floor(h.w / 18))
  for (let i = 0; i < count; i++) {
    const ax = h.x + 8 + i * ((h.w - 16) / count)
    const aw = 10
    arches.push(`M${ax} ${G - 12}V${G - 36}Q${ax + aw / 2} ${G - 48} ${ax + aw} ${G - 36}V${G - 12}`)
  }
  return [
    `M${h.x} ${G}H${h.x + h.w}V${y}H${h.x}Z`,
    `M${h.x - 4} ${y}H${h.x + h.w + 4}V${y - h.roof}H${h.x - 4}Z`,
    ...arches,
  ].join('')
}

export const modernTowers: Tower[] = [
  { x: 6248, w: 36, h: 168, cols: 3, rows: 8 },
  { x: 6292, w: 28, h: 214, cols: 2, rows: 11 },
  { x: 6328, w: 52, h: 286, cols: 4, rows: 14 },
  { x: 6390, w: 24, h: 196, cols: 2, rows: 10 },
  { x: 6424, w: 44, h: 340, cols: 4, rows: 16, crown: true },
  { x: 6478, w: 30, h: 232, cols: 2, rows: 12 },
  { x: 6516, w: 62, h: 392, cols: 5, rows: 18, crown: true },
  { x: 6588, w: 26, h: 176, cols: 2, rows: 9 },
  { x: 6622, w: 40, h: 268, cols: 3, rows: 13 },
  { x: 6672, w: 34, h: 318, cols: 3, rows: 15 },
  { x: 6716, w: 48, h: 246, cols: 4, rows: 12 },
  { x: 6774, w: 22, h: 188, cols: 2, rows: 10 },
  { x: 6804, w: 56, h: 360, cols: 4, rows: 17, crown: true },
  { x: 6870, w: 32, h: 210, cols: 3, rows: 11 },
  { x: 6912, w: 42, h: 292, cols: 3, rows: 14 },
  { x: 6964, w: 28, h: 164, cols: 2, rows: 8 },
  { x: 7000, w: 50, h: 248, cols: 4, rows: 12 },
  { x: 7060, w: 36, h: 198, cols: 3, rows: 10 },
]

export function pinePath(x: number, h: number) {
  const top = G - h
  let d = `M${x} ${G}L${x} ${top + 16}`
  const layers = 4
  for (let i = 0; i < layers; i++) {
    const y = top + 18 + i * (h * 0.16)
    const spread = 10 + i * 7
    d += `M${x - spread} ${y + 16}L${x} ${y - 6}L${x + spread} ${y + 16}`
  }
  return d
}

export function willowPath(x: number, h: number) {
  const top = G - h
  let d = `M${x} ${G}C${x - 4} ${G - h * 0.4} ${x + 6} ${G - h * 0.7} ${x} ${top}`
  for (let i = -3; i <= 3; i++) {
    const dx = i * 11
    d += `M${x} ${top + 8}C${x + dx * 0.4} ${top + h * 0.25} ${x + dx} ${top + h * 0.55} ${x + dx * 1.05} ${G - 18}`
  }
  return d
}

export function cypressPath(x: number, h: number) {
  const top = G - h
  return [
    `M${x} ${G}L${x} ${G - h * 0.2}`,
    `M${x - 16} ${G - h * 0.18}Q${x} ${top} ${x + 16} ${G - h * 0.18}Z`,
    `M${x - 11} ${G - h * 0.42}Q${x} ${top + 12} ${x + 11} ${G - h * 0.42}`,
  ].join('')
}

export function leafTreePath(x: number, h: number) {
  const top = G - h
  return [
    `M${x} ${G}V${top + h * 0.38}`,
    `M${x} ${top + h * 0.42}m${-20} ${12}a22 18 0 1 1 40 0a18 16 0 1 1 -36 8a16 14 0 1 1 28 -14`,
  ].join('')
}

export function treePath(t: Tree) {
  if (t.kind === 'pine') return pinePath(t.x, t.h)
  if (t.kind === 'willow') return willowPath(t.x, t.h)
  if (t.kind === 'cypress') return cypressPath(t.x, t.h)
  return leafTreePath(t.x, t.h)
}

export const midTrees: Tree[] = [
  { x: 64, h: 96, kind: 'pine' },
  { x: 108, h: 72, kind: 'pine' },
  { x: 980, h: 88, kind: 'pine' },
  { x: 1044, h: 110, kind: 'pine' },
  { x: 1128, h: 76, kind: 'willow' },
  { x: 1210, h: 92, kind: 'willow' },
  { x: 2320, h: 84, kind: 'willow' },
  { x: 2396, h: 102, kind: 'cypress' },
  { x: 2468, h: 118, kind: 'cypress' },
  { x: 3376, h: 108, kind: 'cypress' },
  { x: 3448, h: 90, kind: 'cypress' },
  { x: 3524, h: 76, kind: 'leaf' },
  { x: 3588, h: 94, kind: 'willow' },
  { x: 4388, h: 80, kind: 'leaf' },
  { x: 4728, h: 98, kind: 'willow' },
  { x: 4804, h: 86, kind: 'willow' },
  { x: 4920, h: 74, kind: 'leaf' },
  { x: 5948, h: 90, kind: 'leaf' },
  { x: 6010, h: 72, kind: 'leaf' },
  { x: 7148, h: 78, kind: 'leaf' },
]

export const nearTrees: Tree[] = [
  { x: 220, h: 64, kind: 'pine' },
  { x: 430, h: 52, kind: 'pine' },
  { x: 1266, h: 58, kind: 'willow' },
  { x: 1988, h: 48, kind: 'willow' },
  { x: 2512, h: 62, kind: 'cypress' },
  { x: 3488, h: 54, kind: 'leaf' },
  { x: 4710, h: 60, kind: 'willow' },
  { x: 5890, h: 50, kind: 'leaf' },
  { x: 6388, h: 46, kind: 'leaf' },
  { x: 7024, h: 52, kind: 'leaf' },
]

export function grassTufts(start: number, end: number, step: number) {
  const d: string[] = []
  for (let x = start; x < end; x += step) {
    const h = 8 + ((x * 13) % 7)
    d.push(`M${x} ${G}l${-3} ${-h}M${x} ${G}l${3} ${-h * 0.85}M${x} ${G}v${-h * 0.6}`)
  }
  return d
}

export const midGrass = [
  ...grassTufts(40, 200, 26),
  ...grassTufts(860, 1120, 28),
  ...grassTufts(2260, 2480, 30),
  ...grassTufts(3480, 3660, 28),
  ...grassTufts(4680, 4980, 26),
  ...grassTufts(5880, 6080, 30),
]

export const nearRoad = {
  fill: `M0 ${G + 8}H${W}V${SCROLL_ART.height}H0Z`,
  edges: [
    `M0 ${G + 10}H${W}`,
    `M0 ${G + 28}H${W}`,
    `M0 ${SCROLL_ART.height - 36}H${W}`,
  ],
  dashes: Array.from({ length: 72 }, (_, i) => {
    const x = i * 100
    return `M${x} ${G + 88}H${x + 42}`
  }),
}

export function railing(x: number, w: number) {
  const posts: string[] = [`M${x} ${G + 10}H${x + w}M${x} ${G - 22}H${x + w}`]
  for (let px = x; px <= x + w; px += 22) {
    posts.push(`M${px} ${G + 10}V${G - 22}`)
  }
  return posts.join('')
}

export const nearRailings = [railing(1628, 220), railing(4748, 210), railing(2148, 140)]

export const lamps = [6280, 6460, 6640, 6820, 7000].map((x) => {
  return `M${x} ${G + 12}V${G - 64}H${x + 16}M${x + 16} ${G - 64}l${8} ${10}M${x + 16} ${G - 64}l${8} ${-6}`
})

export const eraMarks = [
  { x: 430, y: 128, id: 'ancient-shu' },
  { x: 1680, y: 128, id: 'qin' },
  { x: 2860, y: 128, id: 'shu-han' },
  { x: 4060, y: 128, id: 'tang-song' },
  { x: 5280, y: 128, id: 'ming-qing' },
  { x: 6480, y: 128, id: 'modern' },
]
