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

vi.mock('@/client/generated', () => sdk)

import {
	createSavedFilter,
	createSavedFilterDraft,
	deleteSavedFilter,
	getProjectIdFromSavedFilterId,
	getSavedFilterIdFromProjectId,
	isSavedFilterProject,
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

describe('saved filter queries', () => {
	beforeEach(() => {
		queryClient.clear()
		Object.values(sdk).forEach(mock => mock.mockReset())
	})

	it('keys details by id and rich-text format', () => {
		expect(savedFilterKeys.all).toEqual(['saved-filters'])
		expect(savedFilterKeys.details()).toEqual(['saved-filters', 'detail'])
		expect(savedFilterKeys.detailRoot(42)).toEqual(['saved-filters', 'detail', 42])
		expect(savedFilterKeys.detail(42)).toEqual(['saved-filters', 'detail', 42, 'html'])
		expect(savedFilterKeys.detail(42, 'markdown')).toEqual(['saved-filters', 'detail', 42, 'markdown'])
	})

	it('reads and normalizes a saved filter', async () => {
		sdk.filtersRead.mockResolvedValue({
			data: {id: 42, title: 'Upcoming'},
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
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter: 'done = false',
				filter_include_nulls: true,
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

		await expect(createSavedFilter({title: 'Created'}, 'markdown')).resolves.toEqual(created)

		expect(sdk.filtersCreate).toHaveBeenCalledWith({
			body: {title: 'Created'},
			query: {format: 'markdown'},
		})
		expect(queryClient.getQueryData(savedFilterKeys.detail(8, 'markdown'))).toEqual(created)
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('updates only the selected detail format and invalidates project navigation', async () => {
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
		await expect(updateSavedFilter({id: 8, ...writable}, 'markdown')).resolves.toMatchObject({
			id: 8,
			title: 'After',
		})

		expect(sdk.filtersUpdate).toHaveBeenCalledWith({
			path: {filter: 8},
			body: writable,
			query: {format: 'markdown'},
		})
		expect(queryClient.getQueryData(savedFilterKeys.detail(8))).toBeUndefined()
		expect(queryClient.getQueryData<SavedFilterResponse>(savedFilterKeys.detail(8, 'markdown'))?.title).toBe('After')
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
	})

	it('patches only the favorite field in every cached format', async () => {
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
	})

	it('deletes every cached detail format and invalidates project navigation', async () => {
		const listKey = projectKeys.list()
		queryClient.setQueryData(listKey, emptyProjectList)
		queryClient.setQueryData(savedFilterKeys.detail(8), serverSavedFilter({id: 8}))
		queryClient.setQueryData(savedFilterKeys.detail(8, 'markdown'), serverSavedFilter({id: 8}))
		sdk.filtersDelete.mockResolvedValue({data: undefined})

		await deleteSavedFilter(8)

		expect(sdk.filtersDelete).toHaveBeenCalledWith({path: {filter: 8}})
		expect(queryClient.getQueriesData({queryKey: savedFilterKeys.detailRoot(8)})).toEqual([])
		expect(queryClient.getQueryState(listKey)?.isInvalidated).toBe(true)
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
