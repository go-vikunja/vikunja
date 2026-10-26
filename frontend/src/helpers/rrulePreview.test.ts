import {afterEach, describe, expect, it} from 'vitest'

import type {TaskRepeat} from '@/client/generated'
import {firstOccurrenceAfter, previewDates} from './rrulePreview'
import {dayInZone} from './rrule'
import {describeRepeatRule} from './rruleDescribe'

const NOW = new Date('2026-01-01T00:00:00Z')
const NEW_YORK = 'America/New_York'
const BERLIN = 'Europe/Berlin'

function repeat(fields: TaskRepeat & {freq: string}): TaskRepeat {
	return {interval: 1, ...fields}
}

function preview(r: TaskRepeat, due: string, zone: string | null): string[] | null {
	return previewDates(r, new Date(due), NOW, false, zone).dates?.map(d => d.toISOString()) ?? null
}

function median(values: number[]): number {
	const sorted = [...values].sort((a, b) => a - b)
	return sorted[Math.floor(sorted.length / 2)]!
}

function medianMs(run: () => unknown, times = 20): number {
	return median(Array.from({length: times}, () => {
		const started = performance.now()
		run()
		return performance.now() - started
	}))
}

const originalTZ = process.env.TZ
afterEach(() => {
	process.env.TZ = originalTZ
})

// Expected chains follow completion for a due date in the future: each completion anchors the rule
// on the due date's wall clock in the user's zone and takes the first occurrence strictly after it.
describe('previewDates sequences', () => {
	it('weekly on Mon, Wed, Fri from a Friday', () => {
		// Fri Oct 23 09:00 EDT -> Mon Oct 26 -> Wed Oct 28, same wall-clock time.
		expect(preview(repeat({freq: 'weekly', by_day: ['mo', 'we', 'fr']}), '2026-10-23T13:00:00Z', NEW_YORK)).toEqual([
			'2026-10-23T13:00:00.000Z',
			'2026-10-26T13:00:00.000Z',
			'2026-10-28T13:00:00.000Z',
		])
	})

	it('weekly every 2 weeks counts Monday-start weeks from the due date', () => {
		// Sun Nov 1 belongs to the week of Mon Oct 26, so the next eligible week starts Mon Nov 9.
		expect(preview(repeat({freq: 'weekly', interval: 2, by_day: ['mo', 'su']}), '2026-11-01T14:00:00Z', NEW_YORK)).toEqual([
			'2026-11-01T14:00:00.000Z',
			'2026-11-09T14:00:00.000Z',
			'2026-11-15T14:00:00.000Z',
		])
		// Fri Oct 30 is in the due date's week; the following week is skipped.
		expect(preview(repeat({freq: 'weekly', interval: 2, by_day: ['mo', 'we', 'fr']}), '2026-10-28T13:00:00Z', NEW_YORK)).toEqual([
			'2026-10-28T13:00:00.000Z',
			'2026-10-30T13:00:00.000Z',
			'2026-11-09T14:00:00.000Z',
		])
	})

	it('weekly on a weekday the due date is not on starts at the next matching day', () => {
		expect(preview(repeat({freq: 'weekly', by_day: ['th']}), '2026-10-23T13:00:00Z', NEW_YORK)).toEqual([
			'2026-10-23T13:00:00.000Z',
			'2026-10-29T13:00:00.000Z',
			'2026-11-05T14:00:00.000Z',
		])
	})

	it('monthly on the 31st skips months without one', () => {
		// Implicit day (DTSTART's day): Feb and Apr have no 31st, so they are skipped, not clamped.
		expect(preview(repeat({freq: 'monthly'}), '2027-01-31T09:00:00Z', BERLIN)).toEqual([
			'2027-01-31T09:00:00.000Z',
			'2027-03-31T08:00:00.000Z',
			'2027-05-31T08:00:00.000Z',
		])
		expect(preview(repeat({freq: 'monthly', by_month_day: [31]}), '2026-08-31T08:00:00Z', BERLIN)).toEqual([
			'2026-08-31T08:00:00.000Z',
			'2026-10-31T09:00:00.000Z',
			'2026-12-31T09:00:00.000Z',
		])
	})

	it('monthly on the last day', () => {
		expect(preview(repeat({freq: 'monthly', by_month_day: [-1]}), '2027-01-31T14:00:00Z', NEW_YORK)).toEqual([
			'2027-01-31T14:00:00.000Z',
			'2027-02-28T14:00:00.000Z',
			'2027-03-31T13:00:00.000Z',
		])
	})

	it('distinguishes the 4th Friday from the last Friday in a five-Friday month', () => {
		// January 2027 has Fridays on the 1st, 8th, 15th, 22nd and 29th.
		expect(preview(repeat({freq: 'monthly', by_day: ['4fr']}), '2026-12-25T14:00:00Z', NEW_YORK)).toEqual([
			'2026-12-25T14:00:00.000Z',
			'2027-01-22T14:00:00.000Z',
			'2027-02-26T14:00:00.000Z',
		])
		expect(preview(repeat({freq: 'monthly', by_day: ['-1fr']}), '2026-12-25T14:00:00Z', NEW_YORK)).toEqual([
			'2026-12-25T14:00:00.000Z',
			'2027-01-29T14:00:00.000Z',
			'2027-02-26T14:00:00.000Z',
		])
	})

	it('yearly on Feb 29 skips non-leap years', () => {
		expect(preview(repeat({freq: 'yearly'}), '2028-02-29T12:00:00Z', 'UTC')).toEqual([
			'2028-02-29T12:00:00.000Z',
			'2032-02-29T12:00:00.000Z',
			'2036-02-29T12:00:00.000Z',
		])
		// Every 3 years from 2028: 2031, 2034 and 2037 have no Feb 29.
		expect(preview(repeat({freq: 'yearly', interval: 3}), '2028-02-29T12:00:00Z', 'UTC')).toEqual([
			'2028-02-29T12:00:00.000Z',
			'2040-02-29T12:00:00.000Z',
			'2052-02-29T12:00:00.000Z',
		])
	})

	it('daily with an interval and an end date', () => {
		expect(preview(repeat({freq: 'daily', interval: 2, until: '2027-01-01T00:00:00Z'}), '2026-10-24T21:30:00Z', BERLIN)).toEqual([
			'2026-10-24T21:30:00.000Z',
			'2026-10-26T22:30:00.000Z',
			'2026-10-28T22:30:00.000Z',
		])
	})

	it('daily with a large interval keeps the local time', () => {
		expect(preview(repeat({freq: 'daily', interval: 36525}), '2026-10-23T13:00:00Z', NEW_YORK)).toEqual([
			'2026-10-23T13:00:00.000Z',
			'2126-10-24T13:00:00.000Z',
			'2226-10-25T13:00:00.000Z',
		])
	})

	it.each(['UTC', 'Asia/Tokyo', NEW_YORK])('uses the given zone, not the browser zone (TZ=%s)', tz => {
		process.env.TZ = tz
		// Mon Oct 26 20:00 EDT is already Tuesday in UTC; the rule must still pick Wed and Fri in New York.
		const dates = previewDates(repeat({freq: 'weekly', by_day: ['mo', 'we', 'fr']}), new Date('2026-10-27T00:00:00Z'), NOW, false, NEW_YORK).dates!
		expect(dates.map(d => d.toISOString())).toEqual([
			'2026-10-27T00:00:00.000Z',
			'2026-10-29T00:00:00.000Z',
			'2026-10-31T00:00:00.000Z',
		])
		expect(dates.map(d => dayInZone(d, NEW_YORK).weekday)).toEqual(['mo', 'we', 'fr'])
	})
})

// Calendar repeats keep the due date's local time across DST: a wall time in a gap moves forward by
// the gap, one that occurs twice takes the earlier instant, and the next date returns to the intended time.
// New York springs forward on Sun 2091-03-11 at 02:00 and falls back on Sun 2091-11-04 at 02:00;
// Berlin changes on Sun 2091-03-25 and Sun 2091-10-28.
describe('previewDates across DST changes', () => {
	it('daily 09:00 stays 09:00 on both sides of spring-forward', () => {
		expect(preview(repeat({freq: 'daily'}), '2091-03-10T14:00:00Z', NEW_YORK)).toEqual([
			'2091-03-10T14:00:00.000Z', // Sat 09:00 EST
			'2091-03-11T13:00:00.000Z', // Sun 09:00 EDT
			'2091-03-12T13:00:00.000Z', // Mon 09:00 EDT
		])
	})

	it.each([
		['plain daily', repeat({freq: 'daily'})],
		['daily with a count', repeat({freq: 'daily', count: 5})],
	])('%s at 02:30 moves forward by the gap, then returns to 02:30', (_name, r) => {
		expect(preview(r, '2091-03-10T07:30:00Z', NEW_YORK)).toEqual([
			'2091-03-10T07:30:00.000Z', // Sat 02:30 EST
			'2091-03-11T07:30:00.000Z', // Sun 03:30 EDT: 02:30 does not exist
			'2091-03-12T06:30:00.000Z', // Mon 02:30 EDT
		])
	})

	it('weekly on Sunday at 02:30 returns to 02:30 after the gap', () => {
		expect(preview(repeat({freq: 'weekly', by_day: ['su']}), '2091-03-04T07:30:00Z', NEW_YORK)).toEqual([
			'2091-03-04T07:30:00.000Z', // Sun 02:30 EST
			'2091-03-11T07:30:00.000Z', // Sun 03:30 EDT
			'2091-03-18T06:30:00.000Z', // Sun 02:30 EDT
		])
	})

	it.each([
		['plain daily', repeat({freq: 'daily'})],
		['daily with a count', repeat({freq: 'daily', count: 5})],
	])('%s at 01:30 takes the earlier instant on fall-back night', (_name, r) => {
		expect(preview(r, '2091-11-03T05:30:00Z', NEW_YORK)).toEqual([
			'2091-11-03T05:30:00.000Z', // Sat 01:30 EDT
			'2091-11-04T05:30:00.000Z', // Sun 01:30 EDT, not 01:30 EST
			'2091-11-05T06:30:00.000Z', // Mon 01:30 EST
		])
	})

	it('daily with a count takes the earlier instant of an ambiguous due time', () => {
		// Due at the first 02:30 (CEST) on 2026-10-25; the next is Monday 02:30 CET.
		expect(preview(repeat({freq: 'daily', count: 5}), '2026-10-25T00:30:00Z', BERLIN)).toEqual([
			'2026-10-25T00:30:00.000Z',
			'2026-10-26T01:30:00.000Z',
			'2026-10-27T01:30:00.000Z',
		])
	})

	it('plain daily 00:30 shows three consecutive dates across fall-back', () => {
		const dates = previewDates(repeat({freq: 'daily'}), new Date('2026-10-31T04:30:00Z'), NOW, false, NEW_YORK).dates!
		expect(dates.map(d => d.toISOString())).toEqual([
			'2026-10-31T04:30:00.000Z', // Sat 00:30 EDT
			'2026-11-01T04:30:00.000Z', // Sun 00:30 EDT
			'2026-11-02T05:30:00.000Z', // Mon 00:30 EST
		])
		expect(dates.map(d => dayInZone(d, NEW_YORK).day)).toEqual([31, 1, 2])
	})

	it('weekly 09:00 keeps 09:00 across a Berlin change', () => {
		expect(preview(repeat({freq: 'weekly'}), '2091-10-21T07:00:00Z', BERLIN)).toEqual([
			'2091-10-21T07:00:00.000Z', // Sun 09:00 CEST
			'2091-10-28T08:00:00.000Z', // Sun 09:00 CET
			'2091-11-04T08:00:00.000Z',
		])
	})

	it('monthly on the 2nd Tuesday keeps 09:00 across a Berlin change', () => {
		expect(preview(repeat({freq: 'monthly', by_day: ['2tu']}), '2091-02-13T08:00:00Z', BERLIN)).toEqual([
			'2091-02-13T08:00:00.000Z', // Tue 09:00 CET
			'2091-03-13T08:00:00.000Z', // Tue 09:00 CET
			'2091-04-10T07:00:00.000Z', // Tue 09:00 CEST
		])
	})

	it('a due date already moved by a gap anchors the preview at the moved time', () => {
		// Its intended 02:30 is known only to the server.
		expect(preview(repeat({freq: 'daily'}), '2091-03-11T07:30:00Z', NEW_YORK)).toEqual([
			'2091-03-11T07:30:00.000Z',
			'2091-03-12T07:30:00.000Z',
			'2091-03-13T07:30:00.000Z',
		])
	})
})

describe('previewDates near the scheduler limits', () => {
	it('ends the list at a date past the year 9999', () => {
		// 22:00 in New York on 9999-12-31 is already the year 10000 in UTC, which cannot be stored.
		expect(preview(repeat({freq: 'yearly'}), '9999-01-01T03:00:00Z', NEW_YORK)).toEqual(['9999-01-01T03:00:00.000Z'])
		expect(preview(repeat({freq: 'daily'}), '9999-12-31T12:00:00Z', 'UTC')).toEqual(['9999-12-31T12:00:00.000Z'])
		expect(preview(repeat({freq: 'daily'}), '9999-12-30T12:00:00Z', 'UTC')).toEqual([
			'9999-12-30T12:00:00.000Z',
			'9999-12-31T12:00:00.000Z',
		])
	})

	it('ends the list where rrule-go\'s search window (DTSTART + about 292 years) leaves no occurrence', () => {
		// 2500, 2600 and 2700 have no Feb 29; 2800 is 400 years away.
		expect(preview(repeat({freq: 'yearly', interval: 100}), '2400-02-29T12:00:00Z', 'UTC')).toEqual(['2400-02-29T12:00:00.000Z'])
		expect(preview(repeat({freq: 'yearly', interval: 100}), '2096-02-29T12:00:00Z', 'UTC')).toEqual([
			'2096-02-29T12:00:00.000Z',
			'2196-02-29T12:00:00.000Z',
			'2296-02-29T12:00:00.000Z',
		])
	})

	it('hides the preview for an interval beyond the 100-year horizon', () => {
		const due = new Date('2026-10-23T13:00:00Z')
		expect(previewDates(repeat({freq: 'yearly', interval: 101}), due, NOW, false, NEW_YORK)).toEqual({dates: null, hiddenReason: 'notEditable'})
		expect(previewDates(repeat({freq: 'daily', interval: 36526}), due, NOW, false, NEW_YORK)).toEqual({dates: null, hiddenReason: 'notEditable'})
		expect(previewDates(repeat({freq: 'yearly', interval: 100}), due, NOW, false, NEW_YORK).hiddenReason).toBeNull()
	})
})

describe('previewDates end conditions', () => {
	const due = '2026-10-23T13:00:00Z'
	const fridays = ['2026-10-23T13:00:00.000Z', '2026-10-30T13:00:00.000Z', '2026-11-06T14:00:00.000Z']

	it('count includes the current occurrence', () => {
		expect(preview(repeat({freq: 'weekly', count: 3}), due, NEW_YORK)).toEqual(fridays)
		expect(preview(repeat({freq: 'weekly', count: 2}), '2026-10-30T13:00:00Z', NEW_YORK)).toEqual(fridays.slice(1))
		// COUNT=1 is cleared on completion instead of rescheduled.
		expect(preview(repeat({freq: 'weekly', count: 1}), due, NEW_YORK)).toEqual(fridays.slice(0, 1))
	})

	it('until is inclusive of an end date picked as the last second of that day', () => {
		expect(preview(repeat({freq: 'weekly', until: '2026-11-07T04:59:59Z'}), due, NEW_YORK)).toEqual(fridays)
		expect(preview(repeat({freq: 'weekly', until: '2026-10-31T03:59:59Z'}), due, NEW_YORK)).toEqual(fridays.slice(0, 2))
	})

	it('until equal to an occurrence keeps it; one second earlier drops it', () => {
		expect(preview(repeat({freq: 'weekly', until: '2026-11-06T14:00:00Z'}), due, NEW_YORK)).toEqual(fridays)
		expect(preview(repeat({freq: 'weekly', until: '2026-11-06T13:59:59Z'}), due, NEW_YORK)).toEqual(fridays.slice(0, 2))
	})

	it('until before the due date leaves only the current due date', () => {
		expect(preview(repeat({freq: 'weekly', until: '2026-10-01T00:00:00Z'}), due, NEW_YORK)).toEqual(fridays.slice(0, 1))
	})
})

describe('previewDates visibility', () => {
	const due = new Date('2026-10-23T13:00:00Z')
	const plainDaily = repeat({freq: 'daily'})
	const weekdays = repeat({freq: 'weekly', by_day: ['mo', 'we', 'fr']})

	it.each([
		['notEditable', null, due, false, NEW_YORK],
		['notEditable', repeat({freq: 'yearly', by_month: [3, 9]}), due, false, NEW_YORK],
		['hourly', repeat({freq: 'hourly', interval: 2}), due, false, NEW_YORK],
		['fromCompletion', weekdays, due, true, NEW_YORK],
		['noDueDate', weekdays, null, false, NEW_YORK],
		['overdue', weekdays, new Date('2025-12-31T23:00:00Z'), false, NEW_YORK],
		['overdue', weekdays, NOW, false, NEW_YORK],
		['noZone', weekdays, due, false, null],
		['noZone', weekdays, due, false, ''],
		['noZone', weekdays, due, false, 'Not/AZone'],
		['noZone', plainDaily, due, false, null],
		['noZone', repeat({freq: 'weekly', interval: 2}), due, false, ''],
		['noZone', repeat({freq: 'weekly', count: 3}), due, false, null],
	] as const)('hides the preview: %s', (reason, r, dueDate, fromCompletion, zone) => {
		expect(previewDates(r, dueDate, NOW, fromCompletion, zone)).toEqual({dates: null, hiddenReason: reason})
	})

	it('reports the first failing condition', () => {
		expect(previewDates(repeat({freq: 'hourly'}), null, NOW, true, null).hiddenReason).toBe('hourly')
		expect(previewDates(weekdays, null, NOW, true, null).hiddenReason).toBe('fromCompletion')
		expect(previewDates(weekdays, new Date('2025-01-01T00:00:00Z'), NOW, false, null).hiddenReason).toBe('overdue')
	})

	it('returns a copy of the due date first', () => {
		const {dates} = previewDates(weekdays, due, NOW, false, NEW_YORK)
		expect(dates![0]).toEqual(due)
		expect(dates![0]).not.toBe(due)
	})
})

describe('timing', () => {
	const describeOptions = {t: (key: string) => key, locale: 'en', weekStart: 1, zone: NEW_YORK}

	it('previews and summarises each slow fixture in a median under 50 ms', () => {
		const fixtures: [TaskRepeat, string, string][] = [
			[repeat({freq: 'monthly', by_month_day: [31]}), '2026-08-31T08:00:00Z', BERLIN],
			[repeat({freq: 'monthly', interval: 12, by_month_day: [31]}), '2027-04-30T08:00:00Z', BERLIN],
			[repeat({freq: 'yearly', interval: 100}), '2400-02-29T12:00:00Z', 'UTC'],
			[repeat({freq: 'yearly', interval: 3}), '2028-02-29T12:00:00Z', NEW_YORK],
			[repeat({freq: 'weekly', by_day: ['mo', 'tu', 'we', 'th', 'fr', 'sa', 'su'], count: 10000}), '2091-03-10T07:30:00Z', NEW_YORK],
			[repeat({freq: 'daily', interval: 36525}), '2026-10-23T13:00:00Z', NEW_YORK],
		]
		const medians = fixtures.map(([r, due, zone]) => medianMs(() => {
			previewDates(r, new Date(due), NOW, false, zone)
			describeRepeatRule(r, {...describeOptions, zone, dueDate: new Date(due)})
		}))
		expect(Math.max(...medians)).toBeLessThan(50)
	})

	it('finds the first date for the worst creation searches in a median under 50 ms', () => {
		const anchor = new Date('2400-02-29T14:00:00Z')
		const after = new Date('2400-02-29T20:00:00Z')
		const searches: [TaskRepeat, Date, Date][] = [
			[repeat({freq: 'yearly', interval: 100}), anchor, after],
			[repeat({freq: 'yearly'}), anchor, after],
			[repeat({freq: 'monthly', interval: 12, by_month_day: [31]}), new Date('2026-04-30T13:00:00Z'), new Date('2026-04-30T20:00:00Z')],
		]
		const medians = searches.map(([r, a, b]) => medianMs(() => firstOccurrenceAfter(r, a, b, NEW_YORK)))
		expect(Math.max(...medians)).toBeLessThan(50)
	})
})

describe('firstOccurrenceAfter', () => {
	// Thu Oct 22 2026 in New York (EDT, UTC-4); the caller's anchor is "now" at 09:00 local.
	const nineAm = new Date('2026-10-22T13:00:00Z')
	const afternoon = new Date('2026-10-22T18:00:00Z')

	it('finds the next matching weekday from a Thursday', () => {
		expect(firstOccurrenceAfter(repeat({freq: 'weekly', by_day: ['mo']}), nineAm, afternoon, NEW_YORK)?.toISOString())
			.toBe('2026-10-26T13:00:00.000Z')
	})

	it('skips months without the chosen day', () => {
		const feb10 = new Date('2027-02-10T15:00:00Z')
		const noonFeb10 = new Date('2027-02-10T17:00:00Z')
		expect(firstOccurrenceAfter(repeat({freq: 'monthly', by_month_day: [31]}), noonFeb10, feb10, NEW_YORK)?.toISOString())
			.toBe('2027-03-31T16:00:00.000Z')
	})

	it('uses the anchor time of day, rolling to the next match when it has passed', () => {
		const lateEvening = new Date('2026-10-23T02:30:00Z')
		expect(firstOccurrenceAfter(repeat({freq: 'daily'}), nineAm, lateEvening, NEW_YORK)?.toISOString())
			.toBe('2026-10-23T13:00:00.000Z')
		// Implicit weekday is the anchor's (Thursday), so the next one is a week later.
		expect(firstOccurrenceAfter(repeat({freq: 'weekly'}), nineAm, new Date('2026-10-23T02:00:00Z'), NEW_YORK)?.toISOString())
			.toBe('2026-10-29T13:00:00.000Z')
		expect(firstOccurrenceAfter(repeat({freq: 'weekly', by_day: ['mo', 'we', 'fr']}), nineAm, lateEvening, NEW_YORK)?.toISOString())
			.toBe('2026-10-23T13:00:00.000Z')
	})

	it('returns the anchor itself when it is still ahead', () => {
		const early = new Date('2026-10-22T11:00:00Z')
		expect(firstOccurrenceAfter(repeat({freq: 'weekly'}), nineAm, early, NEW_YORK)?.toISOString())
			.toBe('2026-10-22T13:00:00.000Z')
	})

	it('moves a time in a DST gap forward and takes the earlier of a repeated time', () => {
		expect(firstOccurrenceAfter(repeat({freq: 'daily'}), new Date('2091-03-10T07:30:00Z'), new Date('2091-03-10T12:00:00Z'), NEW_YORK)?.toISOString())
			.toBe('2091-03-11T07:30:00.000Z')
		expect(firstOccurrenceAfter(repeat({freq: 'daily'}), new Date('2091-11-03T05:30:00Z'), new Date('2091-11-03T12:00:00Z'), NEW_YORK)?.toISOString())
			.toBe('2091-11-04T05:30:00.000Z')
	})

	it('steps hourly repeats from the anchor time', () => {
		const lateEvening = new Date('2026-10-23T02:30:00Z')
		expect(firstOccurrenceAfter(repeat({freq: 'hourly', interval: 4}), nineAm, lateEvening, NEW_YORK)?.toISOString())
			.toBe('2026-10-23T05:00:00.000Z')
	})

	it('ignores count for the first occurrence', () => {
		expect(firstOccurrenceAfter(repeat({freq: 'weekly', count: 1}), nineAm, afternoon, NEW_YORK)?.toISOString())
			.toBe('2026-10-29T13:00:00.000Z')
	})

	it('returns null when until leaves no occurrence', () => {
		expect(firstOccurrenceAfter(repeat({freq: 'weekly', until: '2026-10-01T03:59:59Z'}), nineAm, afternoon, NEW_YORK)).toBeNull()
		expect(firstOccurrenceAfter(repeat({freq: 'daily', until: '2026-10-23T03:59:59Z'}), nineAm, new Date('2026-10-23T02:30:00Z'), NEW_YORK)).toBeNull()
		expect(firstOccurrenceAfter(repeat({freq: 'hourly', until: '2026-10-22T20:00:00Z'}), nineAm, new Date('2026-10-22T20:30:00Z'), NEW_YORK)).toBeNull()
	})

	it('has no fixed year limit', () => {
		expect(firstOccurrenceAfter(repeat({freq: 'yearly', interval: 6}), nineAm, afternoon, NEW_YORK)?.toISOString())
			.toBe('2032-10-22T13:00:00.000Z')
		// Feb 29 09:00 EST: 2029 to 2031 have none.
		expect(firstOccurrenceAfter(repeat({freq: 'yearly'}), new Date('2028-02-29T14:00:00Z'), new Date('2028-02-29T18:00:00Z'), NEW_YORK)?.toISOString())
			.toBe('2032-02-29T14:00:00.000Z')
	})

	it('returns null beyond the scheduler\'s search window or the last storable year', () => {
		// The next Feb 29 a multiple of 100 years after 2400 is in 2800, beyond about 292 years.
		expect(firstOccurrenceAfter(repeat({freq: 'yearly', interval: 100}), new Date('2400-02-29T14:00:00Z'), new Date('2400-02-29T20:00:00Z'), NEW_YORK)).toBeNull()
		expect(firstOccurrenceAfter(repeat({freq: 'yearly'}), new Date('9998-12-31T12:00:00Z'), new Date('9999-06-01T00:00:00Z'), 'UTC')?.toISOString())
			.toBe('9999-12-31T12:00:00.000Z')
		expect(firstOccurrenceAfter(repeat({freq: 'yearly'}), new Date('9999-01-01T03:00:00Z'), new Date('9999-06-01T00:00:00Z'), NEW_YORK)).toBeNull()
	})

	it('returns null for a repeat the editor cannot represent', () => {
		expect(firstOccurrenceAfter(repeat({freq: 'yearly', by_month: [3]}), nineAm, nineAm, NEW_YORK)).toBeNull()
	})

	it.each([null, ''])('falls back to the browser zone when the zone is %j', zone => {
		process.env.TZ = 'Asia/Tokyo'
		// Thu Oct 22 09:00 JST; the next Monday or Wednesday at 09:00 JST is Mon Oct 26.
		const anchor = new Date('2026-10-22T00:00:00Z')
		expect(firstOccurrenceAfter(repeat({freq: 'weekly', by_day: ['mo', 'we']}), anchor, anchor, zone)?.toISOString())
			.toBe('2026-10-26T00:00:00.000Z')
	})
})
