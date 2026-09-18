// src/utils/waterRipple.ts
export const DEFAULT_RADIUS = 2
export const DEFAULT_STRENGTH = 256

export type RippleState = {
  width: number
  height: number
  rippleMap: Int16Array
  lastMap: Int16Array
  oldIndex: number
  newIndex: number
}

export function createRippleState(width: number, height: number): RippleState {
  const size = width * (height + 2) * 2
  return {
    width,
    height,
    rippleMap: new Int16Array(size),
    lastMap: new Int16Array(width * height),
    oldIndex: width,
    newIndex: width * (height + 3),
  }
}

export function disturb(
  state: RippleState,
  x: number,
  y: number,
  radius = DEFAULT_RADIUS,
  strength = DEFAULT_STRENGTH,
): void {
  const { width, height, rippleMap, oldIndex, newIndex } = state
  const xi = x | 0
  const yi = y | 0
  const half = strength >> 1
  const singleBufferOnly = half === 0 && strength > 0
  for (let j = yi - radius; j < yi + radius; j++) {
    for (let k = xi - radius; k < xi + radius; k++) {
      if (j >= 0 && j < height && k >= 0 && k < width) {
        const offset = j * width + k
        if (singleBufferOnly) {
          rippleMap[oldIndex + offset]! += strength
        } else {
          rippleMap[oldIndex + offset]! += half
          rippleMap[newIndex + offset]! += half
        }
      }
    }
  }
}

/** 经典双缓冲振幅 + 按振幅差折射采样；返回是否仍有非零振幅差 */
export function stepRipple(
  state: RippleState,
  texture: Uint8ClampedArray,
  output: Uint8ClampedArray,
): boolean {
  const { width, height, rippleMap, lastMap } = state
  let i = state.oldIndex
  state.oldIndex = state.newIndex
  state.newIndex = i

  i = 0
  let mapIndex = state.oldIndex
  let active = false
  const halfW = width >> 1
  const halfH = height >> 1

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const data =
        ((rippleMap[mapIndex - width]! +
          rippleMap[mapIndex + width]! +
          rippleMap[mapIndex - 1]! +
          rippleMap[mapIndex + 1]!) >>
          1) - rippleMap[state.newIndex + i]!

      rippleMap[state.newIndex + i] = data - (data >> 5)

      let amplitude = 1024 - data
      const oldAmp = lastMap[i]
      lastMap[i] = amplitude
      mapIndex++

      if (oldAmp !== amplitude) {
        active = true
        let dx = (((x - halfW) * amplitude) / 1024 + halfW) | 0
        let dy = (((y - halfH) * amplitude) / 1024 + halfH) | 0
        if (dx >= width) dx = width - 1
        if (dx < 0) dx = 0
        if (dy >= height) dy = height - 1
        if (dy < 0) dy = 0
        const src = (dx + dy * width) * 4
        const dst = i * 4
        output[dst]! = texture[src]!
        output[dst + 1]! = texture[src + 1]!
        output[dst + 2]! = texture[src + 2]!
        output[dst + 3]! = texture[src + 3]!
      } else {
        const dst = i * 4
        output[dst]! = texture[dst]!
        output[dst + 1]! = texture[dst + 1]!
        output[dst + 2]! = texture[dst + 2]!
        output[dst + 3]! = texture[dst + 3]!
      }
      i++
    }
  }
  return active
}
