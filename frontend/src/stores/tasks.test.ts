import {describe, expect, it} from 'vitest'
import {buildDefaultRemindersForQuickAdd} from './tasks'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

const aDefault: ITaskReminder = {
	reminder: null,
	relativePeriod: -3600,
	relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
} as ITaskReminder

describe('buildDefaultRemindersForQuickAdd', () => {
	it('returns empty array when due date is null', () => {
		expect(buildDefaultRemindersForQuickAdd([aDefault], null)).toEqual([])
	})

	it('returns empty array when defaults are undefined', () => {
		expect(buildDefaultRemindersForQuickAdd(undefined, '2026-05-01T00:00:00.000Z')).toEqual([])
	})

	it('returns empty array when defaults are empty', () => {
		expect(buildDefaultRemindersForQuickAdd([], '2026-05-01T00:00:00.000Z')).toEqual([])
	})

	it('clones defaults with relativeTo locked to due_date', () => {
		const result = buildDefaultRemindersForQuickAdd([aDefault], '2026-05-01T00:00:00.000Z')
		expect(result).toHaveLength(1)
		expect(result[0].relativePeriod).toBe(-3600)
		expect(result[0].relativeTo).toBe(REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE)
		expect(result[0].reminder).toBeNull()
	})

	it('does not share references with the input array', () => {
		const defaults = [aDefault]
		const result = buildDefaultRemindersForQuickAdd(defaults, '2026-05-01T00:00:00.000Z')
		expect(result[0]).not.toBe(defaults[0])
	})

	it('forces relativeTo to due_date even if a default somehow had another value', () => {
		const weird = {...aDefault, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE} as ITaskReminder
		const result = buildDefaultRemindersForQuickAdd([weird], '2026-05-01T00:00:00.000Z')
		expect(result[0].relativeTo).toBe(REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE)
	})
})
