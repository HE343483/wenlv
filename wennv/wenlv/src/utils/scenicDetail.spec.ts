import { describe, it, expect } from 'vitest'
import { splitList, parseSections, estimatedSet, displayFact } from './scenicDetail'

describe('scenicDetail 展示工具', () => {
  it('splitList 拆分逗号串并去空', () => {
    expect(splitList('a.jpg, b.jpg ,,')).toEqual(['a.jpg', 'b.jpg'])
    expect(splitList(undefined)).toEqual([])
    expect(splitList('')).toEqual([])
  })

  it('parseSections 对非法 JSON 返回空数组', () => {
    expect(parseSections('not-json')).toEqual([])
    expect(parseSections('{"a":1}')).toEqual([])
    expect(parseSections(undefined)).toEqual([])
  })

  it('parseSections 过滤缺标题或正文的段落', () => {
    const raw = JSON.stringify([
      { title: '街区沿革', text: '甲', image: 'g1.jpg' },
      { title: '缺正文' },
      { title: '缺配图', text: '丙' },
      null,
    ])
    expect(parseSections(raw)).toEqual([
      { title: '街区沿革', text: '甲', image: 'g1.jpg' },
      { title: '缺配图', text: '丙', image: '' },
    ])
  })

  it('estimatedSet 解析参考值字段', () => {
    expect([...estimatedSet('ticket_price,recommend_hours')]).toEqual(['ticket_price', 'recommend_hours'])
    expect([...estimatedSet('')]).toEqual([])
  })

  it('displayFact 空值回退占位符', () => {
    expect(displayFact('   ', '暂无数据')).toBe('暂无数据')
    expect(displayFact(undefined, '暂无数据')).toBe('暂无数据')
    expect(displayFact('09:00-17:00', '暂无数据')).toBe('09:00-17:00')
  })
})
