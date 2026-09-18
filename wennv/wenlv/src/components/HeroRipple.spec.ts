// src/components/HeroRipple.spec.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import HeroRipple from './HeroRipple.vue'

describe('HeroRipple', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn().mockImplementation((query: string) => ({
        matches: query.includes('prefers-reduced-motion'),
        media: query,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn(),
        onchange: null,
      })),
    )
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('reduced-motion: root is not data-active', () => {
    const wrapper = mount(HeroRipple)
    expect(wrapper.find('.hero-ripple').attributes('data-active')).toBeUndefined()
  })
})
