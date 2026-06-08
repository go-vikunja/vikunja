import {describe, it, expect} from 'vitest'

import {smartFillStart} from './smartFillStart'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

function entry(startTime: Date, endTime: Date | null): ITimeEntry {
	return {
		id: 1,
		userId: 1,
		taskId: 0,
		projectId: 0,
		startTime,
		endTime,
		comment: '',
		created: startTime,
		updated: startTime,
		maxPermission: null,
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

	it('falls back to the default start time on the current day when there are no entries', () => {
		expect(smartFillStart([], '08:15', now)).toEqual(new Date('2026-06-07T08:15:00'))
	})

	it('falls back to 09:00 when no default is configured', () => {
		expect(smartFillStart([], '', now)).toEqual(new Date('2026-06-07T09:00:00'))
	})
})
