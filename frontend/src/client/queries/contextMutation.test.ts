import {QueryClient} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {error, success} from '@/message'

const session = vi.hoisted(() => ({epoch: 1}))
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/helpers/auth', () => ({
	getAuthSessionEpoch: () => session.epoch,
	getToken: () => null,
	getTokenIdentity: () => null,
}))
vi.mock('@/helpers/fetcher', () => ({getApiV2BaseUrl: () => '/api/v2/'}))

import {contextMutationOptions} from './contextMutation'

let client: QueryClient
beforeEach(() => {
	vi.resetAllMocks()
	session.epoch = 1
	client = new QueryClient()
})

function execute<TData, TOptimistic = undefined>(
	options: Parameters<typeof contextMutationOptions<TData, number, TOptimistic>>[0],
) {
	return client.getMutationCache()
		.build(client, contextMutationOptions(options))
		.execute(1)
}

const listKey = ['items', 'list'] as const
const detailKey = ['items', 'detail', 1] as const
const absentKey = ['items', 'detail', 2] as const

function optimistic(update = vi.fn()) {
	return {
		queryKeys: () => [listKey, detailKey, absentKey],
		update: (input: number, queryClient: QueryClient) => {
			queryClient.setQueryData<string[]>(listKey, current => current?.filter(() => false))
			queryClient.setQueryData<string>(detailKey, current => current ? 'optimistic' : current)
			queryClient.setQueryData<string>(absentKey, current => current ? 'optimistic' : current)
			return update(input)
		},
	}
}

describe('contextMutationOptions', () => {
	it('toasts errors while the context is current', async () => {
		const cause = new Error('denied')
		await expect(execute({mutationFn: async () => { throw cause }})).rejects.toBe(cause)
		expect(error).toHaveBeenCalledWith(cause)
	})

	it('suppresses the error toast after the context changed', async () => {
		await expect(execute({mutationFn: async () => {
			session.epoch++
			throw new Error('denied')
		}})).rejects.toThrow('denied')
		expect(error).not.toHaveBeenCalled()
	})

	it('skips onSettled when the context is stale', async () => {
		const onSettled = vi.fn(async () => {})
		await expect(execute({
			mutationFn: async () => { session.epoch++ },
			onSettled,
		})).rejects.toThrow('Client request context changed')
		expect(onSettled).not.toHaveBeenCalled()
	})

	it('does not toast success without a successMessage', async () => {
		const onSettled = vi.fn(async () => {})
		await execute({mutationFn: async () => 'done', onSettled})
		expect(onSettled).toHaveBeenCalledWith(1, client)
		expect(success).not.toHaveBeenCalled()
	})

	it('rejects when the context changes while settling', async () => {
		await expect(execute({
			mutationFn: async () => 'done',
			onSettled: async () => { session.epoch++ },
		})).rejects.toMatchObject({name: 'AbortError'})
	})

	it('suppresses the error toast when toastError declines', async () => {
		await expect(execute({
			mutationFn: async () => { throw new Error('denied') },
			toastError: input => input !== 1,
		})).rejects.toThrow('denied')
		expect(error).not.toHaveBeenCalled()
	})
})

describe('contextMutationOptions optimistic updates', () => {
	beforeEach(() => {
		client.setQueryData(listKey, ['before'])
		client.setQueryData(detailKey, 'before')
	})

	it('passes the optimistic update result to onSuccess', async () => {
		const onSuccess = vi.fn()
		await execute({
			mutationFn: async () => 'done',
			optimistic: optimistic(vi.fn(() => 'snapshot')),
			onSuccess,
		})
		expect(onSuccess).toHaveBeenCalledWith('done', 1, client, 'snapshot')
	})

	it('rolls back only the keys the factory fenced off', async () => {
		await expect(execute({
			mutationFn: async () => { throw new Error('denied') },
			optimistic: {
				queryKeys: () => [listKey],
				update: (_input: number, queryClient: QueryClient) => {
					queryClient.setQueryData<string[]>(listKey, [])
					queryClient.setQueryData<string>(detailKey, 'optimistic')
				},
			},
		})).rejects.toThrow('denied')

		expect(client.getQueryData(listKey)).toEqual(['before'])
		expect(client.getQueryData(detailKey)).toBe('optimistic')
	})

	it('rolls back cached queries after a failure in the same context', async () => {
		const cause = new Error('denied')
		const onSettled = vi.fn(async () => {})
		await expect(execute({
			mutationFn: async () => {
				expect(client.getQueryData(listKey)).toEqual([])
				expect(client.getQueryData(detailKey)).toBe('optimistic')
				throw cause
			},
			optimistic: optimistic(),
			onSettled,
		})).rejects.toBe(cause)

		expect(client.getQueryData(listKey)).toEqual(['before'])
		expect(client.getQueryData(detailKey)).toBe('before')
		expect(client.getQueryState(absentKey)).toBeUndefined()
		expect(error).toHaveBeenCalledWith(cause)
		expect(onSettled).toHaveBeenCalledOnce()
	})

	it('neither rolls back, toasts nor settles after the context changed', async () => {
		const onSettled = vi.fn(async () => {})
		await expect(execute({
			mutationFn: async () => {
				session.epoch++
				client.clear()
				client.setQueryData(listKey, ['next session'])
				throw new Error('denied')
			},
			optimistic: optimistic(),
			onSettled,
		})).rejects.toThrow('denied')

		expect(client.getQueryData(listKey)).toEqual(['next session'])
		expect(client.getQueryState(detailKey)).toBeUndefined()
		expect(error).not.toHaveBeenCalled()
		expect(onSettled).not.toHaveBeenCalled()
	})

	it('skips the optimistic write and request when the context changes while cancelling', async () => {
		const mutationFn = vi.fn(async () => 'done')
		const update = vi.fn()
		vi.spyOn(client, 'cancelQueries').mockImplementation(async () => { session.epoch++ })

		await expect(execute({mutationFn, optimistic: optimistic(update)}))
			.rejects.toMatchObject({name: 'AbortError'})

		expect(update).not.toHaveBeenCalled()
		expect(mutationFn).not.toHaveBeenCalled()
		expect(client.getQueryData(listKey)).toEqual(['before'])
		expect(error).not.toHaveBeenCalled()
	})
})
