import {describe, expect, it} from 'vitest'

import {getHumanSize} from './getHumanSize'

describe('getHumanSize', () => {
	it('keeps sizes below 1024 in their unit', () => {
		expect(getHumanSize(0)).toBe('0 B')
		expect(getHumanSize(1023)).toBe('1023 B')
		expect(getHumanSize(1536)).toBe('1.5 KB')
	})

	it('moves exactly 1024 of a unit to the next unit', () => {
		expect(getHumanSize(1024)).toBe('1 KB')
		expect(getHumanSize(1024 ** 2)).toBe('1 MB')
		expect(getHumanSize(1024 ** 3)).toBe('1 GB')
		expect(getHumanSize(1024 ** 4)).toBe('1 TB')
	})

	it('stays in TB beyond the largest unit', () => {
		expect(getHumanSize(1024 ** 5)).toBe('1024 TB')
		expect(getHumanSize(2 * 1024 ** 5)).toBe('2048 TB')
	})
})
