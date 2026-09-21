import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {SupportedLocale} from '@/i18n'
import {createUserSettingsDraft} from './userSettings'

describe('createUserSettingsDraft', () => {
	beforeEach(() => {
		vi.stubGlobal('navigator', {language: 'de-DE'})
	})

	it('falls back to the browser language when the api returns an empty language', () => {
		// The api sends a plain string, which can be outside the SupportedLocale union
		const settings = createUserSettingsDraft({language: '' as SupportedLocale})

		expect(settings.language).toBe('de-DE')
	})

	it('falls back to the browser language when the api returns null', () => {
		const settings = createUserSettingsDraft({language: null as never})

		expect(settings.language).toBe('de-DE')
	})

	it('falls back to the browser language when no language is passed', () => {
		const settings = createUserSettingsDraft({})

		expect(settings.language).toBe('de-DE')
	})

	it('keeps the language returned by the api', () => {
		const settings = createUserSettingsDraft({language: 'fr-FR'})

		expect(settings.language).toBe('fr-FR')
	})

	it('drops extra settings links whose url is not http(s)', () => {
		const settings = createUserSettingsDraft({
			extra_settings_links: {
				a: {
					text: 'x',
					url: 'javascript:alert(1)',
				},
				b: {
					text: 'y',
					url: '/settings',
				},
			},
		})

		expect(settings.extra_settings_links).toEqual({
			b: {
				text: 'y',
				url: '/settings',
			},
		})
	})
})
