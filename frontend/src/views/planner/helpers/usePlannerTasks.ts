import {computed, watch, type Ref} from 'vue'
import {useQuery, useQueryClient} from '@tanstack/vue-query'
import dayjs from 'dayjs'

import {allTasksQuery, taskKeys, type TaskFilterParams, type TaskResponse} from '@/client/queries/tasks'
import type {Task} from '@/client/generated'
import {useUpdateTaskMutation} from '@/client/queries/taskMutations'

import {useAuthStore} from '@/stores/auth'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {success} from '@/message'
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

// Dates reach the planner as `Date`s from drag gestures; null clears the field.
export interface PlannerTaskDates {
	id: number
	start_date?: Date | null
	end_date?: Date | null
	due_date?: Date | null
}

type SortableSidebarSort = Exclude<PlannerSidebarSort, 'none' | 'random'>

function createdAt(task: TaskResponse): number {
	return parseDateOrNull(task.created)?.getTime() ?? 0
}

// Mirrors the server's `sort_by=[field, id]` ordering. Sorting client-side
// means switching the sort doesn't refetch the whole unscheduled list.
const SIDEBAR_COMPARATORS: Record<SortableSidebarSort, (a: TaskResponse, b: TaskResponse) => number> = {
	'priority:desc': (a, b) => b.priority - a.priority,
	'priority:asc': (a, b) => a.priority - b.priority,
	'title:asc': (a, b) => a.title.localeCompare(b.title),
	'title:desc': (a, b) => b.title.localeCompare(a.title),
	'created:desc': (a, b) => createdAt(b) - createdAt(a),
	'created:asc': (a, b) => createdAt(a) - createdAt(b),
	'percent_done:desc': (a, b) => b.percent_done - a.percent_done,
	'percent_done:asc': (a, b) => a.percent_done - b.percent_done,
}

function hasAnyDate(task: TaskResponse): boolean {
	return Boolean(
		parseDateOrNull(task.start_date)
		?? parseDateOrNull(task.end_date)
		?? parseDateOrNull(task.due_date),
	)
}

// The api clears a date when it is sent empty (taskWriteBody maps '' to the
// api's zero date); an omitted field is left untouched by the patch.
function dateForApi(date: Date | null): string {
	return date === null ? '' : date.toISOString()
}

export function usePlannerTasks(
	range: Ref<PlannerRange>,
	sidebarFilter: Ref<TaskFilterParams>,
	sidebarSort: Ref<PlannerSidebarSort>,
	overdueEnabled: Ref<boolean>,
) {
	const authStore = useAuthStore()
	const queryClient = useQueryClient()
	const timezone = computed(() => authStore.settings.timezone)

	const gridQuery = useQuery(computed(() => {
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

		return allTasksQuery({
			params: {
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
				filter_timezone: timezone.value,
				q: '',
				expand: ['subtasks'],
			},
		})
	}))

	const sidebarQuery = useQuery(computed(() => {
		// Combine the user's filter (already API-form from the Filters component)
		// with done=false. The filter grammar can't express "date is null", so the
		// unscheduled partition below keeps only tasks lacking a start/end/due.
		const userFilter = sidebarFilter.value.filter?.trim()
		const filter = userFilter ? `(${userFilter}) && done = false` : 'done = false'

		return allTasksQuery({
			params: {
				filter,
				// The sidebar's own null-date filtering happens client-side, so
				// include_nulls stays the user's choice from the filter popup — the
				// backend applies it to every condition of their filter.
				filter_include_nulls: sidebarFilter.value.filter_include_nulls ?? false,
				filter_timezone: timezone.value,
				q: sidebarFilter.value.q ?? '',
				expand: ['subtasks'],
			},
		})
	}))

	const overdueQuery = useQuery(computed(() => {
		// Superset fetch: any not-done task dated before today. The precise
		// "overdue" definition (a schedule reaching into today or later isn't
		// overdue) mixes the three date fields, which the filter grammar can't
		// express, so it is applied client-side via isOverdue below.
		const cutoff = isoToKebabDate(overdueCutoff().toISOString())
		return {
			...allTasksQuery({
				params: {
					filter: `(due_date < "${cutoff}" || end_date < "${cutoff}" || start_date < "${cutoff}") && done = false`,
					filter_include_nulls: false,
					filter_timezone: timezone.value,
					q: '',
					expand: ['subtasks'],
				},
			}),
			enabled: overdueEnabled.value,
		}
	}))

	// One pool for all three lists. The query cache patches an edited task inside
	// whichever list already holds it (replaceTaskEverywhere) but never moves it
	// between them, so partitioning the union client-side is what re-files a task
	// when an edit elsewhere adds or clears its dates.
	const tasksById = computed(() => {
		const map = new Map<TaskResponse['id'], TaskResponse>()
		for (const list of [gridQuery.data.value, sidebarQuery.data.value, overdueQuery.data.value]) {
			for (const task of list ?? []) {
				map.set(task.id, task)
			}
		}
		return map
	})

	const gridTasks = computed(() => {
		const map = new Map<TaskResponse['id'], TaskResponse>()
		for (const task of tasksById.value.values()) {
			if (hasAnyDate(task)) {
				map.set(task.id, task)
			}
		}
		return map
	})

	// 'random' keys each task once so the order is stable across recomputes and
	// a newly added task gets a slot without reshuffling the rest; re-selecting
	// 'random' clears the keys for a fresh shuffle.
	const randomKeys = new Map<TaskResponse['id'], number>()
	watch(sidebarSort, sort => {
		if (sort === 'random') {
			randomKeys.clear()
		}
	})

	function randomKey(id: TaskResponse['id']): number {
		let key = randomKeys.get(id)
		if (key === undefined) {
			key = Math.random()
			randomKeys.set(id, key)
		}
		return key
	}

	// Truly unscheduled = no start, end or due date. Due-only tasks already
	// render in the all-day row, so keep them out of the sidebar.
	const unscheduledTasks = computed(() =>
		[...tasksById.value.values()].filter(task => !task.done && !hasAnyDate(task)),
	)

	const sidebarTasks = computed<TaskResponse[]>(() => {
		// Guard against a stale/garbage stored value.
		const sort = PLANNER_SIDEBAR_SORTS.includes(sidebarSort.value) ? sidebarSort.value : DEFAULT_PLANNER_SIDEBAR_SORT
		if (sort === 'none') {
			return unscheduledTasks.value
		}
		if (sort === 'random') {
			return [...unscheduledTasks.value].sort((a, b) => randomKey(a.id) - randomKey(b.id))
		}
		const compare = SIDEBAR_COMPARATORS[sort]
		// id desc as the final tiebreaker so the chosen column drives the order.
		return [...unscheduledTasks.value].sort((a, b) => compare(a, b) || b.id - a.id)
	})

	// Additive listing of grid tasks whose anchor already passed, not an
	// alternative to being on the grid.
	const overdueTasks = computed<TaskResponse[]>(() => {
		if (!overdueEnabled.value) {
			return []
		}
		return [...tasksById.value.values()]
			.filter(task => isOverdue(task))
			.sort((a, b) => (overdueAnchor(a)?.getTime() ?? 0) - (overdueAnchor(b)?.getTime() ?? 0))
	})

	const isLoading = computed(() =>
		gridQuery.isFetching.value || sidebarQuery.isFetching.value || overdueQuery.isFetching.value,
	)
	const loadError = computed(() =>
		gridQuery.isError.value || sidebarQuery.isError.value || overdueQuery.isError.value,
	)

	// The shared mutation applies the change to every cached list up front and
	// rolls it back (with an error toast) if the request fails.
	const updateMutation = useUpdateTaskMutation(true)

	async function patchDates({id, ...dates}: PlannerTaskDates) {
		// Only send the fields the caller actually set: the optimistic update
		// merges this object over the cached task, and a key present as undefined
		// would blank that date until the server response restored it.
		const payload: Task & {id: number} = {id}
		for (const field of ['start_date', 'end_date', 'due_date'] as const) {
			const date = dates[field]
			if (date !== undefined) {
				payload[field] = dateForApi(date)
			}
		}
		await updateMutation.mutateAsync(payload)
		success({message: i18n.global.t('planner.saved')})
	}

	async function updateTask(dates: PlannerTaskDates) {
		if (!tasksById.value.has(dates.id)) {
			return
		}
		await patchDates(dates)
	}

	// Schedule a freshly created task (not in any planner list yet) into the
	// painted slot. Used by the create-by-gesture flow, where AddTask creates a
	// dateless task that we then place.
	async function scheduleTask(taskId: number, dates: {start_date: Date | null, end_date: Date | null}) {
		await patchDates({id: taskId, ...dates})
	}

	// Pull the lists back from the server — needed after creating tasks, which
	// no cached planner list contains yet.
	async function refresh() {
		await queryClient.invalidateQueries({queryKey: taskKeys.allLists, refetchType: 'active'})
	}

	return {
		gridTasks,
		sidebarTasks,
		overdueTasks,
		isLoading,
		loadError,
		updateTask,
		scheduleTask,
		refresh,
	}
}
