import type {Ref} from 'vue'
import type {RouteLocationNormalized, RouteLocationRaw} from 'vue-router'

import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {parseDateProp} from '@/helpers/time/parseDateProp'
import {parseBooleanProp} from '@/helpers/time/parseBooleanProp'
import {useRouteFilters} from '@/composables/useRouteFilters'
import {useGanttTaskList} from './useGanttTaskList'

import type {IProject} from '@/modelTypes/IProject'
import type {GetAllTasksParams} from '@/services/taskCollection'

import type {DateISO} from '@/types/DateISO'
import type {DateKebab} from '@/types/DateKebab'

// convenient internal filter object
export interface GanttFilters {
	projectId: IProject['id']
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

// FIXME: use zod for this
function ganttRouteToFilters(route: Partial<RouteLocationNormalized>): GanttFilters {
	const ganttRoute = route
	return {
		projectId: Number(ganttRoute.params?.projectId),
		dateFrom: parseDateProp(ganttRoute.query?.dateFrom as DateKebab) || getDefaultDateFrom(),
		dateTo: parseDateProp(ganttRoute.query?.dateTo as DateKebab) || getDefaultDateTo(),
		showTasksWithoutDates: parseBooleanProp(ganttRoute.query?.showTasksWithoutDates as string) || DEFAULT_SHOW_TASKS_WITHOUT_DATES,
	}
}

function ganttGetDefaultFilters(route: Partial<RouteLocationNormalized>): GanttFilters {
	return ganttRouteToFilters({params: {projectId: route.params?.projectId as string}})
}

// FIXME: use zod for this
function ganttFiltersToRoute(filters: GanttFilters): RouteLocationRaw {
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
		name: 'project.gantt',
		params: {projectId: filters.projectId},
		query,
	}
}

function ganttFiltersToApiParams(filters: GanttFilters): GetAllTasksParams {
	return {
		sort_by: ['start_date', 'done', 'id'],
		order_by: ['asc', 'asc', 'desc'],
		filter_by: ['start_date', 'start_date'],
		filter_comparator: ['greater_equals', 'less_equals'],
		filter_value: [isoToKebabDate(filters.dateFrom), isoToKebabDate(filters.dateTo)],
		filter_concat: 'and',
		filter_include_nulls: filters.showTasksWithoutDates,
	}
}

export type UseGanttFiltersReturn =
	ReturnType<typeof useRouteFilters<GanttFilters>> &
	ReturnType<typeof useGanttTaskList<GanttFilters>>

export function useGanttFilters(route: Ref<RouteLocationNormalized>): UseGanttFiltersReturn {
	const {
		filters,
		hasDefaultFilters,
		setDefaultFilters,
	} = useRouteFilters<GanttFilters>(
		route,
		ganttGetDefaultFilters,
		ganttRouteToFilters,
		ganttFiltersToRoute,
		['project.gantt'],
	)

	const {
		tasks,
		loadTasks,

		isLoading,
		addTask,
		updateTask,
	} = useGanttTaskList<GanttFilters>(filters, ganttFiltersToApiParams)

	return {
		filters,
		hasDefaultFilters,
		setDefaultFilters,

		tasks,
		loadTasks,

		isLoading,
		addTask,
		updateTask,
	}
}