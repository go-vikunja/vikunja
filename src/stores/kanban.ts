import {defineStore, acceptHMRUpdate} from 'pinia'
import cloneDeep from 'lodash.clonedeep'

import {findById, findIndexById} from '@/helpers/utils'
import {i18n} from '@/i18n'
import {success} from '@/message'

import BucketService from '@/services/bucket'
import TaskCollectionService from '@/services/taskCollection'

import {setModuleLoading} from '@/stores/helper'

import type { ITask } from '@/modelTypes/ITask'
import type { IList } from '@/modelTypes/IList'
import type { IBucket } from '@/modelTypes/IBucket'

const TASKS_PER_BUCKET = 25

function getTaskIndicesById(state: KanbanState, taskId: ITask['id']) {
	let taskIndex
	const bucketIndex = state.buckets.findIndex(({ tasks }) => {
		taskIndex = findIndexById(tasks, taskId)
		return taskIndex !== -1
	})

	return {
		bucketIndex: bucketIndex !== -1 ? bucketIndex : null,
		taskIndex: taskIndex !== -1 ? taskIndex : null,
	}	
}

const addTaskToBucketAndSort = (state: KanbanState, task: ITask) => {
	const bucketIndex = findIndexById(state.buckets, task.bucketId)
	if(typeof state.buckets[bucketIndex] === 'undefined') {
		return
	}
	state.buckets[bucketIndex].tasks.push(task)
	state.buckets[bucketIndex].tasks.sort((a, b) => a.kanbanPosition > b.kanbanPosition ? 1 : -1)
}

export interface KanbanState {
	buckets: IBucket[],
	listId: IList['id'],
	bucketLoading: {
		[id: IBucket['id']]: boolean
	},
	taskPagesPerBucket: {
		[id: IBucket['id']]: number
	},
	allTasksLoadedForBucket: {
		[id: IBucket['id']]: boolean
	},
	isLoading: boolean,
}

/**
 * This store is intended to hold the currently active kanban view.
 * It should hold only the current buckets.
 */
export const useKanbanStore = defineStore('kanban', {
	state: () : KanbanState => ({
		buckets: [],
		listId: 0,
		bucketLoading: {},
		taskPagesPerBucket: {},
		allTasksLoadedForBucket: {},
		isLoading: false,
	}),

	getters: {
		getBucketById(state) {
			return (bucketId: IBucket['id']) => findById(state.buckets, bucketId)
		},

		getTaskById(state) {
			return (id: ITask['id']) => {
				const { bucketIndex, taskIndex } = getTaskIndicesById(state, id)

				
				return {
					bucketIndex,
					taskIndex,
					task: bucketIndex && taskIndex && state.buckets[bucketIndex]?.tasks?.[taskIndex] || null,
				}
			}
		},
	},

	actions: {
		setIsLoading(isLoading: boolean) {
			this.isLoading = isLoading
		},

		setListId(listId: IList['id']) {
			this.listId = Number(listId)
		},

		setBuckets(buckets: IBucket[]) {
			this.buckets = buckets
			buckets.forEach(b => {
				this.taskPagesPerBucket[b.id] = 1
				this.allTasksLoadedForBucket[b.id] = false
			})
		},

		addBucket(bucket: IBucket) {
			this.buckets.push(bucket)
		},

		removeBucket(bucket: IBucket) {
			const bucketIndex = findIndexById(this.buckets, bucket.id)
			this.buckets.splice(bucketIndex, 1)
		},

		setBucketById(bucket: IBucket) {
			const bucketIndex = findIndexById(this.buckets, bucket.id)
			this.buckets[bucketIndex] = bucket
		},

		setBucketByIndex({
			bucketIndex,
			bucket,
		} : {
			bucketIndex: number,
			bucket: IBucket
		}) {
			this.buckets[bucketIndex] = bucket
		},

		setTaskInBucketByIndex({
			bucketIndex,
			taskIndex,
			task,
		} : {
			bucketIndex: number,
			taskIndex: number,
			task: ITask
		}) {
			const bucket = this.buckets[bucketIndex]
			bucket.tasks[taskIndex] = task
			this.buckets[bucketIndex] = bucket
		},

		setTasksInBucketByBucketId({
			bucketId,
			tasks,
		} : {
			bucketId: IBucket['id'],
			tasks: ITask[],
		}) {
			const bucketIndex = findIndexById(this.buckets, bucketId)
			this.buckets[bucketIndex] = {
				...this.buckets[bucketIndex],
				tasks,
			}
		},
		
		setTaskInBucket(task: ITask) {
			// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
			if (this.buckets.length === 0) {
				return
			}

			let found = false

			const findAndUpdate = b => {
				for (const t in this.buckets[b].tasks) {
					if (this.buckets[b].tasks[t].id === task.id) {
						const bucket = this.buckets[b]
						bucket.tasks[t] = task

						if (bucket.id !== task.bucketId) {
							bucket.tasks.splice(t, 1)
							addTaskToBucketAndSort(this, task)
						}

						this.buckets[b] = bucket

						found = true
						return
					}
				}
			}

			for (const b in this.buckets) {
				if (this.buckets[b].id === task.bucketId) {
					findAndUpdate(b)
					if (found) {
						return
					}
				}
			}

			for (const b in this.buckets) {
				findAndUpdate(b)
				if (found) {
					return
				}
			}
		},

		addTaskToBucket(task: ITask) {
			const bucketIndex = findIndexById(this.buckets, task.bucketId)
			const oldBucket = this.buckets[bucketIndex]
			const newBucket = {
				...oldBucket,
				tasks: [
					...oldBucket.tasks,
					task,
				],
			}
			this.buckets[bucketIndex] = newBucket
		},

		addTasksToBucket({tasks, bucketId}: {
			tasks: ITask[];
			bucketId: IBucket['id'];
		}) {
			const bucketIndex = findIndexById(this.buckets, bucketId)
			const oldBucket = this.buckets[bucketIndex]
			const newBucket = {
				...oldBucket,
				tasks: [
					...oldBucket.tasks,
					...tasks,
				],
			}
			this.buckets[bucketIndex] = newBucket
		},

		removeTaskInBucket(task: ITask) {
			// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
			if (this.buckets.length === 0) {
				return
			}

			const { bucketIndex, taskIndex } = getTaskIndicesById(this, task.id)

			if (
				!bucketIndex || 
				this.buckets[bucketIndex]?.id !== task.bucketId ||
				!taskIndex ||
				(this.buckets[bucketIndex]?.tasks[taskIndex]?.id !== task.id)
			) {
				return
			}
			
			this.buckets[bucketIndex].tasks.splice(taskIndex, 1)
		},

		setBucketLoading({bucketId, loading}: {bucketId: IBucket['id'], loading: boolean}) {
			this.bucketLoading[bucketId] = loading
		},

		setTasksLoadedForBucketPage({bucketId, page}: {bucketId: IBucket['id'], page: number}) {
			this.taskPagesPerBucket[bucketId] = page
		},

		setAllTasksLoadedForBucket(bucketId: IBucket['id']) {
			this.allTasksLoadedForBucket[bucketId] = true
		},

		async loadBucketsForList({listId, params}: {listId: IList['id'], params}) {
			const cancel = setModuleLoading(this)

			// Clear everything to prevent having old buckets in the list if loading the buckets from this list takes a few moments
			this.setBuckets([])

			params.per_page = TASKS_PER_BUCKET

			const bucketService = new BucketService()
			try {
				const buckets = await bucketService.getAll({listId}, params)
				this.setBuckets(buckets)
				this.setListId(listId)
				return buckets
			} finally {
				cancel()
			}
		},

		async loadNextTasksForBucket(
			{listId, ps = {}, bucketId} :
			{listId: IList['id'], ps, bucketId: IBucket['id']},
		) {
			const isLoading = this.bucketLoading[bucketId] ?? false
			if (isLoading) {
				return
			}

			const page = (this.taskPagesPerBucket[bucketId] ?? 1) + 1

			const alreadyLoaded = this.allTasksLoadedForBucket[bucketId] ?? false
			if (alreadyLoaded) {
				return
			}

			const cancel = setModuleLoading(this)
			this.setBucketLoading({bucketId: bucketId, loading: true})

			const params = JSON.parse(JSON.stringify(ps))

			params.sort_by = 'kanban_position'
			params.order_by = 'asc'

			let hasBucketFilter = false
			for (const f in params.filter_by) {
				if (params.filter_by[f] === 'bucket_id') {
					hasBucketFilter = true
					if (params.filter_value[f] !== bucketId) {
						params.filter_value[f] = bucketId
					}
					break
				}
			}

			if (!hasBucketFilter) {
				params.filter_by = [...(params.filter_by ?? []), 'bucket_id']
				params.filter_value = [...(params.filter_value ?? []), bucketId]
				params.filter_comparator = [...(params.filter_comparator ?? []), 'equals']
			}

			params.per_page = TASKS_PER_BUCKET

			const taskService = new TaskCollectionService()
			try {
				const tasks = await taskService.getAll({listId}, params, page)
				this.addTasksToBucket({tasks, bucketId: bucketId})
				this.setTasksLoadedForBucketPage({bucketId, page})
				if (taskService.totalPages <= page) {
					this.setAllTasksLoadedForBucket(bucketId)
				}
				return tasks
			} finally {
				cancel()
				this.setBucketLoading({bucketId, loading: false})
			}
		},

		async createBucket(bucket: IBucket) {
			const cancel = setModuleLoading(this)

			const bucketService = new BucketService()
			try {
				const createdBucket = await bucketService.create(bucket)
				this.addBucket(createdBucket)
				return createdBucket
			} finally {
				cancel()
			}
		},

		async deleteBucket({bucket, params}: {bucket: IBucket, params}) {
			const cancel = setModuleLoading(this)

			const bucketService = new BucketService()
			try {
				const response = await bucketService.delete(bucket)
				this.removeBucket(bucket)
				// We reload all buckets because tasks are being moved from the deleted bucket
				this.loadBucketsForList({listId: bucket.listId, params})
				return response
			} finally {
				cancel()
			}
		},

		async updateBucket(updatedBucketData: IBucket) {
			const cancel = setModuleLoading(this)

			const bucketIndex = findIndexById(this.buckets, updatedBucketData.id)
			const oldBucket = cloneDeep(this.buckets[bucketIndex])

			const updatedBucket = {
				...oldBucket,
				...updatedBucketData,
			}

			this.setBucketByIndex({bucketIndex, bucket: updatedBucket})
			
			const bucketService = new BucketService()
			try {
				const returnedBucket = await bucketService.update(updatedBucket)
				this.setBucketByIndex({bucketIndex, bucket: returnedBucket})
				return returnedBucket
			} catch(e) {
				// restore original state
				this.setBucketByIndex({bucketIndex, bucket: oldBucket})

				throw e
			} finally {
				cancel()
			}
		},

		async updateBucketTitle({ id, title }: { id: IBucket['id'], title: IBucket['title'] }) {
			const bucket = findById(this.buckets, id)

			if (bucket?.title === title) {
				// bucket title has not changed
				return
			}

			await this.updateBucket({ id, title })
			success({message: i18n.global.t('list.kanban.bucketTitleSavedSuccess')})
		},
	},
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useKanbanStore, import.meta.hot))
}