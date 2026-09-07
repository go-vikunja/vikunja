import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createI18n} from 'vue-i18n'

import en from '@/i18n/lang/en.json'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'

const hoisted = vi.hoisted(() => {
	return {
		filters: {
			projectId: 1,
			viewId: 10,
			dateFrom: '2026-09-01T00:00:00.000Z',
			dateTo: '2026-09-10T00:00:00.000Z',
			showTasksWithoutDates: false,
		},
	}
})

vi.mock('@/views/project/helpers/useGanttFilters', async () => {
	const {ref} = await import('vue')
	const filters = ref(hoisted.filters)
	return {
		useGanttFilters: () => ({
			filters,
			hasDefaultFilters: ref(true),
			setDefaultFilters: vi.fn(),
			tasks: ref(new Map()),
			isLoading: ref(false),
			addTask: vi.fn(),
			updateTask: vi.fn(),
		}),
	}
})

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({currentProject: {id: 1, maxPermission: 2}}),
}))

import ProjectGantt from './ProjectGantt.vue'
import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
import JalaliCalendarGrid from '@/components/input/JalaliCalendarGrid.vue'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function setLocaleAndTimezone(locale: string, timezone: string) {
	globalI18n.global.locale.value = locale as never
	useAuthStore().setUserSettings({timezone} as never)
}

function mountGantt() {
	return mount(ProjectGantt, {
		props: {
			isLoadingProject: false,
			projectId: 1,
			route: {query: {}, params: {}} as never,
			viewId: 10,
		},
		global: {
			plugins: [i18n],
			stubs: {
				ProjectWrapper: {template: '<div><slot name="default" /></div>'},
				GanttChart: {template: '<div />'},
				TaskForm: {template: '<div />'},
				// Stub the flatpickr wrapper so tests never init flatpickr DOM
				// (mirrors DatepickerWithRange tests stubbing flat-pickr).
				Foo: {template: '<input />'},
			},
		},
	})
}

describe('ProjectGantt Jalali range picker', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		hoisted.filters.dateFrom = '2026-09-01T00:00:00.000Z'
		hoisted.filters.dateTo = '2026-09-10T00:00:00.000Z'
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en'
	})

	it('renders flatpickr for en', () => {
		setLocaleAndTimezone('en', 'UTC')
		const wrapper = mountGantt()
		expect(wrapper.findComponent(Foo).exists()).toBe(true)
		expect(wrapper.findComponent(JalaliCalendarGrid).exists()).toBe(false)
	})

	it('renders the Jalali grid in range mode for fa', () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const wrapper = mountGantt()
		expect(wrapper.findComponent(JalaliCalendarGrid).exists()).toBe(true)
		expect(wrapper.findComponent(Foo).exists()).toBe(false)
		expect(wrapper.findComponent(JalaliCalendarGrid).props('mode')).toBe('range')
	})

	it('converts grid picks to ISO through the existing filter flow (Asia/Tehran)', async () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const wrapper = mountGantt()
		const grid = wrapper.findComponent(JalaliCalendarGrid)
		expect(grid.props('timeZone')).toBe('Asia/Tehran')

		const from = new Date('2026-09-02T20:30:00.000Z')
		const to = new Date('2026-09-05T20:29:00.000Z')
		await grid.vm.$emit('update:modelValue', [from, to])
		await wrapper.vm.$nextTick()

		expect(hoisted.filters.dateFrom).toBe(from.toISOString())
		expect(hoisted.filters.dateTo).toBe(to.toISOString())
		// Kebab wire unchanged: ISO still flows through isoToKebabDate
		expect(isoToKebabDate(hoisted.filters.dateFrom as never)).toBe(isoToKebabDate(from.toISOString() as never))
		expect(isoToKebabDate(hoisted.filters.dateTo as never)).toBe(isoToKebabDate(to.toISOString() as never))
	})

	it('converts grid picks to ISO in UTC', async () => {
		setLocaleAndTimezone('fa-IR', 'UTC')
		const wrapper = mountGantt()

		const from = new Date('2026-03-21T00:00:00Z')
		const to = new Date('2026-03-25T00:00:00Z')
		await wrapper.findComponent(JalaliCalendarGrid).vm.$emit('update:modelValue', [from, to])
		await wrapper.vm.$nextTick()

		expect(hoisted.filters.dateFrom).toBe(from.toISOString())
		expect(hoisted.filters.dateTo).toBe(to.toISOString())
	})

	it('ignores partial (single-click) grid selections like flatpickr', async () => {
		setLocaleAndTimezone('fa-IR', 'Asia/Tehran')
		const wrapper = mountGantt()

		await wrapper.findComponent(JalaliCalendarGrid).vm.$emit('update:modelValue', [new Date('2026-09-02T20:30:00.000Z')])
		await wrapper.vm.$nextTick()

		expect(hoisted.filters.dateFrom).toBe('2026-09-01T00:00:00.000Z')
		expect(hoisted.filters.dateTo).toBe('2026-09-10T00:00:00.000Z')
	})
})
