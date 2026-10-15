import {describe, it, expect, vi, afterEach} from 'vitest'
import {ref} from 'vue'
import {mount} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'

import TimeControl from './TimeControl.vue'
import en from '@/i18n/lang/en.json'
import {TIME_FORMAT, type TimeFormat} from '@/constants/timeFormat'

const timeFormat = ref<TimeFormat>(TIME_FORMAT.HOURS_24)

vi.mock('@/composables/useTimeFormat', () => ({
	useTimeFormat: () => ({store: timeFormat}),
}))

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

function mountTimeControl(format: TimeFormat = TIME_FORMAT.HOURS_24) {
	timeFormat.value = format
	const wrapper = mount(TimeControl, {
		props: {
			hours: 8,
			minutes: 30,
			onUpdate: (value: {hours: number, minutes: number}) => wrapper.setProps(value),
		},
		attachTo: document.body,
		global: {
			plugins: [i18n],
			stubs: {Icon: true},
			directives: {tooltip: {}},
		},
	})
	return wrapper
}

afterEach(() => {
	document.body.innerHTML = ''
})

describe('TimeControl', () => {
	it('keeps the step buttons out of the tab order', () => {
		const wrapper = mountTimeControl()
		const steps = wrapper.findAll('.time-control__step')

		expect(steps).toHaveLength(4)
		for (const step of steps) {
			expect(step.attributes('tabindex')).toBe('-1')
			expect(step.attributes('aria-label')).toBeTruthy()
		}
	})
})
