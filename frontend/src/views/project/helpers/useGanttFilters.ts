import {watch, type Ref} from 'vue'
import type {RouteLocationNormalized, RouteLocationRaw, LocationQueryRaw} from 'vue-router'

import {useViewFiltersStore} from '@/stores/viewFilters'

import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {parseDateProp} from '@/helpers/time/parseDateProp'
import {parseBooleanProp} from '@/helpers/time/parseBooleanProp'
import {useRouteFilters, type UseRouteFiltersReturn} from '@/composables/useRouteFilters'
import {useGanttTaskList, type UseGanttTaskListReturn} from './useGanttTaskList'

import type {IProject} from '@/modelTypes/IProject'
import type {TaskFilterParams} from '@/services/taskCollection'

import type {DateISO} from '@/types/DateISO'
import type {DateKebab} from '@/types/DateKebab'
import type {IProjectView} from '@/modelTypes/IProjectView'

// convenient internal filter object
export interface GanttFilters {
	projectId: IProject['id']
	viewId: IProjectView['id'],
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
function ganttRouteToFilters(route: Partial<RouteLocationNormalized>, projectId: IProject['id'], viewId: IProjectView['id']): GanttFilters {
	return {
		projectId,
		viewId,
		dateFrom: parseDateProp(route.query?.dateFrom as DateKebab) || getDefaultDateFrom(),
		dateTo: parseDateProp(route.query?.dateTo as DateKebab) || getDefaultDateTo(),
		showTasksWithoutDates: parseBooleanProp(route.query?.showTasksWithoutDates as string) || DEFAULT_SHOW_TASKS_WITHOUT_DATES,
	}
}

function ganttGetDefaultFilters(projectId: IProject['id'], viewId: IProjectView['id']): GanttFilters {
	return ganttRouteToFilters({}, projectId, viewId)
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
		name: 'project.view',
		params: {
			projectId: filters.projectId,
			viewId: filters.viewId,
		},
		query,
	}
}

const GANTT_QUERY_PARAMS = ['dateFrom', 'dateTo', 'showTasksWithoutDates']

// useRouteFilters rebuilds the url from gantt filters alone; carry the rest over, in place.
function ganttFiltersToRoutePreservingQuery(
	filters: GanttFilters,
	currentQuery: RouteLocationNormalized['query'],
): RouteLocationRaw {
	const routeLocation = ganttFiltersToRoute(filters) as {query: LocationQueryRaw}
	const ganttQuery = routeLocation.query
	const query: LocationQueryRaw = {...currentQuery, ...ganttQuery}
	for (const key of GANTT_QUERY_PARAMS) {
		if (!(key in ganttQuery)) {
			delete query[key]
		}
	}

	return {...routeLocation, query}
}

function ganttFiltersToApiParams(filters: GanttFilters, includeSubprojects: boolean): TaskFilterParams {
	const dateFrom = isoToKebabDate(filters.dateFrom)
	const dateTo = isoToKebabDate(filters.dateTo)

	return {
		include_subprojects: includeSubprojects,
		sort_by: ['start_date', 'done', 'id'],
		order_by: ['asc', 'asc', 'desc'],
		filter: '(' +
			'(start_date >= "' + dateFrom + '" && start_date <= "' + dateTo + '") || ' +
			'(end_date >= "' + dateFrom + '" && end_date <= "' + dateTo + '") || ' +
			'(due_date >= "' + dateFrom + '" && due_date <= "' + dateTo + '") || ' +
			'(start_date <= "' + dateFrom + '" && end_date >= "' + dateTo + '")' +
			')',
		filter_include_nulls: filters.showTasksWithoutDates,
		expand: 'subtasks',
	}
}

export type UseGanttFiltersReturn =
	UseRouteFiltersReturn<GanttFilters> &
	UseGanttTaskListReturn

export function useGanttFilters(
	route: Ref<RouteLocationNormalized>,
	projectId: Ref<IProject['id']>,
	viewId: Ref<IProjectView['id']>,
	includeSubprojects: Ref<boolean>,
): UseGanttFiltersReturn {
	const viewFiltersStore = useViewFiltersStore()

	// Ids come from the props and not from the route: while the task detail modal is open, gantt
	// renders as its backdrop and the current route is the task, without any project params.
	const {
		filters,
		hasDefaultFilters,
		setDefaultFilters,
	} = useRouteFilters<GanttFilters>(
		route,
		() => ganttGetDefaultFilters(projectId.value, viewId.value),
		r => ganttRouteToFilters(r, projectId.value, viewId.value),
		filters => ganttFiltersToRoutePreservingQuery(filters, route.value.query),
		['project.view'],
	)

	// Sync filters to store whenever they change (for view tab navigation)
	watch(
		filters,
		(newFilters) => {
			const routeLocation = ganttFiltersToRoute(newFilters)
			const query = routeLocation.query as LocationQueryRaw
			if (query && Object.keys(query).length > 0) {
				viewFiltersStore.setViewQuery(viewId.value, query)
			} else {
				viewFiltersStore.clearViewQuery(viewId.value)
			}
		},
		{immediate: true, deep: true},
	)

	const {
		tasks,
		loadTasks,

		isLoading,
		addTask,
		updateTask,
	} = useGanttTaskList<GanttFilters>(
		filters,
		currentFilters => ganttFiltersToApiParams(currentFilters, includeSubprojects.value),
		viewId,
	)

	watch(includeSubprojects, () => {
		loadTasks()
	})

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
