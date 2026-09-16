import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient, type MutationOptions} from '@tanstack/vue-query'

import {queryClient} from '@/client/queryClient'
import type {ProjectView} from '@/client/generated'
import {normalizeProject, projectKeys, type ProjectListResult, type ProjectResponse} from './projects'
import {error, success} from '@/message'

const sdk = vi.hoisted(() => ({
	projectViewsCreate: vi.fn(),
	projectViewsUpdate: vi.fn(),
	projectViewsDelete: vi.fn(),
}))

const requestContext = vi.hoisted(() => ({
	identity: {id: 1, type: 1} as {id: number; type: number} | null,
	sessionEpoch: 1,
	apiV2BaseUrl: 'https://identity-a.example/api/v2/',
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/helpers/auth', () => ({
	getAuthSessionEpoch: () => requestContext.sessionEpoch,
	getToken: () => null,
	getTokenIdentity: () => requestContext.identity,
}))
vi.mock('@/helpers/fetcher', () => ({
	getApiV2BaseUrl: () => requestContext.apiV2BaseUrl,
}))

import {
	createProjectViewMutationOptions,
	createProjectViewDraft,
	createProjectViewUpdate,
	deleteProjectViewMutationOptions,
	sortProjectViewsByPosition,
	updateProjectViewMutationOptions,
	type CreateProjectViewInput,
	type UpdateProjectViewInput,
	type DeleteProjectViewInput,
} from './projectViews'

function execute<TData, TVariables, TContext>(
	options: MutationOptions<TData, Error, TVariables, TContext>,
	variables: TVariables,
	client = queryClient,
) {
	return client.getMutationCache().build(client, options).execute(variables)
}

const createProjectView = (input: CreateProjectViewInput) => execute(createProjectViewMutationOptions(), input)
const updateProjectView = (input: UpdateProjectViewInput) => execute(updateProjectViewMutationOptions(), input)
const deleteProjectView = (input: DeleteProjectViewInput) => execute(deleteProjectViewMutationOptions(), input)

const views: ProjectView[] = [
	{id: 3, project_id: 7, title: 'Table', view_kind: 'table', position: 30},
	{id: 1, project_id: 7, title: 'List', view_kind: 'list', position: 10},
	{id: 2, project_id: 7, title: 'Board', view_kind: 'kanban', position: 20},
]

function embeddedViews(client = queryClient) {
	return client.getQueryData<ProjectListResult>(projectKeys.list())?.projects[0].views
}

beforeEach(() => {
	requestContext.identity = {id: 1, type: 1}
	requestContext.sessionEpoch = 1
	requestContext.apiV2BaseUrl = 'https://identity-a.example/api/v2/'
})

describe('project view queries', () => {
	beforeEach(() => {
		queryClient.clear()
		vi.clearAllMocks()
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
	const delayedMutationCases = [
		{
			name: 'create',
			mock: sdk.projectViewsCreate,
			run: () => createProjectView({
				projectId: 7,
				view: createProjectViewDraft({title: 'Identity A view'}),
			}),
			response: {data: {...views[0], id: 4, title: 'Identity A view'}},
		},
		{
			name: 'update',
			mock: sdk.projectViewsUpdate,
			run: () => updateProjectView({
				projectId: 7,
				viewId: 2,
				view: createProjectViewUpdate({...views[2], title: 'Identity A update'}),
			}),
			response: {data: {...views[2], title: 'Identity A update'}},
		},
		{
			name: 'delete',
			mock: sdk.projectViewsDelete,
			run: () => deleteProjectView({projectId: 7, viewId: 2}),
			response: {data: undefined},
		},
	]
	beforeEach(() => {
		queryClient.clear()
		vi.clearAllMocks()
		queryClient.setQueryData<ProjectListResult>(projectKeys.list(), {
			projects: [normalizeProject({id: 7, title: 'Project', max_permission: 2, views})],
			favoriteProject: null,
			savedFilterProjects: [],
		})
		queryClient.setQueryData(projectKeys.detail(7), normalizeProject({id: 7, title: 'Project', views}))
	})

	describe('after the identity changes', () => {
		it.each(delayedMutationCases)('discards a delayed $name completion', async ({mock, run, response}) => {
			let resolveRequest: (value: unknown) => void = () => {}
			mock.mockReturnValue(new Promise(resolve => {
				resolveRequest = resolve
			}))

			const mutation = run()
			await vi.waitFor(() => expect(mock).toHaveBeenCalledOnce())
			requestContext.identity = {id: 2, type: 1}
			queryClient.clear()
			const identityBViews = views.map(view => ({...view, title: `Identity B ${view.title}`}))
			queryClient.setQueryData(projectKeys.detail(7), normalizeProject({id: 7, views: identityBViews}))
			resolveRequest(response)

			await expect(mutation).rejects.toMatchObject({name: 'AbortError'})
			expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(7))?.views).toEqual(sortProjectViewsByPosition(identityBViews))
			expect(queryClient.getQueryState(projectKeys.detail(7))?.isInvalidated).toBe(false)
			expect(success).not.toHaveBeenCalled()
			expect(error).not.toHaveBeenCalled()
		})
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
		expect(embeddedViews()?.map(view => view.id))
			.toEqual([1, 4, 2, 3])
		expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(7))?.views).toEqual(embeddedViews())
		expect(queryClient.getQueryCache().findAll({queryKey: ['project-views']})).toEqual([])
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
		expect(embeddedViews()?.map(view => view.id))
			.toEqual([2, 1, 3])
		expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(7))?.views).toEqual(embeddedViews())
	})

	it('removes a deleted view from view and nested project caches', async () => {
		sdk.projectViewsDelete.mockResolvedValue({data: undefined})

		await deleteProjectView({projectId: 7, viewId: 2})

		expect(sdk.projectViewsDelete).toHaveBeenCalledWith({path: {project: 7, view: 2}})
		expect(embeddedViews()?.map(view => view.id))
			.toEqual([1, 3])
		expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(7))?.views).toEqual(embeddedViews())
	})

	it.each(['update', 'delete'] as const)('rolls back an optimistic %s in views and embedded project views', async operation => {
		const originalProjects = queryClient.getQueryData(projectKeys.list())
		const originalDetail = queryClient.getQueryData(projectKeys.detail(7))
		const failure = new Error('Request failed')
		const mock = operation === 'update' ? sdk.projectViewsUpdate : sdk.projectViewsDelete
		mock.mockImplementation(async () => {
			const expected = operation === 'update'
				? expect.arrayContaining([expect.objectContaining({id: 2, default_bucket_id: 42, done_bucket_id: 43})])
				: expect.not.arrayContaining([expect.objectContaining({id: 2})])
			expect(embeddedViews()).toEqual(expected)
			expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(7))?.views).toEqual(expected)
			throw failure
		})

		const mutation = operation === 'update'
			? updateProjectView({projectId: 7, viewId: 2, view: {default_bucket_id: 42, done_bucket_id: 43}})
			: deleteProjectView({projectId: 7, viewId: 2})
		await expect(mutation).rejects.toBe(failure)

		expect(queryClient.getQueryData(projectKeys.list())).toEqual(originalProjects)
		expect(queryClient.getQueryData(projectKeys.detail(7))).toEqual(originalDetail)
		expect(queryClient.getQueryState(projectKeys.list())?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(projectKeys.detail(7))?.isInvalidated).toBe(true)
	})

	it('marks the project list stale instead of reloading every project', async () => {
		sdk.projectViewsCreate.mockResolvedValue({data: {...views[0], id: 4}})
		const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

		try {
			await createProjectView({projectId: 7, view: {title: 'New'}})

			expect(invalidate).toHaveBeenCalledWith({queryKey: projectKeys.list(), refetchType: 'none'})
			expect(invalidate).toHaveBeenCalledWith({queryKey: projectKeys.detail(7)})
		} finally {
			invalidate.mockRestore()
		}
	})

	it.each(delayedMutationCases)('does not materialize unmounted caches on $name', async ({mock, run, response}) => {
		queryClient.clear()
		mock.mockResolvedValue(response)

		await run()

		expect(queryClient.getQueryCache().getAll()).toEqual([])
	})

	it('writes through the executing client and preserves project metadata and bucket settings', async () => {
		const client = new QueryClient()
		const project = normalizeProject({id: 7, title: 'Board project', max_permission: 2, views})
		client.setQueryData<ProjectListResult>(projectKeys.list(), {
			projects: [project], favoriteProject: null, savedFilterProjects: [],
		})
		client.setQueryData(projectKeys.detail(7), project)
		const updated = {...views[2], default_bucket_id: 42, done_bucket_id: 43}
		sdk.projectViewsUpdate.mockResolvedValue({data: updated})

		await execute(updateProjectViewMutationOptions('Bucket updated'), {
			projectId: 7, viewId: 2, view: createProjectViewUpdate(updated),
		}, client)

		expect(embeddedViews(client)?.find(view => view.id === 2)).toEqual(updated)
		expect(client.getQueryData<ProjectListResult>(projectKeys.list())?.projects[0])
			.toMatchObject({title: 'Board project', max_permission: 2})
		expect(client.getQueryData<ProjectResponse>(projectKeys.detail(7)))
			.toMatchObject({title: 'Board project', views: expect.arrayContaining([updated])})
		expect(embeddedViews()).toEqual(sortProjectViewsByPosition(views))
		expect(success).toHaveBeenCalledWith({message: 'Bucket updated'})
		client.clear()
	})

	it('suppresses notifications for a project the caller has left', async () => {
		sdk.projectViewsCreate.mockResolvedValue({data: {...views[0], id: 4}})
		await execute(createProjectViewMutationOptions(() => false), {projectId: 7, view: {title: 'New'}})
		sdk.projectViewsUpdate.mockRejectedValue(new Error('Failed'))
		await expect(execute(updateProjectViewMutationOptions(undefined, () => false), {
			projectId: 7, viewId: 2, view: {title: 'Edit'},
		})).rejects.toThrow('Failed')

		expect(success).not.toHaveBeenCalled()
		expect(error).not.toHaveBeenCalled()
	})
})
