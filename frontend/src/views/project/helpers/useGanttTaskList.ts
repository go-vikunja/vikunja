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
			const fullTask = await taskService.get(new TaskModel({id: updatedTask.id}))
			const precedesTasks = fullTask?.relatedTasks?.precedes
			const followsTasks = fullTask?.relatedTasks?.follows

			const hasPrecedes = precedesTasks && Array.isArray(precedesTasks) && precedesTasks.length > 0
			const hasFollows = followsTasks && Array.isArray(followsTasks) && followsTasks.length > 0

			if (!hasPrecedes && !hasFollows) return

			const oldStart = oldTask.startDate ? new Date(oldTask.startDate).getTime() : 0
			const newStart = updatedTask.startDate ? new Date(updatedTask.startDate).getTime() : 0
			if (oldStart === 0 || newStart === 0) return

			const deltaDays = Math.round((newStart - oldStart) / (1000 * 60 * 60 * 24))
			if (deltaDays === 0) return

			const movedEarlier = deltaDays < 0
			const absDays = Math.abs(deltaDays)
			const today = new Date()
			today.setHours(0, 0, 0, 0)

			// Upstream collision: task moved before its predecessor
			if (movedEarlier && hasFollows) {
				for (const pred of followsTasks!) {
					if (!pred.startDate) continue
					if (new Date(newStart) <= new Date(pred.startDate)) {
						// Check past-date safety
						const earliest = findEarliestDate(followsTasks!)
						if (earliest) {
							const shifted = new Date(earliest.getTime() + deltaDays * 24 * 60 * 60 * 1000)
							if (shifted < today) {
								window.alert(`Cannot shift upstream — would move tasks to ${shifted.toLocaleDateString()}, which is in the past.`)
								return
							}
						}
						const confirmedUp = window.confirm(`This task is now before its predecessor. Shift upstream task(s) ${absDays} day(s) back?`)
						if (confirmedUp) {
							await cascadeShiftTasks(followsTasks!, deltaDays, 'follows')
						}
						break
					}
				}
			}

			// Downstream cascade
			if (hasPrecedes) {
				if (movedEarlier) {
					const earliest = findEarliestDate(precedesTasks!)
					if (earliest) {
						const shifted = new Date(earliest.getTime() + deltaDays * 24 * 60 * 60 * 1000)
						if (shifted < today) {
							window.alert(`Cannot shift downstream — would move tasks to ${shifted.toLocaleDateString()}, which is in the past.`)
							return
						}
					}
				}
				const direction = deltaDays > 0 ? 'forward' : 'back'
				const confirmed = window.confirm(`Shift ${precedesTasks!.length} downstream task(s) ${absDays} day(s) ${direction}?`)
				if (confirmed) {
					await cascadeShiftTasks(precedesTasks!, deltaDays, 'precedes')
				}
			}
		} catch (e) {
			console.error('Failed to check cascade:', e)
		}
	}

	function findEarliestDate(tasks: ITask[]): Date | null {
		let earliest: Date | null = null
		for (const t of tasks) {
			if (t.startDate) {
				const d = new Date(t.startDate)
				if (!earliest || d < earliest) earliest = d
			}
		}
		return earliest
	}

	async function cascadeShiftTasks(chainTasks: ITask[], deltaDays: number, direction: 'precedes' | 'follows') {
		const deltaMs = deltaDays * 24 * 60 * 60 * 1000

		for (const t of chainTasks) {
			const shiftedTask: Record<string, any> = {id: t.id}

			if (t.startDate) {
				shiftedTask.startDate = new Date(new Date(t.startDate).getTime() + deltaMs)
			}
			if (t.endDate) {
				shiftedTask.endDate = new Date(new Date(t.endDate).getTime() + deltaMs)
			}
			if (t.dueDate) {
				shiftedTask.dueDate = new Date(new Date(t.dueDate).getTime() + deltaMs)
			}

			try {
				const updated = await taskService.update({...t, ...shiftedTask})
				tasks.value.set(updated.id, updated)

				try {
					const full = await taskService.get(new TaskModel({id: updated.id}))
					const nextTasks = full?.relatedTasks?.[direction]
					if (nextTasks && Array.isArray(nextTasks) && nextTasks.length > 0) {
						await cascadeShiftTasks(nextTasks, deltaDays, direction)
					}
				} catch {
					// End of chain
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
