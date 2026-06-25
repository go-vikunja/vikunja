import {describe, it, expect} from 'vitest'

import {parseTimeEntry} from './timeEntry'

describe('parseTimeEntry', () => {
	it('maps snake_case keys and coerces dates', () => {
		const e = parseTimeEntry({
			id: 1,
			user_id: 2,
			task_id: 3,
			project_id: 0,
			start_time: '2020-01-01T09:00:00Z',
			end_time: '2020-01-01T10:00:00Z',
			comment: 'work',
		})
		expect(e.userId).toBe(2)
		expect(e.taskId).toBe(3)
		expect(e.comment).toBe('work')
		expect(e.startTime).toBeInstanceOf(Date)
		expect(e.endTime).toBeInstanceOf(Date)
	})

	it('treats a null end time as a running timer', () => {
		const e = parseTimeEntry({
			id: 1,
			user_id: 1,
			task_id: 1,
			start_time: '2020-01-01T09:00:00Z',
			end_time: null,
		})
		expect(e.endTime).toBeNull()
	})
})
