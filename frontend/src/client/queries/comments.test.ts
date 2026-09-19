import {
	beforeEach,
	expect,
	it,
	vi,
} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	commentsQuery,
	commentKeys,
	deleteCommentMutationOptions,
	updateCommentMutationOptions,
} from './comments'
import {
	normalizeTask,
	taskKeys,
} from './tasks'
const sdk = vi.hoisted(() => ({
	taskCommentsList: vi.fn(),
	taskCommentsDelete: vi.fn(),
	taskCommentsUpdate: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))
let client: QueryClient
beforeEach(() => { vi.clearAllMocks(); client = new QueryClient() })

it('requests the selected task, order and page', async () => {
	sdk.taskCommentsList.mockResolvedValue({data: {
		items: [{id: 4}],
		page: 2,
		total: 4,
		total_pages: 2,
		per_page: 3,
	}})
	const result = await client.fetchQuery(commentsQuery(1, 'desc', 2))
	expect(result.items).toEqual([{
		id: 4,
		comment: '',
		reactions: {},
	}])
	expect(sdk.taskCommentsList).toHaveBeenCalledWith({
		path: {task: 1},
		query: {
			order_by: 'desc',
			page: 2,
		},
		signal: expect.any(AbortSignal),
	})
})

it('decrements every cached page total and only expanded task counts', async () => {
	for (const page of [1, 2]) client.setQueryData(commentKeys.page(1, 'asc', page), {
		items: [{id: page}],
		total: 2,
		total_pages: 2,
		per_page: 1,
		page,
	})
	client.setQueryData(taskKeys.detail(1), normalizeTask({
		id: 1,
		comment_count: 2,
	}))
	sdk.taskCommentsDelete.mockResolvedValue({data: undefined})
	await client.getMutationCache().build(client, deleteCommentMutationOptions()).execute({
		taskId: 1,
		id: 2,
	})
	for (const page of [1, 2]) expect(client.getQueryData(commentKeys.page(1, 'asc', page))).toMatchObject({
		total: 1,
		total_pages: 1,
	})
	expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({comment_count: 1})
})

it('keeps author, dates and reactions the update response drops', async () => {
	client.setQueryData(commentKeys.page(1, 'asc', 1), {
		items: [{
			id: 2,
			comment: 'before',
			author: {id: 3},
			created: '2024-01-01T10:00:00Z',
			updated: '2024-01-02T10:00:00Z',
			reactions: {'👍': [{id: 1}]},
		}],
		total: 1,
		total_pages: 1,
		per_page: 50,
		page: 1,
	})
	sdk.taskCommentsUpdate.mockResolvedValue({data: {
		id: 2,
		comment: 'after',
		author: null,
		reactions: null,
		created: '0001-01-01T00:00:00Z',
		updated: '0001-01-01T00:00:00Z',
	}})
	await client.getMutationCache().build(client, updateCommentMutationOptions()).execute({
		taskId: 1,
		id: 2,
		comment: 'after',
	})
	expect(client.getQueryData(commentKeys.page(1, 'asc', 1))).toMatchObject({
		items: [{
			id: 2,
			comment: 'after',
			author: {id: 3},
			created: '2024-01-01T10:00:00Z',
			updated: '2024-01-02T10:00:00Z',
			reactions: {'👍': [{id: 1}]},
		}],
	})
})
