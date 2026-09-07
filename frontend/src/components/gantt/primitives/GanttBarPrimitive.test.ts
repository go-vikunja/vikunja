import {afterEach, beforeEach, describe, expect, it} from 'vitest'
import {mount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import GanttBarPrimitive from './GanttBarPrimitive.vue'
import en from '@/i18n/lang/en.json'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {formatDate} from '@/helpers/time/formatDate'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function setLocaleAndTimezone(locale: string, timezone: string) {
	globalI18n.global.locale.value = locale as never
	useAuthStore().setUserSettings({timezone} as never)
}

const START = new Date('2026-09-07T12:00:00Z')
const END = new Date('2026-09-10T12:00:00Z')
const TIMELINE_START = new Date('2026-09-01T00:00:00Z')
const TIMELINE_END = new Date('2026-09-15T00:00:00Z')

function mountPrimitive() {
	return mount(GanttBarPrimitive, {
		props: {
			model: {id: '1', start: new Date(START), end: new Date(END), meta: {label: 'Test'}},
			timelineStart: TIMELINE_START,
			timelineEnd: TIMELINE_END,
		},
		global: {plugins: [i18n]},
	})
}

describe('GanttBarPrimitive Jalali aria', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	it('uses the central Jalali formatter in Asia/Tehran', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const wrapper = mountPrimitive()
		const text = wrapper.attributes('aria-valuetext') ?? ''

		expect(text).toBe(`${formatDate(START, 'LLL')} – ${formatDate(END, 'LLL')}`)
		expect(text).not.toBe(`${START.toLocaleString()} – ${END.toLocaleString()}`)
	})

	it('uses the central Jalali formatter in UTC', () => {
		setLocaleAndTimezone('fa-IR', 'UTC')
		const wrapper = mountPrimitive()
		const text = wrapper.attributes('aria-valuetext') ?? ''

		expect(text).toBe(`${formatDate(START, 'LLL')} – ${formatDate(END, 'LLL')}`)
	})

	it('keeps Gregorian aria byte-identical', () => {
		setLocaleAndTimezone('en', 'UTC')
		const wrapper = mountPrimitive()

		expect(wrapper.attributes('aria-valuetext')).toBe(`${START.toLocaleString()} – ${END.toLocaleString()}`)
	})

	it('keeps min/max/now as instants in both locales', () => {
		for (const [locale, tz] of [['en', 'UTC'], ['fa-IR', 'Asia/Tehran']] as const) {
			setLocaleAndTimezone(locale, tz)
			const wrapper = mountPrimitive()
			expect(wrapper.attributes('aria-valuemin')).toBe(String(TIMELINE_START.valueOf()))
			expect(wrapper.attributes('aria-valuemax')).toBe(String(TIMELINE_END.valueOf()))
			expect(wrapper.attributes('aria-valuenow')).toBe(String(START.valueOf()))
		}
	})
})
