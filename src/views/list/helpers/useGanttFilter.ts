import type {Ref} from 'vue'
import type {RouteLocationNormalized, RouteLocationRaw} from 'vue-router'

import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {parseDateProp} from '@/helpers/time/parseDateProp'
import {parseBooleanProp} from '@/helpers/time/parseBooleanProp'
import {useRouteFilter} from '@/composables/useRouteFilter'

import type {IList} from '@/modelTypes/IList'
import type {DateISO} from '@/types/DateISO'
import type {DateKebab} from '@/types/DateKebab'

// convenient internal filter object
export interface GanttFilter {
	listId: IList['id']
	dateFrom: DateISO
	dateTo: DateISO
	showTasksWithoutDates: boolean
}

const DEFAULT_SHOW_TASKS_WITHOUT_DATES = false

const DEFAULT_DATEFROM_DAY_OFFSET = -15
const DEFAULT_DATETO_DAY_OFFSET = +55

const now = new Date()

function getDefaultDateFrom() {
	return new Date(now.getFullYear(), now.getMonth(), now.getDate() + DEFAULT_DATEFROM_DAY_OFFSET).toISOString()
}

function getDefaultDateTo() {
	return new Date(now.getFullYear(), now.getMonth(), now.getDate() + DEFAULT_DATETO_DAY_OFFSET).toISOString()
}

function routeToFilter(route: RouteLocationNormalized): GanttFilter {
	return {
		listId: Number(route.params.listId as string),
		dateFrom: parseDateProp(route.query.dateFrom as DateKebab) || getDefaultDateFrom(),
		dateTo: parseDateProp(route.query.dateTo as DateKebab) || getDefaultDateTo(),
		showTasksWithoutDates: parseBooleanProp(route.query.showTasksWithoutDates as string) || DEFAULT_SHOW_TASKS_WITHOUT_DATES,
	}
}

function filterToRoute(filters: GanttFilter): RouteLocationRaw {
	let query: Record<string, string> = {}
	if (
		filters.dateFrom !== getDefaultDateFrom() ||
		filters.dateTo !== getDefaultDateTo()
	) {
		query = {
			dateFrom: isoToKebabDate(filters.dateFrom),
			dateTo: isoToKebabDate(filters.dateTo),
		}
	}

	if (filters.showTasksWithoutDates) {
		query.showTasksWithoutDates = String(filters.showTasksWithoutDates)
	}

	return {
		name: 'list.gantt',
		params: {listId: filters.listId},
		query,
	}
}

export function useGanttFilter(route: Ref<RouteLocationNormalized>) {
	return useRouteFilter<GanttFilter>(route, routeToFilter, filterToRoute)
}