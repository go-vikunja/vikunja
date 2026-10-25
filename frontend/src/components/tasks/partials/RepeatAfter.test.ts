import {describe, it, expect, vi, beforeEach} from 'vitest'
import {mount} from '@vue/test-utils'

import type {Task, TaskRepeat} from '@/client/generated'
import XButton from '@/components/input/Button.vue'

const message = vi.hoisted(() => ({
	error: vi.fn(),
	success: vi.fn(),
}))

vi.mock('@/message', () => message)
vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t: (key: string) => key}),
}))

import RepeatAfter from './RepeatAfter.vue'

function taskWith(repeat?: TaskRepeat, repeatsFromCurrentDate = false): Task {
	return {
		id: 7,
		project_id: 1,
		title: 'Water the plants',
		repeat,
		repeats_from_current_date: repeatsFromCurrentDate,
	}
}

function mountEditor(modelValue: Task) {
	return mount(RepeatAfter, {
		props: {modelValue},
		global: {
			components: {XButton},
			stubs: {
				Icon: true,
				RouterLink: true,
			},
			directives: {cy: {}},
			mocks: {$t: (key: string) => key},
		},
	})
}

function lastEmitted(wrapper: ReturnType<typeof mountEditor>): Task {
	const events = wrapper.emitted('update:modelValue') ?? []
	return events[events.length - 1]![0] as Task
}

function presetButton(wrapper: ReturnType<typeof mountEditor>, label: string) {
	return wrapper.findAll('button').find(button => button.text() === label)!
}

const advancedRule: TaskRepeat = {
	freq: 'yearly',
	interval: 1,
	read_only: true,
}

describe('RepeatAfter', () => {
	beforeEach(() => {
		message.error.mockReset()
	})

	it.each([
		['task.repeat.everyDay', 'daily'],
		['task.repeat.everyWeek', 'weekly'],
		['task.repeat.everyMonth', 'monthly'],
		['task.repeat.everyYear', 'yearly'],
	])('sets a single-interval rule from the %s preset', async (label, freq) => {
		const task = taskWith(undefined, true)
		const wrapper = mountEditor(task)

		await presetButton(wrapper, label).trigger('click')

		expect(lastEmitted(wrapper)).toMatchObject({
			id: 7,
			repeat: {
				freq,
				interval: 1,
			},
			repeats_from_current_date: true,
		})
		expect(task.repeat).toBeUndefined()
	})

	it('builds the rule from the interval and unit', async () => {
		const wrapper = mountEditor(taskWith({
			freq: 'daily',
			interval: 1,
		}))

		await wrapper.get('#repeatInterval').setValue(2)

		expect(lastEmitted(wrapper).repeat).toEqual({
			freq: 'daily',
			interval: 2,
		})
	})

	it('sets the day of month only for a monthly rule', async () => {
		const wrapper = mountEditor(taskWith({
			freq: 'daily',
			interval: 1,
		}))
		expect(wrapper.find('#repeatDay').exists()).toBe(false)

		await wrapper.get('.repeat-interval-unit select').setValue('months')
		await wrapper.get('#repeatDay').setValue('15')

		expect(lastEmitted(wrapper).repeat).toEqual({
			freq: 'monthly',
			interval: 1,
			by_month_day: [15],
		})
	})

	it('keeps the due date day when the day of month is reset', async () => {
		const wrapper = mountEditor(taskWith({
			freq: 'monthly',
			interval: 1,
			by_month_day: [15],
		}))

		await wrapper.get('#repeatDay').setValue('0')

		expect(lastEmitted(wrapper).repeat).toEqual({
			freq: 'monthly',
			interval: 1,
		})
	})

	it('toggles repeating from the completion date', async () => {
		const wrapper = mountEditor(taskWith({
			freq: 'weekly',
			interval: 3,
		}))

		await wrapper.get('.repeat-from-current-date input').setValue(true)

		expect(lastEmitted(wrapper)).toMatchObject({
			repeat: {
				freq: 'weekly',
				interval: 3,
			},
			repeats_from_current_date: true,
		})
	})

	it('rejects an interval below one', async () => {
		const wrapper = mountEditor(taskWith({
			freq: 'daily',
			interval: 1,
		}))

		await wrapper.get('#repeatInterval').setValue(0)

		expect(message.error).toHaveBeenCalledExactlyOnceWith({message: 'task.repeat.invalidAmount'})
		expect(wrapper.emitted('update:modelValue')).toBeUndefined()
	})

	it.each([
		['a server read-only rule', advancedRule],
		['a rule with parts the editor cannot show', {
			freq: 'weekly',
			interval: 1,
			by_day: ['MO', 'WE'],
		}],
	])('shows %s read-only', async (_name, repeat: TaskRepeat) => {
		const wrapper = mountEditor(taskWith(repeat))

		expect(wrapper.text()).toContain('task.repeat.advancedReadonly')
		for (const button of wrapper.findAll('.button-group button')) {
			expect(button.attributes('disabled')).toBeDefined()
		}
		expect(wrapper.get('#repeatInterval').attributes('disabled')).toBeDefined()
		expect(wrapper.get('.repeat-interval-unit select').attributes('disabled')).toBeDefined()
		expect(wrapper.get('.repeat-from-current-date input').attributes('disabled')).toBeUndefined()

		await presetButton(wrapper, 'task.repeat.everyDay').trigger('click')
		await wrapper.get('#repeatInterval').trigger('change')
		expect(wrapper.emitted('update:modelValue')).toBeUndefined()
	})

	it('keeps a read-only rule whole when the completion date flag changes', async () => {
		const wrapper = mountEditor(taskWith(advancedRule))

		await wrapper.get('.repeat-from-current-date input').setValue(true)

		expect(lastEmitted(wrapper)).toMatchObject({
			repeat: advancedRule,
			repeats_from_current_date: true,
		})
	})

	it.each([
		['the last day of the month', {freq: 'monthly', interval: 1, by_month_day: [-1]}],
		['a yearly rule on a day of the month', {freq: 'yearly', interval: 1, by_month_day: [15]}],
		['a simple monthly rule', {freq: 'monthly', interval: 2, by_month_day: [15], read_only: false}],
	])('keeps %s whole when the completion date flag changes', async (_name, repeat: TaskRepeat) => {
		const wrapper = mountEditor(taskWith({...repeat}))

		await wrapper.get('.repeat-from-current-date input').setValue(true)

		const emitted = lastEmitted(wrapper)
		expect(emitted.repeat).toEqual(repeat)
		expect(emitted.repeats_from_current_date).toBe(true)
	})

	it('shows no editable rule for a simple one', () => {
		const wrapper = mountEditor(taskWith({
			freq: 'daily',
			interval: 1,
		}))

		expect(wrapper.text()).not.toContain('task.repeat.advancedReadonly')
		expect(wrapper.get('#repeatInterval').attributes('disabled')).toBeUndefined()
	})

	it('hides the interval row for a frequency it has no unit for', () => {
		const wrapper = mountEditor(taskWith({
			freq: 'minutely',
			interval: 30,
		}))

		expect(wrapper.find('#repeatInterval').exists()).toBe(false)
		expect(wrapper.text()).toContain('task.repeat.advancedReadonly')
	})
})
