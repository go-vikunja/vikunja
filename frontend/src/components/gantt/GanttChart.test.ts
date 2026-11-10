import {describe, it, expect} from 'vitest'
import {nextTick} from 'vue'
import {mount} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'

import GanttChart from './GanttChart.vue'
import GanttTimelineHeader from './GanttTimelineHeader.vue'
import en from '@/i18n/lang/en.json'
import {i18n as globalI18n} from '@/i18n'
import {normalizeTask, type TaskResponse} from '@/client/queries/tasks'
import type {GanttFilters} from '@/views/project/helpers/useGanttFilters'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

// the dayjs locale sync reads the app-wide i18n instance, not the one installed on the wrapper
globalI18n.global.locale.value = 'en'
const router = createRouter({history: createMemoryHistory(), routes: [{path: '/', component: {template: '<div/>'}}]})

const FILTERS: GanttFilters = {
	projectId: 1,
	viewId: 1,
	dateFrom: '2026-08-01T00:00:00.000Z',
	dateTo: '2026-10-01T00:00:00.000Z',
	showTasksWithoutDates: false,
	order: 'startDate',
}

function mountChart(isLoading: boolean) {
	return mount(GanttChart, {
		shallow: true,
		props: {
			isLoading,
			filters: FILTERS,
			tasks: new Map<number, TaskResponse>(),
			defaultTaskStartDate: FILTERS.dateFrom,
			defaultTaskEndDate: FILTERS.dateTo,
		},
		global: {
			plugins: [i18n, router],
		},
	})
}

describe('GanttChart.vue', () => {
	it('measures the day width once the chart replaces the loading state', async () => {
		const wrapper = mountChart(true)
		expect(wrapper.findComponent(GanttTimelineHeader).exists()).toBe(false)

		await wrapper.setProps({isLoading: false})

		expect(wrapper.findComponent(GanttTimelineHeader).props('dayWidthPixels')).toBeGreaterThan(0)
	})

	describe('manual row order', () => {
		function task(id: number, position: number): TaskResponse {
			return normalizeTask({
				id,
				title: `Task ${id}`,
				start_date: '2026-08-10T00:00:00.000Z',
				end_date: '2026-08-20T00:00:00.000Z',
				position,
				related_tasks: {},
			})
		}

		function mountReorderable(reorderable: boolean) {
			return mount(GanttChart, {
				shallow: true,
				props: {
					isLoading: false,
					filters: {...FILTERS, order: 'manual'},
					tasks: new Map([[1, task(1, 100)], [2, task(2, 200)], [3, task(3, 300)]]),
					defaultTaskStartDate: FILTERS.dateFrom,
					defaultTaskEndDate: FILTERS.dateTo,
					reorderable,
				},
				global: {
					plugins: [i18n, router],
					stubs: {GanttChartBody: false, GanttChartPrimitive: false, GanttRow: false, GanttRowPrimitive: false},
				},
			})
		}

		it('shows a grip per row only when rows can be reordered', () => {
			expect(mountReorderable(true).findAll('.gantt-row-grip')).toHaveLength(3)
			expect(mountReorderable(false).findAll('.gantt-row-grip')).toHaveLength(0)
		})

		it('emits the new position when a row is dropped on another slot', async () => {
			const wrapper = mountReorderable(true)

			await wrapper.findAll('.gantt-row-grip')[2].trigger('pointerdown', {clientY: 100})
			document.dispatchEvent(new PointerEvent('pointermove', {clientY: 0}))
			await nextTick()
			expect(wrapper.find('.gantt-row-drop-indicator').exists()).toBe(true)

			document.dispatchEvent(new PointerEvent('pointerup'))
			await nextTick()
			expect(wrapper.emitted('update:position')).toEqual([[3, 50]])
			expect(wrapper.find('.gantt-row-drop-indicator').exists()).toBe(false)
		})
	})
})
