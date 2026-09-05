import {beforeEach, describe, expect, it} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'

import {dateIsValid, formatDate, formatISO} from './formatDate'

describe('dateIsValid', () => {
	it.each([
		['a valid date', new Date('2021-02-06T12:00:00Z'), true],
		['a valid date string', '2021-02-06 12:00', true],
		['an invalid date', new Date('not a date'), false],
		['an unparseable string', 'not a date', false],
		['null', null, false],
		['undefined', undefined, false],
		['the api zero time', '0001-01-01T00:00:00Z', false],
	])('returns %s', (_name, date, expected) => {
		expect(dateIsValid(date)).toBe(expected)
	})
})

describe('formatISO', () => {
	it('formats a valid date', () => {
		expect(formatISO(new Date('2021-02-06T12:00:00Z'))).toBe('2021-02-06T12:00:00.000Z')
	})

	it.each([
		['an invalid date', new Date('not a date')],
		['null', null],
		['undefined', undefined],
		['the api zero time', '0001-01-01T00:00:00Z'],
	])('returns an empty string for %s', (_name, date) => {
		expect(formatISO(date)).toBe('')
	})
})

describe('formatDate', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('formats a valid date', () => {
		expect(formatDate(new Date(2021, 1, 6, 12, 0), 'YYYY-MM-DD')).toBe('2021-02-06')
	})

	it.each([
		['an invalid date', new Date('not a date')],
		['null', null],
		['undefined', undefined],
		['the api zero time', '0001-01-01T00:00:00Z'],
	])('returns an empty string for %s', (_name, date) => {
		expect(formatDate(date, 'YYYY-MM-DD')).toBe('')
	})
})
