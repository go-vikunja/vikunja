import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient, QueryObserver} from '@tanstack/vue-query'
import {bucketHasMore, bucketKeys, kanbanKeys, type BoardData} from './kanban'
import {projectKeys} from './projects'
import {
	createBucketMutationOptions,
	deleteBucketMutationOptions,
	loadBucketPageMutationOptions,
	updateBucketMutationOptions,
} from './kanbanMutations'

const sdk = vi.hoisted(() => ({
	bucketsCreate: vi.fn(),
	bucketsDelete: vi.fn(),
	bucketsUpdate: vi.fn(),
	projectViewTasksList: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/client/requestContext', () => ({
	captureClientRequestContext: () => 1,
	assertClientRequestContext: vi.fn(),
	isClientRequestContextCurrent: () => true,
}))

describe('board mutations', () => {
	beforeEach(() => vi.resetAllMocks())

	it('appends a bucket page and preserves other bucket metadata', async () => {
		const client = new QueryClient()
		const key = kanbanKeys.board(-1, 2)
		client.setQueryData(key, {
			buckets: [
				{id: 3, count: 2, tasks: [{id: 1}]},
				{id: 4, count: 0, tasks: []},
			],
			pages: {3: 1, 4: 1},
		})
		sdk.projectViewTasksList.mockResolvedValue({data: {items: [{id: 2}], total_pages: 2}})
		await client.getMutationCache()
			.build(client, loadBucketPageMutationOptions())
			.execute({project: -1, view: 2, params: {}, bucket: 3, page: 2})
		const board = client.getQueryData<BoardData>(key)!
		expect(board).toMatchObject({
			buckets: [{tasks: [{id: 1}, {id: 2}]}, {tasks: []}],
			pages: {3: 2, 4: 1},
		})
		expect(board.buckets.map(bucketHasMore)).toEqual([false, false])
	})

	it('ignores a bucket page that arrives out of order', async () => {
		const client = new QueryClient()
		const key = kanbanKeys.board(-1, 2)
		const board = {buckets: [{id: 3, count: 5, tasks: [{id: 1}]}], pages: {3: 1}}
		client.setQueryData(key, board)
		sdk.projectViewTasksList.mockResolvedValue({data: {items: [{id: 7}], total_pages: 3}})
		await client.getMutationCache()
			.build(client, loadBucketPageMutationOptions())
			.execute({project: -1, view: 2, params: {}, bucket: 3, page: 3})
		expect(client.getQueryData<BoardData>(key)).toEqual(board)
	})

	it('invalidates the cached bucket list after adding a column', async () => {
		const client = new QueryClient()
		client.setQueryData(bucketKeys.list(1, 2), [{id: 3}])
		sdk.bucketsCreate.mockResolvedValue({data: {id: 5, title: 'new'}})
		await client.getMutationCache()
			.build(client, createBucketMutationOptions())
			.execute({project: 1, view: 2, bucket: {title: 'new'}})
		expect(client.getQueryState(bucketKeys.list(1, 2))?.isInvalidated).toBe(true)
	})

	it('refetches the project after a delete, because the backend clears its default and done bucket', async () => {
		const client = new QueryClient()
		client.setQueryData(kanbanKeys.board(1, 2), {buckets: [{id: 3, tasks: []}], pages: {3: 1}})
		const readProject = vi.fn().mockResolvedValue({id: 1})
		const unsubscribe = new QueryObserver(client, {
			queryKey: projectKeys.detail(1),
			queryFn: readProject,
		}).subscribe(() => {})
		await vi.waitFor(() => expect(readProject).toHaveBeenCalledTimes(1))
		sdk.bucketsDelete.mockResolvedValue({})
		await client.getMutationCache()
			.build(client, deleteBucketMutationOptions())
			.execute({project: 1, view: 2, bucket: 3})
		unsubscribe()
		expect(readProject).toHaveBeenCalledTimes(2)
	})

	it('sends limit and position along with a renamed bucket, because the PUT would reset them', async () => {
		const client = new QueryClient()
		client.setQueryData(kanbanKeys.board(1, 2), {
			buckets: [{id: 3, title: 'old', count: 1, tasks: [{id: 1}]}],
			pages: {3: 1},
		})
		sdk.bucketsUpdate.mockResolvedValue({data: {id: 3, title: 'new', count: 0, tasks: null}})
		await client.getMutationCache()
			.build(client, updateBucketMutationOptions())
			.execute({project: 1, view: 2, bucket: {id: 3, title: 'new', limit: 5, position: 100}})
		expect(sdk.bucketsUpdate).toHaveBeenCalledWith({
			path: {project: 1, view: 2, bucket: 3},
			body: {title: 'new', limit: 5, position: 100},
		})
	})
})
