import {describe, it, expect} from 'vitest'
import {packColumns} from './packColumns'

interface Interval {
	id: string
	start: number
	end: number
}

function pack(items: Interval[]) {
	return packColumns(items, i => i.start, i => i.end)
		.reduce((acc, p) => {
			acc[p.item.id] = {col: p.col, cols: p.cols}
			return acc
		}, {} as Record<string, {col: number, cols: number}>)
}

describe('packColumns', () => {
	it('gives a single non-overlapping item one full column', () => {
		const out = pack([{id: 'a', start: 0, end: 60}])
		expect(out.a).toEqual({col: 0, cols: 1})
	})

	it('treats touching intervals as non-overlapping', () => {
		const out = pack([
			{id: 'a', start: 0, end: 60},
			{id: 'b', start: 60, end: 120},
		])
		expect(out.a).toEqual({col: 0, cols: 1})
		expect(out.b).toEqual({col: 0, cols: 1})
	})

	it('splits two overlapping intervals into two columns', () => {
		const out = pack([
			{id: 'a', start: 0, end: 60},
			{id: 'b', start: 30, end: 90},
		])
		expect(out.a).toEqual({col: 0, cols: 2})
		expect(out.b).toEqual({col: 1, cols: 2})
	})

	it('reuses a freed column within the same cluster', () => {
		// a+b overlap (2 cols); c starts after a ends but still overlaps b,
		// so the whole run is one cluster of width 2 and c reuses column 0.
		const out = pack([
			{id: 'a', start: 0, end: 30},
			{id: 'b', start: 10, end: 90},
			{id: 'c', start: 40, end: 80},
		])
		expect(out.b.cols).toBe(2)
		expect(out.a.col).toBe(0)
		expect(out.b.col).toBe(1)
		expect(out.c.col).toBe(0)
	})

	it('keeps separate clusters independent', () => {
		const out = pack([
			{id: 'a', start: 0, end: 60},
			{id: 'b', start: 10, end: 70},
			{id: 'c', start: 200, end: 260},
		])
		expect(out.a.cols).toBe(2)
		expect(out.b.cols).toBe(2)
		expect(out.c).toEqual({col: 0, cols: 1})
	})
})
