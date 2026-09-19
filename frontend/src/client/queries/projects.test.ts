import type {MutationOptions} from '@tanstack/vue-query'
import {QueryClient} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {Project} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import type {ProjectListResult, ProjectResponse} from './projects'

const sdk = vi.hoisted(() => ({
	projectsList: vi.fn(),
	projectsRead: vi.fn(),
	projectsCreate: vi.fn(),
	projectsUpdate: vi.fn(),
	projectsDelete: vi.fn(),
	projectsDuplicate: vi.fn(),
	patchProjectsRead: vi.fn(),
	subscriptionsCreate: vi.fn(),
	subscriptionsDelete: vi.fn(),
}))

const requestContext = vi.hoisted(() => ({
	identity: {id: 1, type: 1} as {id: number; type: number} | null,
	sessionEpoch: 1,
	apiV2BaseUrl: 'https://identity-a.example/api/v2/',
}))

vi.mock('@/message', () => ({success: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/helpers/auth', () => ({
	getAuthSessionEpoch: () => requestContext.sessionEpoch,
	getToken: () => null,
	getTokenIdentity: () => requestContext.identity,
}))
vi.mock('@/helpers/fetcher', () => ({
	getApiV2BaseUrl: () => requestContext.apiV2BaseUrl,
}))

import {
	createProjectMutationOptions,
	createProjectDraft,
	deleteProjectMutationOptions,
	duplicateProjectMutationOptions,
	findProjectByExactTitle,
	getCachedProject,
	getProjectIdFromSavedFilterId,
	getSavedFilterIdFromProjectId,
	isSavedFilterProject,
	normalizeProject,
	projectKeys,
	projectQuery,
	patchProjectFavoriteMutationOptions,
	projectsQuery,
	setProjectSubscriptionMutationOptions,
	updateProjectMutationOptions,
} from './projects'
import {taskKeys} from './tasks'

function execute<TData, TVariables, TContext>(
	options: MutationOptions<TData, Error, TVariables, TContext>,
	variables: TVariables,
	client = queryClient,
) {
	return client.getMutationCache().build(client, options).execute(variables)
}

const createProject = (input: Parameters<NonNullable<ReturnType<typeof createProjectMutationOptions>['mutationFn']>>[0]) =>
	execute(createProjectMutationOptions(), input)
const updateProject = (input: Parameters<NonNullable<ReturnType<typeof updateProjectMutationOptions>['mutationFn']>>[0]) =>
	execute(updateProjectMutationOptions(), input)
const patchProjectFavorite = (id: number, isFavorite: boolean) =>
	execute(patchProjectFavoriteMutationOptions(), {id, isFavorite})
const deleteProject = (id: number) => execute(deleteProjectMutationOptions(), id)
const duplicateProject = (input: Parameters<NonNullable<ReturnType<typeof duplicateProjectMutationOptions>['mutationFn']>>[0]) =>
	execute(duplicateProjectMutationOptions(), input)

function serverProject(overrides: Partial<ProjectResponse> = {}): ProjectResponse {
	return {
		id: 1,
		title: 'Project',
		description: '',
		hex_color: '',
		identifier: '',
		is_archived: false,
		is_favorite: false,
		parent_project_id: 0,
		position: 0,
		views: [],
		...overrides,
	}
}

beforeEach(() => {
	requestContext.identity = {id: 1, type: 1}
	requestContext.sessionEpoch = 1
	requestContext.apiV2BaseUrl = 'https://identity-a.example/api/v2/'
})

describe('project queries', () => {
	beforeEach(() => {
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
	})

	it('reads the detail before falling back to the navigation list', () => {
		const listed = serverProject({title: 'Listed'})
		const html = serverProject({title: 'Detailed', description: '<p>HTML</p>'})
		queryClient.setQueryData(projectKeys.list(), {projects: [listed], favoriteProject: null, savedFilterProjects: []})
		expect(getCachedProject(1)).toEqual(listed)
		queryClient.setQueryData(projectKeys.detail(1), html)
		expect(getCachedProject(1)).toEqual(html)
	})

	it('reads pseudo projects from the navigation list', () => {
		const favorites = serverProject({id: -1, title: 'Favorites'})
		const filter = serverProject({id: -2, title: 'Filter'})
		queryClient.setQueryData(projectKeys.list(), {
			projects: [],
			favoriteProject: favorites,
			savedFilterProjects: [filter],
		})

		expect(getCachedProject(-1)).toEqual(favorites)
		expect(getCachedProject(-2)).toEqual(filter)
		expect(getCachedProject(-3)).toBeUndefined()
	})

	it('returns undefined for an uncached project without fetching or creating query state', () => {
		expect(getCachedProject(1)).toBeUndefined()
		expect(queryClient.getQueryCache().getAll()).toEqual([])
		expect(sdk.projectsRead).not.toHaveBeenCalled()
		expect(sdk.projectsList).not.toHaveBeenCalled()
	})

	it('requests projects with the maximum page size and partitions pseudo projects from real ones, deduping repeated ids', async () => {
		sdk.projectsList.mockResolvedValue({
			data: {
				items: [
					serverProject({id: 2, title: 'Second', position: 200}),
					serverProject({id: -1, title: 'Favorites', is_favorite: true, position: -1}),
					serverProject({id: -3, title: 'Zulu filter'}),
					serverProject({id: 1, title: 'First', position: 100}),
					serverProject({id: -2, title: 'Alpha filter'}),
					serverProject({id: -3, title: 'Zulu filter'}),
				],
				total_pages: 1,
			},
		})

		const result = await queryClient.fetchQuery(projectsQuery())

		expect(result.projects.map(project => project.id)).toEqual([1, 2])
		expect(result.favoriteProject?.id).toBe(-1)
		expect(result.savedFilterProjects.map(project => project.id)).toEqual([-2, -3])
		expect(sdk.projectsList).toHaveBeenCalledExactlyOnceWith({
			query: {is_archived: true, expand: 'permissions', page: 1, per_page: 1000},
		})
	})

	it('normalizes fields required by project consumers at the read boundary', async () => {
		sdk.projectsList.mockResolvedValue({
			data: {items: [{id: 1, title: 'Minimal'}], total_pages: 1},
		})

		const result = await queryClient.fetchQuery(projectsQuery())

		expect(result.projects[0]).toMatchObject({
			id: 1,
			title: 'Minimal',
			description: '',
			hex_color: '',
			identifier: '',
			is_archived: false,
			is_favorite: false,
			parent_project_id: 0,
			position: 0,
			views: [],
		})
	})

	it('normalizes project colors for CSS consumers', () => {
		expect(normalizeProject({id: 1, hex_color: '00db60'}).hex_color).toBe('#00db60')
		expect(normalizeProject({id: 1, hex_color: '#00db60'}).hex_color).toBe('#00db60')
		expect(normalizeProject({id: 1, hex_color: ''}).hex_color).toBe('')
	})

	it('loads pseudo-project details but rejects an unset id', () => {
		expect(projectQuery(-2).enabled).toBe(true)
		expect(projectQuery(-1).enabled).toBe(true)
		expect(projectQuery(0).enabled).toBe(false)
	})

})

describe('project lookups', () => {
	const projects = [
		serverProject({id: 1, title: 'Root'}),
		serverProject({id: 2, title: 'Child'}),
	]

	it('finds projects by exact title case-insensitively', () => {
		expect(findProjectByExactTitle(projects, 'root')?.id).toBe(1)
		expect(findProjectByExactTitle(projects, 'missing')).toBeNull()
	})
})

describe('project drafts and cache mutations', () => {
	const listKey = projectKeys.list()
	const delayedMutationCases = [
		{
			name: 'create',
			mock: sdk.projectsCreate,
			run: () => createProject({title: 'Identity A project'}),
			response: {data: serverProject({id: 5, title: 'Identity A project'})},
		},
		{
			name: 'update',
			mock: sdk.projectsUpdate,
			run: () => updateProject({...serverProject(), title: 'Identity A update'}),
			response: {data: serverProject({title: 'Identity A update'})},
		},
		{
			name: 'favorite',
			mock: sdk.patchProjectsRead,
			run: () => patchProjectFavorite(1, true),
			response: {data: serverProject({is_favorite: true})},
		},
		{
			name: 'subscribe',
			mock: sdk.subscriptionsCreate,
			run: () => execute(setProjectSubscriptionMutationOptions(), {projectId: 1, subscribed: true}),
			response: {data: {id: 7, entity: 'project', entity_id: 1}},
		},
		{
			name: 'delete',
			mock: sdk.projectsDelete,
			run: () => deleteProject(1),
			response: {data: undefined},
		},
		{
			name: 'duplicate',
			mock: sdk.projectsDuplicate,
			run: () => duplicateProject({projectId: 1}),
			response: {data: {duplicated_project: serverProject({id: 9, title: 'Identity A copy'})}},
		},
	]

	beforeEach(() => {
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
	})

	describe.each([
		['identity', () => { requestContext.identity = {id: 2, type: 1} }],
		['session', () => { requestContext.sessionEpoch++ }],
		['API URL', () => { requestContext.apiV2BaseUrl = 'https://identity-b.example/api/v2/' }],
	] as const)('after the %s changes', (_name, switchContext) => {
		it.each(delayedMutationCases)('discards a delayed $name completion', async ({mock, run, response}) => {
			const identityAProject = serverProject({title: 'Identity A project'})
			const identityBProject = serverProject({title: 'Identity B project'})
			queryClient.setQueryData(listKey, {
				projects: [identityAProject],
				favoriteProject: null,
				savedFilterProjects: [],
			})
			queryClient.setQueryData(projectKeys.detail(1), identityAProject)
			let resolveRequest: (value: unknown) => void = () => {}
			mock.mockReturnValue(new Promise(resolve => {
				resolveRequest = resolve
			}))

			const mutation = run()
			await vi.waitFor(() => expect(mock).toHaveBeenCalledOnce())
			switchContext()
			queryClient.clear()
			const identityBList = {
				projects: [identityBProject],
				favoriteProject: null,
				savedFilterProjects: [],
			}
			queryClient.setQueryData(listKey, identityBList)
			queryClient.setQueryData(projectKeys.detail(1), identityBProject)
			resolveRequest(response)

			await expect(mutation).rejects.toMatchObject({name: 'AbortError'})
			expect(queryClient.getQueryData(listKey)).toEqual(identityBList)
			expect(queryClient.getQueryData(projectKeys.detail(1))).toEqual(identityBProject)
			expect(queryClient.getQueryData(projectKeys.detail(5))).toBeUndefined()
			expect(queryClient.getQueryData(projectKeys.detail(9))).toBeUndefined()
			expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(false)
		})
	})

	it.each(['update', 'favorite', 'delete'] as const)('rolls back an optimistic %s after failure', async operation => {
		const cached = serverProject({title: 'Before', description: '<p>html</p>'})
		const previous = {projects: [cached], favoriteProject: null, savedFilterProjects: []}
		queryClient.setQueryData(listKey, previous)
		queryClient.setQueryData(projectKeys.detail(1), cached)
		queryClient.setQueryData(projectKeys.detail(99), serverProject({id: 99}))
		const failure = new Error('Request failed')
		const selected = delayedMutationCases.find(test => test.name === operation)!
		selected.mock.mockImplementation(async () => {
			queryClient.setQueryData(projectKeys.detail(99), serverProject({id: 99, title: 'Independent update'}))
			const current = queryClient.getQueryData<ProjectListResult>(listKey)!
			if (operation === 'delete') {
				expect(current.projects).toEqual([])
			} else if (operation === 'favorite') {
				expect(current.projects[0].is_favorite).toBe(true)
			} else {
				expect(current.projects[0].title).toBe('Identity A update')
			}
			throw failure
		})

		await expect(selected.run()).rejects.toThrow(failure)

		expect(queryClient.getQueryData(listKey)).toEqual(previous)
		expect(queryClient.getQueryData(projectKeys.detail(1))).toEqual(cached)
		expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(99))?.title).toBe('Independent update')
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it.each(delayedMutationCases)('does not materialize absent caches during $name', async ({mock, run, response}) => {
		mock.mockResolvedValue(response)
		await run()
		expect(queryClient.getQueryCache().getAll()).toEqual([])
	})

	it('writes to the mutation client rather than the singleton', async () => {
		const client = new QueryClient()
		client.setQueryData(listKey, {projects: [], favoriteProject: null, savedFilterProjects: []})
		sdk.projectsCreate.mockResolvedValue({data: serverProject({id: 5})})
		await execute(createProjectMutationOptions(), {title: 'New'}, client)
		expect(client.getQueryData<ProjectListResult>(listKey)?.projects[0].id).toBe(5)
		expect(queryClient.getQueryData(listKey)).toBeUndefined()
		client.clear()
	})

	it('creates a generated-type draft with stable UI defaults', () => {
		expect(createProjectDraft()).toEqual({
			title: '',
			description: '',
			hex_color: '',
			identifier: '',
			is_archived: false,
			is_favorite: false,
			parent_project_id: 0,
			position: 0,
		})
	})

	it('does not seed the detail cache from a sparse create response', async () => {
		const created = {id: 5, title: 'Created', hex_color: 'abcdef'}
		const hydrated = serverProject({id: 5, title: 'Created', hex_color: 'abcdef'})
		const normalizedCreated = {...hydrated, hex_color: '#abcdef'}
		queryClient.setQueryData(listKey, {projects: [], favoriteProject: null, savedFilterProjects: []})
		sdk.projectsCreate.mockResolvedValue({data: created})
		sdk.projectsRead.mockResolvedValue({data: hydrated})

		await expect(createProject({title: 'Created', hex_color: '#abcdef'})).resolves.toMatchObject(normalizedCreated)

		expect(sdk.projectsCreate).toHaveBeenCalledWith({
			body: {title: 'Created', hex_color: 'abcdef'},
		})
		expect(sdk.projectsRead).not.toHaveBeenCalled()
		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects).toContainEqual(normalizedCreated)
		expect(queryClient.getQueryData(projectKeys.detail(5))).toBeUndefined()
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('updates list and detail caches and preserves read-only permission state', async () => {
		const cached = serverProject({id: 1, title: 'Before', max_permission: 2})
		queryClient.setQueryData(listKey, {projects: [cached], favoriteProject: null, savedFilterProjects: []})
		queryClient.setQueryData(projectKeys.detail(1), cached)
		sdk.projectsUpdate.mockResolvedValue({
			data: serverProject({id: 1, title: 'After', max_permission: null}),
		})

		await updateProject({...cached, title: 'After', hex_color: '#ff00aa'})

		expect(sdk.projectsUpdate).toHaveBeenCalledWith({
			path: {id: 1},
			body: {
				title: 'After',
				description: '',
				hex_color: 'ff00aa',
				identifier: '',
				is_archived: false,
				is_favorite: false,
				parent_project_id: 0,
				position: 0,
			},
		})
		expect(queryClient.getQueryData<Project>(projectKeys.detail(1))).toMatchObject({
			title: 'After',
			max_permission: 2,
		})
		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects[0]).toMatchObject({
			title: 'After',
			max_permission: 2,
		})
		expect(queryClient.getQueryState(projectKeys.detail(1))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('preserves list-only permission and views when the update response is sparse', async () => {
		const views = [{id: 12, project_id: 1, title: 'List', view_kind: 'list'}] as const
		const cached = serverProject({id: 1, title: 'Before', max_permission: 2, views: [...views]})
		queryClient.setQueryData(listKey, {projects: [cached], favoriteProject: null, savedFilterProjects: []})
		sdk.projectsUpdate.mockResolvedValue({
			data: {id: 1, title: 'After', max_permission: null, views: null},
		})

		const updated = await updateProject({...cached, title: 'After'})

		expect(updated.title).toBe('After')
		expect(queryClient.getQueryData(projectKeys.detail(1))).toBeUndefined()
		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects[0]).toMatchObject({
			title: 'After',
			max_permission: 2,
			views,
		})
	})

	it('patches only the favorite field', async () => {
		const cached = serverProject({id: 1, title: 'Project', is_favorite: false})
		queryClient.setQueryData(listKey, {projects: [cached], favoriteProject: null, savedFilterProjects: []})
		queryClient.setQueryData(projectKeys.detail(1), cached)
		sdk.patchProjectsRead.mockResolvedValue({data: {...cached, is_favorite: true}})

		await patchProjectFavorite(1, true)

		expect(sdk.patchProjectsRead).toHaveBeenCalledWith({
			path: {id: 1},
			body: [{op: 'replace', path: '/is_favorite', value: true}],
		})
		expect(queryClient.getQueryData<Project>(projectKeys.detail(1))).toMatchObject({
			title: 'Project',
			is_favorite: true,
		})
	})

	it('removes a deleted project, all descendants, and their detail entries', async () => {
		queryClient.setQueryData(listKey, {
			projects: [
				serverProject({id: 1}),
				serverProject({id: 2, parent_project_id: 1}),
				serverProject({id: 3, parent_project_id: 2}),
				serverProject({id: 4}),
			],
			favoriteProject: null,
			savedFilterProjects: [serverProject({id: -2})],
		})
		for (const id of [1, 2, 3, 4]) {
			queryClient.setQueryData(projectKeys.detail(id), serverProject({id}))
		}
		sdk.projectsDelete.mockResolvedValue({data: undefined})

		await deleteProject(1)

		expect(sdk.projectsDelete).toHaveBeenCalledWith({path: {id: 1}})
		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects.map(project => project.id)).toEqual([4])
		expect(queryClient.getQueryData(projectKeys.detail(1))).toBeUndefined()
		expect(queryClient.getQueryData(projectKeys.detail(2))).toBeUndefined()
		expect(queryClient.getQueryData(projectKeys.detail(3))).toBeUndefined()
		expect(queryClient.getQueryData(projectKeys.detail(4))).toBeDefined()
	})

	it('duplicates a project under a parent and caches the duplicate as admin', async () => {
		queryClient.setQueryData(listKey, {projects: [], favoriteProject: null, savedFilterProjects: []})
		sdk.projectsDuplicate.mockResolvedValue({
			data: {duplicated_project: serverProject({id: 9, title: 'Copy', parent_project_id: 4})},
		})

		const duplicate = await duplicateProject({
			projectId: 1,
			parentProjectId: 4,
			duplicateShares: true,
		})

		expect(sdk.projectsDuplicate).toHaveBeenCalledWith({
			path: {projectid: 1},
			body: {parent_project_id: 4, duplicate_shares: true},
		})
		expect(duplicate).toMatchObject({id: 9, max_permission: 2})
		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects).toContainEqual(duplicate)
		expect(queryClient.getQueryData(projectKeys.detail(9))).toBeUndefined()
	})

	it('patches the subscription into the cached project and marks sub projects and tasks stale without refetching the list', async () => {
		const cached = serverProject({id: 1})
		const child = serverProject({id: 2, parent_project_id: 1})
		queryClient.setQueryData(listKey, {projects: [cached, child], favoriteProject: null, savedFilterProjects: []})
		queryClient.setQueryData(projectKeys.detail(1), cached)
		queryClient.setQueryData(projectKeys.detail(2), child)
		queryClient.setQueryData(taskKeys.detail(5), {id: 5})
		const subscription = {id: 7, entity: 'project', entity_id: 1} as const
		sdk.subscriptionsCreate.mockResolvedValue({data: subscription})

		await execute(setProjectSubscriptionMutationOptions(), {projectId: 1, subscribed: true})

		expect(sdk.subscriptionsCreate).toHaveBeenCalledWith({path: {entity: 'project', entityID: 1}})
		expect(queryClient.getQueryData<ProjectListResult>(listKey)?.projects[0].subscription).toEqual(subscription)
		expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(1))?.subscription).toEqual(subscription)
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(projectKeys.detail(1))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(projectKeys.detail(2))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(taskKeys.detail(5))?.isInvalidated).toBe(true)
		expect(sdk.projectsList).not.toHaveBeenCalled()

		sdk.subscriptionsDelete.mockResolvedValue({data: undefined})
		await execute(setProjectSubscriptionMutationOptions(), {projectId: 1, subscribed: false})

		expect(sdk.subscriptionsDelete).toHaveBeenCalledWith({path: {entity: 'project', entityID: 1}})
		expect(queryClient.getQueryData<ProjectListResult>(listKey)?.projects[0].subscription).toBeUndefined()
		expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(1))?.subscription).toBeUndefined()
	})
})

describe('saved filter project ids', () => {
	it('maps saved filters to negative pseudo-project ids and back', () => {
		expect(getProjectIdFromSavedFilterId(1)).toBe(-2)
		expect(getProjectIdFromSavedFilterId(42)).toBe(-43)
		expect(getProjectIdFromSavedFilterId(0)).toBe(0)
		expect(getProjectIdFromSavedFilterId(-1)).toBe(0)
		expect(getSavedFilterIdFromProjectId(-2)).toBe(1)
		expect(getSavedFilterIdFromProjectId(-43)).toBe(42)
		expect(getSavedFilterIdFromProjectId(-1)).toBe(0)
		expect(getSavedFilterIdFromProjectId(1)).toBe(0)
	})

	it('recognizes only saved-filter pseudo-projects', () => {
		expect(isSavedFilterProject({id: -2})).toBe(true)
		expect(isSavedFilterProject({id: -1})).toBe(false)
		expect(isSavedFilterProject({id: 1})).toBe(false)
		expect(isSavedFilterProject(null)).toBe(false)
	})
})
