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

function inputs(wrapper: ReturnType<typeof mountTimeControl>) {
	return {
		hours: wrapper.get<HTMLInputElement>('input[aria-label="Hours"]'),
		minutes: wrapper.get<HTMLInputElement>('input[aria-label="Minutes"]'),
	}
}

// setValue() also fires change, which would commit the hour without auto-advance.
async function type(input: ReturnType<typeof inputs>['hours'], value: string) {
	input.element.value = value
	await input.trigger('input')
}

function expectFullySelected(input: HTMLInputElement) {
	expect(document.activeElement).toBe(input)
	expect(input.selectionStart).toBe(0)
	expect(input.selectionEnd).toBe(input.value.length)
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

	it('selects the value on focus so typing replaces it', async () => {
		const {hours} = inputs(mountTimeControl())

		hours.element.focus()

		expectFullySelected(hours.element)
	})

	it.each([
		{format: TIME_FORMAT.HOURS_24, typed: ['1', '14'], hours: 14},
		{format: TIME_FORMAT.HOURS_24, typed: ['2', '23'], hours: 23},
		{format: TIME_FORMAT.HOURS_24, typed: ['0', '07'], hours: 7},
		{format: TIME_FORMAT.HOURS_24, typed: ['3'], hours: 3},
		{format: TIME_FORMAT.HOURS_24, typed: ['9'], hours: 9},
		{format: TIME_FORMAT.HOURS_12, typed: ['1', '11'], hours: 11},
		{format: TIME_FORMAT.HOURS_12, typed: ['0', '05'], hours: 5},
		{format: TIME_FORMAT.HOURS_12, typed: ['2'], hours: 2},
	])('moves to minutes once $typed is a complete hour in $format mode', async ({format, typed, hours}) => {
		const wrapper = mountTimeControl(format)
		const {hours: hoursInput, minutes} = inputs(wrapper)
		hoursInput.element.focus()

		for (const [i, value] of typed.entries()) {
			await type(hoursInput, value)
			if (i < typed.length - 1) {
				expect(document.activeElement).toBe(hoursInput.element)
				expect(wrapper.emitted('update')).toBeUndefined()
			}
		}

		expect(wrapper.emitted('update')).toEqual([[{hours, minutes: 30}]])
		expectFullySelected(minutes.element)
	})

	it('keeps the meridiem when auto-advancing in 12h mode', async () => {
		const wrapper = mountTimeControl(TIME_FORMAT.HOURS_12)
		await wrapper.setProps({hours: 20})
		const {hours} = inputs(wrapper)
		hours.element.focus()

		await type(hours, '9')

		expect(wrapper.emitted('update')).toEqual([[{hours: 21, minutes: 30}]])
	})

	it('confirms the typed time with Enter', async () => {
		const wrapper = mountTimeControl()
		const {hours, minutes} = inputs(wrapper)

		hours.element.value = '1'
		await hours.trigger('keydown', {key: 'Enter'})
		await vi.waitFor(() => expect(wrapper.emitted('confirm')).toHaveLength(1))
		expect(wrapper.emitted('update')).toEqual([[{hours: 1, minutes: 30}]])

		minutes.element.value = '45'
		await minutes.trigger('keydown', {key: 'Enter'})
		await vi.waitFor(() => expect(wrapper.emitted('confirm')).toHaveLength(2))
		expect(wrapper.emitted('update')?.at(-1)).toEqual([{hours: 1, minutes: 45}])
	})
})
