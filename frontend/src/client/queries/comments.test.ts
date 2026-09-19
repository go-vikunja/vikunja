import {
	beforeEach,
	expect,
	it,
	vi,
} from 'vitest'
import {
	QueryClient,
	QueryObserver,
} from '@tanstack/vue-query'
import {
	commentsQuery,
	commentKeys,
	createCommentMutationOptions,
	deleteCommentMutationOptions,
	updateCommentMutationOptions,
	type CommentPage,
} from './comments'
import {
	normalizeTask,
	taskKeys,
	type TaskResponse,
} from './tasks'

const sdk = vi.hoisted(() => ({
	taskCommentsList: vi.fn(),
	taskCommentsCreate: vi.fn(),
	taskCommentsDelete: vi.fn(),
	taskCommentsUpdate: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

let client: QueryClient
beforeEach(() => {
	vi.clearAllMocks()
	client = new QueryClient({defaultOptions: {queries: {retry: false}}})
})

it('requests the selected task, order, page and page size', async () => {
	sdk.taskCommentsList.mockResolvedValue({data: {
		items: [{id: 4}],
		page: 2,
		total: 4,
		total_pages: 2,
		per_page: 3,
	}})
	const result = await client.fetchQuery(commentsQuery(1, 'desc', 2, 3))
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
			per_page: 3,
		},
		signal: expect.any(AbortSignal),
	})
})

it('clamps per_page to the v2 maximum', async () => {
	sdk.taskCommentsList.mockResolvedValue({data: {
		items: [],
		page: 1,
		total: 0,
		total_pages: 0,
		per_page: 1000,
	}})
	await client.fetchQuery(commentsQuery(1, 'desc', 1, 5000))
	expect(sdk.taskCommentsList).toHaveBeenCalledWith({
		path: {task: 1},
		query: {
			order_by: 'desc',
			page: 1,
			per_page: 1000,
		},
		signal: expect.any(AbortSignal),
	})
})

it('keeps the previous page as placeholder for the same task only', () => {
	const previousData: CommentPage = {
		items: [{
			id: 1,
			comment: 'first',
			reactions: {},
		}],
		total: 2,
		total_pages: 2,
		per_page: 1,
		page: 1,
	}
	const placeholderData = commentsQuery(1, 'asc', 2, 1).placeholderData as (
		previousData: CommentPage | undefined,
		previousQuery: {queryKey: ReturnType<typeof commentKeys.page>} | undefined,
	) => CommentPage | undefined

	expect(placeholderData(previousData, {queryKey: commentKeys.page(1, 'asc', 1, 1)})).toBe(previousData)
	expect(placeholderData(previousData, {queryKey: commentKeys.page(2, 'asc', 1, 1)})).toBeUndefined()
	expect(placeholderData(undefined, undefined)).toBeUndefined()
})

it('decrements every cached page total and only expanded task counts', async () => {
	for (const page of [1, 2]) client.setQueryData(commentKeys.page(1, 'asc', page, 1), {
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
	client.setQueryData(taskKeys.detail(1, ['reactions']), normalizeTask({id: 1}))
	sdk.taskCommentsDelete.mockResolvedValue({data: undefined})
	await client.getMutationCache().build(client, deleteCommentMutationOptions()).execute({
		taskId: 1,
		id: 2,
	})
	expect(client.getQueryData(commentKeys.page(1, 'asc', 1, 1))).toEqual({
		items: [{id: 1}],
		total: 1,
		total_pages: 1,
		per_page: 1,
		page: 1,
	})
	expect(client.getQueryData(commentKeys.page(1, 'asc', 2, 1))).toEqual({
		items: [],
		total: 1,
		total_pages: 1,
		per_page: 1,
		page: 2,
	})
	expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({comment_count: 1})
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(1, ['reactions']))?.comment_count).toBeUndefined()
})

it('keeps author, dates and reactions the update response drops', async () => {
	client.setQueryData(commentKeys.page(1, 'asc', 1, 50), {
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
	expect(client.getQueryData(commentKeys.page(1, 'asc', 1, 50))).toMatchObject({
		items: [{
			id: 2,
			comment: 'after',
			author: {id: 3},
			created: '2024-01-01T10:00:00Z',
			updated: '2024-01-02T10:00:00Z',
			reactions: {'👍': [{id: 1}]},
		}],
	})
	expect(client.getQueryState(commentKeys.page(1, 'asc', 1, 50))?.isInvalidated).toBe(true)
})

it('prepends onto a full first desc page and drops the last item', async () => {
	client.setQueryData(commentKeys.page(1, 'desc', 1, 2), {
		items: [
			{
				id: 2,
				comment: 'newer',
				reactions: {},
			},
			{
				id: 1,
				comment: 'older',
				reactions: {},
			},
		],
		total: 2,
		total_pages: 1,
		per_page: 2,
		page: 1,
	})
	sdk.taskCommentsCreate.mockResolvedValue({data: {
		id: 3,
		comment: 'newest',
	}})
	await client.getMutationCache().build(client, createCommentMutationOptions()).execute({
		taskId: 1,
		comment: 'newest',
	})
	expect(client.getQueryData(commentKeys.page(1, 'desc', 1, 2))).toEqual({
		items: [
			{
				id: 3,
				comment: 'newest',
				reactions: {},
			},
			{
				id: 2,
				comment: 'newer',
				reactions: {},
			},
		],
		total: 3,
		total_pages: 2,
		per_page: 2,
		page: 1,
	})
})

it('appends onto the last asc page', async () => {
	client.setQueryData(commentKeys.page(1, 'asc', 2, 50), {
		items: [{
			id: 99,
			comment: 'last',
			reactions: {},
		}],
		total: 99,
		total_pages: 2,
		per_page: 50,
		page: 2,
	})
	sdk.taskCommentsCreate.mockResolvedValue({data: {
		id: 100,
		comment: 'newest',
	}})
	await client.getMutationCache().build(client, createCommentMutationOptions()).execute({
		taskId: 1,
		comment: 'newest',
	})
	expect(client.getQueryData(commentKeys.page(1, 'asc', 2, 50))).toEqual({
		items: [
			{
				id: 99,
				comment: 'last',
				reactions: {},
			},
			{
				id: 100,
				comment: 'newest',
				reactions: {},
			},
		],
		total: 100,
		total_pages: 2,
		per_page: 50,
		page: 2,
	})
})

it('stales a mounted page after create without refetching over the insert', async () => {
	const key = commentKeys.page(1, 'desc', 1, 50)
	const queryFn = vi.fn(() => ({
		items: [],
		total: 0,
		total_pages: 0,
		per_page: 50,
		page: 1,
	}))
	const unsubscribe = new QueryObserver(client, {
		queryKey: key,
		queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
	sdk.taskCommentsCreate.mockResolvedValue({data: {
		id: 3,
		comment: 'newest',
	}})
	await client.getMutationCache().build(client, createCommentMutationOptions()).execute({
		taskId: 1,
		comment: 'newest',
	})
	expect(client.getQueryData<CommentPage>(key)?.items).toEqual([{
		id: 3,
		comment: 'newest',
		reactions: {},
	}])
	expect(client.getQueryState(key)?.isInvalidated).toBe(true)
	expect(queryFn).toHaveBeenCalledTimes(1)
	unsubscribe()
})

it('bumps totals without touching items on a non-last asc page', async () => {
	client.setQueryData(commentKeys.page(1, 'asc', 1, 50), {
		items: [{
			id: 1,
			comment: 'first',
			reactions: {},
		}],
		total: 51,
		total_pages: 2,
		per_page: 50,
		page: 1,
	})
	sdk.taskCommentsCreate.mockResolvedValue({data: {
		id: 52,
		comment: 'newest',
	}})
	await client.getMutationCache().build(client, createCommentMutationOptions()).execute({
		taskId: 1,
		comment: 'newest',
	})
	expect(client.getQueryData(commentKeys.page(1, 'asc', 1, 50))).toEqual({
		items: [{
			id: 1,
			comment: 'first',
			reactions: {},
		}],
		total: 52,
		total_pages: 2,
		per_page: 50,
		page: 1,
	})
})
