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
import {isOverdue, overdueAnchor, overdueCutoff} from './overdue'

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

// Page-fetch batch size: enough to hide latency without hammering the API when
// a filter matches thousands of tasks (hundreds of pages).
const PAGE_FETCH_CONCURRENCY = 4

type SortableSidebarSort = Exclude<PlannerSidebarSort, 'none' | 'random'>

// Mirrors the server's `sort_by=[field, id]` ordering. Sorting client-side
// means switching the sort doesn't refetch the whole unscheduled list.
const SIDEBAR_COMPARATORS: Record<SortableSidebarSort, (a: ITask, b: ITask) => number> = {
	'priority:desc': (a, b) => b.priority - a.priority,
	'priority:asc': (a, b) => a.priority - b.priority,
	'title:asc': (a, b) => a.title.localeCompare(b.title),
	'title:desc': (a, b) => b.title.localeCompare(a.title),
	'created:desc': (a, b) => b.created.getTime() - a.created.getTime(),
	'created:asc': (a, b) => a.created.getTime() - b.created.getTime(),
	'percent_done:desc': (a, b) => b.percentDone - a.percentDone,
	'percent_done:asc': (a, b) => a.percentDone - b.percentDone,
}

export function usePlannerTasks(
	range: Ref<PlannerRange>,
	sidebarFilter: Ref<TaskFilterParams>,
	sidebarSort: Ref<PlannerSidebarSort>,
	overdueEnabled: Ref<boolean>,
) {
	const authStore = useAuthStore()
	const taskStore = useTaskStore()

	const gridService = shallowReactive(new TaskService())
	const sidebarService = shallowReactive(new TaskService())
	const overdueService = shallowReactive(new TaskService())

	const gridTasks = ref<Map<ITask['id'], ITask>>(new Map())
	// Server-order list; the exposed sidebarTasks computed applies the sort.
	const sidebarTasksRaw = ref<ITask[]>([])
	const overdueTasks = ref<ITask[]>([])

	// 'random' keys each task once so the order is stable across recomputes and
	// a newly added task gets a slot without reshuffling the rest; re-selecting
	// 'random' clears the keys for a fresh shuffle.
	const randomKeys = new Map<ITask['id'], number>()
	watch(sidebarSort, sort => {
		if (sort === 'random') {
			randomKeys.clear()
		}
	})

	function randomKey(id: ITask['id']): number {
		let key = randomKeys.get(id)
		if (key === undefined) {
			key = Math.random()
			randomKeys.set(id, key)
		}
		return key
	}

	const sidebarTasks = computed<ITask[]>(() => {
		// Guard against a stale/garbage stored value.
		const sort = PLANNER_SIDEBAR_SORTS.includes(sidebarSort.value) ? sidebarSort.value : DEFAULT_PLANNER_SIDEBAR_SORT
		if (sort === 'none') {
			return sidebarTasksRaw.value
		}
		if (sort === 'random') {
			return [...sidebarTasksRaw.value].sort((a, b) => randomKey(a.id) - randomKey(b.id))
		}
		const compare = SIDEBAR_COMPARATORS[sort]
		// id desc as the final tiebreaker so the chosen column drives the order.
		return [...sidebarTasksRaw.value].sort((a, b) => compare(a, b) || b.id - a.id)
	})

	const isLoading = computed(() => gridService.loading || sidebarService.loading || overdueService.loading)
	const loadError = ref(false)

	// Monotonic tokens so a slow earlier load can't overwrite a newer one when the
	// user navigates faster than requests resolve.
	let gridLoadId = 0
	let sidebarLoadId = 0
	let overdueLoadId = 0

	async function fetchAll(service: TaskService, params: TaskFilterParams): Promise<ITask[]> {
		const first = await service.getAll({} as ITask, params, 1) as ITask[]
		// totalPages is known after the first page; snapshot it before later
		// responses overwrite it on the shared service instance.
		const totalPages = service.totalPages
		if (totalPages <= 1) {
			return first
		}
		const pages = [first]
		for (let page = 2; page <= totalPages; page += PAGE_FETCH_CONCURRENCY) {
			pages.push(...await Promise.all(Array.from(
				{length: Math.min(PAGE_FETCH_CONCURRENCY, totalPages - page + 1)},
				(_, i) => service.getAll({} as ITask, params, page + i) as Promise<ITask[]>,
			)))
		}
		return pages.flat()
	}

	async function loadGrid() {
		// Workaround for a backend bug: a date-only filter's `>=`/`=` comparison
		// excludes a value landing exactly on the boundary instead of treating it
		// as inclusive, so a task starting at local midnight on the first visible
		// day gets silently dropped. Query from the day before and let the
		// client-side day layout (which clamps to the real visible range) discard
		// anything that doesn't belong.
		const from = isoToKebabDate(dayjs(range.value.from).subtract(1, 'day').toISOString())
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
			// repeat_after is filterable until feat-filterable-repeat-mode lands
			// (see plans/feat-filterable-repeat-mode.md), so month-mode tasks with
			// repeat_after = 0 aren't caught here — add `|| repeat_mode != 0` then.
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

		const id = ++gridLoadId
		try {
			const loaded = await fetchAll(gridService, params)
			if (id !== gridLoadId) {
				return
			}
			const map = new Map<ITask['id'], ITask>()
			loaded.forEach(t => map.set(t.id, t))
			gridTasks.value = map
			loadError.value = false
		} catch (_) {
			if (id === gridLoadId) {
				loadError.value = true
				error({message: i18n.global.t('planner.loadError')})
			}
		}
	}

	async function loadSidebar() {
		// Combine the user's filter (already API-form from the Filters component)
		// with done=false. The v1 filter can't express "date is null", so we keep
		// only tasks lacking a start/end client-side.
		const userFilter = sidebarFilter.value.filter?.trim()
		const filter = userFilter ? `(${userFilter}) && done = false` : 'done = false'

		const params: TaskFilterParams = {
			filter,
			// The sidebar's own null-date filtering happens client-side below, so
			// include_nulls stays the user's choice from the filter popup — the
			// backend applies it to every condition of their filter.
			filter_include_nulls: sidebarFilter.value.filter_include_nulls ?? false,
			filter_timezone: authStore.settings.timezone,
			s: sidebarFilter.value.s ?? '',
			expand: 'subtasks',
		} as TaskFilterParams

		// Truly unscheduled = no start, end or due date. Due-only tasks already
		// render in the all-day row, so keep them out of the sidebar.
		const id = ++sidebarLoadId
		try {
			const loaded = await fetchAll(sidebarService, params)
			if (id !== sidebarLoadId) {
				return
			}
			sidebarTasksRaw.value = loaded.filter(task => !task.startDate && !task.endDate && !task.dueDate)
			loadError.value = false
		} catch (_) {
			if (id === sidebarLoadId) {
				loadError.value = true
				error({message: i18n.global.t('planner.loadError')})
			}
		}
	}

	async function loadOverdue() {
		const id = ++overdueLoadId
		if (!overdueEnabled.value) {
			overdueTasks.value = []
			return
		}

		// Superset fetch: any not-done task dated before today. The precise
		// "overdue" definition (a schedule reaching into today or later isn't
		// overdue) mixes the three date fields, which the filter grammar can't
		// express, so it is applied client-side via isOverdue below.
		const cutoff = isoToKebabDate(overdueCutoff().toISOString())
		const params: TaskFilterParams = {
			filter: `(due_date < "${cutoff}" || end_date < "${cutoff}" || start_date < "${cutoff}") && done = false`,
			filter_include_nulls: false,
			filter_timezone: authStore.settings.timezone,
			s: '',
			expand: 'subtasks',
		} as TaskFilterParams

		try {
			const loaded = await fetchAll(overdueService, params)
			if (id !== overdueLoadId) {
				return
			}
			overdueTasks.value = loaded
				.filter(task => isOverdue(task))
				.sort((a, b) => (overdueAnchor(a)?.getTime() ?? 0) - (overdueAnchor(b)?.getTime() ?? 0))
			loadError.value = false
		} catch (_) {
			if (id === overdueLoadId) {
				loadError.value = true
				error({message: i18n.global.t('planner.loadError')})
			}
		}
	}

	watch(range, () => loadGrid(), {immediate: true, deep: true})
	// Sorting happens client-side (the sidebarTasks computed), so only filter
	// changes need a refetch.
	watch(sidebarFilter, () => loadSidebar(), {immediate: true, deep: true})
	watch(overdueEnabled, () => loadOverdue(), {immediate: true})

	// Keep the lists in sync with edits made elsewhere (e.g. the task detail
	// modal): re-file the task into the grid or sidebar, or drop it if it's now
	// done. Untracked tasks are placed too when they carry a date (e.g. one just
	// scheduled into the visible week from quick search); untracked dateless
	// tasks are skipped since the sidebar's server-side filter can't be
	// evaluated client-side.
	watch(
		() => taskStore.lastUpdatedTask,
		updatedTask => {
			if (!updatedTask) {
				return
			}
			const known = gridTasks.value.has(updatedTask.id)
				|| sidebarTasksRaw.value.some(t => t.id === updatedTask.id)
				|| overdueTasks.value.some(t => t.id === updatedTask.id)
			if (known || updatedTask.startDate || updatedTask.endDate || updatedTask.dueDate) {
				placeTask(updatedTask)
			}
		},
	)

	// Drop a task deleted elsewhere (e.g. the task detail modal opened over the
	// planner) from all lists, since the planner stays mounted underneath.
	watch(
		() => taskStore.lastDeletedTask,
		deletedTask => {
			if (!deletedTask) {
				return
			}
			gridTasks.value.delete(deletedTask.id)
			removeFromList(sidebarTasksRaw.value, deletedTask.id)
			removeFromList(overdueTasks.value, deletedTask.id)
		},
	)

	function removeFromList(list: ITask[], taskId: ITask['id']) {
		const index = list.findIndex(t => t.id === taskId)
		if (index >= 0) {
			list.splice(index, 1)
		}
	}

	// Put a task into whichever list(s) it now belongs to. A dated task always
	// goes on the grid (matching the date-range fetch in loadGrid, which knows
	// nothing about overdue status); the overdue sidebar section is a separate,
	// additive listing of the same task when it's enabled and still not done, not
	// an alternative to being on the grid. A dateless task falls back to the
	// unscheduled sidebar.
	function placeTask(task: ITask) {
		gridTasks.value.delete(task.id)
		removeFromList(sidebarTasksRaw.value, task.id)
		removeFromList(overdueTasks.value, task.id)

		if (task.startDate || task.endDate || task.dueDate) {
			gridTasks.value.set(task.id, task)
		} else if (!task.done) {
			sidebarTasksRaw.value.unshift(task)
		}

		if (overdueEnabled.value && isOverdue(task)) {
			overdueTasks.value.push(task)
			overdueTasks.value.sort((a, b) => (overdueAnchor(a)?.getTime() ?? 0) - (overdueAnchor(b)?.getTime() ?? 0))
		}
	}

	async function updateTask(partial: ITaskPartialWithId) {
		const base = gridTasks.value.get(partial.id)
			?? sidebarTasksRaw.value.find(t => t.id === partial.id)
			?? overdueTasks.value.find(t => t.id === partial.id)
		if (!base) return

		const oldTask = klona(base)
		const newTask: ITask = {...oldTask, ...partial}

		placeTask(newTask)

		try {
			const updated = await taskStore.update(newTask)
			placeTask(updated)
			success({message: i18n.global.t('planner.saved')})
		} catch (_) {
			error({message: i18n.global.t('planner.saveError')})
			placeTask(oldTask)
		}
	}

	// Place a freshly created task (not yet tracked) onto the grid with the given
	// dates and persist them. Used by the create-by-gesture flow, where AddTask
	// creates a dateless task that we then schedule into the painted slot.
	async function scheduleTask(task: ITask, dates: {startDate: Date | null, endDate: Date | null}) {
		const oldTask = klona(task)
		const newTask: ITask = {...task, ...dates}
		placeTask(newTask)
		try {
			const updated = await taskStore.update(newTask)
			placeTask(updated)
			success({message: i18n.global.t('planner.saved')})
		} catch (_) {
			error({message: i18n.global.t('planner.saveError')})
			// The task does exist (created dateless just before), so re-file it
			// as it was instead of leaving never-persisted dates painted.
			placeTask(oldTask)
		}
	}

	return {
		gridTasks,
		sidebarTasks,
		overdueTasks,
		isLoading,
		loadError,
		updateTask,
		scheduleTask,
	}
}
