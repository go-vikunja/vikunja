import {describe, it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	deleteProjectMutationOptions,
	deleteTaskPermanentlyMutationOptions,
	orderTrashedSubProjects,
	projectInTrashFromError,
	restoreProjectMutationOptions,
	restoreTaskMutationOptions,
	taskInTrashFromError,
	trashDaysLeft,
	trashedTasksQuery,
	trashKeys,
} from './trash'
import {timeEntryKeys} from './timeEntries'
import {
	normalizeProject,
	projectKeys,
	type ProjectListResult,
} from './projects'

const sdk = vi.hoisted(() => ({
	tasksTrashList: vi.fn(),
	tasksRestore: vi.fn(),
	tasksDeletePermanently: vi.fn(),
	tasksTrashEmpty: vi.fn(),
	projectsDelete: vi.fn(),
	projectsRestore: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

function seedTrash(client: QueryClient) {
	client.setQueryData(trashKeys.taskList(1), {
		items: [{id: 1, title: 'keep'}, {id: 2, title: 'restore'}],
		page: 1,
		per_page: 2,
		total: 3,
		total_pages: 2,
	})
	client.setQueryData(timeEntryKeys.list('', 'UTC', 1), {items: [], page: 1, per_page: 50, total: 0, total_pages: 0})
}

it('normalizes the trashed tasks of a page', async () => {
	const client = new QueryClient()
	sdk.tasksTrashList.mockResolvedValue({data: {items: [{id: 3}], page: 2, per_page: 50, total: 51, total_pages: 2}})

	const page = await client.fetchQuery(trashedTasksQuery(2))

	expect(sdk.tasksTrashList).toHaveBeenCalledWith(expect.objectContaining({query: {page: 2}}))
	expect(page.total).toBe(51)
	expect(page.items[0]).toMatchObject({id: 3, title: '', labels: []})
})

it('drops a restored task from the trash and marks time entries stale', async () => {
	const client = new QueryClient()
	seedTrash(client)
	sdk.tasksRestore.mockResolvedValue({data: {id: 2, title: 'restore'}})

	await client.getMutationCache().build(client, restoreTaskMutationOptions()).execute(2)

	expect(sdk.tasksRestore).toHaveBeenCalledWith({path: {task: 2}})
	expect(client.getQueryData(trashKeys.taskList(1))).toMatchObject({items: [{id: 1}], total: 2})
	expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
})

it('drops a permanently deleted task from the trash', async () => {
	const client = new QueryClient()
	seedTrash(client)
	sdk.tasksDeletePermanently.mockResolvedValue({})

	await client.getMutationCache().build(client, deleteTaskPermanentlyMutationOptions()).execute(1)

	expect(sdk.tasksDeletePermanently).toHaveBeenCalledWith({path: {task: 1}})
	expect(client.getQueryData(trashKeys.taskList(1))).toMatchObject({items: [{id: 2}], total: 2})
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
	sdk.projectsRestore.mockResolvedValue({data: {id: 2, title: 'restore'}})

	await client.getMutationCache().build(client, restoreProjectMutationOptions()).execute(2)

	expect(sdk.projectsRestore).toHaveBeenCalledWith({path: {id: 2}})
	expect(client.getQueryData(trashKeys.projectList(1))).toMatchObject({items: [{id: 1}], total: 2})
	expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
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

	it('marks time entries stale', async () => {
		const client = new QueryClient()
		seedTrash(client)
		sdk.projectsDelete.mockResolvedValue({data: undefined})

		await client.getMutationCache().build(client, deleteProjectMutationOptions()).execute(1)

		expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
	})
})

describe('taskInTrashFromError', () => {
	it('reads the details of a trashed task', () => {
		const details = {project_id: 1, deleted_at: '2026-01-01T00:00:00Z'}
		expect(taskInTrashFromError({status: 404, code: 4040, details})).toEqual(details)
	})
	it('ignores a plain not found', () => {
		expect(taskInTrashFromError({status: 404, code: 4002})).toBeNull()
	})
})

describe('projectInTrashFromError', () => {
	it('reads the details of a trashed project', () => {
		const details = {id: 1, deleted_at: '2026-01-01T00:00:00Z', can_restore: true}
		expect(projectInTrashFromError({status: 404, code: 3018, details})).toEqual(details)
	})
	it('ignores a plain not found', () => {
		expect(projectInTrashFromError({status: 404, code: 3001})).toBeNull()
	})
	it('ignores a trashed task', () => {
		expect(projectInTrashFromError({status: 404, code: 4040, details: {project_id: 1}})).toBeNull()
	})
})

it('orders trashed sub-projects below their parents, siblings in API order', () => {
	const ordered = orderTrashedSubProjects(1, [
		{id: 2, parent_project_id: 1, depth: 1},
		{id: 3, parent_project_id: 1, depth: 1},
		{id: 4, parent_project_id: 3, depth: 2},
		{id: 5, parent_project_id: 2, depth: 2},
	])

	expect(ordered.map(({id, depth}) => ({id, depth}))).toEqual([
		{id: 2, depth: 1},
		{id: 5, depth: 2},
		{id: 3, depth: 1},
		{id: 4, depth: 2},
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
