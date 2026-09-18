import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import { CULTURE_SCROLL_SEGMENTS } from '@/data/cultureScroll'
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

  it('renders far SVG with transparent lineart only (no mid scenery)', () => {
    const wrapper = mount(CultureScroll)
    expect(wrapper.find('[data-scroll-art="far"]').exists()).toBe(true)
    expect(wrapper.find('[data-scroll-art="mid-scenery"]').exists()).toBe(false)
    expect(wrapper.findAll('.culture-scroll__scenery')).toHaveLength(0)
    expect(wrapper.find('[data-scroll-art="mid-lineart"]').exists()).toBe(true)
    expect(wrapper.findAll('.culture-scroll__lineart')).toHaveLength(4)
    expect(wrapper.find('.culture-scroll__lineart').attributes('src')).toContain(
      'era-scroll-lineart-transparent-v1.png',
    )
    expect(wrapper.find('canvas').exists()).toBe(false)
  })

  it('places era media cards opposite text cards with confirmed srcs', () => {
    const wrapper = mount(CultureScroll)
    const markers = wrapper.findAll('.culture-scroll__marker')
    expect(markers).toHaveLength(6)

    const expectedSrcs = [
      '/images/home/1.jpg',
      '/images/home/carousel-xiling.jpg',
      '/images/home/hero-chengdu.jpg',
      '/images/home/154b682806f848688077dbabd5bcac3f_720.jpg',
      '/images/culture-scroll/era-scenery-shanshui-v1.jpg',
      '/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg',
    ]

    const medias = wrapper.findAll('.culture-scroll__era-media img')
    expect(medias).toHaveLength(6)
    expect(medias.map((img) => img.attributes('src'))).toEqual(expectedSrcs)

    // even index: text above → media below (no --media-above class)
    expect(markers[0]!.find('.culture-scroll__era-card').exists()).toBe(true)
    expect(markers[0]!.classes()).toContain('culture-scroll__marker--above')
    expect(markers[0]!.find('.culture-scroll__era-media').exists()).toBe(true)

    // odd: text below → media above
    expect(markers[1]!.classes()).toContain('culture-scroll__marker--below')
    expect(markers[1]!.find('.culture-scroll__era-media').exists()).toBe(true)
  })

  it('renders decorative photos along both sides of the timeline', () => {
    const wrapper = mount(CultureScroll)
    const decors = wrapper.findAll('.culture-scroll__decor')
    expect(decors.length).toBeGreaterThanOrEqual(6)
    expect(wrapper.findAll('.culture-scroll__decor--above').length).toBeGreaterThan(0)
    expect(wrapper.findAll('.culture-scroll__decor--below').length).toBeGreaterThan(0)
    expect(decors[0]!.find('img').attributes('src')).toBeTruthy()
  })

  it('segments expose imageUrl for every era', () => {
    expect(CULTURE_SCROLL_SEGMENTS.map((s) => s.imageUrl)).toEqual([
      '/images/home/1.jpg',
      '/images/home/carousel-xiling.jpg',
      '/images/home/hero-chengdu.jpg',
      '/images/home/154b682806f848688077dbabd5bcac3f_720.jpg',
      '/images/culture-scroll/era-scenery-shanshui-v1.jpg',
      '/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg',
    ])
  })

  it('renders mid-layer timeline with six always-visible era cards', () => {
    const wrapper = mount(CultureScroll)
    expect(wrapper.find('.culture-scroll__timeline').exists()).toBe(true)
    expect(wrapper.findAll('.culture-scroll__hotspot')).toHaveLength(0)
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)

    const markers = wrapper.findAll('.culture-scroll__marker')
    expect(markers).toHaveLength(6)
    expect(markers.map((m) => m.attributes('data-marker'))).toEqual([
      'ancient-shu',
      'qin',
      'shu-han',
      'tang-song',
      'ming-qing',
      'modern',
    ])

    const first = markers[0]!
    expect(first.attributes('style')).toContain('%')
    expect(first.classes()).toContain('culture-scroll__marker--above')
    expect(markers[1]!.classes()).toContain('culture-scroll__marker--below')

    const card = first.find('.culture-scroll__era-card')
    expect(card.exists()).toBe(true)
    expect(card.text()).toContain('古蜀时期')
    expect(card.text()).toContain('约公元前1600年')
    expect(card.text()).toContain('三星堆')

    const axisTime = first.find('.culture-scroll__axis-time')
    expect(axisTime.exists()).toBe(true)
    expect(axisTime.text()).toContain('约公元前1600年')
  })

  it('timeline cards follow language store without dialog', async () => {
    const wrapper = mount(CultureScroll)
    useLanguageStore().setLang('en')
    await flushPromises()
    const first = wrapper.find('[data-marker="ancient-shu"]')
    expect(first.text()).toContain('Ancient Shu')
    expect(first.text()).toContain('1600 BC')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })
})
