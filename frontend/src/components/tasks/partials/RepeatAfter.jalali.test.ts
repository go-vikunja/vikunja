import {describe, expect, it, beforeEach, vi} from 'vitest'
import {mount, shallowMount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import RepeatAfter from './RepeatAfter.vue'
import KanbanCard from './KanbanCard.vue'
import SingleTaskInProject from './SingleTaskInProject.vue'
import SingleTaskInlineReadonly from './SingleTaskInlineReadonly.vue'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import en from '@/i18n/lang/en.json'
import faIR from '@/i18n/lang/fa-IR.json'

vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

vi.mock('vue-router', async importOriginal => ({
	...await importOriginal<typeof import('vue-router')>(),
	useRouter: () => ({push: vi.fn(), currentRoute: {value: {fullPath: '/'}}}),
	useRoute: () => ({fullPath: '/', params: {}, query: {}}),
	onBeforeRouteUpdate: () => {},
}))

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({currentProject: {id: 1, title: 'Test'}}),
}))

vi.mock('@/stores/projects', () => ({
	useProjectStore: () => ({projects: {}}),
}))

vi.mock('@/stores/tasks', () => ({
	useTaskStore: () => ({update: vi.fn(), toggleFavorite: vi.fn()}),
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({settings: {frontendSettings: {}}}),
}))

const JALALI_MONTH = (TASK_REPEAT_MODES as unknown as Record<string, number>).REPEAT_MODE_JALALI_MONTH ?? 3
const JALALI_YEAR = (TASK_REPEAT_MODES as unknown as Record<string, number>).REPEAT_MODE_JALALI_YEAR ?? 4

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function makeRepeatTask(overrides: Record<string, unknown> = {}) {
	return {
		id: 1,
		title: 'Test',
		done: false,
		dueDate: null,
		startDate: null,
		endDate: null,
		repeatAfter: {amount: 0, type: 'days'},
		repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
		priority: 0,
		labels: [],
		assignees: [],
		attachments: [],
		description: '',
		hexColor: '',
		percentDone: 0,
		projectId: 1,
		bucketId: 0,
		coverImageAttachmentId: null,
		isFavorite: false,
		relatedTasks: {},
		...overrides,
	} as never
}

function mountRepeatAfter(overrides: Record<string, unknown> = {}) {
	return mount(RepeatAfter, {
		props: {modelValue: makeRepeatTask(overrides)},
		global: {
			plugins: [i18n],
			stubs: {XButton: true},
		},
	})
}

describe('RepeatAfter Jalali modes', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('renders 5 repeat mode options', () => {
		const wrapper = mountRepeatAfter()
		const options = wrapper.find('#repeatMode').findAll('option')
		expect(options).toHaveLength(5)
		expect(options.map(o => o.element.value)).toEqual(['0', '1', '2', '3', '4'])
		expect(options[3].text()).toBe('Monthly (Jalali)')
		expect(options[4].text()).toBe('Yearly (Jalali)')
		wrapper.unmount()
	})

	it('selecting Jalali monthly with amount 0 emits a valid model (not a no-op)', async () => {
		const wrapper = mountRepeatAfter()
		await wrapper.find('#repeatMode').setValue(String(JALALI_MONTH))
		const emitted = wrapper.emitted('update:modelValue')
		expect(emitted).toBeTruthy()
		expect((emitted![0][0] as unknown as {repeatMode: number}).repeatMode).toBe(JALALI_MONTH)
		wrapper.unmount()
	})

	it('selecting Jalali yearly with amount 0 emits a valid model (not a no-op)', async () => {
		const wrapper = mountRepeatAfter()
		await wrapper.find('#repeatMode').setValue(String(JALALI_YEAR))
		const emitted = wrapper.emitted('update:modelValue')
		expect(emitted).toBeTruthy()
		expect((emitted![0][0] as unknown as {repeatMode: number}).repeatMode).toBe(JALALI_YEAR)
		wrapper.unmount()
	})

	it.each([
		['jalali monthly', JALALI_MONTH, false],
		['jalali yearly', JALALI_YEAR, false],
		['gregorian monthly', TASK_REPEAT_MODES.REPEAT_MODE_MONTH, false],
		['default', TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT, true],
		['from current date', TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE, true],
	])('amount block %s visibility: visible=%s', async (_label, mode, visible) => {
		const wrapper = mountRepeatAfter({repeatMode: mode})
		await wrapper.vm.$nextTick()
		expect(wrapper.find('input[type="number"]').exists()).toBe(visible)
		wrapper.unmount()
	})

	it('keeps Gregorian no-op and error behavior unchanged', async () => {
		const noopDefault = mountRepeatAfter({repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT})
		await noopDefault.find('#repeatMode').trigger('change')
		expect(noopDefault.emitted('update:modelValue')).toBeUndefined()
		noopDefault.unmount()

		const noopFromCurrent = mountRepeatAfter({
			repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE,
			repeatAfter: {amount: 0, type: 'days'},
		})
		await noopFromCurrent.find('#repeatMode').trigger('change')
		expect(noopFromCurrent.emitted('update:modelValue')).toBeUndefined()
		noopFromCurrent.unmount()

		const monthly = mountRepeatAfter({repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_MONTH})
		await monthly.find('#repeatMode').trigger('change')
		expect(monthly.emitted('update:modelValue')).toBeTruthy()
		monthly.unmount()

		const daily = mountRepeatAfter({repeatAfter: {amount: 1, type: 'days'}})
		await daily.find('#repeatMode').trigger('change')
		expect(daily.emitted('update:modelValue')).toBeTruthy()
		daily.unmount()
	})

	it('keeps quick buttons rendered unchanged', () => {
		const wrapper = mountRepeatAfter()
		expect(wrapper.findAll('x-button-stub')).toHaveLength(3)
		expect(wrapper.find('#repeatMode').exists()).toBe(true)
		wrapper.unmount()
	})
})

describe('Jalali recurrence badges', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('KanbanCard shows history icon for Jalali modes with amount 0', () => {
		for (const mode of [JALALI_MONTH, JALALI_YEAR]) {
			const wrapper = shallowMount(KanbanCard, {
				props: {task: makeRepeatTask({repeatMode: mode}), projectId: 1},
				global: {plugins: [i18n]},
			})
			expect(wrapper.html()).toContain('history')
			wrapper.unmount()
		}
	})

	it('KanbanCard keeps Gregorian badge behavior unchanged', () => {
		const hidden = shallowMount(KanbanCard, {
			props: {task: makeRepeatTask({repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT}), projectId: 1},
			global: {plugins: [i18n]},
		})
		expect(hidden.html()).not.toContain('history')
		hidden.unmount()

		const shown = shallowMount(KanbanCard, {
			props: {task: makeRepeatTask({repeatAfter: {amount: 1, type: 'days'}}), projectId: 1},
			global: {plugins: [i18n]},
		})
		expect(shown.html()).toContain('history')
		shown.unmount()
	})

	it('SingleTaskInProject shows repeating icon for Jalali modes with amount 0', () => {
		for (const mode of [JALALI_MONTH, JALALI_YEAR]) {
			const wrapper = shallowMount(SingleTaskInProject, {
				props: {theTask: makeRepeatTask({repeatMode: mode})},
				global: {plugins: [i18n]},
			})
			expect(wrapper.html()).toContain('history')
			wrapper.unmount()
		}
	})

	it('SingleTaskInProject keeps Gregorian repeating behavior unchanged', () => {
		const monthly = shallowMount(SingleTaskInProject, {
			props: {theTask: makeRepeatTask({repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_MONTH})},
			global: {plugins: [i18n]},
		})
		expect(monthly.html()).toContain('history')
		monthly.unmount()

		const plain = shallowMount(SingleTaskInProject, {
			props: {theTask: makeRepeatTask({repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT})},
			global: {plugins: [i18n]},
		})
		expect(plain.html()).not.toContain('history')
		plain.unmount()
	})

	it('SingleTaskInlineReadonly shows history icon for Jalali modes with amount 0', () => {
		for (const mode of [JALALI_MONTH, JALALI_YEAR]) {
			const wrapper = shallowMount(SingleTaskInlineReadonly, {
				props: {task: makeRepeatTask({repeatMode: mode})},
				global: {plugins: [i18n]},
			})
			expect(wrapper.html()).toContain('history')
			wrapper.unmount()
		}
	})

	it('SingleTaskInlineReadonly keeps Gregorian badge behavior unchanged', () => {
		const plain = shallowMount(SingleTaskInlineReadonly, {
			props: {task: makeRepeatTask({repeatMode: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT})},
			global: {plugins: [i18n]},
		})
		expect(plain.html()).not.toContain('history')
		plain.unmount()

		const amount = shallowMount(SingleTaskInlineReadonly, {
			props: {task: makeRepeatTask({repeatAfter: {amount: 2, type: 'weeks'}})},
			global: {plugins: [i18n]},
		})
		expect(amount.html()).toContain('history')
		amount.unmount()
	})
})

describe('Jalali repeat i18n labels', () => {
	it('provides en source strings and fa-IR translations for the same keys', () => {
		expect(i18n.global.t('task.repeat.jalaliMonthly')).toBe('Monthly (Jalali)')
		expect(i18n.global.t('task.repeat.jalaliYearly')).toBe('Yearly (Jalali)')
		const i18nFa = createI18n({legacy: false, locale: 'fa-IR', messages: {'fa-IR': faIR}})
		expect(i18nFa.global.t('task.repeat.jalaliMonthly')).toBe('ماهانه جلالی')
		expect(i18nFa.global.t('task.repeat.jalaliYearly')).toBe('سالانه جلالی')
		expect(i18nFa.global.t('task.repeat.monthly')).toBe('ماهانه')
	})
})
