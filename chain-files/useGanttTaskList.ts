import {computed, ref, type Ref, shallowReactive, watch, type ComputedRef} from 'vue'
import {klona} from 'klona/lite'

import type {Filters} from '@/composables/useRouteFilters'
import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'

import TaskCollectionService, {type TaskFilterParams} from '@/services/taskCollection'
import TaskService from '@/services/task'

import TaskModel from '@/models/task'
import {error, success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'
import type {IProjectView} from '@/modelTypes/IProjectView'

export interface UseGanttTaskListReturn {
	tasks: Ref<Map<ITask['id'], ITask>>
	isLoading: ComputedRef<boolean>
	loadTasks: () => Promise<void>
	addTask: (task: Partial<ITask>) => Promise<ITask>
	updateTask: (task: ITaskPartialWithId) => Promise<void>
}

// FIXME: unify with general `useTaskList`
export function useGanttTaskList<F extends Filters>(
	filters: Ref<F>,
	filterToApiParams: (filters: F) => TaskFilterParams,
	viewId: Ref<IProjectView['id']>,
	loadAll: boolean = true,
	extraParams?: Ref<Record<string, unknown>>,
) : UseGanttTaskListReturn {
	const taskCollectionService = shallowReactive(new TaskCollectionService())
	const taskService = shallowReactive(new TaskService())
	const authStore = useAuthStore()

	const isLoading = computed(() => taskCollectionService.loading)

	const tasks = ref<Map<ITask['id'], ITask>>(new Map())

	async function fetchTasks(params: TaskFilterParams, page = 1): Promise<ITask[]> {

		if (params.filter_timezone === '') {
			params.filter_timezone = authStore.settings.timezone
		}

		// Merge any extra params (e.g. include_subprojects, exclude_project_ids)
		const mergedParams = extraParams?.value
			? {...params, ...extraParams.value}
			: params
		
		const tasks = await taskCollectionService.getAll({projectId: filters.value.projectId, viewId: viewId.value}, mergedParams, page) as ITask[]
		if (loadAll && page < taskCollectionService.totalPages) {
			const nextTasks = await fetchTasks(params, page + 1)
			return tasks.concat(nextTasks)
		}
		return tasks
	}

	/**
	 * Load and assign new tasks
	 * Normally there is no need to trigger this manually
	 */
	async function loadTasks() {
		const params: TaskFilterParams = filterToApiParams(filters.value)

		const loadedTasks = await fetchTasks(params)
		tasks.value = new Map()
		loadedTasks.forEach(t => tasks.value.set(t.id, t))
	}

	/**
	 * Load tasks when filters change
	 */
	watch(
		filters,
		() => loadTasks(),
		{immediate: true, deep: true},
	)

	// Sync task updates from other views (e.g. task detail modal)
	const taskStore = useTaskStore()
	watch(
		() => taskStore.lastUpdatedTask,
		(updatedTask) => {
			if (updatedTask && tasks.value.has(updatedTask.id)) {
				tasks.value.set(updatedTask.id, updatedTask)
			}
		},
	)

	async function addTask(task: Partial<ITask>) {
		const newTask = await taskService.create(new TaskModel({...task}))
		tasks.value.set(newTask.id, newTask)

		return newTask
	}

	async function updateTask(task: ITaskPartialWithId) {
		const oldTask = klona(tasks.value.get(task.id))

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

			// Check for date cascade: if start or end date changed, check for downstream chain tasks
			const startChanged = oldTask.startDate?.toString() !== newTask.startDate?.toString()
			const endChanged = oldTask.endDate?.toString() !== newTask.endDate?.toString()

			if (startChanged || endChanged) {
				await checkCascadeDownstream(updatedTask, oldTask)
			}
		} catch (_) {
			error('Something went wrong saving the task')
			// roll back changes
			tasks.value.set(task.id, oldTask)
		}
	}

	async function checkCascadeDownstream(updatedTask: ITask, oldTask: ITask) {
		try {
			// Fetch the full task with relations
			const fullTask = await taskService.get(new TaskModel({id: updatedTask.id}))
			const precedesTasks = fullTask?.relatedTasks?.precedes

			if (!precedesTasks || !Array.isArray(precedesTasks) || precedesTasks.length === 0) return

			// Calculate the delta in days
			const oldStart = oldTask.startDate ? new Date(oldTask.startDate).getTime() : 0
			const newStart = updatedTask.startDate ? new Date(updatedTask.startDate).getTime() : 0
			if (oldStart === 0 || newStart === 0) return

			const deltaDays = Math.round((newStart - oldStart) / (1000 * 60 * 60 * 24))
			if (deltaDays === 0) return

			const direction = deltaDays > 0 ? 'forward' : 'back'
			const absDays = Math.abs(deltaDays)

			// Prompt user
			const confirmed = window.confirm(
				`This task is part of a chain. Shift ${precedesTasks.length} downstream task(s) ${absDays} day(s) ${direction}?`,
			)

			if (!confirmed) return

			// Cascade: shift all downstream tasks by the same delta
			await cascadeShiftTasks(precedesTasks, deltaDays)
		} catch (e) {
			console.error('Failed to check cascade:', e)
		}
	}

	async function cascadeShiftTasks(downstreamTasks: ITask[], deltaDays: number) {
		const deltaMs = deltaDays * 24 * 60 * 60 * 1000

		for (const downstream of downstreamTasks) {
			const shiftedTask: Record<string, any> = {id: downstream.id}

			if (downstream.startDate) {
				shiftedTask.startDate = new Date(new Date(downstream.startDate).getTime() + deltaMs)
			}
			if (downstream.endDate) {
				shiftedTask.endDate = new Date(new Date(downstream.endDate).getTime() + deltaMs)
			}
			if (downstream.dueDate) {
				shiftedTask.dueDate = new Date(new Date(downstream.dueDate).getTime() + deltaMs)
			}

			try {
				const updated = await taskService.update({...downstream, ...shiftedTask})
				tasks.value.set(updated.id, updated)

				// Recursively check if this task also precedes others
				try {
					const fullDownstream = await taskService.get(new TaskModel({id: updated.id}))
					const nextTasks = fullDownstream?.relatedTasks?.precedes
					if (nextTasks && Array.isArray(nextTasks) && nextTasks.length > 0) {
						await cascadeShiftTasks(nextTasks, deltaDays)
					}
				} catch {
					// No relations or fetch failed — end of chain
				}
			} catch (e) {
				console.error(`Failed to cascade task ${downstream.id}:`, e)
			}
		}
	}


	return {
		tasks,

		isLoading,
		loadTasks,

		addTask,
		updateTask,
	}
}
