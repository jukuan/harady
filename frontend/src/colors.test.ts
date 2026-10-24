import { describe, it, expect } from 'vitest'
import { colorFor } from './colors'

describe('colorFor', () => {
  it('returns the same colour for the same id', () => {
    const a = colorFor('player-abc')
    const b = colorFor('player-abc')
    expect(a).toBe(b)
  })

  it('returns different colours for different ids in most cases', () => {
    const ids = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']
    const colours = ids.map(colorFor)
    const distinct = new Set(colours)
    expect(distinct.size).toBeGreaterThan(2)
  })

  it('returns a palette entry with all expected fields', () => {
    const c = colorFor('anything')
    expect(c).toHaveProperty('bg')
    expect(c).toHaveProperty('border')
    expect(c).toHaveProperty('text')
    expect(c).toHaveProperty('solid')
  })
})
