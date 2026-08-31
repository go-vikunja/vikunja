import {queryOptions, useMutation} from '@tanstack/vue-query'

import {
	filtersCreate,
	filtersDelete,
	filtersRead,
	filtersUpdate,
	patchFiltersRead,
} from '@/client/generated'
import type {
	SavedFilter,
	SavedFilterReadBody,
	SavedFilterWritable,
	TaskCollection,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import type {TaskFilterParams} from '@/services/taskCollection'

import {projectKeys} from './projects'
export {getSavedFilterIdFromProjectId, isSavedFilterProject} from './projects'

export type SavedFilterCollection = Required<Omit<TaskCollection, 'order_by' | 'sort_by'>> &
	Pick<TaskFilterParams, 'order_by' | 'sort_by'>

export type SavedFilterResponse = Omit<SavedFilterReadBody,
	'id' |
	'title' |
	'description' |
	'filters' |
	'is_favorite'
> & {
	id: number
	title: string
	description: string
	filters: SavedFilterCollection
	is_favorite: boolean
}

export type SavedFilterDraft = Required<Omit<SavedFilterWritable, 'filters'>> & {
	filters: SavedFilterCollection
}

export type UpdateSavedFilterInput = SavedFilterDraft & {id: number}

export const savedFilterKeys = {
	all: ['saved-filters'] as const,
	details: () => ['saved-filters', 'detail'] as const,
	detailRoot: (id: number) => ['saved-filters', 'detail', id] as const,
	detail: (id: number, format: 'html' | 'markdown' = 'html') => [
		'saved-filters',
		'detail',
		id,
		format,
	] as const,
}

function normalizeSavedFilterCollection(filters?: TaskCollection): SavedFilterCollection {
	return {
		sort_by: (filters?.sort_by ?? ['done', 'id']) as TaskFilterParams['sort_by'],
		order_by: (filters?.order_by ?? ['asc', 'desc']) as TaskFilterParams['order_by'],
		filter: filters?.filter ?? 'done = false',
		filter_include_nulls: filters?.filter_include_nulls ?? true,
		s: filters?.s ?? '',
	}
}

export function createSavedFilterDraft(filter: Partial<SavedFilterWritable> = {}): SavedFilterDraft {
	return {
		title: filter.title ?? '',
		description: filter.description ?? '',
		filters: normalizeSavedFilterCollection(filter.filters),
		is_favorite: filter.is_favorite ?? false,
	}
}

export function normalizeSavedFilter(filter: SavedFilter | SavedFilterReadBody): SavedFilterResponse {
	if (typeof filter.id !== 'number') {
		throw new Error('Saved filter response is missing an id')
	}

	return {
		...filter,
		id: filter.id,
		title: filter.title ?? '',
		description: filter.description ?? '',
		filters: normalizeSavedFilterCollection(filter.filters),
		is_favorite: filter.is_favorite ?? false,
	}
}

export function savedFilterQuery(id: number, format: 'html' | 'markdown' = 'html') {
	return queryOptions({
		queryKey: savedFilterKeys.detail(id, format),
		queryFn: async () => {
			const {data} = await filtersRead({path: {filter: id}, query: {format}})
			return normalizeSavedFilter(data)
		},
		enabled: id > 0,
	})
}

export function getProjectIdFromSavedFilterId(savedFilterId: number): number {
	return savedFilterId > 0 ? savedFilterId * -1 - 1 : 0
}

async function invalidateProjectNavigation(): Promise<void> {
	await queryClient.invalidateQueries({queryKey: projectKeys.lists()})
}

export async function createSavedFilter(
	filter: SavedFilterWritable,
	format: 'html' | 'markdown' = 'html',
): Promise<SavedFilterResponse> {
	const {data} = await filtersCreate({body: filter, query: {format}})
	const created = normalizeSavedFilter(data)
	queryClient.setQueryData(savedFilterKeys.detail(created.id, format), created)
	await invalidateProjectNavigation()
	return created
}

export async function updateSavedFilter(
	{id, ...filter}: UpdateSavedFilterInput,
	format: 'html' | 'markdown' = 'html',
): Promise<SavedFilterResponse> {
	await queryClient.cancelQueries({queryKey: savedFilterKeys.detailRoot(id)})
	const {data} = await filtersUpdate({
		path: {filter: id},
		body: filter,
		query: {format},
	})
	const updated = normalizeSavedFilter(data)
	queryClient.removeQueries({
		queryKey: savedFilterKeys.detailRoot(id),
		predicate: query => query.queryKey[3] !== format,
	})
	queryClient.setQueryData(savedFilterKeys.detail(id, format), updated)
	await invalidateProjectNavigation()
	return updated
}

export async function patchSavedFilterFavorite(
	id: number,
	isFavorite: boolean,
): Promise<SavedFilterResponse> {
	await queryClient.cancelQueries({queryKey: savedFilterKeys.detailRoot(id)})
	const {data} = await patchFiltersRead({
		path: {filter: id},
		body: [{op: 'replace', path: '/is_favorite', value: isFavorite}],
	})
	const updated = normalizeSavedFilter(data)
	queryClient.setQueriesData<SavedFilterResponse>(
		{queryKey: savedFilterKeys.detailRoot(id)},
		current => current
			? {...current, is_favorite: updated.is_favorite}
			: current,
	)
	await invalidateProjectNavigation()
	return updated
}

export async function deleteSavedFilter(id: number): Promise<void> {
	await queryClient.cancelQueries({queryKey: savedFilterKeys.detailRoot(id)})
	await filtersDelete({path: {filter: id}})
	queryClient.removeQueries({queryKey: savedFilterKeys.detailRoot(id)})
	await invalidateProjectNavigation()
}

export function useCreateSavedFilterMutation() {
	return useMutation({mutationFn: (filter: SavedFilterWritable) => createSavedFilter(filter)})
}

export function useUpdateSavedFilterMutation() {
	return useMutation({mutationFn: (filter: UpdateSavedFilterInput) => updateSavedFilter(filter)})
}

export function useDeleteSavedFilterMutation() {
	return useMutation({mutationFn: deleteSavedFilter})
}
