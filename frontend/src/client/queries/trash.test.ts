import {describe, it, expect, vi} from 'vitest'
import {QueryClient, QueryObserver} from '@tanstack/vue-query'
import {
	deleteProjectMutationOptions,
	deleteTaskPermanentlyMutationOptions,
	emptyTaskTrashMutationOptions,
	projectInTrashFromError,
	restoreProjectMutationOptions,
	restoreTaskMutationOptions,
	taskInTrashFromError,
	trashDaysLeft,
	trashedProjectsQuery,
	trashedTasksQuery,
	trashKeys,
} from './trash'
import {timeEntryKeys} from './timeEntries'
import {kanbanKeys, normalizeBucket, type BoardData} from './kanban'
import {taskKeys} from './tasks'
import {
	normalizeProject,
	projectKeys,
	type ProjectListResult,
} from './projects'
import {ApiError} from '@/client/problem'
import {error} from '@/message'

const sdk = vi.hoisted(() => ({
	tasksTrashList: vi.fn(),
	tasksRestore: vi.fn(),
	tasksRead: vi.fn(),
	tasksDeletePermanently: vi.fn(),
	tasksTrashEmpty: vi.fn(),
	projectsDelete: vi.fn(),
	projectsRestore: vi.fn(),
	projectsTrashList: vi.fn(),
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
	sdk.tasksRead.mockResolvedValue({
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
	sdk.tasksRead.mockResolvedValue({
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

it('patches the restored card with the full task detail', async () => {
	const client = new QueryClient()
	client.setQueryData(projectKeys.detail(5), {
		id: 5,
		views: [{
			id: 6,
			bucket_configuration_mode: 'manual',
			default_bucket_id: 3,
		}],
	})
	client.setQueryData(kanbanKeys.board(5, 6), {
		buckets: [normalizeBucket({
			id: 3,
			count: 1,
			tasks: [{
				id: 1,
				position: 400,
			}],
		})],
		pages: {3: 1},
	})
	sdk.tasksRestore.mockResolvedValue({
		data: {
			id: 2,
			project_id: 5,
		},
	})
	sdk.tasksRead.mockResolvedValue({
		data: {
			id: 2,
			project_id: 5,
			identifier: 'P-2',
			labels: [{
				id: 7,
				title: 'urgent',
			}],
		},
	})

	await client.getMutationCache().build(client, restoreTaskMutationOptions()).execute(2)

	expect(sdk.tasksRead).toHaveBeenCalledWith(expect.objectContaining({path: {task: 2}}))
	expect(client.getQueryData<BoardData>(kanbanKeys.board(5, 6))?.buckets[0].tasks[0]).toMatchObject({
		id: 2,
		identifier: 'P-2',
		labels: [{id: 7}],
		position: 200,
	})
})

it('falls back to the restore response when the task detail fails', async () => {
	const client = new QueryClient()
	seedTrash(client)
	sdk.tasksRestore.mockResolvedValue({
		data: {
			id: 2,
			title: 'restore',
		},
	})
	sdk.tasksRead.mockRejectedValue(new Error('network'))

	const restored = await client.getMutationCache().build(client, restoreTaskMutationOptions()).execute(2)

	expect(restored).toMatchObject({
		id: 2,
		title: 'restore',
		labels: [],
	})
	expect(error).not.toHaveBeenCalled()
	expect(client.getQueryData(trashKeys.taskList(1))).toMatchObject({items: [{id: 1}]})
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

it('drops a restored project from the trash and marks time entries stale', async () => {
	const client = new QueryClient()
	seedTrash(client)
	client.setQueryData(trashKeys.projectList(1), {
		items: [{id: 1, title: 'keep'}, {id: 2, title: 'restore'}],
		page: 1,
		per_page: 2,
		total: 3,
		total_pages: 2,
	})
	client.setQueryData(taskKeys.detail(5), {id: 5})
	sdk.projectsRestore.mockResolvedValue({data: {id: 2, title: 'restore'}})

	await client.getMutationCache().build(client, restoreProjectMutationOptions()).execute(2)

	expect(sdk.projectsRestore).toHaveBeenCalledWith({path: {id: 2}})
	expect(client.getQueryData(trashKeys.projectList(1))).toMatchObject({items: [{id: 1}], total: 2})
	expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(taskKeys.detail(5))?.isInvalidated).toBe(true)
})

describe('deleteProjectMutationOptions', () => {
	const project = (id: number, parent_project_id = 0) => normalizeProject({id, title: 'Project', parent_project_id})

	it('removes a deleted project, all descendants, and their detail entries', async () => {
		const client = new QueryClient()
		client.setQueryData<ProjectListResult>(projectKeys.list(), {
			projects: [
				project(1),
				project(2, 1),
				project(3, 2),
				project(4),
			],
			favoriteProject: null,
			savedFilterProjects: [project(-2)],
		})
		for (const id of [1, 2, 3, 4]) {
			client.setQueryData(projectKeys.detail(id), project(id))
		}
		sdk.projectsDelete.mockResolvedValue({data: undefined})

		await client.getMutationCache().build(client, deleteProjectMutationOptions()).execute(1)

		expect(sdk.projectsDelete).toHaveBeenCalledWith({path: {id: 1}})
		expect(client.getQueryData<ProjectListResult>(projectKeys.list())?.projects.map(p => p.id)).toEqual([4])
		expect(client.getQueryData(projectKeys.detail(1))).toBeUndefined()
		expect(client.getQueryData(projectKeys.detail(2))).toBeUndefined()
		expect(client.getQueryData(projectKeys.detail(3))).toBeUndefined()
		expect(client.getQueryData(projectKeys.detail(4))).toBeDefined()
	})

	it('marks time entries, task lists and task details stale', async () => {
		const client = new QueryClient()
		seedTrash(client)
		client.setQueryData(taskKeys.list({}), [])
		client.setQueryData(taskKeys.detail(5), {id: 5})
		sdk.projectsDelete.mockResolvedValue({data: undefined})

		await client.getMutationCache().build(client, deleteProjectMutationOptions()).execute(1)

		expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
		expect(client.getQueryState(taskKeys.list({}))?.isInvalidated).toBe(true)
		expect(client.getQueryState(taskKeys.detail(5))?.isInvalidated).toBe(true)
	})
})

describe('taskInTrashFromError', () => {
	it('reads the details of a trashed task', () => {
		const details = {
			id: 1,
			project_id: 1,
			deleted_at: '2026-01-01T00:00:00Z',
			can_restore: true,
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

describe('projectInTrashFromError', () => {
	it('reads the details of a trashed project', () => {
		const details = {
			id: 1,
			deleted_at: '2026-01-01T00:00:00Z',
			can_restore: true,
		}
		expect(projectInTrashFromError(new ApiError({
			status: 404,
			code: 3018,
			detail: 'in trash',
			details,
		}))).toEqual(details)
	})
	it('ignores a plain not found', () => {
		expect(projectInTrashFromError(new ApiError({
			status: 404,
			code: 3001,
			detail: 'not found',
		}))).toBeNull()
	})
	it('ignores a trashed task', () => {
		expect(projectInTrashFromError(new ApiError({
			status: 404,
			code: 4040,
			detail: 'in trash',
			details: {project_id: 1},
		}))).toBeNull()
	})
})

it('normalizes trashed projects and orders sub-projects below their parents', async () => {
	const client = new QueryClient()
	sdk.projectsTrashList.mockResolvedValue({
		data: {
			items: [
				{
					id: 1,
					hex_color: 'ff0000',
					trashed_sub_projects: [
						{
							id: 2,
							parent_project_id: 1,
							depth: 1,
						},
						{
							id: 3,
							parent_project_id: 1,
							depth: 1,
						},
						{
							id: 4,
							parent_project_id: 3,
							depth: 2,
						},
						{
							id: 5,
							parent_project_id: 2,
							depth: 2,
						},
					],
				},
			],
			page: 1,
			per_page: 50,
			total: 1,
			total_pages: 1,
		},
	})

	const page = await client.fetchQuery(trashedProjectsQuery(1))

	expect(sdk.projectsTrashList).toHaveBeenCalledWith(expect.objectContaining({query: {page: 1}}))
	expect(page.items[0]?.hex_color).toBe('#ff0000')
	expect(page.items[0]?.trashed_sub_projects?.map(({id, depth}) => ({
		id,
		depth,
	}))).toEqual([
		{
			id: 2,
			depth: 1,
		},
		{
			id: 5,
			depth: 2,
		},
		{
			id: 3,
			depth: 1,
		},
		{
			id: 4,
			depth: 2,
		},
	])
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
