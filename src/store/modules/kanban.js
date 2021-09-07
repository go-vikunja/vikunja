import cloneDeep from 'lodash/cloneDeep'

import {findById, findIndexById} from '@/helpers/find'
import BucketService from '../../services/bucket'
import {setLoading} from '../helper'
import TaskCollectionService from '@/services/taskCollection'

const TASKS_PER_BUCKET = 25

function getTaskPosition(state, task) {
	const bucketIndex = findIndexById(state.buckets, task.bucketId)

	if (!bucketIndex) {
		throw('Bucket not found')
	}

	const bucket = state.buckets[bucketIndex]
	const taskIndex = findIndexById(bucket.tasks, task.id)

	if (!bucketIndex) {
		throw('Task not found')
	}

	return {
		bucketIndex,
		taskIndex,
	}
}

const addTaskToBucketAndSort = (state, task) => {
	const bucketIndex = findIndexById(state.buckets, task.bucketId)
	state.buckets[bucketIndex].tasks.push(task)
	state.buckets[bucketIndex].tasks.sort((a, b) => a.kanbanPosition > b.kanbanPosition ? 1 : -1)
}

/**
 * This store is intended to hold the currently active kanban view.
 * It should hold only the current buckets.
 */
export default {
	namespaced: true,

	state: () => ({
		buckets: [],
		listId: 0,
		bucketLoading: {},
		taskPagesPerBucket: {},
		allTasksLoadedForBucket: {},
	}),

	mutations: {
		setListId(state, listId) {
			state.listId = parseInt(listId)
		},

		setBuckets(state, buckets) {
			state.buckets = buckets
			buckets.forEach(b => {
				state.taskPagesPerBucket[b.id] = 1
				state.allTasksLoadedForBucket[b.id] = false
			})
		},

		addBucket(state, bucket) {
			state.buckets.push(bucket)
		},

		removeBucket(state, bucket) {
			const bucketIndex = findIndexById(state.buckets, bucket.id)
			state.buckets.splice(bucketIndex, 1)
		},

		setBucketById(state, bucket) {
			const bucketIndex = findIndexById(state.buckets, bucket.id)
			state.buckets[bucketIndex] = bucket
		},

		setBucketByIndex(state, {bucketIndex, bucket}) {
			state.buckets[bucketIndex] = bucket
		},

		setTaskInBucketByIndex(state, {bucketIndex, taskIndex, task}) {
			const bucket = state.buckets[bucketIndex]
			bucket.tasks[taskIndex] = task
			state.buckets[bucketIndex] = bucket
		},

		setTasksInBucketByBucketId(state, {bucketId, tasks}) {
			const bucketIndex = findIndexById(state.buckets, bucketId)
			state.buckets[bucketIndex] = {
				...state.buckets[bucketIndex],
				tasks,
			}
		},
		
		setTaskInBucket(state, task) {
			// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
			if (state.buckets.length === 0) {
				return
			}

			let found = false

			const findAndUpdate = b => {
				for (const t in state.buckets[b].tasks) {
					if (state.buckets[b].tasks[t].id === task.id) {
						const bucket = state.buckets[b]
						bucket.tasks[t] = task

						if (bucket.id !== task.bucketId) {
							bucket.tasks.splice(t, 1)
							addTaskToBucketAndSort(state, task)
						}

						state.buckets[b] = bucket

						found = true
						return
					}
				}
			}

			for (const b in state.buckets) {
				if (state.buckets[b].id === task.bucketId) {
					findAndUpdate(b)
					if (found) {
						return
					}
				}
			}

			for (const b in state.buckets) {
				findAndUpdate(b)
				if (found) {
					return
				}
			}
		},

		addTaskToBucket(state, task) {
			const bucketIndex = findIndexById(state.buckets, task.bucketId)
			const oldBucket = state.buckets[bucketIndex]
			const newBucket = {
				...oldBucket,
				tasks: [
					...oldBucket.tasks,
					task,
				],
			}
			state.buckets[bucketIndex] = newBucket
		},

		addTasksToBucket(state, {tasks, bucketId}) {
			const bucketIndex = findIndexById(state.buckets, bucketId)
			const oldBucket = state.buckets[bucketIndex]
			const newBucket = {
				...oldBucket,
				tasks: [
					...oldBucket.tasks,
					tasks,
				],
			}
			state.buckets[bucketIndex] = newBucket
		},

		removeTaskInBucket(state, task) {
			// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
			if (state.buckets.length === 0) {
				return
			}

			const { bucketIndex, taskIndex } = getTaskPosition(state, task)
			
			state.buckets[bucketIndex].tasks.splice(taskIndex, 1)
		},

		setBucketLoading(state, {bucketId, loading}) {
			state.bucketLoading[bucketId] = loading
		},

		setTasksLoadedForBucketPage(state, {bucketId, page}) {
			state.taskPagesPerBucket[bucketId] = page
		},

		setAllTasksLoadedForBucket(state, bucketId) {
			state.allTasksLoadedForBucket[bucketId] = true
		},
	},

	getters: {
		getBucketById(state) {
			return (bucketId) => findById(state.buckets, bucketId)
		},
		getTaskById: state => id => {
			for (const b in state.buckets) {
				for (const t in state.buckets[b].tasks) {
					if (state.buckets[b].tasks[t].id === id) {
						return {
							bucketIndex: b,
							taskIndex: t,
							task: state.buckets[b].tasks[t],
						}
					}
				}
			}
			return {
				bucketIndex: null,
				taskIndex: null,
				task: null,
			}
		},
	},

	actions: {
		loadBucketsForList(ctx, {listId, params}) {
			const cancel = setLoading(ctx, 'kanban')

			// Clear everything to prevent having old buckets in the list if loading the buckets from this list takes a few moments
			ctx.commit('setBuckets', [])

			params.per_page = TASKS_PER_BUCKET

			const bucketService = new BucketService()
			return bucketService.getAll({listId: listId}, params)
				.then(r => {
					ctx.commit('setBuckets', r)
					ctx.commit('setListId', listId)
					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
				.finally(() => {
					cancel()
				})
		},

		loadNextTasksForBucket(ctx, {listId, ps = {}, bucketId}) {
			const bucketIndex = findIndexById(ctx.state.buckets, bucketId)

			const isLoading = ctx.state.bucketLoading[bucketIndex] ?? false
			if (isLoading) {
				return Promise.resolve()
			}

			const page = (ctx.state.taskPagesPerBucket[bucketIndex] ?? 1) + 1

			const alreadyLoaded = ctx.state.allTasksLoadedForBucket[bucketIndex] ?? false
			if (alreadyLoaded) {
				return Promise.resolve()
			}

			const cancel = setLoading(ctx, 'kanban')
			ctx.commit('setBucketLoading', {bucketId: bucketId, loading: true})

			const params = cloneDeep(ps)

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
			return taskService.getAll({listId: listId}, params, page)
				.then(r => {
					ctx.commit('addTasksToBucket', {tasks: r, bucketId: bucketId})
					ctx.commit('setTasksLoadedForBucketPage', {bucketId, page})
					if (taskService.totalPages <= page) {
						ctx.commit('setAllTasksLoadedForBucket', bucketId)
					}
					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
				.finally(() => {
					cancel()
					ctx.commit('setBucketLoading', {bucketId: bucketId, loading: false})
				})
		},
		createBucket(ctx, bucket) {
			const cancel = setLoading(ctx, 'kanban')

			const bucketService = new BucketService()
			return bucketService.create(bucket)
				.then(r => {
					ctx.commit('addBucket', r)
					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
				.finally(() => {
					cancel()
				})
		},
		deleteBucket(ctx, {bucket, params}) {
			const cancel = setLoading(ctx, 'kanban')

			const bucketService = new BucketService()
			return bucketService.delete(bucket)
				.then(r => {
					ctx.commit('removeBucket', bucket)
					// We reload all buckets because tasks are being moved from the deleted bucket
					ctx.dispatch('loadBucketsForList', {listId: bucket.listId, params: params})
					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
				.finally(() => {
					cancel()
				})
		},
		updateBucket(ctx, bucket) {
			const cancel = setLoading(ctx, 'kanban')

			const bucketService = new BucketService()
			return bucketService.update(bucket)
				.then(r => {
					const bi = findById(ctx.state.buckets, r.id)
					const bucket = r
					bucket.tasks = ctx.state.buckets[bi].tasks
					ctx.commit('setBucketByIndex', {bucketIndex: bi, bucket})
					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
				.finally(() => {
					cancel()
				})
		},
	},
}