import {describe, expect, it} from 'vitest'

import {parseDateOrNull} from './parseDateOrNull'

describe('parseDateOrNull', () => {
	it('parses a valid iso string', () => {
		expect(parseDateOrNull('2021-02-06T12:00:00Z')?.toISOString()).toBe('2021-02-06T12:00:00.000Z')
	})

	it('passes through a valid date', () => {
		const date = new Date('2021-02-06T12:00:00Z')
		expect(parseDateOrNull(date)).toBe(date)
	})

	it('returns null for an invalid date', () => {
		expect(parseDateOrNull(new Date('not a date'))).toBeNull()
	})

	it('returns null for an unparseable string', () => {
		expect(parseDateOrNull('12. of never')).toBeNull()
	})

	it('returns null for an empty string', () => {
		expect(parseDateOrNull('')).toBeNull()
	})

	it('returns null for null', () => {
		expect(parseDateOrNull(null)).toBeNull()
	})

	it('returns null for undefined', () => {
		expect(parseDateOrNull(undefined)).toBeNull()
	})

	it('returns null for the api zero time string', () => {
		expect(parseDateOrNull('0001-01-01T00:00:00Z')).toBeNull()
	})

	it('returns null for the api zero time as date', () => {
		expect(parseDateOrNull(new Date('0001-01-01T00:00:00Z'))).toBeNull()
	})
})
