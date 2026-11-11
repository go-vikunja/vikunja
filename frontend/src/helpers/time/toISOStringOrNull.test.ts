import {describe, expect, it} from 'vitest'

import {toISOStringOrNull} from './toISOStringOrNull'

describe('toISOStringOrNull', () => {
	it('formats a valid date', () => {
		expect(toISOStringOrNull(new Date('2021-02-06T12:00:00Z'))).toBe('2021-02-06T12:00:00.000Z')
	})

	it('formats a valid iso string', () => {
		expect(toISOStringOrNull('2021-02-06T12:00:00Z')).toBe('2021-02-06T12:00:00.000Z')
	})

	it('returns null for an invalid date instead of throwing', () => {
		expect(toISOStringOrNull(new Date('not a date'))).toBeNull()
	})

	it('returns null for null', () => {
		expect(toISOStringOrNull(null)).toBeNull()
	})

	it('returns null for undefined', () => {
		expect(toISOStringOrNull(undefined)).toBeNull()
	})

	it('returns null for the api zero time string', () => {
		expect(toISOStringOrNull('0001-01-01T00:00:00Z')).toBeNull()
	})
})
