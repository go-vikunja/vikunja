import {describe, it, expect} from 'vitest'
import dayjs from 'dayjs'

import {allDayItemsByDay, monthItemsByDay, type AllDayItem} from './dayLayout'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import type {TaskResponse} from '@/client/queries/tasks'

// The api sends this year instead of `null` for "no date set".
const NO_DATE = '0001-01-01T00:00:00Z'

function makeTask(overrides: Partial<TaskResponse>): TaskResponse {
	return {
		id: 1,
		done: false,
		repeat_after: 0,
		repeat_mode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
		...overrides,
	} as TaskResponse
}

function daysFrom(start: string, count: number): Date[] {
	return Array.from({length: count}, (_, i) => dayjs(start).add(i, 'day').toDate())
}

// The day keys a task landed on, in grid order.
function daysWithTask(items: Map<string, AllDayItem[]>, taskId: number): string[] {
	return [...items.entries()]
		.filter(([, dayItems]) => dayItems.some(item => item.task.id === taskId))
		.map(([key]) => key)
}

function itemsOn(items: Map<string, AllDayItem[]>, key: string): AllDayItem[] {
	return items.get(key) ?? []
}

describe('monthItemsByDay', () => {
	// A month grid runs over whole weeks, so it always carries days from the
	// neighbouring months. Oct 26 - Nov 8 2026 straddles the November boundary.
	const days = daysFrom('2026-10-26', 14)

	it('seeds every visible day, including the ones with nothing on them', () => {
		const items = monthItemsByDay([], days)
		const keys = [...items.keys()]

		expect(items.size).toBe(14)
		expect(keys[0]).toBe('2026-10-26')
		expect(keys[keys.length - 1]).toBe('2026-11-08')
		expect([...items.values()].every(dayItems => dayItems.length === 0)).toBe(true)
	})

	it('returns an empty map when there are no days to lay out', () => {
		const task = makeTask({due_date: new Date('2026-11-05T12:00:00').toISOString()})

		expect(monthItemsByDay([task], []).size).toBe(0)
	})

	it('spreads a timed task across every day it covers, over a month boundary', () => {
		const task = makeTask({
			start_date: new Date('2026-10-31T14:00:00').toISOString(),
			end_date: new Date('2026-11-02T09:00:00').toISOString(),
		})

		const items = monthItemsByDay([task], days)

		expect(daysWithTask(items, 1)).toEqual(['2026-10-31', '2026-11-01', '2026-11-02'])
		expect(itemsOn(items, '2026-11-01')[0].isGhost).toBe(false)
	})

	// The month grid has no time axis, so it spreads timed tasks that the week
	// grid leaves to the columns. This is the whole difference between the two.
	it('spreads a timed task the all-day row drops', () => {
		const task = makeTask({
			start_date: new Date('2026-10-31T14:00:00').toISOString(),
			end_date: new Date('2026-11-02T09:00:00').toISOString(),
		})

		expect(daysWithTask(monthItemsByDay([task], days), 1)).toHaveLength(3)
		expect(daysWithTask(allDayItemsByDay([task], days), 1)).toHaveLength(0)
	})

	it('clips a span that starts before the visible range', () => {
		const task = makeTask({
			start_date: new Date('2026-10-20T09:00:00').toISOString(),
			end_date: new Date('2026-10-28T17:00:00').toISOString(),
		})

		expect(daysWithTask(monthItemsByDay([task], days), 1))
			.toEqual(['2026-10-26', '2026-10-27', '2026-10-28'])
	})

	it('drops a task that falls entirely outside the visible range', () => {
		const task = makeTask({
			start_date: new Date('2026-09-01T10:00:00').toISOString(),
			end_date: new Date('2026-09-01T11:00:00').toISOString(),
		})

		expect(daysWithTask(monthItemsByDay([task], days), 1)).toEqual([])
	})

	it('anchors a due-only task to its due date', () => {
		const task = makeTask({due_date: new Date('2026-11-05T12:00:00').toISOString()})

		expect(daysWithTask(monthItemsByDay([task], days), 1)).toEqual(['2026-11-05'])
	})

	it('anchors a start-only task to its start date', () => {
		const task = makeTask({start_date: new Date('2026-11-03T08:00:00').toISOString()})

		expect(daysWithTask(monthItemsByDay([task], days), 1)).toEqual(['2026-11-03'])
	})

	it('anchors an end-only task to its end date', () => {
		const task = makeTask({end_date: new Date('2026-11-04T17:00:00').toISOString()})

		expect(daysWithTask(monthItemsByDay([task], days), 1)).toEqual(['2026-11-04'])
	})

	it('ignores a task with no dates at all', () => {
		expect(daysWithTask(monthItemsByDay([makeTask({})], days), 1)).toEqual([])
	})

	it('treats the api zero date as no date rather than a span', () => {
		const task = makeTask({
			start_date: NO_DATE,
			end_date: NO_DATE,
			due_date: new Date('2026-11-06T09:00:00').toISOString(),
		})

		expect(daysWithTask(monthItemsByDay([task], days), 1)).toEqual(['2026-11-06'])
	})

	it('keeps every task that lands on the same day', () => {
		const tasks = [
			makeTask({id: 1, due_date: new Date('2026-11-05T09:00:00').toISOString()}),
			makeTask({id: 2, due_date: new Date('2026-11-05T15:00:00').toISOString()}),
		]

		expect(itemsOn(monthItemsByDay(tasks, days), '2026-11-05')).toHaveLength(2)
	})

	it('marks projected occurrences of a recurring task as ghosts', () => {
		const task = makeTask({
			start_date: new Date('2026-10-27T09:00:00').toISOString(),
			end_date: new Date('2026-10-27T10:00:00').toISOString(),
			repeat_after: 604800, // 1 week
		})

		const items = monthItemsByDay([task], days)

		expect(daysWithTask(items, 1)).toEqual(['2026-10-27', '2026-11-03'])
		expect(itemsOn(items, '2026-10-27')[0].isGhost).toBe(false)
		expect(itemsOn(items, '2026-11-03')[0].isGhost).toBe(true)
	})
})
