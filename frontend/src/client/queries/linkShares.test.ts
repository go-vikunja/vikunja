import {QueryClient} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'

const sdk = vi.hoisted(() => ({sharesList: vi.fn(), sharesCreate: vi.fn(), sharesDelete: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/helpers/fetcher', () => ({getApiV2BaseUrl: () => '/api/v2/'}))
vi.mock('@/helpers/auth', () => ({getAuthSessionEpoch: () => 1, getToken: () => null, getTokenIdentity: () => null}))

import {linkShareKeys, linkSharesQuery, createLinkShareDraft, createLinkShareMutationOptions, deleteLinkShareMutationOptions} from './linkShares'

let client: QueryClient
beforeEach(() => {
	vi.resetAllMocks()
	client = new QueryClient({defaultOptions: {queries: {retry: false}}})
})

describe('link shares', () => {
	it('requests the first page with the project id', async () => {
		sdk.sharesList.mockResolvedValueOnce({data: {items: [{id: 1}], total_pages: 1}})
		expect(await client.fetchQuery(linkSharesQuery(7))).toEqual([{id: 1}])
		expect(sdk.sharesList).toHaveBeenCalledOnce()
		expect(sdk.sharesList).toHaveBeenCalledWith({path: {project: 7}, query: {page: 1, per_page: 1000}, signal: expect.any(AbortSignal)})
	})

	it('creates with a password, caches only the returned share and invalidates only its project', async () => {
		client.setQueryData(linkShareKeys.list(7), [])
		client.setQueryData(linkShareKeys.list(8), [{id: 2}])
		sdk.sharesCreate.mockResolvedValue({data: {id: 1, name: 'Client', hash: 'public-hash', sharing_type: 2, created: '2026-01-01T00:00:00Z'}})
		const share = createLinkShareDraft({name: 'Client', password: 'secret', permission: 1})
		await client.getMutationCache().build(client, createLinkShareMutationOptions()).execute({projectId: 7, share})
		expect(sdk.sharesCreate).toHaveBeenCalledWith({path: {project: 7}, body: share})
		expect(client.getQueryData(linkShareKeys.list(7))).toEqual([{id: 1, name: 'Client', hash: 'public-hash', sharing_type: 2, created: '2026-01-01T00:00:00Z'}])
		expect(client.getQueryState(linkShareKeys.list(7))?.isInvalidated).toBe(true)
		expect(client.getQueryState(linkShareKeys.list(8))?.isInvalidated).toBe(false)
	})

	it('removes only the deleted share', async () => {
		client.setQueryData(linkShareKeys.list(7), [{id: 1}, {id: 2}])
		sdk.sharesDelete.mockResolvedValue({})
		await client.getMutationCache().build(client, deleteLinkShareMutationOptions()).execute({projectId: 7, id: 1})
		expect(sdk.sharesDelete).toHaveBeenCalledWith({path: {project: 7, share: 1}})
		expect(client.getQueryData(linkShareKeys.list(7))).toEqual([{id: 2}])
	})

	it('provides draft defaults', () => {
		expect(createLinkShareDraft()).toEqual({name: '', password: '', permission: 0})
	})
})
