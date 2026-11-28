import {afterEach, describe, expect, it} from 'vitest'

import {
	describeRepeatRule,
	formatInZone,
	formatOrdinal,
	weekdayName,
	type DescribeOptions,
} from './rruleDescribe'
import {classifyRepeat, isStrictInstant} from './rrule'
import type {TaskRepeat} from '@/client/generated'

function repeat(fields: Record<string, unknown>): TaskRepeat {
	return fields as unknown as TaskRepeat
}

const K = 'task.repeat.describe'

// Returns the key, its params and the plural count, so each case asserts exactly what reaches vue-i18n.
const keyT: DescribeOptions['t'] = (k, p, n) => `${k}|${JSON.stringify(p ?? {})}|${n ?? ''}`

// English source strings, rendered the way vue-i18n does for two plural forms.
const EN: Record<string, string> = {
	hourly: 'Every hour | Every {n} hours',
	daily: 'Every day | Every {n} days',
	dailyOn: 'Every day on {days} | Every {n} days on {days}',
	weekly: 'Every week | Every {n} weeks',
	weeklyOn: 'Every week on {days} | Every {n} weeks on {days}',
	weekdays: 'Every weekday',
	monthly: 'Every month | Every {n} months',
	monthlyOnDay: 'Every month on the {day} | Every {n} months on the {day}',
	monthlyOnLastDay: 'Every month on the last day | Every {n} months on the last day',
	monthlyOnDayOrLastDay: 'Every month on the {day}, or the last day of February | Every {n} months on the {day}, or the last day of February',
	monthlyOnNth: 'Every month on the {ordinal} {weekday} | Every {n} months on the {ordinal} {weekday}',
	monthlyOnLastWeekday: 'Every month on the last {weekday} | Every {n} months on the last {weekday}',
	lastWeekdayOfMonth: 'Every month on the last weekday | Every {n} months on the last weekday',
	yearly: 'Every year | Every {n} years',
	yearlyOn: 'Every year on {date} | Every {n} years on {date}',
	endCount: ', {n} time | , {n} times',
	endUntil: ', until {date}',
	'ordinal.one': '{n}st',
	'ordinal.two': '{n}nd',
	'ordinal.few': '{n}rd',
	'ordinal.other': '{n}th',
	custom: 'Custom repeat: {parts}',
	'part.freq': '{freq}',
	'part.every.secondly': 'every second | every {n} seconds',
	'part.every.minutely': 'every minute | every {n} minutes',
	'part.every.hourly': 'every hour | every {n} hours',
	'part.every.daily': 'every day | every {n} days',
	'part.every.weekly': 'every week | every {n} weeks',
	'part.every.monthly': 'every month | every {n} months',
	'part.every.yearly': 'every year | every {n} years',
	'part.byDay': 'on {days}',
	'part.byDayEvery': 'on every {days}',
	'part.nthWeekday': 'the {ordinal} {weekday}',
	'part.lastWeekday': 'the last {weekday}',
	'part.nthLastWeekday': 'the {ordinal} to last {weekday}',
	'part.byMonthDay': 'on day {days}',
	'part.byMonth': 'in {months}',
	'part.byYearDay': 'on day {days} of the year',
	'part.byHour': 'at hour {values}',
	'part.count': '{n} time | {n} times',
	'part.countUnlimited': 'any number of times',
	'part.until': 'until {date}',
	'part.wkst': 'weeks start on {weekday}',
	'freq.secondly': 'secondly',
	'freq.minutely': 'minutely',
	'freq.hourly': 'hourly',
	'freq.daily': 'daily',
	'freq.weekly': 'weekly',
	'freq.monthly': 'monthly',
	'freq.yearly': 'yearly',
}

const englishT: DescribeOptions['t'] = (k, p = {}, n) => {
	const message = EN[k.slice(K.length + 1)]
	if (message === undefined) {
		throw new Error(`missing message for ${k}`)
	}
	const forms = message.split(' | ')
	const form = forms.length === 2 && n !== undefined ? forms[n === 1 ? 0 : 1] : forms[0]
	return form.replace(/\{(\w+)\}/g, (_m, name) => String(p[name]))
}

// Fri 2026-10-23 at noon UTC: the same calendar day in UTC and New York.
const FRI_OCT_23 = new Date('2026-10-23T12:00:00Z')

function opts(overrides: Partial<DescribeOptions> = {}): DescribeOptions {
	return {t: keyT, locale: 'en', weekStart: 1, zone: 'UTC', dueDate: FRI_OCT_23, ...overrides}
}

function english(r: TaskRepeat, overrides: Partial<DescribeOptions> = {}): string {
	return describeRepeatRule(r, opts({t: englishT, ...overrides}))
}

describe('describeRepeatRule', () => {
	const originalTZ = process.env.TZ

	afterEach(() => {
		process.env.TZ = originalTZ
	})

	it('describes no repeat as an empty string', () => {
		expect(describeRepeatRule(null, opts())).toBe('')
		expect(describeRepeatRule(undefined, opts())).toBe('')
		expect(describeRepeatRule(repeat({freq: ''}), opts())).toBe('')
	})

	describe('locked rules', () => {
		it.each([
			['a weekly rule', repeat({freq: 'weekly', interval: 1, by_day: ['mo'], read_only: true})],
			['a projection missing a part', repeat({freq: 'yearly', interval: 1, read_only: true})],
			['a projection without a frequency', repeat({read_only: true})],
		])('gives %s the server marks read-only the generic summary', (_name, r) => {
			expect(describeRepeatRule(r, opts())).toBe(`${K}.locked|{}|`)
		})

		it('describes a rule the server does not mark read-only from its parts', () => {
			expect(describeRepeatRule(repeat({freq: 'weekly', interval: 1, by_day: ['mo'], read_only: false}), opts()))
				.toBe(`${K}.weeklyOn|{"n":1,"days":"Monday"}|1`)
			expect(english(repeat({freq: 'monthly', interval: 1, by_month_day: [28, 29, 30, 31], by_set_pos: [-1], read_only: false})))
				.toBe('Every month on the last day')
		})
	})

	describe('editable shapes', () => {
		const cases: [string, TaskRepeat, string][] = [
			['hourly interval 3', repeat({freq: 'hourly', interval: 3}), `${K}.hourly|{"n":3}|3`],
			['daily', repeat({freq: 'daily', interval: 1}), `${K}.daily|{"n":1}|1`],
			['daily interval 2 count 5', repeat({freq: 'daily', interval: 2, count: 5}), `${K}.daily|{"n":2}|2${K}.endCount|{"n":5}|5`],
			['weekly implicit', repeat({freq: 'weekly', interval: 1}), `${K}.weeklyOn|{"n":1,"days":"Friday"}|1`],
			['weekly mo/we/fr', repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'we', 'fr']}), `${K}.weeklyOn|{"n":1,"days":"Monday, Wednesday, and Friday"}|1`],
			['weekly interval 2 tu/th', repeat({freq: 'weekly', interval: 2, by_day: ['tu', 'th']}), `${K}.weeklyOn|{"n":2,"days":"Tuesday and Thursday"}|2`],
			['weekdays', repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'tu', 'we', 'th', 'fr']}), `${K}.weekdays|{}|`],
			['weekdays every 2 weeks', repeat({freq: 'weekly', interval: 2, by_day: ['mo', 'tu', 'we', 'th', 'fr']}), `${K}.weeklyOn|{"n":2,"days":"Monday, Tuesday, Wednesday, Thursday, and Friday"}|2`],
			['monthly implicit', repeat({freq: 'monthly', interval: 1}), `${K}.monthlyOnDay|{"n":1,"day":"${K}.ordinal.few|{\\"n\\":23}|"}|1`],
			['monthly on the 15th, due the 23rd', repeat({freq: 'monthly', interval: 1, by_month_day: [15]}), `${K}.monthlyOnDay|{"n":1,"day":"${K}.ordinal.other|{\\"n\\":15}|"}|1`],
			['monthly last day', repeat({freq: 'monthly', interval: 1, by_month_day: [-1]}), `${K}.monthlyOnLastDay|{"n":1}|1`],
			['monthly 4th friday', repeat({freq: 'monthly', interval: 1, by_day: ['4fr']}), `${K}.monthlyOnNth|{"n":1,"ordinal":"${K}.ordinal.other|{\\"n\\":4}|","weekday":"Friday"}|1`],
			['monthly last friday', repeat({freq: 'monthly', interval: 1, by_day: ['-1fr']}), `${K}.monthlyOnLastWeekday|{"n":1,"weekday":"Friday"}|1`],
			['yearly', repeat({freq: 'yearly', interval: 1}), `${K}.yearlyOn|{"n":1,"date":"October 23"}|1`],
			['explicit interval 1 and wkst mo', repeat({freq: 'weekly', interval: 1, wkst: 'mo'}), `${K}.weeklyOn|{"n":1,"days":"Friday"}|1`],
			['upper-case codes', repeat({freq: 'WEEKLY', interval: 1, by_day: ['FR', 'MO']}), `${K}.weeklyOn|{"n":1,"days":"Monday and Friday"}|1`],
			['empty advanced lists', repeat({freq: 'daily', interval: 1, by_day: [], by_month: []}), `${K}.daily|{"n":1}|1`],
			['no interval', repeat({freq: 'weekly', by_day: ['mo']}), `${K}.weeklyOn|{"n":1,"days":"Monday"}|1`],
		]

		it.each(cases)('describes %s with its own key', (_name, r, expected) => {
			expect(classifyRepeat(r, FRI_OCT_23)).toBe('editable')
			expect(describeRepeatRule(r, opts())).toBe(expected)
		})

		it('reads the plural forms in English', () => {
			expect(english(repeat({freq: 'daily', interval: 1}))).toBe('Every day')
			expect(english(repeat({freq: 'daily', interval: 3}))).toBe('Every 3 days')
			expect(english(repeat({freq: 'monthly', interval: 2, by_day: ['2tu']}))).toBe('Every 2 months on the 2nd Tuesday')
			expect(english(repeat({freq: 'monthly', interval: 1, by_month_day: [1]}))).toBe('Every month on the 1st')
			expect(english(repeat({freq: 'monthly', interval: 1, by_month_day: [3]}))).toBe('Every month on the 3rd')
		})

		it('describes implicit shapes without a day when there is no due date', () => {
			expect(describeRepeatRule(repeat({freq: 'weekly', interval: 1}), opts({dueDate: null}))).toBe(`${K}.weekly|{"n":1}|1`)
			expect(describeRepeatRule(repeat({freq: 'monthly', interval: 2}), opts({dueDate: null}))).toBe(`${K}.monthly|{"n":2}|2`)
			expect(describeRepeatRule(repeat({freq: 'yearly', interval: 1}), opts({dueDate: undefined}))).toBe(`${K}.yearly|{"n":1}|1`)
		})

		it('reads the due weekday in the recurrence zone, not the browser zone', () => {
			process.env.TZ = 'UTC'
			// Fri 20:00 in New York is already Saturday in UTC.
			const dueFridayEvening = new Date('2026-10-24T00:00:00Z')
			const zoned = opts({t: englishT, zone: 'America/New_York', dueDate: dueFridayEvening})
			expect(describeRepeatRule(repeat({freq: 'weekly', interval: 1}), zoned)).toBe('Every week on Friday')
			expect(describeRepeatRule(repeat({freq: 'monthly', interval: 1}), zoned)).toBe('Every month on the 23rd')
			expect(describeRepeatRule(repeat({freq: 'yearly', interval: 1}), zoned)).toBe('Every year on October 23')
			expect(describeRepeatRule(repeat({freq: 'weekly', interval: 1}), {...zoned, zone: 'UTC'})).toBe('Every week on Saturday')
		})
	})

	describe('end conditions', () => {
		it('appends the count with its plural form', () => {
			expect(english(repeat({freq: 'weekly', interval: 1, by_day: ['fr'], count: 3}))).toBe('Every week on Friday, 3 times')
			expect(english(repeat({freq: 'weekly', interval: 1, by_day: ['fr'], count: 1}))).toBe('Every week on Friday, 1 time')
		})

		it('shows an end date picked in the editor as a date only', () => {
			const r = repeat({freq: 'weekly', interval: 1, by_day: ['fr'], until: '2026-11-06T23:59:59Z'})
			expect(describeRepeatRule(r, opts())).toBe(`${K}.weeklyOn|{"n":1,"days":"Friday"}|1${K}.endUntil|{"date":"Nov 6"}|`)
		})

		it('reads the end date in the recurrence zone when it differs from the browser zone', () => {
			process.env.TZ = 'UTC'
			const r = repeat({freq: 'weekly', interval: 1, by_day: ['fr'], until: '2026-11-07T04:59:59Z'})
			expect(english(r, {zone: 'America/New_York'})).toBe('Every week on Friday, until Nov 6')
			expect(english(r, {zone: 'UTC'})).toBe('Every week on Friday, until Nov 7, 4:59 AM')
		})

		it('keeps the time of an end instant that is not the end of a day', () => {
			const r = repeat({freq: 'daily', interval: 1, until: '2027-06-30T12:00:00Z'})
			expect(english(r, {dueDate: new Date('2027-06-01T12:00:00Z')})).toBe('Every day, until Jun 30, 12:00 PM')
		})

		it('adds the year when the end date is in another year than the due date', () => {
			const r = repeat({freq: 'weekly', interval: 1, by_day: ['fr'], until: '2027-06-30T23:59:59Z'})
			expect(english(r)).toBe('Every week on Friday, until Jun 30, 2027')
		})
	})

	describe('extended shapes', () => {
		it('describes yearly in one month on the due day', () => {
			const r = repeat({freq: 'yearly', interval: 1, by_month: [3]})
			expect(describeRepeatRule(r, opts())).toBe(`${K}.yearlyOn|{"n":1,"date":"March 23"}|1`)
			expect(english(repeat({freq: 'yearly', interval: 2, by_month: [3], count: 4}))).toBe('Every 2 years on March 23, 4 times')
		})

		it('falls back to the generic description when the due day does not exist in that month', () => {
			const r = repeat({freq: 'yearly', interval: 1, by_month: [2]})
			expect(english(r, {dueDate: new Date('2026-10-31T12:00:00Z')})).toBe('Custom repeat: yearly, in February')
			expect(english(r, {dueDate: null})).toBe('Custom repeat: yearly, in February')
			expect(english(repeat({freq: 'yearly', interval: 1, by_month: [13]}))).toBe('Custom repeat: yearly, by_month=13')
		})

		it('describes daily on a list of weekdays', () => {
			const r = repeat({freq: 'daily', interval: 2, by_day: ['fr', 'mo']})
			expect(describeRepeatRule(r, opts())).toBe(`${K}.dailyOn|{"n":2,"days":"Monday and Friday"}|2`)
			expect(english(repeat({freq: 'daily', interval: 1, by_day: ['mo', 'tu', 'we', 'th', 'fr']}))).toBe('Every weekday')
			expect(english(repeat({freq: 'daily', interval: 1, by_day: ['mo', 'mo']}))).toBe('Every day on Monday')
		})

		it('describes the last weekday of the month', () => {
			const r = repeat({freq: 'monthly', interval: 1, by_day: ['MO', 'TU', 'WE', 'TH', 'FR'], by_set_pos: [-1]})
			expect(describeRepeatRule(r, opts())).toBe(`${K}.lastWeekdayOfMonth|{"n":1}|1`)
			expect(english(repeat({freq: 'monthly', interval: 3, by_day: ['mo', 'tu', 'we', 'th', 'fr'], by_set_pos: [-1], until: '2026-12-31T23:59:59Z'})))
				.toBe('Every 3 months on the last weekday, until Dec 31')
		})
	})

	describe('month-end rules from the legacy monthly migration', () => {
		// A legacy monthly task due on the 29th-31st became BYMONTHDAY=28..D;BYSETPOS=-1: the last of those days a month has.
		it.each([
			[[28, 29, 30, 31], 'Every month on the last day'],
			[[28, 29, 30], 'Every month on the 30th, or the last day of February'],
			[[28, 29], 'Every month on the 29th, or the last day of February'],
		])('describes BYMONTHDAY=%j;BYSETPOS=-1', (days, sentence) => {
			expect(english(repeat({freq: 'monthly', interval: 1, by_month_day: days, by_set_pos: [-1]}))).toBe(sentence)
		})

		it('reads the days in any order, with an interval and an ending', () => {
			expect(english(repeat({freq: 'monthly', interval: 1, by_month_day: [31, 28, 30, 29], by_set_pos: [-1]}))).toBe('Every month on the last day')
			expect(english(repeat({freq: 'monthly', interval: 2, by_month_day: [28, 29, 30], by_set_pos: [-1], count: 3})))
				.toBe('Every 2 months on the 30th, or the last day of February, 3 times')
			expect(english(repeat({freq: 'monthly', interval: 1, by_month_day: [28, 29, 30, 31], by_set_pos: [-1], until: '2026-12-31T23:59:59Z'})))
				.toBe('Every month on the last day, until Dec 31')
		})

		it.each([
			['a gap in the days', {by_month_day: [28, 30], by_set_pos: [-1]}],
			['days starting before the 28th', {by_month_day: [27, 28, 29], by_set_pos: [-1]}],
			['only the 28th', {by_month_day: [28], by_set_pos: [-1]}],
			['the first of the set', {by_month_day: [28, 29, 30, 31], by_set_pos: [1]}],
			['another part as well', {by_month_day: [28, 29, 30, 31], by_set_pos: [-1], by_month: [1]}],
		])('keeps the generic description for %s', (_name, parts) => {
			expect(english(repeat({freq: 'monthly', interval: 1, ...parts}))).toMatch(/^Custom repeat: /)
		})
	})

	describe('generic fallback', () => {
		it('names every field of a compound rule in natural fragments', () => {
			const r = repeat({freq: 'yearly', interval: 2, by_month: [3, 9], by_day: ['1mo']})
			expect(english(r)).toBe('Custom repeat: every 2 years, in March and September, on the 1st Monday')
		})

		it('uses the generic key with a fragment per field', () => {
			const out = describeRepeatRule(repeat({freq: 'yearly', interval: 1, by_month: [3, 9], by_day: ['1mo']}), opts())
			expect(out.startsWith(`${K}.custom|`)).toBe(true)
			expect(out).toContain(`${K}.part.freq|`)
			expect(out).toContain(`${K}.freq.yearly|`)
			expect(out).toContain(`${K}.part.byMonth|`)
			expect(out).toContain(`${K}.part.byDay|`)
			expect(out).toContain(`${K}.part.nthWeekday|`)
		})

		it.each([
			[['1mo', '-1fr'], 'monthly, on the 1st Monday and the last Friday'],
			[['+2tu', '2tu'], 'monthly, by_day=+2tu,2tu'],
			[['-2fr', '1mo'], 'monthly, on the 2nd to last Friday and the 1st Monday'],
			[['5fr', '-5mo'], 'monthly, on the 5th Friday and the 5th to last Monday'],
			// Ordinals past 5 make the scheduler index its weekday mask out of range.
			[['6mo', '1tu'], 'monthly, by_day=6mo,1tu'],
			[['-6mo'], 'monthly, by_day=-6mo'],
			[['53mo'], 'monthly, by_day=53mo'],
			[['fr', 'sa'], 'monthly, on every Friday and Saturday'],
			// A day must match the plain and the ordinal weekdays at once.
			[['2mo', 'sa'], 'monthly, by_day=2mo,sa'],
			[['2MO', 'mo'], 'monthly, by_day=2MO,mo'],
			[['0fr', '1mo'], 'monthly, by_day=0fr,1mo'],
			[['54mo', '1tu'], 'monthly, by_day=54mo,1tu'],
			[['xx', 'sa'], 'monthly, by_day=xx,sa'],
		])('describes monthly by_day %j as "%s"', (by_day, text) => {
			expect(english(repeat({freq: 'monthly', interval: 1, by_day}))).toBe(`Custom repeat: ${text}`)
		})

		it.each([
			['weekly', '2mo', 'Custom repeat: weekly, on Monday'],
			['daily', '-1fr', 'Custom repeat: daily, on Friday'],
			['hourly', '1tu', 'Custom repeat: hourly, by_day=1tu'],
			['yearly', '1mo', 'Custom repeat: yearly, on the 1st Monday'],
			['fortnightly', '2mo', 'Custom repeat: freq=fortnightly, by_day=2mo'],
		])('describes a %s rule with the weekday code %s by what the scheduler does', (freq, code, text) => {
			expect(english(repeat({freq, interval: 1, by_day: [code]}))).toBe(text)
		})

		it('never describes an unknown frequency as daily', () => {
			expect(describeRepeatRule(repeat({freq: 'minutely', interval: 1}), opts()))
				.toBe(`${K}.custom|{"parts":"${K}.part.freq|{\\"freq\\":\\"${K}.freq.minutely|{}|\\"}|"}|`)
			expect(english(repeat({freq: 'minutely', interval: 15}))).toBe('Custom repeat: every 15 minutes')
			expect(english(repeat({freq: 'secondly', interval: 1}))).toBe('Custom repeat: secondly')
			expect(english(repeat({freq: 'fortnightly', interval: 2}))).toBe('Custom repeat: freq=fortnightly, interval=2')
			expect(english(repeat({freq: 'fortnightly', interval: 1}))).toBe('Custom repeat: freq=fortnightly')
			expect(describeRepeatRule(repeat({freq: 'fortnightly', interval: 1}), opts())).not.toContain('daily')
		})

		it.each([
			[repeat({freq: 'daily', interval: 1, foo: [1, 2]}), 'Custom repeat: daily, foo=[1,2]'],
			[repeat({freq: 'daily', interval: 1, foo: 'bar'}), 'Custom repeat: daily, foo=bar'],
			[repeat({freq: 'daily', foo: null}), 'Custom repeat: daily, foo=null'],
			[repeat({freq: 'daily', foo: []}), 'Custom repeat: daily, foo=[]'],
			[repeat({freq: 'daily', interval: 1, by_day: ['mo'], foo: null}), 'Custom repeat: daily, on Monday, foo=null'],
			[repeat({freq: 'yearly', interval: 1, by_month: [3], foo: []}), 'Custom repeat: yearly, in March, foo=[]'],
		])('shows the unknown key of %j raw, even when empty', (r, text) => {
			expect(english(r)).toBe(text)
		})

		it('lists every remaining field in a fixed order', () => {
			const r = repeat({
				until: '2027-01-01T12:00:00Z',
				count: 3,
				wkst: 'su',
				by_second: [0],
				by_minute: [30],
				by_hour: [9, 17],
				by_set_pos: [1, -1],
				by_month_day: [1, 15],
				by_year_day: [100],
				by_week_no: [1],
				by_month: [1],
				interval: 1,
				freq: 'monthly',
			})
			expect(english(r)).toBe('Custom repeat: monthly, by_month=1, by_week_no=1, by_year_day=100, by_month_day=1,15, ' +
				'by_set_pos=1,-1, by_hour=9,17, by_minute=30, by_second=0, weeks start on Sunday, 3 times, until Jan 1, 2027, 12:00 PM')
		})

		it('describes count and until together without dropping either', () => {
			expect(english(repeat({freq: 'daily', interval: 1, count: 3, until: '2026-12-31T23:59:59Z'})))
				.toBe('Custom repeat: daily, 3 times, until Dec 31')
		})

		it.each([
			[repeat({freq: 'daily', interval: 1, count: 0}), 'Custom repeat: daily, any number of times'],
			[repeat({freq: 'daily', interval: 1, count: -1}), 'Custom repeat: daily, any number of times'],
			[repeat({freq: 'daily', interval: 1, count: 0, until: '2026-12-31T23:59:59Z'}), 'Custom repeat: daily, any number of times, until Dec 31'],
			[repeat({freq: 'weekly', interval: 1, by_day: ['fr'], count: 0}), 'Custom repeat: weekly, on Friday, any number of times'],
		])('never describes the unlimited or unreadable count of %j as finite', (r, text) => {
			expect(english(r)).toBe(text)
		})

		it.each([
			[repeat({freq: 'daily', interval: 1, until: 'next tuesday'}), 'Custom repeat: daily, until next tuesday'],
			[repeat({freq: 'daily', interval: 1, until: '2026-02-30T12:00:00Z'}), 'Custom repeat: daily, until 2026-02-30T12:00:00Z'],
			[repeat({freq: 'daily', interval: 1, until: '2026-01-01T24:00:00Z'}), 'Custom repeat: daily, until 2026-01-01T24:00:00Z'],
			[repeat({freq: 'daily', interval: 1, until: '2026-01-01T12:00:00'}), 'Custom repeat: daily, until 2026-01-01T12:00:00'],
			[repeat({freq: 'daily', interval: 1, by_day: ['mo', 'fr'], until: '2026-02-30T12:00:00Z'}), 'Custom repeat: daily, on Monday and Friday, until 2026-02-30T12:00:00Z'],
			[repeat({freq: 'yearly', interval: 1, by_month: [3], until: '2026-13-01T12:00:00Z'}), 'Custom repeat: yearly, in March, until 2026-13-01T12:00:00Z'],
		])('shows the end value of %j verbatim when it is not a valid instant', (r, text) => {
			expect(english(r)).toBe(text)
		})

		// The selector combinations described in words; each matches what rrule-go produces for it.
		it.each([
			['daily in a month', {freq: 'daily', interval: 1, by_month: [3]}, 'daily, in March'],
			['weekly in a month', {freq: 'weekly', interval: 1, by_month: [3]}, 'weekly, in March'],
			['monthly in two months', {freq: 'monthly', interval: 1, by_month: [9, 3]}, 'monthly, in March and September'],
			['yearly in two months', {freq: 'yearly', interval: 1, by_month: [3, 9]}, 'yearly, in March and September'],
			['daily ordinal weekdays', {freq: 'daily', interval: 1, by_day: ['2mo', 'fr']}, 'daily, on Monday and Friday'],
			['weekly ordinal weekdays', {freq: 'weekly', interval: 1, by_day: ['2mo', '-1fr']}, 'weekly, on Monday and Friday'],
			['weekly with a negative interval', {freq: 'weekly', interval: -2, by_day: ['mo']}, 'weekly, interval=-2, on Monday'],
			['monthly plain weekdays', {freq: 'monthly', interval: 1, by_day: ['mo', 'fr']}, 'monthly, on every Monday and Friday'],
			['yearly plain weekday', {freq: 'yearly', interval: 1, by_day: ['fr']}, 'yearly, on every Friday'],
			['yearly ordinal weekdays', {freq: 'yearly', interval: 1, by_day: ['-1fr', '20mo']}, 'yearly, on the last Friday and the 20th Monday'],
			['monthly month days', {freq: 'monthly', interval: 1, by_month_day: [31, 1, 15]}, 'monthly, on day 1, 15, and 31'],
			['yearly year days', {freq: 'yearly', interval: 1, by_year_day: [100, 1]}, 'yearly, on day 1 and 100 of the year'],
			['daily hours', {freq: 'daily', interval: 1, by_hour: [17, 9]}, 'daily, at hour 9 and 17'],
			['yearly months with one ordinal weekday', {freq: 'yearly', interval: 1, by_month: [2], by_day: ['-1fr']}, 'yearly, in February, on the last Friday'],
			['yearly months with a 5th weekday', {freq: 'yearly', interval: 1, by_month: [3], by_day: ['-5fr']}, 'yearly, in March, on the 5th to last Friday'],
		])('describes %s in words', (_name, fields, text) => {
			expect(english(repeat(fields))).toBe(`Custom repeat: ${text}`)
		})

		// Everything outside those combinations is shown as stored: stored order, duplicates kept, every selector.
		it.each([
			['negative interval', {freq: 'daily', interval: -1}, 'daily, interval=-1'],
			['non-integer interval', {freq: 'daily', interval: 1.5}, 'daily, interval=1.5'],
			['month out of range', {freq: 'yearly', interval: 1, by_month: [3, 13]}, 'yearly, by_month=3,13'],
			['duplicate months', {freq: 'yearly', interval: 1, by_month: [9, 3, 3]}, 'yearly, by_month=9,3,3'],
			['hourly in a month', {freq: 'hourly', interval: 1, by_month: [3]}, 'hourly, by_month=3'],
			['yearly month day without months', {freq: 'yearly', interval: 1, by_month_day: [1]}, 'yearly, by_month_day=1'],
			['yearly month day in a month', {freq: 'yearly', interval: 1, by_month: [3], by_month_day: [1]}, 'yearly, by_month=3, by_month_day=1'],
			// Fires on the 15th of every month: the weekly frequency does not hold.
			['weekly month day', {freq: 'weekly', interval: 1, by_month_day: [15]}, 'weekly, by_month_day=15'],
			['daily month day', {freq: 'daily', interval: 1, by_month_day: [15]}, 'daily, by_month_day=15'],
			['daily year day', {freq: 'daily', interval: 1, by_year_day: [100]}, 'daily, by_year_day=100'],
			['negative month day', {freq: 'monthly', interval: 1, by_month_day: [-1, 15]}, 'monthly, by_month_day=-1,15'],
			['month day 0', {freq: 'monthly', interval: 1, by_month_day: [0]}, 'monthly, by_month_day=0'],
			['month day 32', {freq: 'monthly', interval: 1, by_month_day: [32]}, 'monthly, by_month_day=32'],
			['duplicate month days', {freq: 'monthly', interval: 1, by_month_day: [15, 1, 15]}, 'monthly, by_month_day=15,1,15'],
			['negative year day', {freq: 'yearly', interval: 1, by_year_day: [-1]}, 'yearly, by_year_day=-1'],
			['year day 367', {freq: 'yearly', interval: 1, by_year_day: [367]}, 'yearly, by_year_day=367'],
			['yearly week number alone', {freq: 'yearly', interval: 1, by_week_no: [1]}, 'yearly, by_week_no=1'],
			['monthly week number alone', {freq: 'monthly', interval: 1, by_week_no: [1]}, 'monthly, by_week_no=1'],
			['weekly week number alone', {freq: 'weekly', interval: 1, by_week_no: [1]}, 'weekly, by_week_no=1'],
			['negative week number', {freq: 'daily', interval: 1, by_week_no: [-1]}, 'daily, by_week_no=-1'],
			// rrule-go marks Jan 1-2 2022 as week 53 although they are ISO week 52.
			['daily week 53', {freq: 'daily', interval: 1, by_week_no: [53]}, 'daily, by_week_no=53'],
			['daily week 52', {freq: 'daily', interval: 1, by_week_no: [52]}, 'daily, by_week_no=52'],
			['daily week numbers', {freq: 'daily', interval: 1, by_week_no: [2, 1]}, 'daily, by_week_no=2,1'],
			['yearly week 53 with a weekday', {freq: 'yearly', interval: 1, by_week_no: [53], by_day: ['sa']}, 'yearly, by_week_no=53, by_day=sa'],
			['yearly week numbers with weekdays', {freq: 'yearly', interval: 1, by_week_no: [20, 1], by_day: ['fr', 'mo']}, 'yearly, by_week_no=20,1, by_day=fr,mo'],
			['yearly month with an ordinal past 5', {freq: 'yearly', interval: 1, by_month: [1], by_day: ['-6mo']}, 'yearly, by_month=1, by_day=-6mo'],
			['weekly duplicate weekday', {freq: 'weekly', interval: 1, by_day: ['mo', 'MO']}, 'weekly, by_day=mo,MO'],
			['weekly ordinals naming one weekday', {freq: 'weekly', interval: 1, by_day: ['2mo', 'mo']}, 'weekly, by_day=2mo,mo'],
			['hourly weekday', {freq: 'hourly', interval: 1, by_day: ['1tu']}, 'hourly, by_day=1tu'],
			['set position', {freq: 'monthly', interval: 1, by_day: ['fr'], by_set_pos: [-1]}, 'monthly, by_day=fr, by_set_pos=-1'],
			['hour out of range', {freq: 'daily', interval: 1, by_hour: [24]}, 'daily, by_hour=24'],
			['duplicate hours', {freq: 'daily', interval: 1, by_hour: [17, 9, 9]}, 'daily, by_hour=17,9,9'],
			['minutes', {freq: 'daily', interval: 1, by_minute: [0, 30]}, 'daily, by_minute=0,30'],
			['seconds', {freq: 'daily', interval: 1, by_second: [15]}, 'daily, by_second=15'],
			['yearly months with two ordinal weekdays', {freq: 'yearly', interval: 1, by_month: [3], by_day: ['1mo', '-1fr']}, 'yearly, by_month=3, by_day=1mo,-1fr'],
			['yearly months with a plain weekday', {freq: 'yearly', interval: 1, by_month: [3], by_day: ['mo']}, 'yearly, by_month=3, by_day=mo'],
			['unknown week start', {freq: 'weekly', interval: 2, wkst: 'xx'}, 'every 2 weeks, wkst=xx'],
			['week start in capitals', {freq: 'weekly', interval: 2, wkst: 'SU'}, 'every 2 weeks, weeks start on Sunday'],
			['non-integer count', {freq: 'daily', interval: 1, count: 1.5}, 'daily, count=1.5'],
		])('describes %s as stored', (_name, fields, text) => {
			expect(english(repeat(fields))).toBe(`Custom repeat: ${text}`)
		})

		it('tells duplicate hours apart when a set position counts them', () => {
			// The scheduler keeps both 09:00 entries, so position 2 is 09:00 here and 17:00 without the duplicate.
			expect(english(repeat({freq: 'daily', interval: 1, by_hour: [9, 9, 17], by_set_pos: [2]})))
				.toBe('Custom repeat: daily, by_set_pos=2, by_hour=9,9,17')
			expect(english(repeat({freq: 'daily', interval: 1, by_hour: [9, 17], by_set_pos: [2]})))
				.toBe('Custom repeat: daily, by_set_pos=2, by_hour=9,17')
		})

		it('shows every selector as stored when one combination has no exact wording', () => {
			expect(english(repeat({freq: 'monthly', interval: 2, by_month_day: [13], by_day: ['fr'], count: 3})))
				.toBe('Custom repeat: every 2 months, by_month_day=13, by_day=fr, 3 times')
			expect(english(repeat({freq: 'yearly', interval: 1, by_month: [9, 3, 3], by_day: ['fr', 'MO'], by_hour: [9]})))
				.toBe('Custom repeat: yearly, by_month=9,3,3, by_day=fr,MO, by_hour=9')
		})

		it('lists plain weekdays in week-start order', () => {
			const r = repeat({freq: 'weekly', interval: 1, by_day: ['1mo', 'su']})
			expect(english(r, {weekStart: 1})).toBe('Custom repeat: weekly, on Monday and Sunday')
			expect(english(r, {weekStart: 0})).toBe('Custom repeat: weekly, on Sunday and Monday')
		})
	})

	describe('field completeness', () => {
		const STORED_FIELDS = ['freq', 'interval', 'by_month', 'by_week_no', 'by_year_day', 'by_month_day', 'by_day', 'by_set_pos',
			'by_hour', 'by_minute', 'by_second', 'wkst', 'count', 'until']

		const PART_KEYS: Record<string, string> = {
			by_month: 'part.byMonth',
			by_year_day: 'part.byYearDay',
			by_month_day: 'part.byMonthDay',
			by_day: 'part.byDay',
			by_hour: 'part.byHour',
			wkst: 'part.wkst',
			count: 'part.count',
			until: 'part.until',
		}

		const foreign: TaskRepeat[] = [
			repeat({freq: 'minutely', interval: 1}),
			repeat({freq: 'secondly', interval: 1}),
			repeat({freq: 'yearly', interval: 1, by_month: [3, 9]}),
			repeat({freq: 'monthly', interval: 1, by_day: ['fr'], by_set_pos: [-1]}),
			repeat({freq: 'monthly', interval: 1, by_day: ['5fr']}),
			repeat({freq: 'monthly', interval: 1, by_day: ['2tu', '4tu']}),
			repeat({freq: 'monthly', interval: 1, by_day: ['2tu'], by_month_day: [15]}),
			repeat({freq: 'monthly', interval: 1, by_month_day: [1, 15]}),
			repeat({freq: 'monthly', interval: 1, by_month_day: [-2]}),
			repeat({freq: 'weekly', interval: 1, by_day: ['2mo']}),
			repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'mo']}),
			repeat({freq: 'weekly', interval: 1, by_month_day: [15]}),
			repeat({freq: 'hourly', interval: 1, by_day: ['mo']}),
			repeat({freq: 'yearly', interval: 1, by_day: ['1mo']}),
			repeat({freq: 'yearly', interval: 1, by_month_day: [1]}),
			repeat({freq: 'daily', interval: 1, count: 3, until: '2027-01-01T00:00:00Z'}),
			repeat({freq: 'daily', interval: 1, count: 0}),
			repeat({freq: 'daily', interval: 1, until: '2026-02-30T12:00:00Z'}),
			repeat({freq: 'daily', interval: 0}),
			repeat({freq: 'daily', interval: -1}),
			repeat({freq: 'weekly', interval: 2, wkst: 'su'}),
			repeat({freq: 'yearly', interval: 1, by_year_day: [100]}),
			repeat({freq: 'yearly', interval: 1, by_week_no: [1]}),
			repeat({freq: 'daily', interval: 1, by_hour: [9]}),
			repeat({freq: 'daily', interval: 1, by_minute: [30]}),
			repeat({freq: 'daily', interval: 1, by_second: [0]}),
			repeat({freq: 'daily', foo: 1}),
			repeat({freq: 'daily', foo: null}),
			repeat({freq: 'daily', interval: 1, by_day: ['mo'], foo: []}),
			repeat({freq: 'daily', interval: 1, count: -1}),
			repeat({freq: 'daily', interval: 1, until: '2026-01-01T24:00:00Z'}),
			repeat({freq: 'yearly', interval: 3, by_month: [1], by_month_day: [1], by_day: ['mo'], wkst: 'tu', count: 2}),
		]

		it.each(foreign.map(r => [JSON.stringify(r), r]))('gives every stored field of %s a fragment', (_name, r) => {
			expect(classifyRepeat(r, FRI_OCT_23)).toBe('foreign')
			const out = describeRepeatRule(r, opts())
			for (const [field, value] of Object.entries(r)) {
				const isEmpty = value === null || value === undefined || (Array.isArray(value) && value.length === 0)
				if (!STORED_FIELDS.includes(field)) {
					expect(out).toContain(`${field}=${JSON.stringify(value)}`)
					continue
				}
				if (isEmpty) {
					continue
				}
				const asStored = out.includes(`${field}=`)
				if (field === 'freq') {
					expect(asStored || /\.part\.(freq|every\.)/.test(out)).toBe(true)
				} else if (field === 'interval') {
					// The backend reads an interval of 0 as 1, the frequency's own default.
					if (value !== 1 && value !== 0) {
						expect(asStored || new RegExp(`part\\.every\\.\\w+\\|\\{\\\\?"n\\\\?":${value}\\}`).test(out)).toBe(true)
					}
				} else if (field === 'count') {
					expect(asStored || /\.part\.count(Unlimited)?\|/.test(out)).toBe(true)
				} else if (field === 'until' && !isStrictInstant(value)) {
					expect(out).toContain(`${K}.part.until|`)
					expect(out).toContain(String(value))
				} else if (field === 'by_day') {
					expect(asStored || /\.part\.byDay(Every)?\|/.test(out)).toBe(true)
				} else {
					expect(asStored || (field in PART_KEYS && out.includes(`${K}.${PART_KEYS[field]}|`))).toBe(true)
				}
			}
		})
	})

	describe('locale', () => {
		it('changes names and list words but not the keys', () => {
			const r = repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'we', 'fr']})
			expect(describeRepeatRule(r, opts({locale: 'de'}))).toBe(`${K}.weeklyOn|{"n":1,"days":"Montag, Mittwoch und Freitag"}|1`)
			expect(describeRepeatRule(r, opts({locale: 'en'}))).toBe(`${K}.weeklyOn|{"n":1,"days":"Monday, Wednesday, and Friday"}|1`)
		})

		it('formats the yearly date for the locale', () => {
			expect(describeRepeatRule(repeat({freq: 'yearly', interval: 1}), opts({locale: 'de'}))).toBe(`${K}.yearlyOn|{"n":1,"date":"23. Oktober"}|1`)
		})

		it('orders weekdays by the week start setting only', () => {
			const r = repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'sa', 'su']})
			expect(describeRepeatRule(r, opts({weekStart: 1}))).toBe(`${K}.weeklyOn|{"n":1,"days":"Monday, Saturday, and Sunday"}|1`)
			expect(describeRepeatRule(r, opts({weekStart: 0}))).toBe(`${K}.weeklyOn|{"n":1,"days":"Sunday, Monday, and Saturday"}|1`)
			expect(describeRepeatRule(r, opts({weekStart: 6}))).toBe(`${K}.weeklyOn|{"n":1,"days":"Saturday, Sunday, and Monday"}|1`)
		})
	})
})

describe('describeRepeatRule timing', () => {
	function medianMs(run: () => unknown, times = 20): number {
		const sorted = Array.from({length: times}, () => {
			const started = performance.now()
			run()
			return performance.now() - started
		}).sort((a, b) => a - b)
		return sorted[Math.floor(sorted.length / 2)]!
	}

	it('summarises each slow fixture in a median under 50 ms', () => {
		const fixtures: [TaskRepeat, string, string][] = [
			[repeat({freq: 'monthly', interval: 1, by_month_day: [31]}), '2026-08-31T08:00:00Z', 'Europe/Berlin'],
			[repeat({freq: 'monthly', interval: 12, by_month_day: [31]}), '2027-04-30T08:00:00Z', 'Europe/Berlin'],
			[repeat({freq: 'yearly', interval: 100}), '2400-02-29T12:00:00Z', 'UTC'],
			[repeat({freq: 'yearly', interval: 3}), '2028-02-29T12:00:00Z', 'America/New_York'],
			[repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'tu', 'we', 'th', 'fr', 'sa', 'su'], count: 10000}), '2091-03-10T07:30:00Z', 'America/New_York'],
			[repeat({freq: 'daily', interval: 36525}), '2026-10-23T13:00:00Z', 'America/New_York'],
		]
		const medians = fixtures.map(([r, due, zone]) => medianMs(() => describeRepeatRule(r, opts({zone, dueDate: new Date(due)}))))
		expect(Math.max(...medians)).toBeLessThan(50)
	})
})

describe('locale formatting helpers', () => {
	it('names weekdays and months', () => {
		expect(weekdayName('mo', 'en')).toBe('Monday')
		expect(weekdayName('SU', 'en', 'short')).toBe('Sun')
		expect(weekdayName('fr', 'de')).toBe('Freitag')
		expect(weekdayName('xx', 'en')).toBe('xx')
	})

	it('names months in the user\'s language', () => {
		const marchAndDecember = repeat({freq: 'yearly', by_month: [12, 3]})
		expect(english(marchAndDecember, {dueDate: null})).toContain('in March and December')
		expect(english(marchAndDecember, {dueDate: null, locale: 'de'})).toContain('in März und Dezember')
	})

	it('joins lists with the locale conjunction', () => {
		expect(english(repeat({freq: 'weekly', by_day: ['mo', 'we', 'fr']}))).toBe('Every week on Monday, Wednesday, and Friday')
		expect(english(repeat({freq: 'weekly', by_day: ['mo', 'fr']}), {locale: 'de'})).toBe('Every week on Montag und Freitag')
	})

	it.each([
		[1, 'one'], [2, 'two'], [3, 'few'], [4, 'other'], [11, 'other'], [12, 'other'], [13, 'other'], [21, 'one'], [22, 'two'], [23, 'few'],
	])('picks the English ordinal category for %i', (n, category) => {
		expect(formatOrdinal(n, {t: keyT, locale: 'en'})).toBe(`${K}.ordinal.${category}|{"n":${n}}|`)
	})

	it('uses the other ordinal form for locales without ordinal categories', () => {
		expect(formatOrdinal(1, {t: keyT, locale: 'de'})).toBe(`${K}.ordinal.other|{"n":1}|`)
	})

	it('formats an instant in the given zone', () => {
		const instant = new Date('2026-11-07T04:59:59Z')
		expect(formatInZone(instant, 'America/New_York', 'en', 'monthDay')).toBe('Nov 6')
		expect(formatInZone(instant, 'UTC', 'en', 'monthDay')).toBe('Nov 7')
		expect(formatInZone(instant, 'America/New_York', 'en')).toBe('Fri, Nov 6')
		expect(formatInZone(instant, 'UTC', 'en', 'monthDayYear')).toBe('Nov 7, 2026')
		expect(formatInZone(instant, 'UTC', 'de', 'monthDay')).toBe('7. Nov.')
	})
})
