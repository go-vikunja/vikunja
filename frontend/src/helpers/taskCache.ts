import {shallowRef} from 'vue'

import type {ITask} from '@/modelTypes/ITask'

const cache = new Map<number, Promise<ITask>>()

// Bumped on eviction: mounted consumers refetch, keeping the stale task visible meanwhile.
export const taskCacheVersion = shallowRef(0)

// Bumped when the authenticated identity changes: consumers must drop what they show.
export const taskCacheIdentityVersion = shallowRef(0)

export function getCachedTask(id: number): Promise<ITask> | undefined {
	return cache.get(id)
}

export function setCachedTask(id: number, task: Promise<ITask>) {
	cache.set(id, task)
}

// Takes the promise so a late rejection cannot evict a newer entry for the same id.
export function deleteCachedTask(id: number, task: Promise<ITask>) {
	if (cache.get(id) === task) {
		cache.delete(id)
	}
}

export function invalidateCachedTask(id: number) {
	cache.delete(id)
	taskCacheVersion.value++
}

export function clearTaskCache() {
	cache.clear()
	taskCacheIdentityVersion.value++
}
