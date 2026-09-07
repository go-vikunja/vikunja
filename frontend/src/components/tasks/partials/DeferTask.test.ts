import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {mount, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import DeferTask from './DeferTask.vue'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

const update = vi.fn(async (task: Record<string, unknown>) => task)

vi.mock('@/services/task', () => ({
	default: class {
		loading = false
		update = update
	},
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountDefer(locale: 'fa-IR' | 'en') {
	globalI18n.global.locale.value = locale
	return mount(DeferTask, {
		props: {modelValue: {id: 1, dueDate: new Date('2026-09-07T12:00:00Z')} as never},
		global: {
			plugins: [i18n],
			stubs: {
				XButton: true,
				'flat-pickr': true,
			},
		},
	})
}

describe('DeferTask Jalali picker', () => {
	let wrapper: VueWrapper | undefined

	beforeEach(() => {
		setActivePinia(createPinia())
		useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
		update.mockClear()
		vi.useFakeTimers()
	})

	afterEach(() => {
		vi.useRealTimers()
		wrapper?.unmount()
		wrapper = undefined
		globalI18n.global.locale.value = 'en'
	})

	it('renders the Jalali grid instead of flatpickr', () => {
		wrapper = mountDefer('fa-IR')
		expect(wrapper.find('.jalali-calendar').exists()).toBe(true)
		expect(wrapper.find('flat-pickr-stub').exists()).toBe(false)
	})

	it('picking a Jalali date persists the correct Gregorian instant', async () => {
		wrapper = mountDefer('fa-IR')

		await wrapper.findAll('.jalali-calendar__day')
			.find((button) => button.text() === '۲۰')!
			.trigger('click')
		wrapper.unmount()
		wrapper = undefined

		// Shahrivar 20 keeps the 15:30 Asia/Tehran wall-clock time.
		expect(update).toHaveBeenCalled()
		const saved = update.mock.calls[0][0] as {dueDate: Date}
		expect(saved.dueDate.toISOString()).toBe('2026-09-11T12:00:00.000Z')
	})
})

describe('DeferTask Gregorian picker', () => {
	let wrapper: VueWrapper | undefined

	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
		globalI18n.global.locale.value = 'en'
	})

	it('renders flatpickr instead of the Jalali grid', () => {
		wrapper = mountDefer('en')
		expect(wrapper.find('flat-pickr-stub').exists()).toBe(true)
		expect(wrapper.find('.jalali-calendar').exists()).toBe(false)
	})
})
