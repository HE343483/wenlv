import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import Carousel from './Carousel.vue'

describe('Carousel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    useLanguageStore().setLang('zh')
  })

  it('renders img when imageUrl is provided', () => {
    const wrapper = mount(Carousel, {
      props: {
        items: [
          {
            id: 'panda',
            imageUrl: '/images/home/carousel-panda.jpg',
            titleZh: '大熊猫繁育研究基地',
            titleEn: 'Giant Panda Base',
          },
        ],
      },
      global: {
        stubs: {},
      },
    })
    const img = wrapper.find('img.carousel__image')
    expect(img.exists()).toBe(true)
    expect(img.attributes('src')).toBe('/images/home/carousel-panda.jpg')
    expect(img.attributes('alt')).toContain('熊猫')
  })

  it('does not render img when imageUrl is empty', () => {
    const wrapper = mount(Carousel, {
      props: {
        items: [
          {
            id: 'empty',
            imageUrl: '',
            titleZh: '占位',
            titleEn: 'Placeholder',
          },
        ],
      },
    })
    expect(wrapper.find('img.carousel__image').exists()).toBe(false)
  })
})
