import {beforeEach, describe, expect, it, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {getDefaultDueDateForDay, getDefaultDueTimeParts, parseUserDefaultDueTime} from './getDefaultDueDateForDay'
import {useAuthStore} from '@/stores/auth'

function setDefaultDueTime(defaultDueTime?: string) {
	const authStore = useAuthStore()
	authStore.setUserSettings({
		...authStore.settings,
		frontendSettings: {
			...authStore.settings.frontendSettings,
			defaultDueTime,
		},
	})
}

describe('getDefaultDueDateForDay', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		useAuthStore()
	})

	it('uses the user default due time when configured', () => {
		setDefaultDueTime('14:30')

		const date = new Date(2026, 7, 4, 9, 15)
		const result = getDefaultDueDateForDay(date)

		expect(result.getFullYear()).toBe(2026)
		expect(result.getMonth()).toBe(7)
		expect(result.getDate()).toBe(4)
		expect(result.getHours()).toBe(14)
		expect(result.getMinutes()).toBe(30)
	})

	it('falls back to the current nearest-hours default when the setting is empty', () => {
		setDefaultDueTime(undefined)

		const date = new Date(2026, 7, 4, 10, 15)
		const result = getDefaultDueDateForDay(date)

		expect(result.getHours()).toBe(12)
		expect(result.getMinutes()).toBe(0)
	})

	it('falls back when the stored time cannot be parsed', () => {
		setDefaultDueTime('25:99')

		const date = new Date(2026, 7, 4, 10, 15)
		const result = getDefaultDueDateForDay(date)

		expect(result.getHours()).toBe(12)
		expect(result.getMinutes()).toBe(0)
	})
})

describe('parseUserDefaultDueTime', () => {
	it('returns fallback parts when no user value is configured', () => {
		setDefaultDueTime(undefined)

		expect(getDefaultDueTimeParts(new Date(2026, 7, 4, 10, 15))).toStrictEqual({hours: 12, minutes: 0})
	})

	it('parses valid times', () => {
		expect(parseUserDefaultDueTime('09:45')).toStrictEqual({hours: 9, minutes: 45})
	})

	it('rejects invalid values', () => {
		expect(parseUserDefaultDueTime(undefined)).toBeNull()
		expect(parseUserDefaultDueTime('9:45')).toBeNull()
		expect(parseUserDefaultDueTime('24:00')).toBeNull()
		expect(parseUserDefaultDueTime('12:60')).toBeNull()
	})
})