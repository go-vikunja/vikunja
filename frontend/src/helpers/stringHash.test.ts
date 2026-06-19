import {describe, it, expect} from 'vitest'
import {stringHash} from './stringHash'

describe('stringHash', () => {
	it('returns a non-negative integer', () => {
		expect(stringHash('hello')).toBeGreaterThanOrEqual(0)
		expect(Number.isInteger(stringHash('hello'))).toBe(true)
	})

	it('is deterministic for the same input', () => {
		expect(stringHash('foo')).toBe(stringHash('foo'))
	})

	it('returns different values for different inputs', () => {
		expect(stringHash('foo')).not.toBe(stringHash('bar'))
	})

	it('handles the empty string', () => {
		expect(stringHash('')).toBeGreaterThanOrEqual(0)
	})
})
