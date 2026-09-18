// src/utils/waterRipple.spec.ts
import { describe, it, expect } from 'vitest'
import { createRippleState, disturb, stepRipple } from './waterRipple'

function solidTexture(w: number, h: number, rgba: [number, number, number, number]) {
  const data = new Uint8ClampedArray(w * h * 4)
  for (let i = 0; i < w * h; i++) {
    data[i * 4] = rgba[0]
    data[i * 4 + 1] = rgba[1]
    data[i * 4 + 2] = rgba[2]
    data[i * 4 + 3] = rgba[3]
  }
  return data
}

describe('waterRipple', () => {
  it('disturb + step changes center-ish pixels vs flat texture', () => {
    const w = 32
    const h = 32
    const state = createRippleState(w, h)
    const texture = solidTexture(w, h, [10, 20, 30, 255])
    // 给纹理一个水平色带，便于检测偏移
    for (let x = 0; x < w; x++) {
      const i = ((h / 2 | 0) * w + x) * 4
      texture[i] = 200
    }
    const out = new Uint8ClampedArray(texture)
    disturb(state, 16, 16, 3, 512)
    stepRipple(state, texture, out)
    stepRipple(state, texture, out)
    let changed = 0
    for (let i = 0; i < out.length; i++) if (out[i] !== texture[i]) changed++
    expect(changed).toBeGreaterThan(0)
  })

  it('without disturb, step keeps output equal to texture', () => {
    const w = 16
    const h = 16
    const state = createRippleState(w, h)
    const texture = solidTexture(w, h, [1, 2, 3, 255])
    const out = new Uint8ClampedArray(w * h * 4)
    stepRipple(state, texture, out)
    expect(Array.from(out)).toEqual(Array.from(texture))
  })
})
