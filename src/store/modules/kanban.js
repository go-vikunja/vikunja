import Vue from 'vue'

import BucketService from '../../services/bucket'
import {filterObject} from '../../helpers/filterObject'
import {setLoading} from '../helper'

/**
 * This store is intended to hold the currently active kanban view.
 * It should hold only the current buckets.
 */
export default {
	namespaced: true,
	state: () => ({
		buckets: [],
		listId: 0,
	}),
	mutations: {
		setListId(state, listId) {
			state.listId = listId
		},
		setBuckets(state, buckets) {
			state.buckets = buckets
		},
		addBucket(state, bucket) {
			state.buckets.push(bucket)
		},
		removeBucket(state, bucket) {
			for (const b in state.buckets) {
				if (state.buckets[b].id === bucket.id) {
					state.buckets.splice(b, 1)
				}
			}
		},
		setBucketById(state, bucket) {
			for (const b in state.buckets) {
				if (state.buckets[b].id === bucket.id) {
					Vue.set(state.buckets, b, bucket)
					return
				}
			}
		},
		setBucketByIndex(state, {bucketIndex, bucket}) {
			Vue.set(state.buckets, bucketIndex, bucket)
		},
		setTaskInBucketByIndex(state, {bucketIndex, taskIndex, task}) {
			const bucket = state.buckets[bucketIndex]
			bucket.tasks[taskIndex] = task
			Vue.set(state.buckets, bucketIndex, bucket)
		},
		setTaskInBucket(state, task) {
			// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
			if (state.buckets.length === 0) {
				return
			}

			for (const b in state.buckets) {
				if (state.buckets[b].id === task.bucketId) {
					for (const t in state.buckets[b].tasks) {
						if (state.buckets[b].tasks[t].id === task.id) {
							const bucket = state.buckets[b]
							bucket.tasks[t] = task
							Vue.set(state.buckets, b, bucket)
							return
						}
					}
					return
				}
			}
		},
		addTaskToBucket(state, task) {
			const bi = filterObject(state.buckets, b => b.id === task.bucketId)
			state.buckets[bi].tasks.push(task)
		},
		removeTaskInBucket(state, task) {
			// If this gets invoked without any tasks actually loaded, we can save the hassle of finding the task
			if (state.buckets.length === 0) {
				return
			}

			for (const b in state.buckets) {
				if (state.buckets[b].id === task.bucketId) {
					for (const t in state.buckets[b].tasks) {
						if (state.buckets[b].tasks[t].id === task.id) {
							const bucket = state.buckets[b]
							bucket.tasks.splice(t, 1)
							Vue.set(state.buckets, b, bucket)
							return
						}
					}
					return
				}
			}
		}
	},
	getters: {
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
		loadBucketsForList(ctx, listId) {
			const cancel = setLoading(ctx)

			// Clear everything to prevent having old buckets in the list if loading the buckets from this list takes a few moments
			ctx.commit('setBuckets', [])

			const bucketService = new BucketService()
			return bucketService.getAll({listId: listId})
				.then(r => {
					ctx.commit('setBuckets', r)
					ctx.commit('setListId', listId)
					return Promise.resolve()
				})
				.catch(e => {
					return Promise.reject(e)
				})
				.finally(() => {
					cancel()
				})
		},
		createBucket(ctx, bucket) {
			const cancel = setLoading(ctx)

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
		deleteBucket(ctx, bucket) {
			const cancel = setLoading(ctx)

			const bucketService = new BucketService()
			return bucketService.delete(bucket)
				.then(r => {
					ctx.commit('removeBucket', bucket)
					// We reload all buckets because tasks are being moved from the deleted bucket
					ctx.dispatch('loadBucketsForList', bucket.listId)
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
			const cancel = setLoading(ctx)

			const bucketService = new BucketService()
			return bucketService.update(bucket)
				.then(r => {
					const bi = filterObject(ctx.state.buckets, b => b.id === r.id)
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