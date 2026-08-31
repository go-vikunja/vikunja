import {mutationOptions, queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'

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
import {assertClientRequestContext, captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import type {ClientRequestContext} from '@/client/requestContext'
import {removeProjectFromHistory} from '@/modules/projectHistory'
import {i18n} from '@/i18n'
import {error, success} from '@/message'
import type {EditableTaskCollection} from '@/types/EditableTaskCollection'

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

async function snapshotSavedFilter(client: QueryClient, id: number) {
	const request = captureClientRequestContext()
	const projectId = getProjectIdFromSavedFilterId(id)
	await Promise.all([
		client.cancelQueries({queryKey: savedFilterKeys.detail(id)}),
		client.cancelQueries({queryKey: projectKeys.list()}),
		client.cancelQueries({queryKey: projectKeys.detail(projectId)}),
	])
	assertClientRequestContext(request)
	return {
		request,
		previous: [
			...client.getQueriesData<SavedFilterResponse>({queryKey: savedFilterKeys.detail(id)}),
			...client.getQueriesData<ProjectListResult>({queryKey: projectKeys.list()}),
			...client.getQueriesData<ProjectResponse>({queryKey: projectKeys.detail(projectId)}),
		],
	}
}

type SavedFilterSnapshot = Awaited<ReturnType<typeof snapshotSavedFilter>>
type UpdateNotify = (input: UpdateSavedFilterInput) => boolean
type DeleteNotify = (id: number) => boolean

function restoreSavedFilter(client: QueryClient, snapshot: SavedFilterSnapshot | undefined) {
	if (!snapshot || !isClientRequestContextCurrent(snapshot.request)) {
		return
	}
	for (const [key, previous] of snapshot.previous) {
		if (previous) {
			client.setQueryData(key, previous)
		}
	}
}

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

async function settleSavedFilter(
	client: QueryClient,
	context: {request: ClientRequestContext} | undefined,
	id?: number,
) {
	if (!context || !isClientRequestContextCurrent(context.request)) {
		return
	}
	await Promise.all([
		client.invalidateQueries({queryKey: projectKeys.list()}),
		...(id === undefined ? [] : [
			client.invalidateQueries({queryKey: savedFilterKeys.detail(id)}),
			client.invalidateQueries({queryKey: projectKeys.detail(getProjectIdFromSavedFilterId(id))}),
		]),
	])
	assertClientRequestContext(context.request)
}

export function createSavedFilterMutationOptions(shouldNotify: () => boolean = () => true) {
	return mutationOptions({
		onMutate: () => ({request: captureClientRequestContext()}),
		mutationFn: async (filter: SavedFilterWritable) => {
			const request = captureClientRequestContext()
			const {data} = await filtersCreate({body: filter})
			assertClientRequestContext(request)
			return normalizeSavedFilter(data)
		},
		onError: (cause, _filter, context) => {
			if (context && isClientRequestContextCurrent(context.request) && shouldNotify()) {
				error(cause)
			}
		},
		onSettled: (_data, _error, _filter, context, {client}) => settleSavedFilter(client, context),
	})
}

export function updateSavedFilterMutationOptions(shouldNotify: UpdateNotify = () => true) {
	return mutationOptions({
		mutationFn: async ({id, ...filter}: UpdateSavedFilterInput) => {
			const request = captureClientRequestContext()
			const {data} = await filtersUpdate({path: {filter: id}, body: filter})
			assertClientRequestContext(request)
			return normalizeSavedFilter(data)
		},
		onMutate: async ({id, ...filter}, {client}) => {
			const snapshot = await snapshotSavedFilter(client, id)
			client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(id), current =>
				current ? {...current, ...filter} : current,
			)
			updateNavigation(client, id, {title: filter.title, is_favorite: filter.is_favorite})
			return snapshot
		},
		onError: (cause, input, context, {client}) => {
			restoreSavedFilter(client, context)
			if (context && isClientRequestContextCurrent(context.request) && shouldNotify(input)) {
				error(cause)
			}
		},
		onSuccess: (updated, input, context, {client}) => {
			assertClientRequestContext(context.request)
			client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(input.id), current =>
				current ? updated : current,
			)
			updateNavigation(client, input.id, {title: updated.title, is_favorite: updated.is_favorite})
			if (shouldNotify(input)) {
				success({message: i18n.global.t('filters.edit.success')})
			}
		},
		onSettled: (_data, _error, {id}, context, {client}) => settleSavedFilter(client, context, id),
	})
}

export function patchSavedFilterFavoriteMutationOptions() {
	return mutationOptions({
		mutationFn: async ({id, isFavorite}: {id: number; isFavorite: boolean}) => {
			const request = captureClientRequestContext()
			const {data} = await patchFiltersRead({
				path: {filter: id},
				body: [{op: 'replace', path: '/is_favorite', value: isFavorite}],
			})
			assertClientRequestContext(request)
			return normalizeSavedFilter(data)
		},
		onMutate: async ({id, isFavorite}, {client}) => {
			const snapshot = await snapshotSavedFilter(client, id)
			client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(id), current =>
				current ? {...current, is_favorite: isFavorite} : current,
			)
			updateNavigation(client, id, {is_favorite: isFavorite})
			return snapshot
		},
		onError: (cause, _input, context, {client}) => {
			restoreSavedFilter(client, context)
			if (context && isClientRequestContextCurrent(context.request)) {
				error(cause)
			}
		},
		onSuccess: (updated, {id}, context, {client}) => {
			assertClientRequestContext(context.request)
			client.setQueryData<SavedFilterResponse>(savedFilterKeys.detail(id), current =>
				current ? {...current, is_favorite: updated.is_favorite} : current,
			)
			updateNavigation(client, id, {is_favorite: updated.is_favorite})
		},
		onSettled: (_data, _error, {id}, context, {client}) => settleSavedFilter(client, context, id),
	})
}

export function deleteSavedFilterMutationOptions(shouldNotify: DeleteNotify = () => true) {
	return mutationOptions({
		mutationFn: async (id: number) => {
			const request = captureClientRequestContext()
			await filtersDelete({path: {filter: id}})
			assertClientRequestContext(request)
		},
		onMutate: async (id, {client}) => {
			const snapshot = await snapshotSavedFilter(client, id)
			const projectId = getProjectIdFromSavedFilterId(id)
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? {...current, savedFilterProjects: current.savedFilterProjects.filter(project => project.id !== projectId)} : current,
			)
			return snapshot
		},
		onError: (cause, id, context, {client}) => {
			restoreSavedFilter(client, context)
			if (context && isClientRequestContextCurrent(context.request) && shouldNotify(id)) {
				error(cause)
			}
		},
		onSuccess: (_data, id, context, {client}) => {
			assertClientRequestContext(context.request)
			const projectId = getProjectIdFromSavedFilterId(id)
			client.removeQueries({queryKey: savedFilterKeys.detail(id)})
			client.removeQueries({queryKey: projectKeys.detail(projectId)})
			removeProjectFromHistory({id: projectId})
			if (shouldNotify(id)) {
				success({message: i18n.global.t('filters.delete.success')})
			}
		},
		onSettled: (_data, _error, id, context, {client}) => settleSavedFilter(client, context, id),
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
