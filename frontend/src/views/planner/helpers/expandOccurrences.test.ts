import {describe, it, expect} from 'vitest'
import {expandOccurrences} from './expandOccurrences'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import type {ITask} from '@/modelTypes/ITask'

function makeTask(overrides: Partial<ITask>): ITask {
	return {
		id: 1,
		startDate: null,
		endDate: null,
		repeatAfter: 0,
		repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
		...overrides,
	} as ITask
}

describe('expandOccurrences', () => {
	it('returns nothing for a task without start or end', () => {
		const task = makeTask({startDate: new Date('2026-06-22T10:00:00')})
		expect(expandOccurrences(task, new Date('2026-06-22'), new Date('2026-06-29'))).toHaveLength(0)
	})

	it('returns a single non-recurring instance, not a ghost', () => {
		const task = makeTask({
			startDate: new Date('2026-06-23T10:00:00'),
			endDate: new Date('2026-06-23T11:00:00'),
		})
		const out = expandOccurrences(task, new Date('2026-06-22T00:00:00'), new Date('2026-06-29T00:00:00'))
		expect(out).toHaveLength(1)
		expect(out[0].isGhost).toBe(false)
	})

	it('skips an instance entirely outside the range', () => {
		const task = makeTask({
			startDate: new Date('2026-01-01T10:00:00'),
			endDate: new Date('2026-01-01T11:00:00'),
		})
		const out = expandOccurrences(task, new Date('2026-06-22T00:00:00'), new Date('2026-06-29T00:00:00'))
		expect(out).toHaveLength(0)
	})

	it('projects weekly ghosts across a month, only the first is real', () => {
		const task = makeTask({
			startDate: new Date('2026-06-01T09:00:00'),
			endDate: new Date('2026-06-01T10:00:00'),
			repeatAfter: {type: 'weeks', amount: 1},
		})
		const out = expandOccurrences(task, new Date('2026-06-01T00:00:00'), new Date('2026-06-29T00:00:00'))
		// Jun 1, 8, 15, 22 (29 is excluded by the open upper bound)
		expect(out.map(o => o.start.getDate())).toEqual([1, 8, 15, 22])
		expect(out[0].isGhost).toBe(false)
		expect(out.slice(1).every(o => o.isGhost)).toBe(true)
	})

	it('preserves duration on ghost occurrences', () => {
		const task = makeTask({
			startDate: new Date('2026-06-01T09:00:00'),
			endDate: new Date('2026-06-01T10:30:00'),
			repeatAfter: {type: 'days', amount: 1},
		})
		const out = expandOccurrences(task, new Date('2026-06-01T00:00:00'), new Date('2026-06-04T00:00:00'))
		out.forEach(o => expect(o.end.getTime() - o.start.getTime()).toBe(90 * 60 * 1000))
	})

	it('projects into a window far past the cap from a long-untouched daily task', () => {
		// Stored start is well over a year before the window; stepping naively from
		// the start would exhaust the iteration cap before reaching it.
		const task = makeTask({
			startDate: new Date('2024-01-01T09:00:00'),
			endDate: new Date('2024-01-01T10:00:00'),
			repeatAfter: {type: 'days', amount: 1},
		})
		const out = expandOccurrences(task, new Date('2026-06-22T00:00:00'), new Date('2026-06-24T00:00:00'))
		expect(out.map(o => o.start.getDate())).toEqual([22, 23])
		expect(out.every(o => o.isGhost)).toBe(true)
	})

	it('honours monthly repeat mode regardless of repeatAfter', () => {
		const task = makeTask({
			startDate: new Date('2026-01-15T09:00:00'),
			endDate: new Date('2026-01-15T10:00:00'),
			repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_MONTH,
			repeatAfter: 0,
		})
		const out = expandOccurrences(task, new Date('2026-01-01T00:00:00'), new Date('2026-04-01T00:00:00'))
		expect(out.map(o => o.start.getMonth())).toEqual([0, 1, 2])
	})

	it('projects a monthly task with a far-past start into the window', () => {
		// A ms-based coarse jump computed from the first (short February) interval
		// used to overshoot the window entirely and drop the occurrence.
		const task = makeTask({
			startDate: new Date('2020-02-15T09:00:00'),
			endDate: new Date('2020-02-15T10:00:00'),
			repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_MONTH,
			repeatAfter: 0,
		})
		const out = expandOccurrences(task, new Date('2026-06-01T00:00:00'), new Date('2026-07-01T00:00:00'))
		expect(out).toHaveLength(1)
		expect(out[0].start.getDate()).toBe(15)
		expect(out[0].isGhost).toBe(true)
	})

	it('overflows month-end dates like the backend instead of clamping', () => {
		// Backend addOneMonthToDate: Jan 31 + 1 month = Mar 3 (2026 is not a leap
		// year), never Feb 28.
		const task = makeTask({
			startDate: new Date('2026-01-31T09:00:00'),
			endDate: new Date('2026-01-31T10:00:00'),
			repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_MONTH,
			repeatAfter: 0,
		})
		const out = expandOccurrences(task, new Date('2026-02-01T00:00:00'), new Date('2026-03-10T00:00:00'))
		expect(out).toHaveLength(1)
		expect(out[0].start.getMonth()).toBe(2)
		expect(out[0].start.getDate()).toBe(3)
	})

	it('does not project ghosts for from-current-date repeats', () => {
		// The next occurrence depends on when the user completes the task, so any
		// projection would be fiction; only the real stored instance shows.
		const task = makeTask({
			startDate: new Date('2026-06-23T10:00:00'),
			endDate: new Date('2026-06-23T11:00:00'),
			repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE,
			repeatAfter: {type: 'days', amount: 1},
		})
		const out = expandOccurrences(task, new Date('2026-06-22T00:00:00'), new Date('2026-06-29T00:00:00'))
		expect(out).toHaveLength(1)
		expect(out[0].isGhost).toBe(false)
	})
})
