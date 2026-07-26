import {describe, it, expect} from 'vitest'

import {calculateItemPosition, calculateItemPositions} from './calculateItemPosition'

describe('calculateItemPosition', () => {
	it('should calculate the task position', () => {
		const result = calculateItemPosition(10, 100)
		expect(result).toBe(55)
	})

	it('should return 0 if no position was provided', () => {
		const result = calculateItemPosition(null, null)
		expect(result).toBe(0)
	})

	it('should calculate the task position for the first task', () => {
		const result = calculateItemPosition(null, 100)
		expect(result).toBe(50)
	})

	it('should calculate the task position for the last task', () => {
		const result = calculateItemPosition(10, null)
		expect(result).toBe(65546)
	})

	it('should handle equal positions (conflict) by nudging above', () => {
		const result = calculateItemPosition(100, 100)
		expect(result).toBeGreaterThan(100)
		expect(result).toBeLessThan(101)
	})

	it('should handle equal positions at zero', () => {
		const result = calculateItemPosition(0, 0)
		expect(result).toBeGreaterThan(0)
	})

	it('should preserve precision after JSON round-trip', () => {
		const position = calculateItemPosition(100, 100)
		const serialized = JSON.stringify(position)
		const deserialized = JSON.parse(serialized)
		expect(deserialized).toBe(position)
		expect(deserialized).toBeGreaterThan(100)
	})
})

describe('calculateItemPositions', () => {
	it('should return nothing for an empty batch', () => {
		expect(calculateItemPositions(0, null, 100)).toEqual([])
	})

	it('should match calculateItemPosition for a single item', () => {
		expect(calculateItemPositions(1, 10, 100)).toEqual([calculateItemPosition(10, 100)])
		expect(calculateItemPositions(1, null, 100)).toEqual([calculateItemPosition(null, 100)])
		expect(calculateItemPositions(1, 10, null)).toEqual([calculateItemPosition(10, null)])
	})

	it('should spread the batch over the gap between both neighbors', () => {
		expect(calculateItemPositions(3, 0, 100)).toEqual([25, 50, 75])
	})

	it('should stay below the following item when there is nothing before', () => {
		const positions = calculateItemPositions(6, null, 100)
		expect(positions).toEqual([...positions].sort((a, b) => a - b))
		expect(positions[0]).toBeGreaterThan(0)
		expect(positions.at(-1)).toBeLessThan(100)
	})

	it('should append when there is no following item', () => {
		expect(calculateItemPositions(2, 10, null)).toEqual([65546, 131082])
		expect(calculateItemPositions(2, null, null)).toEqual([65536, 131072])
	})

	it('should stay ordered when the neighbors conflict', () => {
		const positions = calculateItemPositions(3, 100, 100)
		expect(positions).toEqual([...positions].sort((a, b) => a - b))
		expect(positions[0]).toBeGreaterThan(100)
	})

	it('should stay ordered when the gap is too small to subdivide', () => {
		const positions = calculateItemPositions(5, 100, 100.001)
		expect(new Set(positions).size).toBe(5)
		expect(positions).toEqual([...positions].sort((a, b) => a - b))
	})

	it('should keep distinct positions after a JSON round-trip', () => {
		const positions = calculateItemPositions(10, null, 1)
		const deserialized = JSON.parse(JSON.stringify(positions))
		expect(deserialized).toEqual(positions)
		expect(new Set(deserialized).size).toBe(positions.length)
	})
})
