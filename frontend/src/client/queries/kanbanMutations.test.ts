import {describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {kanbanKeys, type BoardData} from './kanban'
import {loadBucketPageMutationOptions, updateBucketMutationOptions} from './kanbanMutations'
const sdk = vi.hoisted(() => ({projectViewTasksList: vi.fn(), bucketsUpdate: vi.fn()}))
vi.mock('@/client/generated', async original => ({...await original<object>(), ...sdk}))
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/client/requestContext', () => ({captureClientRequestContext: () => 1, assertClientRequestContext: vi.fn(), isClientRequestContextCurrent: () => true}))

describe('board mutations', () => {
	it('appends a bucket page and preserves other bucket metadata', async () => {
		const client = new QueryClient()
		const key = kanbanKeys.board(-1, 2)
		client.setQueryData(key, {buckets: [{id: 3, tasks: [{id: 1}]}, {id: 4, tasks: []}], pages: {3: 1, 4: 1}, hasMore: {3: true, 4: false}})
		sdk.projectViewTasksList.mockResolvedValue({data: {items: [{id: 2}], total_pages: 2}})
		await client.getMutationCache().build(client, loadBucketPageMutationOptions()).execute({project: -1, view: 2, params: {}, bucket: 3, page: 2})
		expect(client.getQueryData<BoardData>(key)).toMatchObject({buckets: [{tasks: [{id: 1}, {id: 2}]}, {tasks: []}], pages: {3: 2, 4: 1}, hasMore: {3: false, 4: false}})
	})
	it('retains embedded tasks when a bucket update returns metadata only', async () => {
		const client = new QueryClient()
		const key = kanbanKeys.board(1, 2)
		client.setQueryData(key, {buckets: [{id: 3, title: 'old', tasks: [{id: 1}]}], pages: {3: 1}, hasMore: {3: false}})
		sdk.bucketsUpdate.mockResolvedValue({data: {id: 3, title: 'new', tasks: null}})
		await client.getMutationCache().build(client, updateBucketMutationOptions()).execute({project: 1, view: 2, bucket: {id: 3, title: 'new'}})
		expect(client.getQueryData<BoardData>(key)?.buckets[0]).toMatchObject({title: 'new', tasks: [{id: 1}]})
	})
})
