import {it, expect} from 'vitest'

import {calculateItemPosition} from './calculateItemPosition'

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
