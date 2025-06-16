import {computed, ref, type Ref, shallowReactive, watch, type ComputedRef} from 'vue'
import {klona} from 'klona/lite'

import type {Filters} from '@/composables/useRouteFilters'
import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'

import TaskCollectionService, {type TaskFilterParams} from '@/services/taskCollection'
import TaskService from '@/services/task'

import TaskModel from '@/models/task'
import {error, success} from '@/message'
import {useAuthStore} from '@/stores/auth'
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
		
		const tasks = await taskCollectionService.getAll({projectId: filters.value.projectId, viewId: viewId.value}, params, page) as ITask[]
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
		} catch (_) {
			error('Something went wrong saving the task')
			// roll back changes
			tasks.value.set(task.id, oldTask)
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
