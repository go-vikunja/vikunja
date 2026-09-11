import {computed, readonly, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {klona} from 'klona/lite'

import {findById, findIndexById} from '@/helpers/utils'

import BucketService from '@/services/bucket'
import TaskCollectionService from '@/services/taskCollection'
import type {TaskFilterParams} from '@/types/TaskFilterParams'

import {setModuleLoading} from '@/stores/helper'

import type {ITask} from '@/modelTypes/ITask'
import type {IBucket} from '@/modelTypes/IBucket'
import {useAuthStore} from '@/stores/auth'
import type {ProjectView} from '@/client/generated'
import {useBaseStore} from '@/stores/base'
import {useProjectNavigation} from '@/composables/useProjectNavigation'
import {useCurrentProject} from '@/composables/useCurrentProject'

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
	const projectNavigation = useProjectNavigation()
	const {currentProject} = useCurrentProject()

	const buckets = ref<IBucket[]>([])
	const projectId = ref(0)
	const bucketLoading = ref<{ [id: IBucket['id']]: boolean }>({})
	const taskPagesPerBucket = ref<{ [id: IBucket['id']]: number }>({})
	const allTasksLoadedForBucket = ref<{ [id: IBucket['id']]: boolean }>({})
	const isLoading = ref(false)

	const getBucketById = computed(() => (bucketId: IBucket['id']): IBucket | undefined => findById(buckets.value, bucketId))
	const getTaskById = computed(() => {
		return (id: ITask['id']) => {
			const {bucketIndex, taskIndex} = getTaskIndicesById(buckets.value, id)

			return {
				bucketIndex,
				taskIndex,
				task: bucketIndex !== null && taskIndex !== null && buckets.value[bucketIndex]?.tasks?.[taskIndex] || null,
			}
		}
	})

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setProjectId(newProjectId: number) {
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
		if (bucketIndex === -1) {
			return
		}

		if (!setTasks) {
			newBucket.tasks = [
				...buckets.value[bucketIndex].tasks,
			]
		}
		buckets.value[bucketIndex] = newBucket
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
			for (const [t, taskInBucket] of buckets.value[b].tasks.entries()) {
				if (taskInBucket.id === task.id) {
					const bucket = buckets.value[b]
					bucket.tasks[t] = task

					buckets.value[b] = bucket

					found = true
					return
				}
			}
		}

		for (let b = 0; b < buckets.value.length; b++) {
			findAndUpdate(b)
			if (found) {
				return
			}
		}
	}
	
	// This function is an exact clone of the logic in the api
	function getDefaultBucketId(view: ProjectView): IBucket['id'] {
		if (view.default_bucket_id) {
			return view.default_bucket_id
		}
		
		return buckets.value[0]?.id
	}
	
	function ensureTaskIsInCorrectBucket(task: ITask) {
		if (buckets.value.length === 0) {
			return
		}
		
		const {bucketIndex} = getTaskIndicesById(buckets.value, task.id)
		if (bucketIndex === null) return
		const currentTaskBucket = buckets.value[bucketIndex]
		
		const currentView = currentProject.value?.views.find(view => view.id === baseStore.currentProjectViewId)
		if(typeof currentView === 'undefined') return
		const doneBucketId = currentView.done_bucket_id ?? 0
		
		// If the task is done, make sure it is in the done bucket
		if (task.done && doneBucketId !== 0 && currentTaskBucket.id !== doneBucketId) {
			moveTaskToBucket(task, doneBucketId)
		}

		// If the task is not done but was in the done bucket before, move it to the default bucket
		if(!task.done && doneBucketId !== 0 && currentTaskBucket.id === doneBucketId) {
			const defaultBucketId = getDefaultBucketId(currentView)
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
		// The target bucket can belong to a kanban view other than the loaded one (the task detail
		// view lets users move tasks between buckets of any view). Removing the task here would drop
		// it from the board with no bucket to put it back into.
		if (findIndexById(buckets.value, bucketId) === -1) {
			return
		}
		removeTaskInBucket(task)
		task.bucketId = bucketId
		addTaskToBucket(task)
	}

	function addTaskToBucket(task: ITask) {
		const bucketIndex = findIndexById(buckets.value, task.bucketId)
		const oldBucket = buckets.value[bucketIndex]
		if (typeof oldBucket === 'undefined') {
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
		if (typeof oldBucket === 'undefined') {
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
			(buckets.value[bucketIndex]?.tasks[taskIndex]?.id !== task.id)
		) {
			return
		}

		buckets.value[bucketIndex].tasks.splice(taskIndex, 1)
		buckets.value[bucketIndex].count--
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

	async function loadBucketsForProject(projectId: number, viewId: number, params) {
		const cancel = setModuleLoading(setIsLoading)

		// Clear everything to prevent having old buckets in the project if loading the buckets from this project takes a few moments
		setBuckets([])

		const taskCollectionService = new TaskCollectionService()
		try {
			const newBuckets = await taskCollectionService.getAll({projectId, viewId}, {
				...params,
				expand: ['comment_count', 'is_unread'],
				per_page: TASKS_PER_BUCKET,
			})
			setBuckets(newBuckets)
			setProjectId(projectId)
			return newBuckets
		} finally {
			cancel()
		}
	}

	async function loadNextTasksForBucket(
		projectId: number,
		viewId: number,
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
		params.expand = ['comment_count', 'is_unread']

		const taskService = new TaskCollectionService()
		try {
			const tasks = await taskService.getAll({projectId, viewId}, params, page)
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

	async function deleteBucket({bucket, params}: { bucket: IBucket, params }) {
		const cancel = setModuleLoading(setIsLoading)

		const bucketService = new BucketService()
		try {
			const response = await bucketService.delete(bucket)
			removeBucket(bucket)

			// Mirrors Bucket.Delete on the server, which zeroes these when they pointed at the deleted bucket.
			const view = projectNavigation.projects[bucket.projectId]?.views?.find(v => v.id === bucket.projectViewId)
			if (view && (view.default_bucket_id === bucket.id || view.done_bucket_id === bucket.id)) {
				projectNavigation.setProjectView({
					...view,
					default_bucket_id: view.default_bucket_id === bucket.id ? 0 : view.default_bucket_id,
					done_bucket_id: view.done_bucket_id === bucket.id ? 0 : view.done_bucket_id,
				})
			}

			// We reload all buckets because tasks are being moved from the deleted bucket
			loadBucketsForProject(bucket.projectId, bucket.projectViewId, params)
			return response
		} finally {
			cancel()
		}
	}

	async function updateBucket(updatedBucketData: Partial<IBucket>) {
		const bucket = findById(buckets.value, updatedBucketData.id)
		if (typeof bucket === 'undefined') {
			return
		}

		const cancel = setModuleLoading(setIsLoading)

		const oldBucket = klona(bucket)
		const updatedBucket = {
			...oldBucket,
			...updatedBucketData,
		}

		setBucketById(updatedBucket)

		const bucketService = new BucketService()
		try {
			// The board can be replaced while the request is in flight, for example when navigating to
			// another view. All writes go by id so the response never lands in another view's buckets.
			const returnedBucket = await bucketService.update(updatedBucket)
			setBucketById(returnedBucket, false)
			return returnedBucket
		} catch (e) {
			// restore original state
			setBucketById(oldBucket)

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
