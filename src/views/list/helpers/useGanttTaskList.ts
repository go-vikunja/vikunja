import {computed, ref, shallowReactive, watchEffect} from 'vue'
import cloneDeep from 'lodash.clonedeep'

import type {Filter} from '@/composables/useRouteFilter'
import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'

import TaskCollectionService, { type GetAllTasksParams } from '@/services/taskCollection'
import TaskService from '@/services/task'

import TaskModel from '@/models/task'
import {error, success} from '@/message'

// FIXME: unify with general `useTaskList`
export function useGanttTaskList<F extends Filter>(
	filters: F,
	filterToApiParams: (filters: F) => GetAllTasksParams,
	options: {
		loadAll?: boolean,
	} = {
		loadAll: true,
	}) {
	const taskCollectionService = shallowReactive(new TaskCollectionService())
	const taskService = shallowReactive(new TaskService())

	const isLoading = computed(() => taskCollectionService.loading)

	const tasks = ref<Map<ITask['id'], ITask>>(new Map())

	async function fetchTasks(params: GetAllTasksParams, page = 1): Promise<ITask[]> {
		const tasks = await taskCollectionService.getAll({listId: filters.listId}, params, page) as ITask[]
		if (options.loadAll && page < taskCollectionService.totalPages) {
			const nextTasks = await fetchTasks(params, page + 1)
			return tasks.concat(nextTasks)
		}
		return tasks
	}

	async function loadTasks(filters: F) {
		const params: GetAllTasksParams = filterToApiParams(filters)

		const loadedTasks = await fetchTasks(params)
		tasks.value = new Map()
		loadedTasks.forEach(t => tasks.value.set(t.id, t))
	}

	watchEffect(() => loadTasks(filters))

	async function addTask(task: Partial<ITask>) {
		const newTask = await taskService.create(new TaskModel({...task}))
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
		tasks,

		isLoading,
		addTask,
		updateTask,
	}
}