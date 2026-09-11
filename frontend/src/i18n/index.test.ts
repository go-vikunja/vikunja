import {afterEach, describe, expect, it} from 'vitest'

import {i18n, isRTLLanguage, setLanguage} from './index'

describe('i18n rtl + fallback', () => {
	afterEach(async () => {
		await setLanguage('en')
		document.documentElement.dir = 'ltr'
	})

	it('treats fa-IR as RTL like ar-SA and he-IL', () => {
		expect(isRTLLanguage('fa-IR')).toBe(true)
		expect(isRTLLanguage('ar-SA')).toBe(true)
		expect(isRTLLanguage('he-IL')).toBe(true)
	})

	it('treats en and de-DE as LTR', () => {
		expect(isRTLLanguage('en')).toBe(false)
		expect(isRTLLanguage('de-DE')).toBe(false)
	})

	it('falls back to English for untranslated keys', () => {
		expect(i18n.global.fallbackLocale.value).toBe('en')
	})

	it('sets document direction from the language', async () => {
		await setLanguage('fa-IR')
		expect(document.documentElement.lang).toBe('fa-IR')
		expect(document.documentElement.dir).toBe('rtl')

		await setLanguage('en')
		expect(document.documentElement.dir).toBe('ltr')
	})
})
