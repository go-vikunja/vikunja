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
}))

const requestContext = vi.hoisted(() => ({
	identity: {id: 1, type: 1} as {id: number; type: number} | null,
	sessionEpoch: 1,
	apiV2BaseUrl: 'https://identity-a.example/api/v2/',
}))

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
	createProject,
	createProjectDraft,
	deleteProject,
	duplicateProject,
	findProjectByExactTitle,
	findProjectByIdentifier,
	getChildProjects,
	getEffectiveParentProjectId,
	getFavoriteNavigationItems,
	getProjectAncestors,
	getRootProjects,
	invalidateProjects,
	normalizeProject,
	projectKeys,
	projectQuery,
	patchProjectFavorite,
	projectsQuery,
	searchProjects,
	updateProject,
} from './projects'

const listArgs = {
	is_archived: true,
	expand: 'permissions',
	format: 'markdown',
	q: 'roadmap',
} as const

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

	it('keys lists by every response-shaping argument and details by format', () => {
		expect(projectKeys.all).toEqual(['projects'])
		expect(projectKeys.lists()).toEqual(['projects', 'list'])
		expect(projectKeys.list(listArgs)).toEqual(['projects', 'list', listArgs])
		expect(projectKeys.details()).toEqual(['projects', 'detail'])
		expect(projectKeys.detail(42)).toEqual(['projects', 'detail', 42, 'html'])
		expect(projectKeys.detail(42, 'markdown')).toEqual(['projects', 'detail', 42, 'markdown'])
	})

	it('invalidates every project list and detail cache', async () => {
		const projectQueryKeys = [
			projectKeys.list(),
			projectKeys.list(listArgs),
			projectKeys.detail(1),
			projectKeys.detail(2, 'markdown'),
		]
		const unrelatedQueryKey = ['tasks', 'list'] as const

		for (const queryKey of [...projectQueryKeys, unrelatedQueryKey]) {
			queryClient.setQueryData(queryKey, {})
		}

		await invalidateProjects()

		for (const queryKey of projectQueryKeys) {
			expect(queryClient.getQueryState(queryKey)?.isInvalidated).toBe(true)
		}
		expect(queryClient.getQueryState(unrelatedQueryKey)?.isInvalidated).toBe(false)
	})

	it('loads every page and partitions pseudo projects from real projects', async () => {
		sdk.projectsList
			.mockResolvedValueOnce({
				data: {
					items: [
						serverProject({id: 2, title: 'Second', position: 200}),
						serverProject({id: -1, title: 'Favorites', is_favorite: true, position: -1}),
						serverProject({id: -3, title: 'Zulu filter'}),
					],
					total_pages: 2,
				},
			})
			.mockResolvedValueOnce({
				data: {
					items: [
						serverProject({id: 1, title: 'First', position: 100}),
						serverProject({id: -2, title: 'Alpha filter'}),
						serverProject({id: -3, title: 'Zulu filter'}),
					],
					total_pages: 2,
				},
			})

		const result = await queryClient.fetchQuery(projectsQuery(listArgs))

		expect(result.projects.map(project => project.id)).toEqual([1, 2])
		expect(result.favoriteProject?.id).toBe(-1)
		expect(result.savedFilterProjects.map(project => project.id)).toEqual([-2, -3])
		expect(sdk.projectsList).toHaveBeenNthCalledWith(1, {
			query: {...listArgs, page: 1, per_page: 1000},
		})
		expect(sdk.projectsList).toHaveBeenNthCalledWith(2, {
			query: {...listArgs, page: 2, per_page: 1000},
		})
	})

	it('normalizes fields required by project consumers at the read boundary', async () => {
		sdk.projectsList.mockResolvedValue({
			data: {items: [{id: 1, title: 'Minimal'}], total_pages: 1},
		})

		const result = await queryClient.fetchQuery(projectsQuery({is_archived: true}))

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

describe('project hierarchy and navigation derivations', () => {
	const projects = [
		serverProject({id: 1, title: 'Root', identifier: 'ROOT', position: 200}),
		serverProject({id: 2, title: 'Child', description: 'nested work', parent_project_id: 1, position: 300}),
		serverProject({id: 3, title: 'Orphan', parent_project_id: 99, position: 100}),
		serverProject({id: 4, title: 'Archived', is_archived: true, position: 400}),
	]

	it('derives roots, children, ancestors, and effective drag parents', () => {
		const grandchild = serverProject({id: 5, title: 'Grandchild', parent_project_id: 2, position: 50})
		const sibling = serverProject({id: 6, title: 'Sibling', parent_project_id: 1, position: 100})
		const tree = [...projects, grandchild, sibling]

		expect(getRootProjects(projects).map(project => project.id)).toEqual([3, 1])
		expect(getChildProjects(tree, 1).map(project => project.id)).toEqual([6, 2])
		expect(getProjectAncestors(tree, grandchild).map(project => project.id)).toEqual([1, 2, 5])
		expect(getProjectAncestors(tree, projects[2]).map(project => project.id)).toEqual([3])
		expect(getEffectiveParentProjectId(projects, projects[2], 0)).toBe(99)
		expect(getEffectiveParentProjectId(projects, projects[2], 1)).toBe(1)
		expect(getEffectiveParentProjectId(projects, projects[1], 0)).toBe(0)
	})

	it('looks up and searches projects case-insensitively', () => {
		expect(findProjectByExactTitle(projects, 'root')?.id).toBe(1)
		expect(findProjectByIdentifier(projects, 'root')?.id).toBe(1)
		expect(findProjectByExactTitle(projects, 'missing')).toBeNull()
		expect(findProjectByIdentifier(projects, 'missing')).toBeNull()
		expect(searchProjects(projects, 'NEST').map(project => project.id)).toEqual([2])
		expect(searchProjects(projects, 'archived')).toEqual([])
		expect(searchProjects(projects, 'archived', true).map(project => project.id)).toEqual([4])
		expect(searchProjects(projects, '')).toEqual([])
	})

	it('combines favorite pseudo, project, and saved-filter navigation items', () => {
		const list: ProjectListResult = {
			projects: [
				serverProject({id: 1, title: 'Favorite project', is_favorite: true, position: 100}),
				serverProject({id: 2, title: 'Archived favorite', is_favorite: true, is_archived: true, position: 200}),
			],
			favoriteProject: serverProject({id: -1, title: 'Favorites', is_favorite: true, position: -1}),
			savedFilterProjects: [
				serverProject({id: -2, title: 'Favorite filter', is_favorite: true}),
				serverProject({id: -3, title: 'Other filter'}),
			],
		}

		expect(getFavoriteNavigationItems(list).map(project => project.id)).toEqual([-1, -2, 1])
	})
})

describe('project drafts and cache mutations', () => {
	const listKey = projectKeys.list({is_archived: true, expand: 'permissions'})
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
	const contextChanges = [
		{
			name: 'authenticated session for the same identity',
			change: () => {
				requestContext.sessionEpoch++
			},
		},
		{
			name: 'authenticated identity',
			change: () => {
				requestContext.identity = {id: 2, type: 1}
			},
		},
		{
			name: 'API origin',
			change: () => {
				requestContext.apiV2BaseUrl = 'https://identity-b.example/api/v2/'
			},
		},
	]

	beforeEach(() => {
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
	})

	describe.each(contextChanges)('after the $name changes', ({change}) => {
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
			change()
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
			query: {format: 'html'},
		})
		expect(sdk.projectsRead).not.toHaveBeenCalled()
		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects).toContainEqual(normalizedCreated)
		expect(queryClient.getQueryData(projectKeys.detail(5))).toBeUndefined()
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('does not let an older list response overwrite a created project', async () => {
		queryClient.setQueryData(listKey, {projects: [], favoriteProject: null, savedFilterProjects: []})
		let resolveList: (value: unknown) => void = () => {}
		sdk.projectsList.mockReturnValue(new Promise(resolve => {
			resolveList = resolve
		}))
		const list = queryClient.fetchQuery({...projectsQuery({is_archived: true, expand: 'permissions'}), staleTime: 0})
		const listSettled = list.catch(() => undefined)
		const created = serverProject({id: 5, title: 'Created'})
		sdk.projectsCreate.mockResolvedValue({data: created})

		await createProject({title: 'Created'})
		resolveList({data: {items: [], total_pages: 1}})
		await listSettled

		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects).toContainEqual(created)
	})

	it('updates the selected response format and preserves read-only permission state', async () => {
		const cached = serverProject({id: 1, title: 'Before', max_permission: 2})
		const markdown = serverProject({id: 1, title: 'Before', description: 'markdown', max_permission: 2})
		queryClient.setQueryData(listKey, {projects: [cached], favoriteProject: null, savedFilterProjects: []})
		queryClient.setQueryData(projectKeys.detail(1), cached)
		queryClient.setQueryData(projectKeys.detail(1, 'markdown'), markdown)
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
			query: {format: 'html'},
		})
		expect(queryClient.getQueryData<Project>(projectKeys.detail(1))).toMatchObject({
			title: 'After',
			max_permission: 2,
		})
		expect(queryClient.getQueryData<{projects: Project[]}>(listKey)?.projects[0]).toMatchObject({
			title: 'After',
			max_permission: 2,
		})
		expect(queryClient.getQueryData(projectKeys.detail(1, 'markdown'))).toBeUndefined()
	})

	it('preserves list-only permission and views when the update response is sparse', async () => {
		const views = [{id: 12, project_id: 1, title: 'List', view_kind: 'list'}] as const
		const cached = serverProject({id: 1, title: 'Before', max_permission: 2, views: [...views]})
		queryClient.setQueryData(listKey, {projects: [cached], favoriteProject: null, savedFilterProjects: []})
		sdk.projectsUpdate.mockResolvedValue({
			data: {id: 1, title: 'After', max_permission: null, views: null},
		})

		const updated = await updateProject({...cached, title: 'After'})

		expect(updated).toMatchObject({
			title: 'After',
			max_permission: 2,
			views,
		})
		expect(queryClient.getQueryData<Project>(projectKeys.detail(1))).toMatchObject({
			title: 'After',
			max_permission: 2,
			views,
		})
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
		expect(queryClient.getQueryData(projectKeys.detail(9))).toEqual(duplicate)
	})
})
