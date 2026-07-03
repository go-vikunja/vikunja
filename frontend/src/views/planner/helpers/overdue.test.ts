import {describe, it, expect} from 'vitest'
import dayjs from 'dayjs'

import {isOverdue} from './overdue'
import type {ITask} from '@/modelTypes/ITask'

const today = dayjs('2026-07-02').startOf('day')

function makeTask(overrides: Partial<ITask>): ITask {
	return {
		id: 1,
		done: false,
		startDate: null,
		endDate: null,
		dueDate: null,
		...overrides,
	} as ITask
}

describe('isOverdue', () => {
	it('is false for a task without any dates', () => {
		expect(isOverdue(makeTask({}), today)).toBe(false)
	})

	it('is true for a block that ended yesterday', () => {
		const task = makeTask({
			startDate: new Date('2026-07-01T10:00:00'),
			endDate: new Date('2026-07-01T11:00:00'),
		})
		expect(isOverdue(task, today)).toBe(true)
	})

	it('is false for a block spanning midnight into today', () => {
		const task = makeTask({
			startDate: new Date('2026-07-01T23:00:00'),
			endDate: new Date('2026-07-02T01:00:00'),
		})
		expect(isOverdue(task, today)).toBe(false)
	})

	it('is true for a due-only task with a past due date', () => {
		expect(isOverdue(makeTask({dueDate: new Date('2026-06-30T09:00:00')}), today)).toBe(true)
	})

	it('is false for a due-only task due today', () => {
		expect(isOverdue(makeTask({dueDate: new Date('2026-07-02T09:00:00')}), today)).toBe(false)
	})

	it('is false for a past-due task that is rescheduled into the future', () => {
		const task = makeTask({
			dueDate: new Date('2026-06-30T09:00:00'),
			startDate: new Date('2026-07-03T10:00:00'),
			endDate: new Date('2026-07-03T11:00:00'),
		})
		expect(isOverdue(task, today)).toBe(false)
	})

	it('is true for a start-only task whose start day passed', () => {
		expect(isOverdue(makeTask({startDate: new Date('2026-06-29T10:00:00')}), today)).toBe(true)
	})

	it('is false for done tasks regardless of dates', () => {
		expect(isOverdue(makeTask({done: true, dueDate: new Date('2026-06-01T09:00:00')}), today)).toBe(false)
	})
})
