import {describe, it, expect} from 'vitest'

import {isComplexRepeat} from './rrule'
import type {ITaskRepeat} from '@/modelTypes/ITask'

describe('isComplexRepeat', () => {
	it('treats null/undefined as not complex', () => {
		expect(isComplexRepeat(null)).toBe(false)
		expect(isComplexRepeat(undefined)).toBe(false)
	})

	it('treats simple freq/interval rules as not complex', () => {
		expect(isComplexRepeat({freq: 'daily', interval: 1})).toBe(false)
		expect(isComplexRepeat({freq: 'weekly', interval: 2})).toBe(false)
	})

	it('treats a single byMonthDay as not complex', () => {
		expect(isComplexRepeat({freq: 'monthly', interval: 1, byMonthDay: [15]})).toBe(false)
	})

	it('treats multiple byMonthDay values as complex', () => {
		expect(isComplexRepeat({freq: 'monthly', interval: 1, byMonthDay: [1, 15]})).toBe(true)
	})

	it('treats advanced fields as complex', () => {
		const advanced: Partial<ITaskRepeat>[] = [
			{byDay: ['MO', 'WE', 'FR']},
			{byMonth: [3]},
			{byYearDay: [100]},
			{byWeekNo: [1]},
			{bySetPos: [-1]},
			{byHour: [9]},
			{byMinute: [30]},
			{bySecond: [0]},
			{count: 5},
			{until: '2026-01-01T00:00:00Z'},
			{wkst: 'MO'},
		]
		for (const extra of advanced) {
			expect(isComplexRepeat({freq: 'weekly', interval: 1, ...extra})).toBe(true)
		}
	})

	it('ignores empty advanced values', () => {
		expect(isComplexRepeat({freq: 'daily', interval: 1, byDay: [], until: '', wkst: ''})).toBe(false)
	})
})
