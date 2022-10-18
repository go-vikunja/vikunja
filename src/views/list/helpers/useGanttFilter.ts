import {computed, ref, shallowReactive, watchEffect, type Ref} from 'vue'
import type {RouteLocationNormalized, RouteLocationRaw} from 'vue-router'
import cloneDeep from 'lodash.clonedeep'

import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {parseDateProp} from '@/helpers/time/parseDateProp'
import {parseBooleanProp} from '@/helpers/time/parseBooleanProp'
import {useRouteFilter} from '@/composables/useRouteFilter'

import type {IList} from '@/modelTypes/IList'
import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'

import type {DateISO} from '@/types/DateISO'
import type {DateKebab} from '@/types/DateKebab'

import TaskCollectionService from '@/services/taskCollection'
import TaskService from '@/services/task'

import TaskModel from '@/models/task'
import {error, success} from '@/message'

// convenient internal filter object
export interface GanttFilter {
	listId: IList['id']
	dateFrom: DateISO
	dateTo: DateISO
	showTasksWithoutDates: boolean
}

// FIXME: unite with other filter params types
interface GetAllTasksParams {
	sort_by: ('start_date' | 'done' | 'id')[],
	order_by: ('asc' | 'asc' | 'desc')[],
	filter_by: 'start_date'[],
	filter_comparator: ('greater_equals' | 'less_equals')[],
	filter_value: [string, string] // [dateFrom, dateTo],
	filter_concat: 'and',
	filter_include_nulls: boolean,
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
	const {filters} = useRouteFilter<GanttFilter>(route, routeToFilter, filterToRoute)

	const taskCollectionService = shallowReactive(new TaskCollectionService())
	const taskService = shallowReactive(new TaskService())

	const isLoading = computed(() => taskCollectionService.loading)

	const tasks = ref<Map<ITask['id'], ITask>>(new Map())

	async function getAllTasks(params: GetAllTasksParams, page = 1): Promise<ITask[]> {
		const tasks = await taskCollectionService.getAll({listId: filters.listId}, params, page) as ITask[]
		if (page < taskCollectionService.totalPages) {
			const nextTasks = await getAllTasks(params, page + 1)
			return tasks.concat(nextTasks)
		}
		return tasks
	}

	async function loadTasks(filters: GanttFilter) {
		const params: GetAllTasksParams = {
			sort_by: ['start_date', 'done', 'id'],
			order_by: ['asc', 'asc', 'desc'],
			filter_by: ['start_date', 'start_date'],
			filter_comparator: ['greater_equals', 'less_equals'],
			filter_value: [isoToKebabDate(filters.dateFrom), isoToKebabDate(filters.dateTo)],
			filter_concat: 'and',
			filter_include_nulls: filters.showTasksWithoutDates,
		}

		const loadedTasks = await getAllTasks(params)
		tasks.value = new Map()
		loadedTasks.forEach(t => tasks.value.set(t.id, t))
	}

	watchEffect(() => loadTasks(filters))

	async function addTask(task: Partial<ITask>) {
		const newTask = await taskService.create(
			new TaskModel({...task})
		)
		tasks.value.set(newTask.id, newTask)
	
		return newTask
	}

	async function updateTask(task: ITaskPartialWithId) {
		const oldTask = cloneDeep(tasks.value.get(task.id))

		if (!oldTask) return

		// we extend the task with potentially missing info
		const newTask: ITask = {
			...oldTask,
			...task,
		}

		// set in expectation that server update works
		tasks.value.set(newTask.id, newTask)

		try {	
			const updatedTask = await taskService.update(newTask)
			// update the task with possible changes from server
			tasks.value.set(updatedTask.id, updatedTask)
			success('Saved')
		} catch(e: any) {
			error('Something went wrong saving the task')
			// roll back changes
			tasks.value.set(task.id, oldTask)
		}
	}

	return {
		filters,

		tasks,

		isLoading,
		addTask,
		updateTask,
	}
}