import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'

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
import {removeProjectFromHistory} from '@/modules/projectHistory'
import {i18n} from '@/i18n'
import type {EditableTaskCollection} from '@/types/EditableTaskCollection'

import {contextMutationOptions} from './contextMutation'
import {getProjectIdFromSavedFilterId, mapProjectNavigationItem, projectKeys, type ProjectListResult, type ProjectResponse} from './projects'

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

// Not a form field. PUT replaces every field, so update callers pass the cached value.
export type SavedFilterDraft = Required<Omit<SavedFilterWritable, 'filters' | 'is_favorite'>> & {
	filters: EditableTaskCollection
}

export type UpdateSavedFilterInput = SavedFilterDraft & Pick<SavedFilterResponse, 'id' | 'is_favorite'>

export const savedFilterKeys = {
	detail: (id: number) => ['saved-filters', 'detail', id] as const,
}

export function newSavedFilterDraft(): SavedFilterDraft {
	return {
		title: '',
		description: '',
		filters: {
			sort_by: ['done', 'id'],
			order_by: ['asc', 'desc'],
			filter: 'done = false',
			filter_include_nulls: true,
			s: '',
		},
	}
}

// Unset fields arrive as null; creation defaults here would be written back on save.
function normalizeSavedFilterCollection(filters?: TaskCollection): EditableTaskCollection {
	return {
		sort_by: filters?.sort_by ?? [],
		order_by: filters?.order_by ?? [],
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

export function savedFilterQuery(id: number) {
	return queryOptions({
		queryKey: savedFilterKeys.detail(id),
		queryFn: async () => {
			const {data} = await filtersRead({path: {filter: id}})
			return normalizeSavedFilter(data)
		},
		enabled: id > 0,
	})
}

type UpdateNotify = (input: UpdateSavedFilterInput) => boolean
type DeleteNotify = (id: number) => boolean

function updateNavigation(
	client: QueryClient,
	id: number,
	fields: Partial<Pick<ProjectResponse, 'title' | 'is_favorite'>>,
) {
	const projectId = getProjectIdFromSavedFilterId(id)
	client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
		current ? mapProjectNavigationItem(current, projectId, project => ({...project, ...fields})) : current,
	)
	client.setQueryData<ProjectResponse>(projectKeys.detail(projectId), current =>
		current ? {...current, ...fields} : current,
	)
}

function invalidateSavedFilter(client: QueryClient, id?: number) {
	return Promise.all([
		client.invalidateQueries({queryKey: projectKeys.list()}),
		...(id === undefined ? [] : [
			client.invalidateQueries({queryKey: savedFilterKeys.detail(id)}),
			client.invalidateQueries({queryKey: projectKeys.detail(getProjectIdFromSavedFilterId(id))}),
		]),
	])
}

function savedFilterQueryKeys(id: number) {
	return [savedFilterKeys.detail(id), projectKeys.list(), projectKeys.detail(getProjectIdFromSavedFilterId(id))]
}

export function createSavedFilterMutationOptions(shouldNotify: () => boolean = () => true) {
	return contextMutationOptions({
		mutationFn: async (filter: SavedFilterWritable) => {
			const {data} = await filtersCreate({body: filter})
			return normalizeSavedFilter(data)
		},
		onSettled: (_filter, client) => invalidateSavedFilter(client),
		toastError: () => shouldNotify(),
	})
}

export function updateSavedFilterMutationOptions(shouldNotify: UpdateNotify = () => true) {
	return contextMutationOptions({
		mutationFn: async ({id, ...filter}: UpdateSavedFilterInput) => {
			const {data} = await filtersUpdate({path: {filter: id}, body: filter})
			return normalizeSavedFilter(data)
		},
		optimistic: {
			queryKeys: ({id}) => savedFilterQueryKeys(id),
			update: ({id, ...filter}, client) => {
				client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(id), current =>
					current ? {...current, ...filter} : current,
				)
				updateNavigation(client, id, {title: filter.title, is_favorite: filter.is_favorite})
			},
		},
		onSuccess: (updated, input, client) => {
			client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(input.id), current =>
				current ? updated : current,
			)
			updateNavigation(client, input.id, {title: updated.title, is_favorite: updated.is_favorite})
		},
		onSettled: ({id}, client) => invalidateSavedFilter(client, id),
		successMessage: (_updated, input) => shouldNotify(input) ? i18n.global.t('filters.edit.success') : undefined,
		toastError: shouldNotify,
	})
}

export function patchSavedFilterFavoriteMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, isFavorite}: {id: number; isFavorite: boolean}) => {
			const {data} = await patchFiltersRead({
				path: {filter: id},
				body: [{op: 'replace', path: '/is_favorite', value: isFavorite}],
			})
			return normalizeSavedFilter(data)
		},
		optimistic: {
			queryKeys: ({id}) => savedFilterQueryKeys(id),
			update: ({id, isFavorite}, client) => {
				client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(id), current =>
					current ? {...current, is_favorite: isFavorite} : current,
				)
				updateNavigation(client, id, {is_favorite: isFavorite})
			},
		},
		onSuccess: (updated, {id}, client) => {
			client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(id), current =>
				current ? {...current, is_favorite: updated.is_favorite} : current,
			)
			updateNavigation(client, id, {is_favorite: updated.is_favorite})
		},
		onSettled: ({id}, client) => invalidateSavedFilter(client, id),
	})
}

export function deleteSavedFilterMutationOptions(shouldNotify: DeleteNotify = () => true) {
	return contextMutationOptions({
		mutationFn: async (id: number) => {
			await filtersDelete({path: {filter: id}})
		},
		optimistic: {
			queryKeys: savedFilterQueryKeys,
			update: (id, client) => {
				const projectId = getProjectIdFromSavedFilterId(id)
				client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
					current ? {...current, savedFilterProjects: current.savedFilterProjects.filter(project => project.id !== projectId)} : current,
				)
			},
		},
		onSuccess: (_data, id, client) => {
			const projectId = getProjectIdFromSavedFilterId(id)
			client.removeQueries({queryKey: savedFilterKeys.detail(id)})
			client.removeQueries({queryKey: projectKeys.detail(projectId)})
			removeProjectFromHistory({id: projectId})
		},
		onSettled: (id, client) => invalidateSavedFilter(client, id),
		successMessage: (_data, id) => shouldNotify(id) ? i18n.global.t('filters.delete.success') : undefined,
		toastError: shouldNotify,
	})
}

export function useCreateSavedFilterMutation(shouldNotify?: () => boolean) {
	return useMutation(createSavedFilterMutationOptions(shouldNotify))
}

export function useUpdateSavedFilterMutation(shouldNotify?: UpdateNotify) {
	return useMutation(updateSavedFilterMutationOptions(shouldNotify))
}

export function usePatchSavedFilterFavoriteMutation() {
	return useMutation(patchSavedFilterFavoriteMutationOptions())
}

export function useDeleteSavedFilterMutation(shouldNotify?: DeleteNotify) {
	return useMutation(deleteSavedFilterMutationOptions(shouldNotify))
}
