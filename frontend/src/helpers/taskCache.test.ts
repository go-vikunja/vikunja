import {describe, it, expect, beforeEach} from 'vitest'

import type {ITask} from '@/modelTypes/ITask'
import {
	clearTaskCache,
	deleteCachedTask,
	getCachedTask,
	invalidateCachedTask,
	setCachedTask,
} from './taskCache'

const task = (id: number) => Promise.resolve({id} as ITask)

describe('taskCache', () => {
	beforeEach(() => {
		clearTaskCache()
	})

	it('invalidates a single id, clears all of them', () => {
		setCachedTask(1, task(1))
		setCachedTask(2, task(2))

		invalidateCachedTask(1)
		expect(getCachedTask(1)).toBeUndefined()
		expect(getCachedTask(2)).toBeDefined()

		clearTaskCache()
		expect(getCachedTask(2)).toBeUndefined()
	})

	it('does not delete an entry that was replaced meanwhile', () => {
		const stale = task(1)
		const fresh = task(1)
		setCachedTask(1, stale)
		setCachedTask(1, fresh)

		deleteCachedTask(1, stale)

		expect(getCachedTask(1)).toBe(fresh)
	})
})
