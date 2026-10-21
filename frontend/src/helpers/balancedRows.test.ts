import {describe, it, expect} from 'vitest'
import {balancedRows} from './balancedRows'

const sizes = (n: number, max: number) => balancedRows(Array.from({length: n}, (_, i) => i), max).map(r => r.length)

describe('balancedRows', () => {
	it.each([
		[8, 8, [8]],
		[8, 5, [4, 4]],
		[7, 5, [4, 3]],
		[7, 3, [3, 2, 2]],
		[8, 3, [3, 3, 2]],
		[5, 1, [1, 1, 1, 1, 1]],
		[3, 0, [1, 1, 1]],
		[0, 4, []],
	])('splits %i items with max %i per row into %j', (n, max, expected) => {
		expect(sizes(n, max)).toEqual(expected)
	})

	it('keeps item order', () => {
		expect(balancedRows(['a', 'b', 'c', 'd', 'e'], 3)).toEqual([['a', 'b', 'c'], ['d', 'e']])
	})
})
