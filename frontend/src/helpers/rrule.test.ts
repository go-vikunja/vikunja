import {afterEach, describe, it, expect, vi} from 'vitest'

import {
	canonicalRepeat,
	classifyRepeat,
	dayInZone,
	defaultCustomForm,
	dueDateOf,
	formToRepeat,
	isRepeating,
	normalizeRepeatForm,
	pickerDateToUntilDate,
	presetIdFor,
	recurrenceZone,
	repeatFromSettings,
	repeatPresets,
	repeatToForm,
	untilDateToPickerDate,
	validateRepeatForm,
	type RepeatForm,
} from './rrule'
import type {TaskRepeat} from '@/client/generated'

describe('repeat model core', () => {
	const originalTZ = process.env.TZ

	afterEach(() => {
		process.env.TZ = originalTZ
	})

	function repeat(fields: Record<string, unknown>): TaskRepeat {
		return fields as unknown as TaskRepeat
	}

	// Fri 2026-10-23 and Thu 2026-10-29 at noon UTC: the same calendar day in UTC and New York.
	const FRI_OCT_23 = new Date('2026-10-23T12:00:00Z')
	const THU_OCT_29 = new Date('2026-10-29T12:00:00Z')

	describe('classifyRepeat', () => {
		const editable: [string, TaskRepeat, Date | null][] = [
			['hourly interval 3', repeat({freq: 'hourly', interval: 3}), FRI_OCT_23],
			['daily', repeat({freq: 'daily', interval: 1}), FRI_OCT_23],
			['daily interval 2 count 5', repeat({freq: 'daily', interval: 2, count: 5}), FRI_OCT_23],
			['weekly implicit', repeat({freq: 'weekly', interval: 1}), FRI_OCT_23],
			['weekly mo/we/fr', repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'we', 'fr']}), FRI_OCT_23],
			['weekly interval 2 tu/th', repeat({freq: 'weekly', interval: 2, by_day: ['tu', 'th']}), FRI_OCT_23],
			['monthly implicit', repeat({freq: 'monthly', interval: 1}), FRI_OCT_23],
			['monthly on the 15th, due the 23rd', repeat({freq: 'monthly', interval: 1, by_month_day: [15]}), FRI_OCT_23],
			['monthly last day', repeat({freq: 'monthly', interval: 1, by_month_day: [-1]}), FRI_OCT_23],
			['monthly 4th friday', repeat({freq: 'monthly', interval: 1, by_day: ['4fr']}), FRI_OCT_23],
			['monthly last friday', repeat({freq: 'monthly', interval: 1, by_day: ['-1fr']}), FRI_OCT_23],
			['yearly', repeat({freq: 'yearly', interval: 1}), FRI_OCT_23],
			['weekly until', repeat({freq: 'weekly', interval: 1, until: '2027-06-30T03:59:59Z'}), FRI_OCT_23],
			['until on a leap day', repeat({freq: 'weekly', interval: 1, until: '2028-02-29T23:59:59Z'}), FRI_OCT_23],
			['until with an offset', repeat({freq: 'weekly', interval: 1, until: '2027-06-29T23:59:59.5-04:00'}), FRI_OCT_23],
			['explicit interval 1 and wkst mo', repeat({freq: 'weekly', interval: 1, wkst: 'mo'}), FRI_OCT_23],
			['no interval', repeat({freq: 'monthly', by_day: ['2tu']}), FRI_OCT_23],
			['no due date', repeat({freq: 'weekly', interval: 1, by_day: ['mo']}), null],
			['upper-case codes', repeat({freq: 'WEEKLY', interval: 1, by_day: ['FR', 'MO']}), FRI_OCT_23],
			['empty advanced lists', repeat({freq: 'daily', interval: 1, by_day: [], by_month: []}), FRI_OCT_23],
		]

		it.each(editable)('classifies %s as editable', (_name, r, due) => {
			expect(classifyRepeat(r, due)).toBe('editable')
		})

		it.each(editable)('round-trips %s losslessly in every zone', (_name, r, due) => {
			for (const zone of ['UTC', 'America/New_York', 'Pacific/Kiritimati', '']) {
				const form = repeatToForm(r, due, zone)
				expect(canonicalRepeat(formToRepeat(form, due, zone))).toEqual(canonicalRepeat(r))
			}
		})

		const foreign: [string, TaskRepeat][] = [
			['minutely', repeat({freq: 'minutely', interval: 1})],
			['secondly', repeat({freq: 'secondly', interval: 1})],
			['yearly in March and September', repeat({freq: 'yearly', interval: 1, by_month: [3, 9]})],
			['by_set_pos', repeat({freq: 'monthly', interval: 1, by_day: ['fr'], by_set_pos: [-1]})],
			['monthly 5th friday', repeat({freq: 'monthly', interval: 1, by_day: ['5fr']})],
			['monthly two weekdays', repeat({freq: 'monthly', interval: 1, by_day: ['2tu', '4tu']})],
			['monthly plain weekday', repeat({freq: 'monthly', interval: 1, by_day: ['fr']})],
			['monthly by_day and by_month_day', repeat({freq: 'monthly', interval: 1, by_day: ['2tu'], by_month_day: [15]})],
			['monthly two month days', repeat({freq: 'monthly', interval: 1, by_month_day: [1, 15]})],
			['monthly day 32', repeat({freq: 'monthly', interval: 1, by_month_day: [32]})],
			['monthly day 0', repeat({freq: 'monthly', interval: 1, by_month_day: [0]})],
			['monthly second-to-last day', repeat({freq: 'monthly', interval: 1, by_month_day: [-2]})],
			['weekly with ordinal', repeat({freq: 'weekly', interval: 1, by_day: ['2mo']})],
			['weekly duplicate days', repeat({freq: 'weekly', interval: 1, by_day: ['mo', 'mo']})],
			['weekly unknown day code', repeat({freq: 'weekly', interval: 1, by_day: ['xx']})],
			['weekly with by_month_day', repeat({freq: 'weekly', interval: 1, by_month_day: [15]})],
			['daily with by_day', repeat({freq: 'daily', interval: 1, by_day: ['mo']})],
			['hourly with by_day', repeat({freq: 'hourly', interval: 1, by_day: ['mo']})],
			['yearly with by_day', repeat({freq: 'yearly', interval: 1, by_day: ['1mo']})],
			['yearly with by_month_day', repeat({freq: 'yearly', interval: 1, by_month_day: [1]})],
			['count and until', repeat({freq: 'daily', interval: 1, count: 3, until: '2027-01-01T00:00:00Z'})],
			['count 0', repeat({freq: 'daily', interval: 1, count: 0})],
			['until not an instant', repeat({freq: 'daily', interval: 1, until: 'next tuesday'})],
			['until on February 30', repeat({freq: 'daily', interval: 1, until: '2026-02-30T12:00:00Z'})],
			['until on April 31', repeat({freq: 'daily', interval: 1, until: '2026-04-31T23:00:00Z'})],
			['until at hour 24', repeat({freq: 'daily', interval: 1, until: '2026-01-01T24:00:00Z'})],
			['until on February 29 of a common year', repeat({freq: 'daily', interval: 1, until: '2027-02-29T12:00:00Z'})],
			['until in month 13', repeat({freq: 'daily', interval: 1, until: '2026-13-01T12:00:00Z'})],
			['until at minute 60', repeat({freq: 'daily', interval: 1, until: '2026-01-01T12:60:00Z'})],
			['until at second 60', repeat({freq: 'daily', interval: 1, until: '2026-12-31T23:59:60Z'})],
			['until with an offset of 24 hours', repeat({freq: 'daily', interval: 1, until: '2026-01-01T12:00:00+24:00'})],
			['until without a zone', repeat({freq: 'daily', interval: 1, until: '2026-01-01T12:00:00'})],
			['interval 0', repeat({freq: 'daily', interval: 0})],
			['interval -1', repeat({freq: 'daily', interval: -1})],
			['wkst su', repeat({freq: 'weekly', interval: 2, wkst: 'su'})],
			['by_year_day', repeat({freq: 'yearly', interval: 1, by_year_day: [100]})],
			['by_week_no', repeat({freq: 'yearly', interval: 1, by_week_no: [1]})],
			['by_hour', repeat({freq: 'daily', interval: 1, by_hour: [9]})],
			['by_minute', repeat({freq: 'daily', interval: 1, by_minute: [30]})],
			['by_second', repeat({freq: 'daily', interval: 1, by_second: [0]})],
			['an unknown key', repeat({freq: 'daily', foo: 1})],
			['an unknown key without a value', repeat({freq: 'daily', foo: null})],
		]

		it.each(foreign)('classifies %s as foreign', (_name, r) => {
			expect(classifyRepeat(r, FRI_OCT_23)).toBe('foreign')
		})

		it('classifies a missing repeat or frequency as none', () => {
			expect(classifyRepeat(null, FRI_OCT_23)).toBe('none')
			expect(classifyRepeat(undefined, FRI_OCT_23)).toBe('none')
			expect(classifyRepeat(repeat({}), FRI_OCT_23)).toBe('none')
			expect(classifyRepeat(repeat({freq: ''}), FRI_OCT_23)).toBe('none')
		})

		it.each([
			['a weekly rule', repeat({freq: 'weekly', interval: 1, by_day: ['mo'], read_only: true})],
			['the migrated month-end rule', repeat({freq: 'monthly', interval: 1, by_month_day: [28, 29, 30, 31], by_set_pos: [-1], read_only: true})],
			['a rule with an unknown key', repeat({freq: 'daily', foo: 1, read_only: true})],
			['a projection without a frequency', repeat({read_only: true})],
		])('classifies %s the server marks read-only as locked', (_name, r) => {
			expect(classifyRepeat(r, FRI_OCT_23)).toBe('locked')
		})

		it.each(editable)('keeps %s editable when the server marks it not read-only', (_name, r, due) => {
			const fromServer = {...r, read_only: false}
			expect(classifyRepeat(fromServer, due)).toBe('editable')
			expect(canonicalRepeat(formToRepeat(repeatToForm(fromServer, due, 'UTC'), due, 'UTC'))).toEqual(canonicalRepeat(fromServer))
		})

		describe('rules the core port already guards', () => {
			it.each([
				['a single monthly day', repeat({freq: 'monthly', interval: 1, by_month_day: [15]})],
				['the 31st every two months', repeat({freq: 'monthly', interval: 2, by_month_day: [31]})],
				['the 1st', repeat({freq: 'monthly', interval: 1, by_month_day: [1]})],
			])('classifies %s as editable', (_name, r) => {
				expect(classifyRepeat(r, FRI_OCT_23)).toBe('editable')
			})

			it.each([
				['the migrated month-end rule', repeat({freq: 'monthly', interval: 1, by_month_day: [28, 29, 30, 31], by_set_pos: [-1]})],
				['monthly day 0', repeat({freq: 'monthly', interval: 1, by_month_day: [0]})],
				['a month day on a yearly rule', repeat({freq: 'yearly', interval: 1, by_month_day: [15]})],
				['a month day on a weekly rule', repeat({freq: 'weekly', interval: 1, by_month_day: [3]})],
				['a month day on a daily rule', repeat({freq: 'daily', interval: 1, by_month_day: [3]})],
				['two month days', repeat({freq: 'monthly', interval: 1, by_month_day: [1, 15]})],
				['minutely', repeat({freq: 'minutely', interval: 1})],
				['secondly', repeat({freq: 'secondly', interval: 1})],
			])('classifies %s as foreign', (_name, r) => {
				expect(classifyRepeat(r, FRI_OCT_23)).toBe('foreign')
			})

			it('classifies the migrated month-end rule as foreign, not locked, when the server allows replacing it', () => {
				const monthEnd = repeat({freq: 'monthly', interval: 1, by_month_day: [28, 29, 30, 31], by_set_pos: [-1], read_only: false})
				expect(classifyRepeat(monthEnd, FRI_OCT_23)).toBe('foreign')
			})
		})
	})

	describe('canonicalRepeat', () => {
		it('fills the backend defaults and drops empty values', () => {
			expect(canonicalRepeat(repeat({freq: 'DAILY', by_day: [], count: null, until: undefined}))).toEqual({
				freq: 'daily',
				interval: 1,
				wkst: 'mo',
			})
		})

		it('sorts weekly days Monday first and compares until as an instant', () => {
			expect(canonicalRepeat(repeat({freq: 'weekly', interval: 2, by_day: ['SU', 'fr', 'mo'], until: '2027-06-29T23:59:59-04:00'}))).toEqual({
				freq: 'weekly',
				interval: 2,
				wkst: 'mo',
				by_day: ['mo', 'fr', 'su'],
				until: Date.parse('2027-06-30T03:59:59Z'),
			})
		})

		it('keeps unknown keys', () => {
			expect(canonicalRepeat(repeat({freq: 'daily', interval: 1, foo: 1}))).toEqual({freq: 'daily', interval: 1, wkst: 'mo', foo: 1})
		})
	})

	describe('repeatToForm', () => {
		it('maps month selectors', () => {
			expect(repeatToForm(repeat({freq: 'monthly', interval: 1, by_day: ['2tu']}), FRI_OCT_23, 'UTC')).toMatchObject({
				unit: 'month', monthMode: 'nthWeekday', ordinal: 2, monthWeekday: 'tu', monthDay: null,
			})
			expect(repeatToForm(repeat({freq: 'monthly', interval: 1, by_month_day: [-1]}), FRI_OCT_23, 'UTC')).toMatchObject({
				monthMode: 'lastDay', monthDay: null,
			})
			expect(repeatToForm(repeat({freq: 'monthly', interval: 1, by_month_day: [15]}), FRI_OCT_23, 'UTC')).toMatchObject({
				monthMode: 'day', monthDay: 15,
			})
		})

		it('keeps implicit selectors implicit and takes defaults from the due date', () => {
			expect(repeatToForm(repeat({freq: 'weekly', interval: 2}), FRI_OCT_23, 'UTC')).toEqual({
				unit: 'week',
				interval: 2,
				weekdays: null,
				monthMode: 'day',
				monthDay: null,
				ordinal: 4,
				monthWeekday: 'fr',
				end: 'never',
				count: 1,
				untilDate: null,
				untilInstant: null,
			})
		})

		it('keeps 4th and last friday distinct', () => {
			const fourth = repeat({freq: 'monthly', interval: 1, by_day: ['4fr']})
			const last = repeat({freq: 'monthly', interval: 1, by_day: ['-1fr']})
			expect(formToRepeat(repeatToForm(fourth, FRI_OCT_23, 'UTC'), FRI_OCT_23, 'UTC').by_day).toEqual(['4fr'])
			expect(formToRepeat(repeatToForm(last, FRI_OCT_23, 'UTC'), FRI_OCT_23, 'UTC').by_day).toEqual(['-1fr'])
		})

		it('reads the end date in the recurrence zone', () => {
			const r = repeat({freq: 'weekly', interval: 1, until: '2027-06-30T03:59:59Z'})
			expect(repeatToForm(r, FRI_OCT_23, 'UTC')).toMatchObject({end: 'until', untilDate: '2027-06-30', untilInstant: '2027-06-30T03:59:59Z'})
			expect(repeatToForm(r, FRI_OCT_23, 'America/New_York')).toMatchObject({end: 'until', untilDate: '2027-06-29'})
		})

		it('reads a missing interval as 1', () => {
			expect(repeatToForm(repeat({freq: 'weekly'}), FRI_OCT_23, 'UTC').interval).toBe(1)
		})

		it('returns a draft that does not alias the stored repeat', () => {
			const r = repeat({freq: 'weekly', interval: 1, by_day: ['mo']})
			const form = repeatToForm(r, FRI_OCT_23, 'UTC')
			form.weekdays?.push('tu')
			expect(r.by_day).toEqual(['mo'])
		})

		it('refuses a repeat the editor cannot represent', () => {
			expect(() => repeatToForm(repeat({freq: 'monthly', interval: 1, by_day: ['5fr']}), FRI_OCT_23, 'UTC')).toThrow()
		})

		it('refuses a locked repeat', () => {
			expect(() => repeatToForm(repeat({freq: 'weekly', interval: 1, read_only: true}), FRI_OCT_23, 'UTC')).toThrow()
		})
	})

	describe('formToRepeat', () => {
		it('writes no defaults', () => {
			expect(formToRepeat(defaultCustomForm(FRI_OCT_23, 'UTC'), FRI_OCT_23, 'UTC')).toStrictEqual({freq: 'weekly'})
		})

		it('writes only the fields of the chosen shape', () => {
			const form: RepeatForm = {
				...defaultCustomForm(FRI_OCT_23, 'UTC'),
				unit: 'month',
				interval: 3,
				weekdays: ['mo'],
				monthMode: 'nthWeekday',
				ordinal: -1,
				monthWeekday: 'fr',
				end: 'count',
				count: 4,
				untilDate: '2027-01-01',
			}
			expect(formToRepeat(form, FRI_OCT_23, 'UTC')).toStrictEqual({freq: 'monthly', interval: 3, by_day: ['-1fr'], count: 4})
		})

		it('never writes an empty weekday list', () => {
			const form: RepeatForm = {...defaultCustomForm(FRI_OCT_23, 'UTC'), weekdays: []}
			expect(formToRepeat(form, FRI_OCT_23, 'UTC')).toStrictEqual({freq: 'weekly'})
		})

		it('sorts weekdays Monday first', () => {
			const form: RepeatForm = {...defaultCustomForm(FRI_OCT_23, 'UTC'), weekdays: ['fr', 'mo', 'we']}
			expect(formToRepeat(form, FRI_OCT_23, 'UTC').by_day).toEqual(['mo', 'we', 'fr'])
		})

		it('writes month day and last day', () => {
			const base: RepeatForm = {...defaultCustomForm(FRI_OCT_23, 'UTC'), unit: 'month'}
			expect(formToRepeat({...base, monthDay: 15}, FRI_OCT_23, 'UTC')).toStrictEqual({freq: 'monthly', by_month_day: [15]})
			expect(formToRepeat({...base, monthMode: 'lastDay'}, FRI_OCT_23, 'UTC')).toStrictEqual({freq: 'monthly', by_month_day: [-1]})
			expect(formToRepeat(base, FRI_OCT_23, 'UTC')).toStrictEqual({freq: 'monthly'})
		})
	})

	describe('end date (UNTIL)', () => {
		const stored = repeat({freq: 'weekly', interval: 1, until: '2027-06-30T03:59:59Z'})

		it('writes the stored instant verbatim while the date is unchanged', () => {
			for (const zone of ['UTC', 'America/New_York', 'Asia/Tokyo']) {
				const form = repeatToForm(stored, FRI_OCT_23, zone)
				expect(formToRepeat(form, FRI_OCT_23, zone).until).toBe('2027-06-30T03:59:59Z')
			}
		})

		it('writes a leap-day end instant back verbatim', () => {
			const leapDay = repeat({freq: 'weekly', interval: 1, until: '2028-02-29T23:59:59Z'})
			expect(formToRepeat(repeatToForm(leapDay, FRI_OCT_23, 'UTC'), FRI_OCT_23, 'UTC').until).toBe('2028-02-29T23:59:59Z')
		})

		it('writes the last second of a changed date in the recurrence zone', () => {
			const form = {...repeatToForm(stored, FRI_OCT_23, 'UTC'), untilDate: '2027-07-15'}
			expect(formToRepeat(form, FRI_OCT_23, 'UTC').until).toBe('2027-07-15T23:59:59.000Z')
		})

		it('uses the recurrence zone, not the browser zone, for a picked date', () => {
			process.env.TZ = 'UTC'
			const form: RepeatForm = {...defaultCustomForm(FRI_OCT_23, 'America/New_York'), end: 'until', untilDate: '2026-11-06'}
			expect(formToRepeat(form, FRI_OCT_23, 'America/New_York').until).toBe('2026-11-07T04:59:59.000Z')
		})

		it('falls back to the browser zone when the recurrence zone is empty', () => {
			process.env.TZ = 'America/Los_Angeles'
			const form: RepeatForm = {...defaultCustomForm(FRI_OCT_23, ''), end: 'until', untilDate: '2026-11-06'}
			expect(formToRepeat(form, FRI_OCT_23, '').until).toBe('2026-11-07T07:59:59.000Z')
		})
	})

	describe('recurrence zone', () => {
		it('uses a loadable timezone setting', () => {
			expect(recurrenceZone({timezone: 'Asia/Tokyo'})).toBe('Asia/Tokyo')
		})

		it('falls back to the browser zone when the setting is empty or unknown', () => {
			const browser = Intl.DateTimeFormat().resolvedOptions().timeZone
			expect(recurrenceZone({})).toBe(browser)
			expect(recurrenceZone({timezone: ''})).toBe(browser)
			expect(recurrenceZone({timezone: 'Not/AZone'})).toBe(browser)
			expect(recurrenceZone({timezone: null})).toBe(browser)
		})

		it('reads calendar facts in the zone', () => {
			// Sat 00:00 UTC is still Friday evening in New York.
			const instant = new Date('2026-10-24T00:00:00Z')
			expect(dayInZone(instant, 'America/New_York')).toEqual({year: 2026, month: 10, day: 23, weekday: 'fr', ordinal: 4})
			expect(dayInZone(instant, 'UTC')).toEqual({year: 2026, month: 10, day: 24, weekday: 'sa', ordinal: 4})
		})

		it.each([
			[1, 1], [7, 1], [8, 2], [21, 3], [22, 4], [28, 4], [29, -1], [31, -1],
		])('gives day %i of the month ordinal %i', (day, ordinal) => {
			expect(dayInZone(new Date(Date.UTC(2026, 9, day, 12)), 'UTC').ordinal).toBe(ordinal)
		})
	})

	describe('dueDateOf', () => {
		it.each([
			['no task', undefined],
			['a null due date', {due_date: null}],
			['a missing due date', {}],
			['an empty due date', {due_date: ''}],
			['the api zero date', {due_date: '0001-01-01T00:00:00Z'}],
			['an unparseable due date', {due_date: 'soon'}],
		])('reads %s as no due date', (_name, task) => {
			expect(dueDateOf(task)).toBeNull()
		})

		it('reads a due date as a date', () => {
			expect(dueDateOf({due_date: '2026-10-23T12:00:00Z'})).toEqual(FRI_OCT_23)
		})
	})

	describe('picker bridges', () => {
		it.each(['America/Los_Angeles', 'UTC', 'Pacific/Kiritimati'])('round-trips a calendar date under TZ=%s', tz => {
			process.env.TZ = tz
			const picked = untilDateToPickerDate('2026-11-06')
			expect(picked).toEqual(new Date(2026, 10, 6, 12))
			expect(pickerDateToUntilDate(picked as Date)).toBe('2026-11-06')
		})

		it('maps no date to no picker date', () => {
			expect(untilDateToPickerDate(null)).toBeNull()
		})
	})

	describe('presets', () => {
		it('derives presets from a due date on the 4th Friday', () => {
			expect(repeatPresets(FRI_OCT_23, 'UTC')).toStrictEqual([
				{id: 'none', labelKey: 'task.repeat.preset.none', labelParams: {}, repeat: null},
				{id: 'daily', labelKey: 'task.repeat.preset.daily', labelParams: {}, repeat: {freq: 'daily'}},
				{id: 'weekdays', labelKey: 'task.repeat.preset.weekdays', labelParams: {}, repeat: {freq: 'weekly', by_day: ['mo', 'tu', 'we', 'th', 'fr']}},
				{id: 'weeklyOnDay', labelKey: 'task.repeat.preset.weeklyOnDay', labelParams: {weekday: 'fr'}, repeat: {freq: 'weekly'}},
				{id: 'monthlyOnDate', labelKey: 'task.repeat.preset.monthlyOnDate', labelParams: {day: 23}, repeat: {freq: 'monthly'}},
				{id: 'monthlyOnNthWeekday', labelKey: 'task.repeat.preset.monthlyOnNthWeekday', labelParams: {ordinal: 4, weekday: 'fr'}, repeat: {freq: 'monthly', by_day: ['4fr']}},
				{id: 'yearlyOnDate', labelKey: 'task.repeat.preset.yearlyOnDate', labelParams: {month: 10, day: 23}, repeat: {freq: 'yearly'}},
			])
		})

		it('uses the last-weekday label on days 29 to 31', () => {
			const presets = repeatPresets(THU_OCT_29, 'UTC')
			expect(presets.find(p => p.id === 'monthlyOnNthWeekday')).toStrictEqual({
				id: 'monthlyOnNthWeekday',
				labelKey: 'task.repeat.preset.monthlyOnLastWeekday',
				labelParams: {ordinal: -1, weekday: 'th'},
				repeat: {freq: 'monthly', by_day: ['-1th']},
			})
			expect(presets.find(p => p.id === 'weeklyOnDay')?.labelParams).toEqual({weekday: 'th'})
			expect(presets.find(p => p.id === 'monthlyOnDate')?.labelParams).toEqual({day: 29})
			expect(presets.find(p => p.id === 'yearlyOnDate')?.labelParams).toEqual({month: 10, day: 29})
		})

		it('derives the weekday in the recurrence zone, not the browser zone', () => {
			process.env.TZ = 'UTC'
			// Fri 2026-10-23 20:00 in New York is Sat 00:00 UTC.
			const anchor = new Date('2026-10-24T00:00:00Z')
			const presets = repeatPresets(anchor, 'America/New_York')
			expect(presets.find(p => p.id === 'weeklyOnDay')?.labelParams).toEqual({weekday: 'fr'})
			expect(presets.find(p => p.id === 'monthlyOnNthWeekday')?.repeat?.by_day).toEqual(['4fr'])
		})

		it.each([FRI_OCT_23, THU_OCT_29])('presetIdFor recognises every preset for %s', anchor => {
			for (const preset of repeatPresets(anchor, 'UTC')) {
				expect(presetIdFor(preset.repeat, anchor, 'UTC')).toBe(preset.id)
			}
		})

		it('presetIdFor compares canonically', () => {
			const stored = repeat({freq: 'weekly', interval: 1, wkst: 'mo', by_day: ['fr', 'th', 'we', 'tu', 'mo']})
			expect(presetIdFor(stored, FRI_OCT_23, 'UTC')).toBe('weekdays')
		})

		it('presetIdFor does not mark a different nth weekday as the current preset', () => {
			const stored = repeat({freq: 'monthly', interval: 1, by_day: ['2tu']})
			expect(presetIdFor(stored, FRI_OCT_23, 'UTC')).toBeNull()
		})

		it('presetIdFor returns null for a non-preset repeat', () => {
			expect(presetIdFor(repeat({freq: 'weekly', interval: 2}), FRI_OCT_23, 'UTC')).toBeNull()
			expect(presetIdFor(repeat({freq: 'weekly', interval: 1, by_day: ['fr']}), FRI_OCT_23, 'UTC')).toBeNull()
		})

		it('defaults Custom to weekly on the due weekday', () => {
			expect(defaultCustomForm(THU_OCT_29, 'UTC')).toEqual({
				unit: 'week',
				interval: 1,
				weekdays: null,
				monthMode: 'day',
				monthDay: null,
				ordinal: -1,
				monthWeekday: 'th',
				end: 'never',
				count: 1,
				untilDate: null,
				untilInstant: null,
			})
		})
	})

	describe('validateRepeatForm', () => {
		const valid = defaultCustomForm(FRI_OCT_23, 'UTC')

		it('accepts the Custom default', () => {
			expect(validateRepeatForm(valid, FRI_OCT_23, 'UTC')).toEqual([])
		})

		it.each([0, -1, 1.5, Number.NaN])('rejects interval %s', interval => {
			expect(validateRepeatForm({...valid, interval}, FRI_OCT_23, 'UTC')).toEqual([{field: 'interval', key: 'task.repeat.error.interval'}])
		})

		it.each([0, 2.5])('rejects count %s when ending after a count', count => {
			expect(validateRepeatForm({...valid, end: 'count', count}, FRI_OCT_23, 'UTC')).toEqual([{field: 'count', key: 'task.repeat.error.count'}])
			expect(validateRepeatForm({...valid, end: 'never', count}, FRI_OCT_23, 'UTC')).toEqual([])
		})

		it('requires an end date when ending on a date', () => {
			expect(validateRepeatForm({...valid, end: 'until', untilDate: null}, FRI_OCT_23, 'UTC')).toEqual([{field: 'untilDate', key: 'task.repeat.error.untilMissing'}])
		})

		it('rejects an end date before the due date', () => {
			expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-10-22'}, FRI_OCT_23, 'UTC')).toEqual([{field: 'untilDate', key: 'task.repeat.error.untilBeforeDue'}])
			expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-10-23'}, FRI_OCT_23, 'UTC')).toEqual([])
		})

		describe('an end date on an undated task', () => {
			const past = {field: 'untilDate', key: 'task.repeat.error.untilPast'}
			afterEach(() => {
				vi.useRealTimers()
			})

			it('refuses a day before today and accepts today and later', () => {
				vi.useFakeTimers({toFake: ['Date']})
				vi.setSystemTime(new Date('2026-10-29T12:00:00Z'))
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-10-28'}, null, 'UTC')).toEqual([past])
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-10-29'}, null, 'UTC')).toEqual([])
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-11-01'}, null, 'UTC')).toEqual([])
			})

			it('takes today in the recurrence zone', () => {
				vi.useFakeTimers({toFake: ['Date']})
				// 02:00 UTC on 29 October is still 28 October in New York.
				vi.setSystemTime(new Date('2026-10-29T02:00:00Z'))
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-10-28'}, null, 'America/New_York')).toEqual([])
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-10-28'}, null, 'UTC')).toEqual([past])
			})

			it('accepts the stored end date unchanged, so an old rule can still be saved', () => {
				vi.useFakeTimers({toFake: ['Date']})
				vi.setSystemTime(new Date('2026-10-29T12:00:00Z'))
				const stored = {freq: 'daily', until: '2020-01-31T23:59:59Z'} as const
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2020-01-31'}, null, 'UTC', stored)).toEqual([])
				// The same stored instant is 1 February in Tokyo.
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2020-02-01'}, null, 'Asia/Tokyo', stored)).toEqual([])
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2020-01-30'}, null, 'UTC', stored)).toEqual([past])
				const offset = {freq: 'daily', until: '2020-01-31T23:59:59+02:00'} as const
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2020-01-31'}, null, 'UTC', offset)).toEqual([])
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2020-01-31'}, null, 'UTC')).toEqual([past])
			})

			it('leaves a dated task to the before-due check', () => {
				vi.useFakeTimers({toFake: ['Date']})
				vi.setSystemTime(new Date('2026-10-29T12:00:00Z'))
				expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-10-24'}, FRI_OCT_23, 'UTC')).toEqual([])
			})
		})

		it('compares the end date with the due date in the recurrence zone', () => {
			process.env.TZ = 'UTC'
			// Fri 2026-11-06 20:00 in New York is Sat 01:00 UTC.
			const due = new Date('2026-11-07T01:00:00Z')
			expect(validateRepeatForm({...valid, end: 'until', untilDate: '2026-11-06'}, due, 'America/New_York')).toEqual([])
		})

		it.each([0, 32, 1.5])('rejects month day %s', monthDay => {
			expect(validateRepeatForm({...valid, unit: 'month', monthDay}, FRI_OCT_23, 'UTC')).toEqual([{field: 'monthDay', key: 'task.repeat.error.monthDay'}])
		})

		it('accepts an implicit month day', () => {
			expect(validateRepeatForm({...valid, unit: 'month', monthDay: null}, FRI_OCT_23, 'UTC')).toEqual([])
		})

		it('does not treat an emptied weekday selection as an error', () => {
			expect(validateRepeatForm({...valid, weekdays: []}, FRI_OCT_23, 'UTC')).toEqual([])
		})

		it.each([
			['hour', 876_600],
			['day', 36_525],
			['week', 5_218],
			['month', 1_200],
			['year', 100],
		] as const)('accepts an interval of %s up to the server limit of %i', (unit, max) => {
			expect(validateRepeatForm({...valid, unit, interval: max}, FRI_OCT_23, 'UTC')).toEqual([])
			expect(validateRepeatForm({...valid, unit, interval: max + 1}, FRI_OCT_23, 'UTC'))
				.toEqual([{field: 'interval', key: 'task.repeat.error.intervalMax', params: {max}}])
		})

		it('accepts a count up to the server limit of 10,000', () => {
			expect(validateRepeatForm({...valid, end: 'count', count: 10_000}, FRI_OCT_23, 'UTC')).toEqual([])
			expect(validateRepeatForm({...valid, end: 'count', count: 10_001}, FRI_OCT_23, 'UTC'))
				.toEqual([{field: 'count', key: 'task.repeat.error.countMax', params: {max: 10_000}}])
			expect(validateRepeatForm({...valid, end: 'never', count: 10_001}, FRI_OCT_23, 'UTC')).toEqual([])
		})

		describe('raising only the count', () => {
			// As the server returns it: interval always present.
			const stored = repeat({freq: 'weekly', interval: 1, by_day: ['fr'], count: 3})
			const form = repeatToForm(stored, FRI_OCT_23, 'UTC')

			it('is refused, because the server would keep the stored count', () => {
				expect(validateRepeatForm({...form, count: 4}, FRI_OCT_23, 'UTC', stored))
					.toEqual([{field: 'count', key: 'task.repeat.error.countRaise'}])
				expect(validateRepeatForm({...form, count: 4}, FRI_OCT_23, 'UTC', {...stored, read_only: false}))
					.toEqual([{field: 'count', key: 'task.repeat.error.countRaise'}])
			})

			it.each([
				['an unchanged count', {count: 3}],
				['a lower count', {count: 2}],
				['removing the count', {end: 'never', count: 4}],
				['a higher count with another interval', {count: 4, interval: 2}],
				['a higher count with other weekdays', {count: 4, weekdays: ['mo', 'fr']}],
				['a higher count on another unit', {count: 4, unit: 'day'}],
			] as [string, Partial<RepeatForm>][])('allows %s', (_name, change) => {
				expect(validateRepeatForm({...form, ...change}, FRI_OCT_23, 'UTC', stored)).toEqual([])
			})

			it('allows adding a count to an unlimited rule', () => {
				const unlimited = repeat({freq: 'weekly', interval: 1, by_day: ['fr']})
				expect(validateRepeatForm({...repeatToForm(unlimited, FRI_OCT_23, 'UTC'), end: 'count', count: 5}, FRI_OCT_23, 'UTC', unlimited)).toEqual([])
			})

			it('is not checked without a stored repeat', () => {
				expect(validateRepeatForm({...form, count: 4}, FRI_OCT_23, 'UTC')).toEqual([])
				expect(validateRepeatForm({...form, count: 4}, FRI_OCT_23, 'UTC', null)).toEqual([])
			})

			it('reports the count limit rather than the raise above 10,000', () => {
				expect(validateRepeatForm({...form, count: 10_001}, FRI_OCT_23, 'UTC', stored))
					.toEqual([{field: 'count', key: 'task.repeat.error.countMax', params: {max: 10_000}}])
			})
		})

		it('normalizes an emptied weekday selection to the implicit due weekday', () => {
			const form: RepeatForm = {...valid, weekdays: []}
			expect(normalizeRepeatForm(form).weekdays).toBeNull()
			expect(form.weekdays).toEqual([])
			const kept = normalizeRepeatForm({...valid, weekdays: ['mo']})
			expect(kept.weekdays).toEqual(['mo'])
		})
	})
})

describe('frequency mapping', () => {
	it('maps editor units to API freqs', () => {
		expect(repeatFromSettings(1, 'years')?.freq).toBe('yearly')
		expect(repeatFromSettings(1, 'hours')?.freq).toBe('hourly')
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
