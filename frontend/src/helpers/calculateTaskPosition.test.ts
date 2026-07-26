import {describe, it, expect} from 'vitest'

import {calculateItemPosition} from './calculateItemPosition'

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
