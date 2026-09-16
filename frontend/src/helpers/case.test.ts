import {describe, expect, it} from 'vitest'

import {objectToCamelCase, objectToSnakeCase} from './case'

describe('objectToCamelCase', () => {
	it('converts keys to camel case', () => {
		expect(objectToCamelCase({due_date: 1, nested_thing: {other_key: 2}})).toEqual({
			dueDate: 1,
			nestedThing: {otherKey: 2},
		})
	})

	it('keeps date instances intact', () => {
		const date = new Date('2021-02-06T12:00:00Z')
		expect(objectToCamelCase({created_at: date}).createdAt).toBe(date)
	})

	it('keeps date instances in arrays intact', () => {
		const date = new Date('2021-02-06T12:00:00Z')
		expect(objectToCamelCase({reminders: [{reminder_date: date}]}).reminders[0].reminderDate).toBe(date)
	})
})

describe('objectToSnakeCase', () => {
	it('keeps date instances intact', () => {
		const date = new Date('2021-02-06T12:00:00Z')
		expect(objectToSnakeCase({createdAt: date}).created_at).toBe(date)
	})
})
