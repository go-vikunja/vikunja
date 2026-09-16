import {describe, it, expect, beforeEach, vi} from 'vitest'

const get = vi.fn()
vi.mock('@/services/task', () => ({
	default: class {
		get = get
	},
}))

import {fetchTaskById} from './fetchTaskById'
import {clearTaskCache, invalidateCachedTask} from './taskCache'

function httpError(status: number) {
	return Object.assign(new Error(`http ${status}`), {response: {status}})
}

describe('fetchTaskById', () => {
	beforeEach(() => {
		get.mockReset()
		clearTaskCache()
	})

	it('fetches a task once for concurrent requests', async () => {
		get.mockResolvedValue({id: 1, title: 'a'})

		const [a, b] = await Promise.all([fetchTaskById(1), fetchTaskById(1)])

		expect(get).toHaveBeenCalledTimes(1)
		expect(a).toBe(b)
	})

	it('caches 403/404 failures so callers do not retry', async () => {
		get.mockRejectedValue(httpError(404))

		await expect(fetchTaskById(1)).rejects.toThrow()
		await expect(fetchTaskById(1)).rejects.toThrow()

		expect(get).toHaveBeenCalledTimes(1)
	})

	it('does not cache transient failures', async () => {
		get.mockRejectedValueOnce(new Error('network'))
		get.mockResolvedValueOnce({id: 1, title: 'a'})

		await expect(fetchTaskById(1)).rejects.toThrow()
		await expect(fetchTaskById(1)).resolves.toEqual({id: 1, title: 'a'})

		expect(get).toHaveBeenCalledTimes(2)
	})

	it('keeps a newer entry when an older request rejects late', async () => {
		let failFirst: (e: Error) => void = () => {}
		get.mockReturnValueOnce(new Promise((_resolve, reject) => {
			failFirst = reject
		}))
		get.mockResolvedValueOnce({id: 1, title: 'fresh'})

		const first = fetchTaskById(1)
		invalidateCachedTask(1)
		const second = fetchTaskById(1)

		failFirst(new Error('network'))
		await expect(first).rejects.toThrow()
		await expect(second).resolves.toEqual({id: 1, title: 'fresh'})
		expect(fetchTaskById(1)).toBe(second)
		expect(get).toHaveBeenCalledTimes(2)
	})

	it('refetches only the invalidated id', async () => {
		get.mockImplementation(async (model: {id: number}) => ({id: model.id}))

		await fetchTaskById(1)
		await fetchTaskById(2)
		invalidateCachedTask(1)
		await fetchTaskById(1)
		await fetchTaskById(2)

		expect(get).toHaveBeenCalledTimes(3)
	})
})
