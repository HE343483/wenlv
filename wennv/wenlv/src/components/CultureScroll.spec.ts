import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import CultureScroll from './CultureScroll.vue'

describe('CultureScroll', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    useLanguageStore().setLang('zh')
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
  afterEach(() => vi.unstubAllGlobals())

  it('reduced-motion: root has culture-scroll--static', () => {
    const wrapper = mount(CultureScroll)
    expect(wrapper.find('.culture-scroll').classes()).toContain('culture-scroll--static')
  })

  it('reduced-motion: no scrub rail height, world scroller', () => {
    const wrapper = mount(CultureScroll)
    const root = wrapper.find('.culture-scroll')

    expect(root.attributes('style')).toBeUndefined()
    expect(wrapper.find('.culture-scroll__world').exists()).toBe(true)
    expect(wrapper.find('.culture-scroll__walker').exists()).toBe(false)
  })

  it('reduced-motion: pointerdown does not enable scrub drag', async () => {
    const wrapper = mount(CultureScroll, { attachTo: document.body })
    wrapper.find('.culture-scroll__stage').element.dispatchEvent(
      new PointerEvent('pointerdown', {
        bubbles: true,
        clientX: 100,
        clientY: 100,
        button: 0,
        pointerId: 1,
      }),
    )
    await flushPromises()
    expect(wrapper.find('.culture-scroll').classes()).not.toContain('culture-scroll--dragging')
    wrapper.unmount()
  })

  it('has no foreground road or walker', () => {
    const wrapper = mount(CultureScroll)
    expect(wrapper.find('.culture-scroll__walker').exists()).toBe(false)
    expect(wrapper.find('[data-scroll-art="near"]').exists()).toBe(false)
  })

  it('renders far SVG with mid shanshui scenery + transparent lineart', () => {
    const wrapper = mount(CultureScroll)
    expect(wrapper.find('[data-scroll-art="far"]').exists()).toBe(true)
    expect(wrapper.find('[data-scroll-art="mid-scenery"]').exists()).toBe(true)
    expect(wrapper.find('[data-scroll-art="mid-lineart"]').exists()).toBe(true)
    expect(wrapper.findAll('.culture-scroll__scenery')).toHaveLength(4)
    expect(wrapper.findAll('.culture-scroll__lineart')).toHaveLength(4)
    expect(wrapper.find('.culture-scroll__scenery').attributes('src')).toContain(
      'era-scenery-shanshui-v1.jpg',
    )
    expect(wrapper.find('.culture-scroll__lineart').attributes('src')).toContain(
      'era-scroll-lineart-transparent-v1.png',
    )
    expect(wrapper.find('canvas').exists()).toBe(false)
  })

  it('places six mid-layer hotspots and opens era dialog', async () => {
    const wrapper = mount(CultureScroll, { attachTo: document.body })
    const buttons = wrapper.findAll('.culture-scroll__hotspot')
    expect(buttons).toHaveLength(6)
    const first = buttons[0]!
    const shuHan = buttons[2]!
    expect(buttons.map((btn) => btn.attributes('data-hotspot'))).toEqual([
      'ancient-shu',
      'qin',
      'shu-han',
      'tang-song',
      'ming-qing',
      'modern',
    ])
    expect(first.attributes('style')).toContain('%')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)

    await first.trigger('click')
    const dialog = wrapper.find('[role="dialog"]')
    expect(dialog.exists()).toBe(true)
    expect(dialog.text()).toContain('古蜀时期')
    expect(dialog.text()).toContain('约公元前1600年')
    expect(dialog.text()).toContain('三星堆')

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)

    await shuHan.trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    await wrapper.find('.culture-scroll__scrim').trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)

    useLanguageStore().setLang('en')
    await first.trigger('click')
    expect(wrapper.find('[role="dialog"]').text()).toContain('Ancient Shu')
    wrapper.unmount()
  })
})
