import {computed, readonly, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {klona} from 'klona/lite'

import {findById, findIndexById} from '@/helpers/utils'

import BucketService from '@/services/bucket'
import TaskCollectionService, {type TaskFilterParams} from '@/services/taskCollection'

import {setModuleLoading} from '@/stores/helper'

import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'
import type {IBucket} from '@/modelTypes/IBucket'
import {useAuthStore} from '@/stores/auth'
import type {IProjectView} from '@/modelTypes/IProjectView'
import {useBaseStore} from '@/stores/base'

const TASKS_PER_BUCKET = 25

function getTaskIndicesById(buckets: IBucket[], taskId: ITask['id']) {
	let taskIndex
	const bucketIndex = buckets.findIndex(({tasks}) => {
		taskIndex = findIndexById(tasks, taskId)
		return taskIndex !== -1
	})

	return {
		bucketIndex: bucketIndex !== -1 ? bucketIndex : null,
		taskIndex: taskIndex !== -1 ? taskIndex : null,
	}
}

/**
 * This store is intended to hold the currently active kanban view.
 * It should hold only the current buckets.
 */
export const useKanbanStore = defineStore('kanban', () => {
	const authStore = useAuthStore()
	const baseStore = useBaseStore()

	const buckets = ref<IBucket[]>([])
	const projectId = ref<IProject['id']>(0)
	const bucketLoading = ref<{ [id: IBucket['id']]: boolean }>({})
	const taskPagesPerBucket = ref<{ [id: IBucket['id']]: number }>({})
	const allTasksLoadedForBucket = ref<{ [id: IBucket['id']]: boolean }>({})
	const isLoading = ref(false)

	const getBucketById = computed(() => (bucketId: IBucket['id']): IBucket | undefined => findById(buckets.value, bucketId))
	const getTaskById = computed(() => {
		return (id: ITask['id']) => {
			const {bucketIndex, taskIndex} = getTaskIndicesById(buckets.value, id)

			if (typeof bucketIndex === 'number' && typeof taskIndex === 'number') {
				const bucket = buckets.value[bucketIndex]
				if (bucket && bucket.tasks && typeof taskIndex === 'number' && bucket.tasks[taskIndex]) {
					return {
						bucketIndex,
						taskIndex,
						task: bucket.tasks[taskIndex],
					}
				}
			}

			return {
				bucketIndex,
				taskIndex,
				task: null,
			}
		}
	})

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setProjectId(newProjectId: IProject['id']) {
		projectId.value = Number(newProjectId)
	}

	function setBuckets(newBuckets: IBucket[]) {
		buckets.value = newBuckets
		newBuckets.forEach(b => {
			taskPagesPerBucket.value[b.id] = 1
			allTasksLoadedForBucket.value[b.id] = false
		})
	}

	function addBucket(bucket: IBucket) {
		buckets.value.push(bucket)
	}

	function removeBucket(newBucket: IBucket) {
		const bucketIndex = findIndexById(buckets.value, newBucket.id)
		buckets.value.splice(bucketIndex, 1)
	}

	function setBucketById(newBucket: IBucket, setTasks: boolean = true) {
		const bucketIndex = findIndexById(buckets.value, newBucket.id)
		const oldBucket = buckets.value[bucketIndex]
		if (!setTasks && oldBucket) {
			newBucket.tasks = [
				...oldBucket.tasks,
			]
		}
		buckets.value[bucketIndex] = newBucket
	}

	function setBucketByIndex(
		bucketIndex: number,
		bucket: IBucket,
	) {
		buckets.value[bucketIndex] = bucket
	}

	function setTaskInBucketByIndex({
		bucketIndex,
		taskIndex,
		task,
	}: {
		bucketIndex: number,
		taskIndex: number,
		task: ITask
	}) {
		const bucket = buckets.value[bucketIndex]
		if (!bucket) {
			return
		}
		bucket.tasks[taskIndex] = task
		buckets.value[bucketIndex] = bucket
	}

	function setTaskInBucket(task: ITask) {
		// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
		if (buckets.value.length === 0) {
			return
		}

		let found = false

		const findAndUpdate = (b: number) => {
			const currentBucket = buckets.value[b]
			if (!currentBucket || !currentBucket.tasks) {
				return
			}

			for (const t in currentBucket.tasks) {
				const taskAtIndex = currentBucket.tasks[t]
				if (taskAtIndex && taskAtIndex.id === task.id) {
					const bucket = buckets.value[b]
					if (!bucket) {
						return
					}
					bucket.tasks[t] = task
					buckets.value[b] = bucket
					found = true
					return
				}
			}
		}

		for (const b in buckets.value) {
			findAndUpdate(Number(b))
			if (found) {
				return
			}
		}
	}
	
	// This function is an exact clone of the logic in the api
	function getDefaultBucketId(view: IProjectView): IBucket['id'] {
		if (view.defaultBucketId) {
			return view.defaultBucketId
		}

		const firstBucket = buckets.value[0]
		if (!firstBucket) {
			throw new Error('No buckets available')
		}
		return firstBucket.id
	}
	
	function ensureTaskIsInCorrectBucket(task: ITask) {
		if (buckets.value.length === 0) {
			return
		}
		
		const {bucketIndex} = getTaskIndicesById(buckets.value, task.id)
		if (bucketIndex === null) return
		const currentTaskBucket = buckets.value[bucketIndex]
		if (!currentTaskBucket) return

		const currentView = baseStore.currentProject?.views.find(v => v.id === baseStore.currentProjectViewId)
		if(!currentView) return
		
		// If the task is done, make sure it is in the done bucket
		if (task.done && currentView.doneBucketId !== 0 && currentTaskBucket.id !== currentView.doneBucketId) {
			moveTaskToBucket(task, currentView.doneBucketId)
		}

		// If the task is not done but was in the done bucket before, move it to the default bucket
		if(!task.done && currentView.doneBucketId !== 0 && currentTaskBucket.id === currentView.doneBucketId) {
			const defaultBucketId = getDefaultBucketId(currentView as IProjectView)
			moveTaskToBucket(task, defaultBucketId)
		}
		
		setTaskInBucket(task)
	}
	
	function moveTaskToBucket(task: ITask, bucketId: IBucket['id']) {
		const {bucketIndex} = getTaskIndicesById(buckets.value, task.id)
		if (bucketIndex === null) return
		const currentTaskBucket = buckets.value[bucketIndex]
		if (typeof currentTaskBucket === 'undefined' || currentTaskBucket.id === bucketId) {
			return
		}		
		removeTaskInBucket(task)
		task.bucketId = bucketId
		addTaskToBucket(task)
	}

	function addTaskToBucket(task: ITask) {
		const bucketIndex = findIndexById(buckets.value, task.bucketId)
		const oldBucket = buckets.value[bucketIndex]
		if (!oldBucket || bucketIndex === -1) {
			return
		}
		const newBucket = {
			...oldBucket,
			count: (oldBucket.count || 0) + 1,
			tasks: [
				task,
				...oldBucket.tasks,
			],
		}
		buckets.value[bucketIndex] = newBucket
	}

	function addTasksToBucket(tasks: ITask[], bucketId: IBucket['id']) {
		const bucketIndex = findIndexById(buckets.value, bucketId)
		const oldBucket = buckets.value[bucketIndex]
		if (!oldBucket || bucketIndex === -1) {
			return
		}
		const newBucket = {
			...oldBucket,
			tasks: [
				...oldBucket.tasks,
				...tasks,
			],
		}
		buckets.value[bucketIndex] = newBucket
	}

	function removeTaskInBucket(task: ITask) {
		// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
		if (buckets.value.length === 0) {
			return
		}

		const {bucketIndex, taskIndex} = getTaskIndicesById(buckets.value, task.id)

		if (
			bucketIndex === null ||
			taskIndex === null ||
			typeof bucketIndex !== 'number' ||
			typeof taskIndex !== 'number' ||
			(buckets.value[bucketIndex]?.tasks[taskIndex]?.id !== task.id)
		) {
			return
		}

		const bucket = buckets.value[bucketIndex]
		if (bucket && bucket.tasks) {
			bucket.tasks.splice(taskIndex, 1)
			bucket.count--
		}
	}

	function setBucketLoading({bucketId, loading}: { bucketId: IBucket['id'], loading: boolean }) {
		bucketLoading.value[bucketId] = loading
	}

	function setTasksLoadedForBucketPage({bucketId, page}: { bucketId: IBucket['id'], page: number }) {
		taskPagesPerBucket.value[bucketId] = page
	}

	function setAllTasksLoadedForBucket(bucketId: IBucket['id']) {
		allTasksLoadedForBucket.value[bucketId] = true
	}

	async function loadBucketsForProject(projectId: IProject['id'], viewId: IProjectView['id'], params: any) {
		const cancel = setModuleLoading(setIsLoading)

		// Clear everything to prevent having old buckets in the project if loading the buckets from this project takes a few moments
		setBuckets([])

		const taskCollectionService = new TaskCollectionService()
		try {
			const result = await taskCollectionService.getAll({projectId, viewId} as any, {
				...params,
				per_page: TASKS_PER_BUCKET,
			})
			// Filter to only get buckets
			const newBuckets = result.filter((item): item is IBucket =>
				'project_view_id' in item || 'projectViewId' in item
			)
			setBuckets(newBuckets)
			setProjectId(projectId)
			return newBuckets
		} finally {
			cancel()
		}
	}

	async function loadNextTasksForBucket(
		projectId: IProject['id'],
		viewId: IProjectView['id'],
		ps: TaskFilterParams,
		bucketId: IBucket['id'],
	) {
		const isLoading = bucketLoading.value[bucketId] ?? false
		if (isLoading) {
			return
		}

		const page = (taskPagesPerBucket.value[bucketId] ?? 1) + 1

		const alreadyLoaded = allTasksLoadedForBucket.value[bucketId] ?? false
		if (alreadyLoaded) {
			return
		}

		const cancel = setModuleLoading(setIsLoading)
		setBucketLoading({bucketId: bucketId, loading: true})

		const params: TaskFilterParams = JSON.parse(JSON.stringify(ps))

		params.sort_by = ['position']
		params.order_by = ['asc']
		params.filter = `${params.filter === '' ? '' : params.filter + ' && '}bucket_id = ${bucketId}`
		params.filter_timezone = authStore.settings.timezone
		params.per_page = TASKS_PER_BUCKET

		const taskService = new TaskCollectionService()
		try {
			const result = await taskService.getAll({projectId, viewId} as any, params, page)
			// Filter to only get tasks, not buckets
			const tasks = result.filter((item): item is ITask =>
				!('project_view_id' in item) && !('projectViewId' in item)
			)
			addTasksToBucket(tasks, bucketId)
			setTasksLoadedForBucketPage({bucketId, page})
			if (taskService.totalPages <= page) {
				setAllTasksLoadedForBucket(bucketId)
			}
			return tasks
		} finally {
			cancel()
			setBucketLoading({bucketId, loading: false})
		}
	}

	async function createBucket(bucket: IBucket) {
		const cancel = setModuleLoading(setIsLoading)

		const bucketService = new BucketService()
		try {
			const createdBucket = await bucketService.create(bucket)
			addBucket(createdBucket)
			return createdBucket
		} finally {
			cancel()
		}
	}

	async function deleteBucket({bucket, params}: { bucket: IBucket, params: any }) {
		const cancel = setModuleLoading(setIsLoading)

		const bucketService = new BucketService()
		try {
			const response = await bucketService.delete(bucket)
			removeBucket(bucket)
			// We reload all buckets because tasks are being moved from the deleted bucket
			loadBucketsForProject(bucket.projectId, bucket.projectViewId, params)
			return response
		} finally {
			cancel()
		}
	}

	async function updateBucket(updatedBucketData: Partial<IBucket>) {
		if (!updatedBucketData.id) {
			throw new Error('Bucket ID is required for update')
		}

		const cancel = setModuleLoading(setIsLoading)

		const bucketIndex = findIndexById(buckets.value, updatedBucketData.id)
		if (bucketIndex === -1) {
			throw new Error('Bucket not found')
		}
		const oldBucket = klona(buckets.value[bucketIndex]) as IBucket

		const updatedBucket: IBucket = {
			...oldBucket,
			...updatedBucketData,
		} as IBucket

		setBucketByIndex(bucketIndex, updatedBucket)

		const bucketService = new BucketService()
		try {
			const returnedBucket = await bucketService.update(updatedBucket)
			setBucketByIndex(bucketIndex, returnedBucket)
			return returnedBucket
		} catch (e) {
			// restore original state
			setBucketByIndex(bucketIndex, oldBucket)

			throw e
		} finally {
			cancel()
		}
	}

	return {
		buckets,
		isLoading: readonly(isLoading),

		getBucketById,
		getTaskById,

		setBuckets,
		setBucketById,
		setTaskInBucketByIndex,
		setTaskInBucket,
		addTaskToBucket,
		removeTaskInBucket,
		moveTaskToBucket,
		loadBucketsForProject,
		loadNextTasksForBucket,
		createBucket,
		deleteBucket,
		updateBucket,
		ensureTaskIsInCorrectBucket,
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useKanbanStore, import.meta.hot))
}
