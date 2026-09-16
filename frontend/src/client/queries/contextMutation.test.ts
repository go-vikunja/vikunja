import {QueryClient} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {error, success} from '@/message'

const session = vi.hoisted(() => ({epoch: 1}))
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/helpers/auth', () => ({getAuthSessionEpoch: () => session.epoch, getToken: () => null, getTokenIdentity: () => null}))
vi.mock('@/helpers/fetcher', () => ({getApiV2BaseUrl: () => '/api/v2/'}))

import {contextMutationOptions} from './contextMutation'

let client: QueryClient
beforeEach(() => {
	vi.resetAllMocks()
	session.epoch = 1
	client = new QueryClient()
})

function execute<TData>(options: Parameters<typeof contextMutationOptions<TData, number>>[0]) {
	return client.getMutationCache().build(client, contextMutationOptions(options)).execute(1)
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
})
