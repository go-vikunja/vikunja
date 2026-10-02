import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'

import GanttChart from './GanttChart.vue'
import GanttTimelineHeader from './GanttTimelineHeader.vue'
import GanttTaskPane from './GanttTaskPane.vue'
import {ZOOM_DAY_WIDTH} from '@/helpers/ganttZoom'
import en from '@/i18n/lang/en.json'
import {i18n as globalI18n} from '@/i18n'
import type {TaskResponse} from '@/client/queries/tasks'
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

	it('uses a fixed day width for a fixed zoom and hands the zoom to the header', async () => {
		const wrapper = mountChart(false)
		await wrapper.setProps({zoom: 'week'})
		await wrapper.vm.$nextTick()

		const header = wrapper.findComponent(GanttTimelineHeader)
		expect(header.props('zoom')).toBe('week')
		expect(header.props('dayWidthPixels')).toBe(ZOOM_DAY_WIDTH.week)
	})

	it('shows the task table by default and hides it on request', async () => {
		const wrapper = mountChart(false)
		expect(wrapper.findComponent(GanttTaskPane).exists()).toBe(true)

		await wrapper.setProps({showPane: false})

		expect(wrapper.findComponent(GanttTaskPane).exists()).toBe(false)
	})

	it('passes the editing right on to the task table', async () => {
		const wrapper = mountChart(false)
		expect(wrapper.findComponent(GanttTaskPane).props('editable')).toBe(false)

		await wrapper.setProps({editable: true})

		expect(wrapper.findComponent(GanttTaskPane).props('editable')).toBe(true)
	})
})
