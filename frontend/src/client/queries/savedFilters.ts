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
import type {TaskFilterParams} from '@/types/TaskFilterParams'

import {getProjectIdFromSavedFilterId, mapProjectNavigationItem, projectKeys, type ProjectListResult, type ProjectResponse} from './projects'

// The literal unions keep the shared Filters.vue v-model typed.
type SavedFilterFilters = EditableTaskCollection & Pick<TaskFilterParams, 'sort_by' | 'order_by'>

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
	filters: SavedFilterFilters
	is_favorite: boolean
}

export type SavedFilterDraft = Required<Omit<SavedFilterWritable, 'filters'>> & {
	filters: SavedFilterFilters
}

export type UpdateSavedFilterInput = SavedFilterDraft & {id: number}

export const savedFilterKeys = {
	detailRoot: (id: number) => ['saved-filters', 'detail', id] as const,
	detail: (id: number, format: 'html' | 'markdown' = 'html') => [
		...savedFilterKeys.detailRoot(id),
		format,
	] as const,
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
		is_favorite: false,
	}
}

// Unset fields arrive as null; creation defaults here would be written back on save.
function normalizeSavedFilterCollection(filters?: TaskCollection): SavedFilterFilters {
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

async function snapshotSavedFilter(client: QueryClient, id: number) {
	const request = captureClientRequestContext()
	const projectId = getProjectIdFromSavedFilterId(id)
	await Promise.all([
		client.cancelQueries({queryKey: savedFilterKeys.detailRoot(id)}),
		client.cancelQueries({queryKey: projectKeys.lists()}),
		client.cancelQueries({queryKey: projectKeys.detailRoot(projectId)}),
	])
	assertClientRequestContext(request)
	return {
		request,
		previous: [
			...client.getQueriesData<SavedFilterResponse>({queryKey: savedFilterKeys.detailRoot(id)}),
			...client.getQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}),
			...client.getQueriesData<ProjectResponse>({queryKey: projectKeys.detailRoot(projectId)}),
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
	client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
		current ? mapProjectNavigationItem(current, projectId, project => ({...project, ...fields})) : current,
	)
	client.setQueriesData<ProjectResponse>({queryKey: projectKeys.detailRoot(projectId)}, current =>
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
		client.invalidateQueries({queryKey: projectKeys.lists()}),
		...(id === undefined ? [] : [
			client.invalidateQueries({queryKey: savedFilterKeys.detailRoot(id)}),
			client.invalidateQueries({queryKey: projectKeys.detailRoot(getProjectIdFromSavedFilterId(id))}),
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
			const {description: _description, ...fields} = filter
			client.setQueriesData<SavedFilterResponse>({queryKey: savedFilterKeys.detailRoot(id)}, current =>
				current ? {...current, ...fields} : current,
			)
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
			client.setQueriesData<SavedFilterResponse>({queryKey: savedFilterKeys.detailRoot(id)}, current =>
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
			client.setQueriesData<SavedFilterResponse>({queryKey: savedFilterKeys.detailRoot(id)}, current =>
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
			client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
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
			client.removeQueries({queryKey: savedFilterKeys.detailRoot(id)})
			client.removeQueries({queryKey: projectKeys.detailRoot(projectId)})
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
