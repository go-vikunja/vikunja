import {queryOptions} from '@tanstack/vue-query'

import {
	filtersCreate,
	filtersDelete,
	filtersRead,
	filtersUpdate,
	patchFiltersRead,
} from '@/client/generated'
import type {
	SavedFilterReadBody,
	SavedFilterWritable,
	TaskCollection,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {assertClientRequestContext, captureClientRequestContext} from '@/client/requestContext'
import type {ClientRequestContext} from '@/client/requestContext'
import {removeProjectFromHistory} from '@/modules/projectHistory'
import type {EditableTaskCollection, TaskFilterParams} from '@/types/TaskFilterParams'

import {getProjectIdFromSavedFilterId, projectKeys} from './projects'

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
	filters: EditableTaskCollection
	is_favorite: boolean
}

export type SavedFilterDraft = Required<Omit<SavedFilterWritable, 'filters'>> & {
	filters: EditableTaskCollection
}

export type UpdateSavedFilterInput = SavedFilterDraft & {id: number}

export const savedFilterKeys = {
	detailRoot: (id: number) => ['saved-filters', 'detail', id] as const,
	detail: (id: number, format: 'html' | 'markdown' = 'html') => [
		...savedFilterKeys.detailRoot(id),
		format,
	] as const,
}

export function createSavedFilterDraft(filter: Partial<SavedFilterWritable> = {}): SavedFilterDraft {
	return {
		title: filter.title ?? '',
		description: filter.description ?? '',
		filters: {
			sort_by: (filter.filters?.sort_by ?? ['done', 'id']) as TaskFilterParams['sort_by'],
			order_by: (filter.filters?.order_by ?? ['asc', 'desc']) as TaskFilterParams['order_by'],
			filter: filter.filters?.filter ?? 'done = false',
			filter_include_nulls: filter.filters?.filter_include_nulls ?? true,
			s: filter.filters?.s ?? '',
		},
		is_favorite: filter.is_favorite ?? false,
	}
}

// Unset fields arrive as null; creation defaults here would be written back on save.
function normalizeSavedFilterCollection(filters?: TaskCollection): EditableTaskCollection {
	return {
		sort_by: (filters?.sort_by ?? []) as TaskFilterParams['sort_by'],
		order_by: (filters?.order_by ?? []) as TaskFilterParams['order_by'],
		filter: filters?.filter ?? '',
		filter_include_nulls: filters?.filter_include_nulls ?? false,
		s: filters?.s ?? '',
	}
}

function normalizeSavedFilter(filter: SavedFilterReadBody): SavedFilterResponse {
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

async function invalidateProjectNavigation(context: ClientRequestContext): Promise<void> {
	await queryClient.invalidateQueries({queryKey: projectKeys.lists()})
	assertClientRequestContext(context)
}

export async function createSavedFilter(filter: SavedFilterWritable): Promise<SavedFilterResponse> {
	const context = captureClientRequestContext()
	const {data} = await filtersCreate({body: filter})
	assertClientRequestContext(context)
	const created = normalizeSavedFilter(data)
	queryClient.setQueryData(savedFilterKeys.detail(created.id), created)
	await invalidateProjectNavigation(context)
	return created
}

export async function updateSavedFilter({id, ...filter}: UpdateSavedFilterInput): Promise<SavedFilterResponse> {
	const context = captureClientRequestContext()
	await queryClient.cancelQueries({queryKey: savedFilterKeys.detailRoot(id)})
	assertClientRequestContext(context)
	const {data} = await filtersUpdate({path: {filter: id}, body: filter})
	assertClientRequestContext(context)
	const updated = normalizeSavedFilter(data)
	queryClient.setQueryData(savedFilterKeys.detail(id), updated)
	await queryClient.invalidateQueries({queryKey: savedFilterKeys.detailRoot(id)})
	await invalidateProjectNavigation(context)
	return updated
}

export async function patchSavedFilterFavorite(
	id: number,
	isFavorite: boolean,
): Promise<SavedFilterResponse> {
	const context = captureClientRequestContext()
	await queryClient.cancelQueries({queryKey: savedFilterKeys.detailRoot(id)})
	assertClientRequestContext(context)
	const {data} = await patchFiltersRead({
		path: {filter: id},
		body: [{op: 'replace', path: '/is_favorite', value: isFavorite}],
	})
	assertClientRequestContext(context)
	const updated = normalizeSavedFilter(data)
	queryClient.setQueriesData<SavedFilterResponse>(
		{queryKey: savedFilterKeys.detailRoot(id)},
		current => current
			? {...current, is_favorite: updated.is_favorite}
			: current,
	)
	await invalidateProjectNavigation(context)
	return updated
}

export async function deleteSavedFilter(id: number): Promise<void> {
	const context = captureClientRequestContext()
	await queryClient.cancelQueries({queryKey: savedFilterKeys.detailRoot(id)})
	assertClientRequestContext(context)
	await filtersDelete({path: {filter: id}})
	assertClientRequestContext(context)
	const projectId = getProjectIdFromSavedFilterId(id)
	queryClient.removeQueries({queryKey: projectKeys.detailRoot(projectId)})
	removeProjectFromHistory({id: projectId})
	queryClient.removeQueries({queryKey: savedFilterKeys.detailRoot(id)})
	await invalidateProjectNavigation(context)
}
