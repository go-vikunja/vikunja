import {describe, it, expect} from 'vitest'

import {
	describeRepeat,
	freqToUiFreq,
	isComplexRepeat,
	isMappableFreq,
	isRepeating,
	repeatFromSettings,
	uiFreqToFreq,
} from './rrule'
import type {TaskRepeat} from '@/client/generated'

describe('isComplexRepeat', () => {
	it('treats null/undefined as not complex', () => {
		expect(isComplexRepeat(null)).toBe(false)
		expect(isComplexRepeat(undefined)).toBe(false)
	})

	it('treats simple freq/interval rules as not complex', () => {
		expect(isComplexRepeat({freq: 'daily', interval: 1})).toBe(false)
		expect(isComplexRepeat({freq: 'weekly', interval: 2})).toBe(false)
	})

	it('treats a single by_month_day as not complex', () => {
		expect(isComplexRepeat({freq: 'monthly', interval: 1, by_month_day: [15]})).toBe(false)
		expect(isComplexRepeat({freq: 'monthly', interval: 2, by_month_day: [31]})).toBe(false)
	})

	// The day-of-month control offers 1-31 and only for monthly rules; anything else would be
	// regenerated without its BYMONTHDAY.
	it('treats a single by_month_day the day control cannot show as complex', () => {
		expect(isComplexRepeat({freq: 'monthly', interval: 1, by_month_day: [-1]})).toBe(true)
		expect(isComplexRepeat({freq: 'monthly', interval: 1, by_month_day: [0]})).toBe(true)
		expect(isComplexRepeat({freq: 'yearly', interval: 1, by_month_day: [15]})).toBe(true)
		expect(isComplexRepeat({freq: 'weekly', interval: 1, by_month_day: [3]})).toBe(true)
		expect(isComplexRepeat({freq: 'daily', interval: 1, by_month_day: [3]})).toBe(true)
	})

	it('treats multiple by_month_day values as complex', () => {
		expect(isComplexRepeat({freq: 'monthly', interval: 1, by_month_day: [1, 15]})).toBe(true)
	})

	it('treats the migrated month-end rule as complex', () => {
		expect(isComplexRepeat({
			freq: 'monthly',
			interval: 1,
			by_month_day: [28, 29, 30, 31],
			by_set_pos: [-1],
		})).toBe(true)
	})

	it('treats a rule the server marks read-only as complex', () => {
		expect(isComplexRepeat({freq: 'daily', interval: 1, read_only: true})).toBe(true)
		expect(isComplexRepeat({freq: 'daily', interval: 1, read_only: false})).toBe(false)
	})

	it('treats advanced fields as complex', () => {
		const advanced: Partial<TaskRepeat>[] = [
			{by_day: ['mo', 'we', 'fr']},
			{by_month: [3]},
			{by_year_day: [100]},
			{by_week_no: [1]},
			{by_set_pos: [-1]},
			{by_hour: [9]},
			{by_minute: [30]},
			{by_second: [0]},
			{count: 5},
			{until: '2026-01-01T00:00:00Z'},
			{wkst: 'mo'},
		]
		for (const extra of advanced) {
			expect(isComplexRepeat({freq: 'weekly', interval: 1, ...extra})).toBe(true)
		}
	})

	it('ignores empty and null advanced values', () => {
		expect(isComplexRepeat({freq: 'daily', interval: 1, by_day: [], until: '', wkst: ''})).toBe(false)
		expect(isComplexRepeat({freq: 'daily', interval: 1, by_day: null, by_month_day: null})).toBe(false)
	})

	it('treats a freq the simple editor cannot map as complex', () => {
		expect(isComplexRepeat({freq: 'minutely', interval: 1})).toBe(true)
		expect(isComplexRepeat({freq: 'secondly', interval: 1})).toBe(true)
	})
})

describe('isMappableFreq', () => {
	it('accepts the freqs the simple editor supports', () => {
		for (const freq of ['hourly', 'daily', 'weekly', 'monthly', 'yearly', 'DAILY']) {
			expect(isMappableFreq(freq)).toBe(true)
		}
	})

	it('rejects freqs with no dropdown unit', () => {
		expect(isMappableFreq('minutely')).toBe(false)
		expect(isMappableFreq('secondly')).toBe(false)
		expect(isMappableFreq('')).toBe(false)
		expect(isMappableFreq(undefined)).toBe(false)
	})
})

describe('frequency mapping', () => {
	it('maps API freqs to editor units and back', () => {
		expect(freqToUiFreq('monthly')).toBe('months')
		expect(freqToUiFreq('WEEKLY')).toBe('weeks')
		expect(uiFreqToFreq('years')).toBe('yearly')
		expect(uiFreqToFreq('hours')).toBe('hourly')
	})
})

describe('repeatFromSettings', () => {
	it('builds a freq/interval rule', () => {
		expect(repeatFromSettings(2, 'weeks')).toEqual({freq: 'weekly', interval: 2})
		expect(repeatFromSettings(1, 'months')).toEqual({freq: 'monthly', interval: 1})
	})

	it('adds a positive month day only', () => {
		expect(repeatFromSettings(1, 'months', 15)).toEqual({freq: 'monthly', interval: 1, by_month_day: [15]})
		expect(repeatFromSettings(1, 'months', 0)).toEqual({freq: 'monthly', interval: 1})
	})

	it('returns null for a non-positive amount', () => {
		expect(repeatFromSettings(0, 'days')).toBeNull()
		expect(repeatFromSettings(-1, 'days')).toBeNull()
	})
})

describe('isRepeating', () => {
	it('needs a rule with a freq', () => {
		expect(isRepeating({freq: 'daily', interval: 1})).toBe(true)
		expect(isRepeating({interval: 1})).toBe(false)
		expect(isRepeating(null)).toBe(false)
		expect(isRepeating(undefined)).toBe(false)
	})
})

describe('describeRepeat', () => {
	const t = (key: string, params?: Record<string, unknown>) => params ? `${key} ${JSON.stringify(params)}` : key

	it('names single-interval rules', () => {
		expect(describeRepeat({freq: 'daily', interval: 1}, t)).toBe('task.repeat.everyDay')
		expect(describeRepeat({freq: 'monthly'}, t)).toBe('task.repeat.everyMonth')
		expect(describeRepeat({freq: 'monthly', interval: 1, by_month_day: [15]}, t))
			.toBe('task.repeat.everyMonthOnDay {"day":15}')
	})

	it('describes longer intervals with the unit', () => {
		expect(describeRepeat({freq: 'weekly', interval: 3}, t))
			.toBe('task.repeat.everyN {"n":3,"unit":"task.repeat.weeks"}')
	})

	it('describes no rule as empty', () => {
		expect(describeRepeat(null, t)).toBe('')
		expect(describeRepeat(undefined, t)).toBe('')
	})
})
