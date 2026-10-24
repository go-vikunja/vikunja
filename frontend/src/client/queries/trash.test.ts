import {describe, it, expect, vi} from 'vitest'
import {QueryClient, QueryObserver} from '@tanstack/vue-query'
import {
	deleteTaskPermanentlyMutationOptions,
	emptyTaskTrashMutationOptions,
	restoreTaskMutationOptions,
	taskInTrashFromError,
	trashDaysLeft,
	trashedTasksQuery,
	trashKeys,
} from './trash'
import {timeEntryKeys} from './timeEntries'
import {kanbanKeys, normalizeBucket, type BoardData} from './kanban'
import {projectKeys} from './projects'
import {deleteTaskMutationOptions} from './taskMutations'
import {ApiError} from '@/client/problem'

const sdk = vi.hoisted(() => ({
	tasksTrashList: vi.fn(),
	tasksRestore: vi.fn(),
	tasksDeletePermanently: vi.fn(),
	tasksTrashEmpty: vi.fn(),
	tasksDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

function seedTrash(client: QueryClient) {
	client.setQueryData(trashKeys.taskList(1), {
		items: [
			{
				id: 1,
				title: 'keep',
			},
			{
				id: 2,
				title: 'restore',
			},
		],
		page: 1,
		per_page: 2,
		total: 3,
		total_pages: 2,
	})
	client.setQueryData(timeEntryKeys.list('', 'UTC', 1), {
		items: [],
		page: 1,
		per_page: 50,
		total: 0,
		total_pages: 0,
	})
}

it('normalizes the trashed tasks of a page', async () => {
	const client = new QueryClient()
	sdk.tasksTrashList.mockResolvedValue({
		data: {
			items: [{id: 3}],
			page: 2,
			per_page: 50,
			total: 51,
			total_pages: 2,
		},
	})

	const page = await client.fetchQuery(trashedTasksQuery(2))

	expect(sdk.tasksTrashList).toHaveBeenCalledWith(expect.objectContaining({query: {page: 2}}))
	expect(page.total).toBe(51)
	expect(page.items[0]).toMatchObject({
		id: 3,
		title: '',
		labels: [],
	})
})

it('drops a restored task from the trash and marks time entries stale', async () => {
	const client = new QueryClient()
	seedTrash(client)
	sdk.tasksRestore.mockResolvedValue({
		data: {
			id: 2,
			title: 'restore',
		},
	})

	await client.getMutationCache().build(client, restoreTaskMutationOptions()).execute(2)

	expect(sdk.tasksRestore).toHaveBeenCalledWith({path: {task: 2}})
	expect(client.getQueryData(trashKeys.taskList(1))).toMatchObject({
		items: [{id: 1}],
		total: 2,
		total_pages: 1,
	})
	expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
})

it('patches the restored card into scrolled boards and refetches only filter boards', async () => {
	const client = new QueryClient()
	seedTrash(client)
	client.setQueryData(projectKeys.detail(5), {
		id: 5,
		views: [
			{
				id: 6,
				bucket_configuration_mode: 'manual',
				default_bucket_id: 3,
			},
			{
				id: 9,
				bucket_configuration_mode: 'filter',
			},
		],
	})
	const readBoard = vi.fn().mockResolvedValue({
		buckets: [normalizeBucket({
			id: 3,
			count: 30,
			tasks: [{id: 1}],
		})],
		pages: {3: 1},
	})
	const readFilterBoard = vi.fn().mockResolvedValue({
		buckets: [],
		pages: {},
	})
	const readOtherBoard = vi.fn().mockResolvedValue({
		buckets: [],
		pages: {},
	})
	const observers = [
		new QueryObserver(client, {
			queryKey: kanbanKeys.board(5, 6),
			queryFn: readBoard,
		}),
		new QueryObserver(client, {
			queryKey: kanbanKeys.board(5, 9),
			queryFn: readFilterBoard,
		}),
		new QueryObserver(client, {
			queryKey: kanbanKeys.board(7, 8),
			queryFn: readOtherBoard,
		}),
	].map(observer => observer.subscribe(() => {}))
	await vi.waitFor(() => {
		expect(readBoard).toHaveBeenCalledTimes(1)
		expect(readFilterBoard).toHaveBeenCalledTimes(1)
		expect(readOtherBoard).toHaveBeenCalledTimes(1)
		expect(client.getQueryData(kanbanKeys.board(5, 6))).toBeDefined()
	})
	client.setQueryData<BoardData>(kanbanKeys.board(5, 6), current => current && {
		...current,
		pages: {3: 2},
	})
	sdk.tasksRestore.mockResolvedValue({
		data: {
			id: 2,
			title: 'restore',
			project_id: 5,
		},
	})

	await client.getMutationCache().build(client, restoreTaskMutationOptions()).execute(2)
	observers.forEach(stop => stop())

	const board = client.getQueryData<BoardData>(kanbanKeys.board(5, 6))!
	expect(board.buckets[0].tasks.map(task => task.id)).toEqual([
		2,
		1,
	])
	expect(board.buckets[0].count).toBe(31)
	expect(board.pages).toEqual({3: 2})
	expect(readBoard).toHaveBeenCalledTimes(1)
	expect(readFilterBoard).toHaveBeenCalledTimes(2)
	expect(readOtherBoard).toHaveBeenCalledTimes(1)
})

it('drops a permanently deleted task from the trash', async () => {
	const client = new QueryClient()
	seedTrash(client)
	sdk.tasksDeletePermanently.mockResolvedValue({})

	await client.getMutationCache().build(client, deleteTaskPermanentlyMutationOptions()).execute(1)

	expect(sdk.tasksDeletePermanently).toHaveBeenCalledWith({path: {task: 1}})
	expect(client.getQueryData(trashKeys.taskList(1))).toMatchObject({
		items: [{id: 2}],
		total: 2,
		total_pages: 1,
	})
})

it('marks the trash and time entries stale after emptying the trash', async () => {
	const client = new QueryClient()
	seedTrash(client)
	sdk.tasksTrashEmpty.mockResolvedValue({})

	await client.getMutationCache().build(client, emptyTaskTrashMutationOptions()).execute(undefined)

	expect(sdk.tasksTrashEmpty).toHaveBeenCalled()
	expect(client.getQueryState(trashKeys.taskList(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
})

it('marks the trash and time entries stale after deleting a task', async () => {
	const client = new QueryClient()
	seedTrash(client)
	sdk.tasksDelete.mockResolvedValue({})

	await client.getMutationCache().build(client, deleteTaskMutationOptions()).execute(5)

	expect(sdk.tasksDelete).toHaveBeenCalledWith({path: {task: 5}})
	expect(client.getQueryState(trashKeys.taskList(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
})

describe('taskInTrashFromError', () => {
	it('reads the details of a trashed task', () => {
		const details = {
			project_id: 1,
			deleted_at: '2026-01-01T00:00:00Z',
		}
		expect(taskInTrashFromError(new ApiError({
			status: 404,
			code: 4040,
			detail: 'in trash',
			details,
		}))).toEqual(details)
	})
	it('ignores a plain not found', () => {
		expect(taskInTrashFromError(new ApiError({
			status: 404,
			code: 4002,
			detail: 'not found',
		}))).toBeNull()
	})
})

describe('trashDaysLeft', () => {
	const now = new Date('2026-01-11T12:00:00Z')
	it('counts down from 30 days', () => {
		expect(trashDaysLeft('2026-01-11T12:00:00Z', now)).toBe(30)
		expect(trashDaysLeft('2026-01-01T12:00:00Z', now)).toBe(20)
	})
	it('never goes below zero', () => {
		expect(trashDaysLeft('2025-01-01T00:00:00Z', now)).toBe(0)
	})
})
