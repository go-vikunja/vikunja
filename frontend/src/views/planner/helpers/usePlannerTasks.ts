import {computed, ref, shallowReactive, watch, type Ref} from 'vue'
import {klona} from 'klona/lite'
import dayjs from 'dayjs'

import TaskService from '@/services/task'
import type {TaskFilterParams} from '@/services/taskCollection'
import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'

import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'
import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {error, success} from '@/message'
import {i18n} from '@/i18n'

export interface PlannerRange {
	from: Date
	to: Date
}

// Sidebar sort is a "<field>:<order>" string, or 'random' (no backend equivalent,
// so we shuffle client-side). Date fields are intentionally excluded: dated tasks
// live in the grid, not the unscheduled sidebar.
export type PlannerSidebarSort =
	| 'none'
	| 'priority:desc' | 'priority:asc'
	| 'title:asc' | 'title:desc'
	| 'created:desc' | 'created:asc'
	| 'percent_done:desc' | 'percent_done:asc'
	| 'random'

export const PLANNER_SIDEBAR_SORTS: PlannerSidebarSort[] = [
	'none',
	'priority:desc', 'priority:asc',
	'title:asc', 'title:desc',
	'created:desc', 'created:asc',
	'percent_done:desc', 'percent_done:asc',
	'random',
]

// Default: no explicit sort — show the order the server returns.
export const DEFAULT_PLANNER_SIDEBAR_SORT: PlannerSidebarSort = 'none'

function shuffle<T>(input: T[]): T[] {
	const arr = [...input]
	for (let i = arr.length - 1; i > 0; i--) {
		const j = Math.floor(Math.random() * (i + 1))
		;[arr[i], arr[j]] = [arr[j], arr[i]]
	}
	return arr
}

export function usePlannerTasks(range: Ref<PlannerRange>, sidebarFilter: Ref<TaskFilterParams>, sidebarSort: Ref<PlannerSidebarSort>) {
	const authStore = useAuthStore()
	const taskStore = useTaskStore()

	const gridService = shallowReactive(new TaskService())
	const sidebarService = shallowReactive(new TaskService())

	const gridTasks = ref<Map<ITask['id'], ITask>>(new Map())
	const sidebarTasks = ref<ITask[]>([])

	const isLoading = computed(() => gridService.loading || sidebarService.loading)

	async function fetchAll(service: TaskService, params: TaskFilterParams, page = 1): Promise<ITask[]> {
		const tasks = await service.getAll({} as ITask, params, page) as ITask[]
		if (page < service.totalPages) {
			return tasks.concat(await fetchAll(service, params, page + 1))
		}
		return tasks
	}

	async function loadGrid() {
		const from = isoToKebabDate(range.value.from.toISOString())
		// The backend parses a date-only filter value as start-of-day, so a `<= to`
		// bound on the last day's date would drop tasks later that same day. Use the
		// day after the last visible day to keep the whole last day inclusive.
		const to = isoToKebabDate(dayjs(range.value.to).add(1, 'day').toISOString())

		const params: TaskFilterParams = {
			sort_by: ['start_date', 'id'],
			order_by: ['asc', 'desc'],
			// Last clause: recurring tasks whose stored start is before the window
			// still project occurrences into it (expandOccurrences walks forward),
			// so fetch any repeater that started on/before the range end. Only
			// repeat_after is filterable (repeat_mode is not), so month-mode tasks
			// with repeat_after = 0 aren't caught here.
			filter: '(' +
				`(start_date >= "${from}" && start_date <= "${to}") || ` +
				`(end_date >= "${from}" && end_date <= "${to}") || ` +
				`(due_date >= "${from}" && due_date <= "${to}") || ` +
				`(start_date <= "${from}" && end_date >= "${to}") || ` +
				`(start_date <= "${to}" && repeat_after > 0)` +
				')',
			filter_include_nulls: false,
			filter_timezone: authStore.settings.timezone,
			s: '',
			expand: 'subtasks',
		}

		const loaded = await fetchAll(gridService, params)
		const map = new Map<ITask['id'], ITask>()
		loaded.forEach(t => map.set(t.id, t))
		gridTasks.value = map
	}

	async function loadSidebar() {
		// Combine the user's filter (already API-form from the Filters component)
		// with done=false. The v1 filter can't express "date is null", so we keep
		// only tasks lacking a start/end client-side.
		const userFilter = sidebarFilter.value.filter?.trim()
		const filter = userFilter ? `(${userFilter}) && done = false` : 'done = false'

		// Guard against a stale/garbage stored value reaching the API as a bad sort.
		const sort = PLANNER_SIDEBAR_SORTS.includes(sidebarSort.value) ? sidebarSort.value : DEFAULT_PLANNER_SIDEBAR_SORT
		// 'random' has no backend sort, so fetch in server order and shuffle below.
		const random = sort === 'random'

		const params: TaskFilterParams = {
			filter,
			filter_include_nulls: true,
			filter_timezone: authStore.settings.timezone,
			s: sidebarFilter.value.s ?? '',
			expand: 'subtasks',
		} as TaskFilterParams

		// 'none'/'random' send no sort_by, so the server returns its own order.
		if (sort !== 'none' && !random) {
			const [field, order] = sort.split(':')
			params.sort_by = (field === 'id' ? ['id'] : [field, 'id']) as TaskFilterParams['sort_by']
			params.order_by = (field === 'id' ? ['desc'] : [order, 'desc']) as TaskFilterParams['order_by']
		}

		// Truly unscheduled = no start, end or due date. Due-only tasks already
		// render in the all-day row, so keep them out of the sidebar.
		const loaded = await fetchAll(sidebarService, params)
		const unscheduled = loaded.filter(task => !task.startDate && !task.endDate && !task.dueDate)
		sidebarTasks.value = random ? shuffle(unscheduled) : unscheduled
	}

	function load() {
		return Promise.all([loadGrid(), loadSidebar()])
	}

	watch(range, () => loadGrid(), {immediate: true, deep: true})
	watch([sidebarFilter, sidebarSort], () => loadSidebar(), {immediate: true, deep: true})

	// Keep both lists in sync with edits made elsewhere (e.g. the task detail
	// modal): re-file the task into the grid or sidebar, or drop it if it's now
	// done. Only react to tasks the planner already tracks.
	watch(
		() => taskStore.lastUpdatedTask,
		updatedTask => {
			if (!updatedTask) {
				return
			}
			const known = gridTasks.value.has(updatedTask.id)
				|| sidebarTasks.value.some(t => t.id === updatedTask.id)
			if (known) {
				placeTask(updatedTask)
			}
		},
	)

	// Put a task into whichever list it now belongs to: the grid if it has any
	// date (timed, all-day or due), otherwise the unscheduled sidebar.
	function placeTask(task: ITask) {
		gridTasks.value.delete(task.id)
		const sidebarIndex = sidebarTasks.value.findIndex(t => t.id === task.id)
		if (sidebarIndex >= 0) {
			sidebarTasks.value.splice(sidebarIndex, 1)
		}

		if (task.startDate || task.endDate || task.dueDate) {
			gridTasks.value.set(task.id, task)
		} else if (!task.done) {
			sidebarTasks.value.unshift(task)
		}
	}

	async function updateTask(partial: ITaskPartialWithId) {
		const base = gridTasks.value.get(partial.id) ?? sidebarTasks.value.find(t => t.id === partial.id)
		if (!base) return

		const oldTask = klona(base)
		const newTask: ITask = {...oldTask, ...partial}

		placeTask(newTask)

		try {
			const updated = await taskStore.update(newTask)
			placeTask(updated)
			success(i18n.global.t('planner.saved'))
		} catch (_) {
			error(i18n.global.t('planner.saveError'))
			placeTask(oldTask)
		}
	}

	// Place a freshly created task (not yet tracked) onto the grid with the given
	// dates and persist them. Used by the create-by-gesture flow, where AddTask
	// creates a dateless task that we then schedule into the painted slot.
	async function scheduleTask(task: ITask, dates: {startDate: Date | null, endDate: Date | null}) {
		const newTask: ITask = {...task, ...dates}
		placeTask(newTask)
		try {
			const updated = await taskStore.update(newTask)
			placeTask(updated)
			success(i18n.global.t('planner.saved'))
		} catch (_) {
			error(i18n.global.t('planner.saveError'))
		}
	}

	return {
		gridTasks,
		sidebarTasks,
		isLoading,
		load,
		updateTask,
		scheduleTask,
	}
}
