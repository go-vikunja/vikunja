import {QueryClient, type MutationOptions} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {ProjectListResult} from './projects'
import {normalizeProject, projectKeys} from './projects'
import type {SavedFilterResponse, UpdateSavedFilterInput} from './savedFilters'
import {queryClient} from '@/client/queryClient'

const sdk = vi.hoisted(() => ({
	filtersCreate: vi.fn(),
	filtersDelete: vi.fn(),
	filtersRead: vi.fn(),
	filtersUpdate: vi.fn(),
	patchFiltersRead: vi.fn(),
}))

const requestContext = vi.hoisted(() => ({
	identity: {id: 1, type: 1} as {id: number; type: number} | null,
	sessionEpoch: 1,
	apiV2BaseUrl: 'https://identity-a.example/api/v2/',
}))

vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
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
	createSavedFilterMutationOptions,
	deleteSavedFilterMutationOptions,
	newSavedFilterDraft,
	patchSavedFilterFavoriteMutationOptions,
	savedFilterKeys,
	savedFilterQuery,
	updateSavedFilterMutationOptions,
} from './savedFilters'

function execute<TData, TVariables, TContext>(
	options: MutationOptions<TData, Error, TVariables, TContext>,
	variables: TVariables,
	client = queryClient,
) {
	return client.getMutationCache().build(client, options).execute(variables)
}

const createSavedFilter = (input: Parameters<NonNullable<ReturnType<typeof createSavedFilterMutationOptions>['mutationFn']>>[0]) =>
	execute(createSavedFilterMutationOptions(), input)
const updateSavedFilter = (input: Parameters<NonNullable<ReturnType<typeof updateSavedFilterMutationOptions>['mutationFn']>>[0]) =>
	execute(updateSavedFilterMutationOptions(), input)
const deleteSavedFilter = (id: number) => execute(deleteSavedFilterMutationOptions(), id)
const patchSavedFilterFavorite = (id: number, isFavorite: boolean) =>
	execute(patchSavedFilterFavoriteMutationOptions(), {id, isFavorite})

function serverSavedFilter(overrides: Partial<SavedFilterResponse> = {}): SavedFilterResponse {
	return {
		id: 1,
		title: 'My filter',
		description: '',
		filters: {
			sort_by: ['done', 'id'],
			order_by: ['asc', 'desc'],
			filter: 'done = false',
			filter_include_nulls: true,
			s: '',
		},
		is_favorite: false,
		...overrides,
	}
}

const emptyProjectList: ProjectListResult = {
	projects: [],
	favoriteProject: null,
	savedFilterProjects: [],
}

beforeEach(() => {
	requestContext.identity = {id: 1, type: 1}
	requestContext.sessionEpoch = 1
	requestContext.apiV2BaseUrl = 'https://identity-a.example/api/v2/'
})

describe('saved filter queries', () => {
	beforeEach(() => {
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
	})

	it('uses the lifecycle client for saved-filter navigation updates', async () => {
		const client = new QueryClient()
		client.setQueryData(projectKeys.list(), {
			...emptyProjectList,
			savedFilterProjects: [normalizeProject({id: -9, title: 'Before'})],
		})
		sdk.patchFiltersRead.mockResolvedValue({data: serverSavedFilter({id: 8, is_favorite: true})})
		await execute(patchSavedFilterFavoriteMutationOptions(), {id: 8, isFavorite: true}, client)
		expect(client.getQueryData<ProjectListResult>(projectKeys.list())?.savedFilterProjects[0].is_favorite).toBe(true)
		expect(queryClient.getQueryData(projectKeys.list())).toBeUndefined()
		client.clear()
	})

	it('reads a saved filter without inventing creation defaults', async () => {
		sdk.filtersRead.mockResolvedValue({
			data: {id: 42, title: 'Upcoming', filters: {sort_by: null, order_by: null}},
		})

		const result = await queryClient.fetchQuery(savedFilterQuery(42))

		expect(sdk.filtersRead).toHaveBeenCalledWith({
			path: {filter: 42},
		})
		expect(result).toEqual({
			id: 42,
			title: 'Upcoming',
			description: '',
			filters: {
				sort_by: [],
				order_by: [],
				filter: '',
				filter_include_nulls: false,
				s: '',
			},
			is_favorite: false,
		})
	})

	it('creates a stable draft', () => {
		expect(newSavedFilterDraft()).toEqual({
			title: '',
			description: '',
			filters: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter: 'done = false',
				filter_include_nulls: true,
				s: '',
			},
		})
	})

	it('creates a saved filter without seeding an absent detail and invalidates project navigation', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		const created = serverSavedFilter({id: 8, title: 'Created'})
		sdk.filtersCreate.mockResolvedValue({data: created})

		await expect(createSavedFilter({title: 'Created'})).resolves.toEqual(created)

		expect(sdk.filtersCreate).toHaveBeenCalledWith({body: {title: 'Created'}})
		expect(queryClient.getQueryData(savedFilterKeys.detail(8))).toBeUndefined()
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('updates and invalidates the detail cache and project navigation', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		queryClient.setQueryData(savedFilterKeys.detail(8), serverSavedFilter({
			id: 8,
			title: 'Before',
			description: '<p>HTML</p>',
		}))
		sdk.filtersUpdate.mockResolvedValue({data: serverSavedFilter({id: 8, title: 'After'})})

		const writable: Omit<UpdateSavedFilterInput, 'id'> = {
			title: 'After',
			description: '',
			filters: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter: 'done = false',
				filter_include_nulls: true,
				s: '',
			},
			is_favorite: false,
		}
		await expect(updateSavedFilter({id: 8, ...writable})).resolves.toMatchObject({
			id: 8,
			title: 'After',
		})

		expect(sdk.filtersUpdate).toHaveBeenCalledWith({
			path: {filter: 8},
			body: writable,
		})
		expect(queryClient.getQueryData<SavedFilterResponse>(savedFilterKeys.detail(8))?.title).toBe('After')
		expect(queryClient.getQueryState(savedFilterKeys.detail(8))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('patches only the favorite field', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		const html = serverSavedFilter({id: 8, description: '<p>HTML</p>'})
		queryClient.setQueryData(savedFilterKeys.detail(8), html)
		sdk.patchFiltersRead.mockResolvedValue({data: {...html, is_favorite: true}})

		await patchSavedFilterFavorite(8, true)

		expect(sdk.patchFiltersRead).toHaveBeenCalledWith({
			path: {filter: 8},
			body: [{op: 'replace', path: '/is_favorite', value: true}],
		})
		expect(queryClient.getQueryData<SavedFilterResponse>(savedFilterKeys.detail(8))).toMatchObject({
			description: '<p>HTML</p>',
			is_favorite: true,
		})
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('deletes the cached detail and pseudo-project and invalidates project navigation', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		queryClient.setQueryData(savedFilterKeys.detail(8), serverSavedFilter({id: 8}))
		queryClient.setQueryData(projectKeys.detail(-9), {id: -9, title: 'Filter'})
		sdk.filtersDelete.mockResolvedValue({data: undefined})

		await deleteSavedFilter(8)

		expect(sdk.filtersDelete).toHaveBeenCalledWith({path: {filter: 8}})
		expect(queryClient.getQueryData(savedFilterKeys.detail(8))).toBeUndefined()
		expect(queryClient.getQueryData(projectKeys.detail(-9))).toBeUndefined()
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})
})

describe('saved filter mutations after the request context changes', () => {
	const listKey = projectKeys.list()
	const delayedMutationCases = [
		{
			name: 'create',
			mock: sdk.filtersCreate,
			run: () => createSavedFilter({title: 'Identity A filter'}),
			response: {data: serverSavedFilter({id: 8, title: 'Identity A filter'})},
		},
		{
			name: 'update',
			mock: sdk.filtersUpdate,
			run: () => updateSavedFilter({id: 8, ...newSavedFilterDraft(), title: 'Identity A update', is_favorite: false}),
			response: {data: serverSavedFilter({id: 8, title: 'Identity A update'})},
		},
		{
			name: 'favorite',
			mock: sdk.patchFiltersRead,
			run: () => patchSavedFilterFavorite(8, true),
			response: {data: serverSavedFilter({id: 8, is_favorite: true})},
		},
		{
			name: 'delete',
			mock: sdk.filtersDelete,
			run: () => deleteSavedFilter(8),
			response: {data: undefined},
		},
	]

	beforeEach(() => {
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
	})

	it.each(delayedMutationCases)('does not materialize an absent cache during $name', async ({mock, run, response}) => {
		mock.mockResolvedValue(response)
		await run()
		expect(queryClient.getQueryCache().getAll()).toEqual([])
	})

	it.each(['update', 'favorite', 'delete'])('rolls back an optimistic %s in the same session', async operation => {
		const selected = delayedMutationCases.find(test => test.name === operation)!
		const previous = serverSavedFilter({id: 8, title: 'Before', description: '<p>HTML</p>'})
		const pseudo = normalizeProject({id: -9, title: 'Before', description: '<p>HTML</p>'})
		const list = {...emptyProjectList, savedFilterProjects: [pseudo]}
		queryClient.setQueryData(listKey, list)
		queryClient.setQueryData(savedFilterKeys.detail(8), previous)
		queryClient.setQueryData(projectKeys.detail(-9), pseudo)
		const failure = new Error('Request failed')
		selected.mock.mockImplementation(async () => {
			const current = queryClient.getQueryData<ProjectListResult>(listKey)!
			if (operation === 'delete') {
				expect(current.savedFilterProjects).toEqual([])
			} else if (operation === 'favorite') {
				expect(current.savedFilterProjects[0].is_favorite).toBe(true)
			} else {
				expect(current.savedFilterProjects[0].title).toBe('Identity A update')
			}
			throw failure
		})

		await expect(selected.run()).rejects.toThrow(failure)

		expect(queryClient.getQueryData(listKey)).toEqual(list)
		expect(queryClient.getQueryData(savedFilterKeys.detail(8))).toEqual(previous)
		expect(queryClient.getQueryData(projectKeys.detail(-9))).toEqual(pseudo)
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	describe.each([
		['identity', () => { requestContext.identity = {id: 2, type: 1} }],
		['session', () => { requestContext.sessionEpoch++ }],
		['API URL', () => { requestContext.apiV2BaseUrl = 'https://identity-b.example/api/v2/' }],
	] as const)('after the %s changes', (_name, switchContext) => {
		it.each(delayedMutationCases)('discards a delayed $name completion', async ({mock, run, response}) => {
			queryClient.setQueryData(listKey, emptyProjectList)
			queryClient.setQueryData(savedFilterKeys.detail(8), serverSavedFilter({id: 8, title: 'Identity A'}))
			let resolveRequest: (value: unknown) => void = () => {}
			mock.mockReturnValue(new Promise(resolve => {
				resolveRequest = resolve
			}))

			const mutation = run()
			await vi.waitFor(() => expect(mock).toHaveBeenCalledOnce())
			switchContext()
			queryClient.clear()
			const identityBFilter = serverSavedFilter({id: 8, title: 'Identity B'})
			const nextList = {...emptyProjectList, savedFilterProjects: [normalizeProject({id: -9, title: 'New session filter'})]}
			queryClient.setQueryData(listKey, nextList)
			queryClient.setQueryData(savedFilterKeys.detail(8), identityBFilter)
			resolveRequest(response)

			await expect(mutation).rejects.toMatchObject({name: 'AbortError'})
			expect(queryClient.getQueryData(savedFilterKeys.detail(8))).toEqual(identityBFilter)
			expect(queryClient.getQueryData(listKey)).toEqual(nextList)
			expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(false)
		})
	})
})
