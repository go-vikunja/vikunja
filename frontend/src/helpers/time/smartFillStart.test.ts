import {describe, it, expect} from 'vitest'

import {smartFillStart} from './smartFillStart'
import type {TimeEntry as ITimeEntry} from '@/client/generated'

function entry(startTime: Date, endTime: Date | null): ITimeEntry {
	return {
		id: 1,
		user_id: 1,
		task_id: 0,
		project_id: 0,
		start_time: startTime.toISOString(),
		end_time: endTime?.toISOString() ?? null,
		comment: '',
		created: startTime.toISOString(),
		updated: startTime.toISOString(),
	}
}

describe('smartFillStart', () => {
	const now = new Date('2026-06-07T15:30:00')

	it('continues from the latest entry end time', () => {
		const entries = [
			entry(new Date('2026-06-07T09:00:00'), new Date('2026-06-07T10:00:00')),
			entry(new Date('2026-06-07T11:00:00'), new Date('2026-06-07T12:30:00')),
		]
		expect(smartFillStart(entries, '09:00', now)).toEqual(new Date('2026-06-07T12:30:00'))
	})

	it('ignores still-running entries (no end) when picking the latest end', () => {
		const entries = [
			entry(new Date('2026-06-07T09:00:00'), new Date('2026-06-07T10:00:00')),
			entry(new Date('2026-06-07T13:00:00'), null),
		]
		expect(smartFillStart(entries, '09:00', now)).toEqual(new Date('2026-06-07T10:00:00'))
	})

	it('ignores entries whose end time is unparseable', () => {
		const entries = [
			{...entry(new Date('2026-06-07T09:00:00'), null), end_time: '0001-01-01T00:00:00Z'},
			{...entry(new Date('2026-06-07T11:00:00'), null), end_time: 'not a date'},
		]
		expect(smartFillStart(entries, '08:15', now)).toEqual(new Date('2026-06-07T08:15:00'))
	})

	it('falls back to the default start time on the current day when there are no entries', () => {
		expect(smartFillStart([], '08:15', now)).toEqual(new Date('2026-06-07T08:15:00'))
	})

	it('falls back to 09:00 when no default is configured', () => {
		expect(smartFillStart([], '', now)).toEqual(new Date('2026-06-07T09:00:00'))
	})

	it('caps the default start at now when it would be in the future (before 09:00)', () => {
		const beforeNine = new Date('2026-06-07T07:30:00')
		expect(smartFillStart([], '09:00', beforeNine)).toEqual(beforeNine)
	})

	it('caps a future last-entry end at now', () => {
		const entries = [entry(new Date('2026-06-07T16:00:00'), new Date('2026-06-07T17:00:00'))]
		expect(smartFillStart(entries, '09:00', now)).toEqual(now)
	})
})
