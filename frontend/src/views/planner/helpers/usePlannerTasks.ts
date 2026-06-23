import {computed, ref, shallowReactive, watch, type Ref} from 'vue'
import {klona} from 'klona/lite'

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

export function usePlannerTasks(range: Ref<PlannerRange>, sidebarFilter: Ref<TaskFilterParams>) {
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
		const to = isoToKebabDate(range.value.to.toISOString())

		const params: TaskFilterParams = {
			sort_by: ['start_date', 'id'],
			order_by: ['asc', 'desc'],
			filter: '(' +
				`(start_date >= "${from}" && start_date <= "${to}") || ` +
				`(end_date >= "${from}" && end_date <= "${to}") || ` +
				`(due_date >= "${from}" && due_date <= "${to}") || ` +
				`(start_date <= "${from}" && end_date >= "${to}")` +
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

		const params: TaskFilterParams = {
			sort_by: ['due_date', 'id'],
			order_by: ['asc', 'desc'],
			filter,
			filter_include_nulls: true,
			filter_timezone: authStore.settings.timezone,
			s: sidebarFilter.value.s ?? '',
			expand: 'subtasks',
		}

		// Truly unscheduled = no start, end or due date. Due-only tasks already
		// render in the all-day row, so keep them out of the sidebar.
		const loaded = await fetchAll(sidebarService, params)
		sidebarTasks.value = loaded.filter(task => !task.startDate && !task.endDate && !task.dueDate)
	}

	function load() {
		return Promise.all([loadGrid(), loadSidebar()])
	}

	watch(range, () => loadGrid(), {immediate: true, deep: true})
	watch(sidebarFilter, () => loadSidebar(), {immediate: true, deep: true})

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

	return {
		gridTasks,
		sidebarTasks,
		isLoading,
		load,
		updateTask,
	}
}
