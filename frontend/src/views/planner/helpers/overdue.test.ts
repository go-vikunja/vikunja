import {describe, it, expect} from 'vitest'
import dayjs from 'dayjs'

import {isOverdue, overdueAnchor} from './overdue'
import type {TaskResponse} from '@/client/queries/tasks'

const today = dayjs('2026-07-02').startOf('day')

function makeTask(overrides: Partial<TaskResponse>): TaskResponse {
	return {
		id: 1,
		done: false,
		...overrides,
	} as TaskResponse
}

describe('isOverdue', () => {
	it('is false for a task without any dates', () => {
		expect(isOverdue(makeTask({}), today)).toBe(false)
	})

	it('is true for a block that ended yesterday', () => {
		const task = makeTask({
			start_date: new Date('2026-07-01T10:00:00').toISOString(),
			end_date: new Date('2026-07-01T11:00:00').toISOString(),
		})
		expect(isOverdue(task, today)).toBe(true)
	})

	it('is false for a block spanning midnight into today', () => {
		const task = makeTask({
			start_date: new Date('2026-07-01T23:00:00').toISOString(),
			end_date: new Date('2026-07-02T01:00:00').toISOString(),
		})
		expect(isOverdue(task, today)).toBe(false)
	})

	it('is true for a due-only task with a past due date', () => {
		expect(isOverdue(makeTask({due_date: new Date('2026-06-30T09:00:00').toISOString()}), today)).toBe(true)
	})

	it('is false for a due-only task due today', () => {
		expect(isOverdue(makeTask({due_date: new Date('2026-07-02T09:00:00').toISOString()}), today)).toBe(false)
	})

	it('is false for a past-due task that is rescheduled into the future', () => {
		const task = makeTask({
			due_date: new Date('2026-06-30T09:00:00').toISOString(),
			start_date: new Date('2026-07-03T10:00:00').toISOString(),
			end_date: new Date('2026-07-03T11:00:00').toISOString(),
		})
		expect(isOverdue(task, today)).toBe(false)
	})

	it('is true for a start-only task whose start day passed', () => {
		expect(isOverdue(makeTask({start_date: new Date('2026-06-29T10:00:00').toISOString()}), today)).toBe(true)
	})

	it('is false for done tasks regardless of dates', () => {
		expect(isOverdue(makeTask({done: true, due_date: new Date('2026-06-01T09:00:00').toISOString()}), today)).toBe(false)
	})
})

describe('overdueAnchor', () => {
	it('is null for a task without any dates', () => {
		expect(overdueAnchor(makeTask({}))).toBeNull()
	})

	it('uses the same precedence as isOverdue: end before due', () => {
		// The past end date is what makes this task overdue; the future due date
		// must not surface as the (red) overdue label.
		const endDate = new Date('2026-06-30T11:00:00')
		const task = makeTask({
			start_date: new Date('2026-06-30T10:00:00').toISOString(),
			end_date: endDate.toISOString(),
			due_date: new Date('2026-07-10T09:00:00').toISOString(),
		})
		expect(isOverdue(task, today)).toBe(true)
		expect(overdueAnchor(task)).toEqual(endDate)
	})

	it('falls back to the start date, then the due date', () => {
		const startDate = new Date('2026-06-29T10:00:00')
		expect(overdueAnchor(makeTask({start_date: startDate.toISOString()}))).toEqual(startDate)

		const dueDate = new Date('2026-06-30T09:00:00')
		expect(overdueAnchor(makeTask({due_date: dueDate.toISOString()}))).toEqual(dueDate)
	})
})
