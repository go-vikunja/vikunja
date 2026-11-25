import {describe, expect, it} from 'vitest'
import {taskRemindersFromSettings} from './userSettings'

describe('taskRemindersFromSettings', () => {
	it('returns no reminders without settings', () => {
		expect(taskRemindersFromSettings(undefined)).toEqual([])
	})

	it('makes stored periods relative to the due date', () => {
		expect(taskRemindersFromSettings([{relative_period: -7200}])).toEqual([{
			relative_period: -7200,
			relative_to: 'due_date',
		}])
	})
})
