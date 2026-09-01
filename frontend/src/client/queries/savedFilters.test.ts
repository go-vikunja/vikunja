import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {ProjectListResult} from './projects'
import {projectKeys} from './projects'
import type {SavedFilterResponse} from './savedFilters'
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
	createSavedFilter,
	createSavedFilterDraft,
	deleteSavedFilter,
	patchSavedFilterFavorite,
	savedFilterKeys,
	savedFilterQuery,
	updateSavedFilter,
} from './savedFilters'

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

	it('reads a saved filter without inventing creation defaults', async () => {
		sdk.filtersRead.mockResolvedValue({
			data: {id: 42, title: 'Upcoming', filters: {sort_by: null, order_by: null}},
		})

		const result = await queryClient.fetchQuery(savedFilterQuery(42, 'markdown'))

		expect(sdk.filtersRead).toHaveBeenCalledWith({
			path: {filter: 42},
			query: {format: 'markdown'},
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

	it('creates a stable draft with overrides', () => {
		expect(createSavedFilterDraft({title: 'Today', is_favorite: true})).toEqual({
			title: 'Today',
			description: '',
			filters: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter: 'done = false',
				filter_include_nulls: true,
				s: '',
			},
			is_favorite: true,
		})
	})

	it('creates a saved filter, caches its detail, and invalidates project navigation', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		const created = serverSavedFilter({id: 8, title: 'Created'})
		sdk.filtersCreate.mockResolvedValue({data: created})

		await expect(createSavedFilter({title: 'Created'})).resolves.toEqual(created)

		expect(sdk.filtersCreate).toHaveBeenCalledWith({body: {title: 'Created'}})
		expect(queryClient.getQueryData(savedFilterKeys.detail(8))).toEqual(created)
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('updates the detail cache, invalidates other formats and project navigation', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		queryClient.setQueryData(savedFilterKeys.detail(8), serverSavedFilter({
			id: 8,
			title: 'Before',
			description: '<p>HTML</p>',
		}))
		queryClient.setQueryData(savedFilterKeys.detail(8, 'markdown'), serverSavedFilter({
			id: 8,
			title: 'Before',
			description: 'Markdown',
		}))
		sdk.filtersUpdate.mockResolvedValue({data: serverSavedFilter({id: 8, title: 'After'})})

		const writable = createSavedFilterDraft({title: 'After'})
		await expect(updateSavedFilter({id: 8, ...writable})).resolves.toMatchObject({
			id: 8,
			title: 'After',
		})

		expect(sdk.filtersUpdate).toHaveBeenCalledWith({
			path: {filter: 8},
			body: writable,
		})
		expect(queryClient.getQueryData<SavedFilterResponse>(savedFilterKeys.detail(8))?.title).toBe('After')
		expect(queryClient.getQueryState(savedFilterKeys.detail(8, 'markdown'))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('patches only the favorite field in every cached format', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		const html = serverSavedFilter({id: 8, description: '<p>HTML</p>'})
		const markdown = serverSavedFilter({id: 8, description: 'Markdown'})
		queryClient.setQueryData(savedFilterKeys.detail(8), html)
		queryClient.setQueryData(savedFilterKeys.detail(8, 'markdown'), markdown)
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
		expect(queryClient.getQueryData<SavedFilterResponse>(savedFilterKeys.detail(8, 'markdown'))).toMatchObject({
			description: 'Markdown',
			is_favorite: true,
		})
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('deletes every cached detail format, the pseudo-project and invalidates project navigation', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		queryClient.setQueryData(savedFilterKeys.detail(8), serverSavedFilter({id: 8}))
		queryClient.setQueryData(savedFilterKeys.detail(8, 'markdown'), serverSavedFilter({id: 8}))
		queryClient.setQueryData(projectKeys.detail(-9), {id: -9, title: 'Filter'})
		sdk.filtersDelete.mockResolvedValue({data: undefined})

		await deleteSavedFilter(8)

		expect(sdk.filtersDelete).toHaveBeenCalledWith({path: {filter: 8}})
		expect(queryClient.getQueriesData({queryKey: savedFilterKeys.detailRoot(8)})).toEqual([])
		expect(queryClient.getQueriesData({queryKey: projectKeys.detailRoot(-9)})).toEqual([])
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
			run: () => updateSavedFilter({id: 8, ...createSavedFilterDraft({title: 'Identity A update'})}),
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
			queryClient.setQueryData(listKey, emptyProjectList)
			queryClient.setQueryData(savedFilterKeys.detail(8), serverSavedFilter({id: 8, title: 'Identity A'}))
			let resolveRequest: (value: unknown) => void = () => {}
			mock.mockReturnValue(new Promise(resolve => {
				resolveRequest = resolve
			}))

			const mutation = run()
			await vi.waitFor(() => expect(mock).toHaveBeenCalledOnce())
			change()
			queryClient.clear()
			const identityBFilter = serverSavedFilter({id: 8, title: 'Identity B'})
			queryClient.setQueryData(listKey, emptyProjectList)
			queryClient.setQueryData(savedFilterKeys.detail(8), identityBFilter)
			resolveRequest(response)

			await expect(mutation).rejects.toMatchObject({name: 'AbortError'})
			expect(queryClient.getQueryData(savedFilterKeys.detail(8))).toEqual(identityBFilter)
			expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(false)
		})
	})
})
