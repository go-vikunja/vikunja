import {describe, expect, it, beforeEach, afterEach} from 'vitest'
import {mount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import DatemathHelp from './DatemathHelp.vue'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'
import faIR from '@/i18n/lang/fa-IR.json'

function makeI18n(locale: string) {
	return createI18n({legacy: false, locale, fallbackLocale: 'en', messages: {en, 'fa-IR': faIR}})
}

function mountHelp(locale = 'en') {
	return mount(DatemathHelp, {
		global: {
			plugins: [makeI18n(locale)],
			stubs: {BaseButton: true},
		},
	})
}

describe('DatemathHelp fa-only note', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en' as never
	})

	it('shows Saturday/Jalali-month note in fa mode', () => {
		globalI18n.global.locale.value = 'fa-IR' as never
		useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
		const wrapper = mountHelp('fa-IR')
		expect(wrapper.text()).toContain('شنبه')
		expect(wrapper.text()).toContain('جلالی')
	})

	it.each([['en'], ['de']])('hides fa note in %s (byte-identical)', (locale) => {
		globalI18n.global.locale.value = locale as never
		useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
		const wrapper = mountHelp(locale)
		expect(wrapper.text()).not.toContain('شنبه')
		expect(wrapper.text()).not.toContain('جلالی')
	})

	it('keeps grammar examples unchanged in fa mode', () => {
		globalI18n.global.locale.value = 'fa-IR' as never
		useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
		const wrapper = mountHelp('fa-IR')
		expect(wrapper.text()).toContain('now/w')
		expect(wrapper.text()).toContain('now/w+1w')
	})
})
