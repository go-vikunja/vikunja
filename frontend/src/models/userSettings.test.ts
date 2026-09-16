import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {SupportedLocale} from '@/i18n'
import UserSettingsModel from './userSettings'

describe('UserSettingsModel', () => {
	beforeEach(() => {
		vi.stubGlobal('navigator', {language: 'de-DE'})
	})

	it('falls back to the browser language when the api returns an empty language', () => {
		// The api sends a plain string, which can be outside the SupportedLocale union
		const settings = new UserSettingsModel({language: '' as SupportedLocale})

		expect(settings.language).toBe('de-DE')
	})

	it('falls back to the browser language when the api returns null', () => {
		const settings = new UserSettingsModel({language: null})

		expect(settings.language).toBe('de-DE')
	})

	it('falls back to the browser language when no language is passed', () => {
		const settings = new UserSettingsModel({})

		expect(settings.language).toBe('de-DE')
	})

	it('keeps the language returned by the api', () => {
		const settings = new UserSettingsModel({language: 'fr-FR'})

		expect(settings.language).toBe('fr-FR')
	})
})
