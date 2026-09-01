import {beforeEach, describe, expect, it, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'
import type {Project, ProjectView} from '@/client/generated'

const sdk = vi.hoisted(() => ({
	projectViewsList: vi.fn(),
	projectViewsRead: vi.fn(),
	projectViewsCreate: vi.fn(),
	projectViewsUpdate: vi.fn(),
	projectViewsDelete: vi.fn(),
}))

const projectCache = vi.hoisted(() => ({
	cancelProjectQueries: vi.fn(),
	updateProjectInCache: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('./projects', () => projectCache)

import {
	createProjectView,
	createProjectViewDraft,
	createProjectViewUpdate,
	deleteProjectView,
	projectViewKeys,
	projectViewQuery,
	projectViewsQuery,
	sortProjectViewsByPosition,
	updateProjectView,
} from './projectViews'

const views: ProjectView[] = [
	{id: 3, project_id: 7, title: 'Table', view_kind: 'table', position: 30},
	{id: 1, project_id: 7, title: 'List', view_kind: 'list', position: 10},
	{id: 2, project_id: 7, title: 'Board', view_kind: 'kanban', position: 20},
]

function applyProjectUpdate(project: Project): Project {
	const calls = projectCache.updateProjectInCache.mock.calls
	const updater = calls[calls.length - 1]?.[1]
	expect(updater).toEqual(expect.any(Function))
	return updater(project)
}

describe('project view queries', () => {
	beforeEach(() => {
		queryClient.clear()
		vi.clearAllMocks()
	})

	it('keys lists by project and response-shaping arguments', () => {
		expect(projectViewsQuery(7, {q: 'board'}).queryKey)
			.toEqual(projectViewKeys.list(7, {q: 'board'}))
		expect(projectViewsQuery(8, {q: 'board'}).queryKey)
			.not.toEqual(projectViewsQuery(7, {q: 'board'}).queryKey)
		expect(projectViewsQuery(7, {q: 'list'}).queryKey)
			.not.toEqual(projectViewsQuery(7, {q: 'board'}).queryKey)
	})

	it('keys details by both project and view', () => {
		expect(projectViewQuery(7, 2).queryKey).toEqual(projectViewKeys.detail(7, 2))
		expect(projectViewQuery(8, 2).queryKey).not.toEqual(projectViewQuery(7, 2).queryKey)
		expect(projectViewQuery(7, 3).queryKey).not.toEqual(projectViewQuery(7, 2).queryKey)
	})

	it('loads the non-paginated view response exactly once', async () => {
		sdk.projectViewsList.mockResolvedValue({data: {items: views, total_pages: 2}})

		await expect(queryClient.fetchQuery(projectViewsQuery(7, {q: 'work'}))).resolves.toEqual(views)
		expect(sdk.projectViewsList).toHaveBeenCalledOnce()
		expect(sdk.projectViewsList).toHaveBeenCalledWith({
			path: {project: 7},
			query: {page: 1, per_page: 1000, q: 'work'},
		})
	})

	it('loads a detail using both parent and resource ids', async () => {
		sdk.projectViewsRead.mockResolvedValue({data: views[1]})

		await expect(queryClient.fetchQuery(projectViewQuery(7, 1))).resolves.toEqual(views[1])
		expect(sdk.projectViewsRead).toHaveBeenCalledWith({path: {project: 7, view: 1}})
	})

	it('sorts by position without changing the source array', () => {
		const sorted = sortProjectViewsByPosition(views)

		expect(sorted.map(view => view.id)).toEqual([1, 2, 3])
		expect(views.map(view => view.id)).toEqual([3, 1, 2])
	})

	it('creates a complete draft with the default task filter', () => {
		expect(createProjectViewDraft()).toEqual({
			title: '',
			view_kind: 'list',
			filter: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter: 'done = false',
				filter_include_nulls: true,
				s: '',
			},
			position: 0,
			bucket_configuration_mode: 'manual',
			bucket_configuration: [],
			default_bucket_id: 0,
			done_bucket_id: 0,
		})
	})

	it('applies draft overrides and normalizes a null bucket configuration', () => {
		const draft = createProjectViewDraft({
			title: 'Board',
			view_kind: 'kanban',
			bucket_configuration: null,
		})

		expect(draft.title).toBe('Board')
		expect(draft.view_kind).toBe('kanban')
		expect(draft.bucket_configuration).toEqual([])
	})

	it('preserves an unfiltered Kanban view when serializing a reorder update', () => {
		const update = createProjectViewUpdate({...views[2], position: 15})
		const serialized = JSON.parse(JSON.stringify(update))

		expect(serialized).toMatchObject({view_kind: 'kanban', position: 15})
		expect(serialized).not.toHaveProperty('filter')
	})
})

describe('project view cache reconciliation', () => {
	beforeEach(() => {
		queryClient.clear()
		vi.clearAllMocks()
		queryClient.setQueryData(projectViewKeys.list(7, {}), views)
		queryClient.setQueryData(projectViewKeys.list(7, {q: 'board'}), [views[2]])
	})

	it('adds a created view to view and nested project caches', async () => {
		const created: ProjectView = {
			id: 4,
			project_id: 7,
			title: 'Gantt',
			view_kind: 'gantt',
			position: 15,
		}
		sdk.projectViewsCreate.mockResolvedValue({data: created})

		await expect(createProjectView({
			projectId: 7,
			view: createProjectViewDraft({title: 'Gantt', view_kind: 'gantt', position: 15}),
		})).resolves.toEqual(created)

		expect(sdk.projectViewsCreate).toHaveBeenCalledWith({
			path: {project: 7},
			body: expect.objectContaining({title: 'Gantt', view_kind: 'gantt', position: 15}),
		})
		expect(queryClient.getQueryData<ProjectView[]>(projectViewKeys.list(7, {}))?.map(view => view.id))
			.toEqual([1, 4, 2, 3])
		expect(queryClient.getQueryData<ProjectView[]>(projectViewKeys.list(7, {q: 'board'}))?.map(view => view.id))
			.toEqual([2])
		expect(queryClient.getQueryState(projectViewKeys.list(7, {q: 'board'}))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryData(projectViewKeys.detail(7, 4))).toEqual(created)
		expect(projectCache.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(projectCache.updateProjectInCache).toHaveBeenCalledWith(7, expect.any(Function))
		expect(applyProjectUpdate({id: 7, views} as Project).views?.map(view => view.id))
			.toEqual([1, 4, 2, 3])
	})

	it('replaces an updated view in view and nested project caches', async () => {
		const updated: ProjectView = {...views[2], title: 'Updated board', position: 5}
		sdk.projectViewsUpdate.mockResolvedValue({data: updated})

		await expect(updateProjectView({
			projectId: 7,
			viewId: 2,
			view: createProjectViewDraft({title: 'Updated board', view_kind: 'kanban', position: 5}),
		})).resolves.toEqual(updated)

		expect(sdk.projectViewsUpdate).toHaveBeenCalledWith({
			path: {project: 7, view: 2},
			body: expect.objectContaining({title: 'Updated board', view_kind: 'kanban', position: 5}),
		})
		expect(queryClient.getQueryData<ProjectView[]>(projectViewKeys.list(7, {}))?.map(view => view.id))
			.toEqual([2, 1, 3])
		expect(queryClient.getQueryData<ProjectView[]>(projectViewKeys.list(7, {q: 'board'}))?.map(view => view.id))
			.toEqual([2])
		expect(queryClient.getQueryState(projectViewKeys.list(7, {q: 'board'}))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryData(projectViewKeys.detail(7, 2))).toEqual(updated)
		expect(projectCache.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(projectCache.updateProjectInCache).toHaveBeenCalledWith(7, expect.any(Function))
		expect(applyProjectUpdate({id: 7, views} as Project).views?.map(view => view.id))
			.toEqual([2, 1, 3])
	})

	it('removes a deleted view from view and nested project caches', async () => {
		queryClient.setQueryData(projectViewKeys.detail(7, 2), views[2])
		sdk.projectViewsDelete.mockResolvedValue({data: undefined})

		await deleteProjectView({projectId: 7, viewId: 2})

		expect(sdk.projectViewsDelete).toHaveBeenCalledWith({path: {project: 7, view: 2}})
		expect(queryClient.getQueryData<ProjectView[]>(projectViewKeys.list(7, {}))?.map(view => view.id))
			.toEqual([1, 3])
		expect(queryClient.getQueryData<ProjectView[]>(projectViewKeys.list(7, {q: 'board'}))?.map(view => view.id))
			.toEqual([2])
		expect(queryClient.getQueryState(projectViewKeys.list(7, {q: 'board'}))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryData(projectViewKeys.detail(7, 2))).toBeUndefined()
		expect(projectCache.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(projectCache.updateProjectInCache).toHaveBeenCalledWith(7, expect.any(Function))
		expect(applyProjectUpdate({id: 7, views} as Project).views?.map(view => view.id))
			.toEqual([1, 3])
	})
})
