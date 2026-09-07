import {beforeEach, describe, expect, it} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'

import {useAuthStore} from '@/stores/auth'
import {useFlatpickrLanguage} from './useFlatpickrLanguage'

function setLanguageWeekStart(language: string, weekStart: number) {
	useAuthStore().setUserSettings({language, weekStart} as never)
}

describe('useFlatpickrLanguage', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('uses Saturday as the first day for fa-IR by default', () => {
		setLanguageWeekStart('fa-IR', 0)
		const language = useFlatpickrLanguage().value

		expect(language.firstDayOfWeek).toBe(6)
		expect(language.weekdays.shorthand).toContain('شنبه')
	})

	it('respects an explicit week start', () => {
		setLanguageWeekStart('fa-IR', 2)
		expect(useFlatpickrLanguage().value.firstDayOfWeek).toBe(2)
	})

	it('leaves english on Sunday by default', () => {
		setLanguageWeekStart('en', 0)
		expect(useFlatpickrLanguage().value.firstDayOfWeek).toBe(0)
	})

	it('leaves german on Monday by default', () => {
		setLanguageWeekStart('de-DE', 0)
		expect(useFlatpickrLanguage().value.firstDayOfWeek).toBe(1)
	})
})
