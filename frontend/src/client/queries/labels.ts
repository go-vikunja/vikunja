import {mutationOptions, queryOptions, useMutation} from '@tanstack/vue-query'
import type {QueryClient} from '@tanstack/vue-query'

import {
	labelsCreate,
	labelsDelete,
	labelsList,
	labelsUpdate,
} from '@/client/generated'
import type {Label, LabelWritable} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {fetchAllPages} from './fetchAllPages'
import {API_MAX_PER_PAGE} from './pagination'
import {colorFromHex} from '@/helpers/color/colorFromHex'
import {i18n} from '@/i18n'
import {success} from '@/message'

export const labelKeys = {
	all: ['labels'] as const,
}

export type LabelDraft = Required<Pick<LabelWritable, 'title' | 'description' | 'hex_color'>>
export type CreateLabelInput = Required<Pick<LabelWritable, 'title'>> &
	Pick<LabelWritable, 'description' | 'hex_color'>
export type UpdateLabelInput = Required<Pick<Label, 'id' | 'title'>> &
	Pick<Label, 'description' | 'hex_color'>

export function createLabelDraft(label: Partial<LabelWritable> = {}): LabelDraft {
	return {
		title: '',
		description: '',
		hex_color: '',
		...label,
	}
}

function fetchAllLabels(): Promise<Label[]> {
	return fetchAllPages(async page => (await labelsList({query: {page, per_page: API_MAX_PER_PAGE}})).data)
}

export function sortLabelsAlphabetically(labels: Label[], locale = i18n.global.locale.value): Label[] {
	return [...labels].sort((a, b) => (a.title ?? '').localeCompare(
		b.title ?? '',
		locale,
		{ignorePunctuation: true},
	))
}

export function labelsQuery() {
	return queryOptions({
		queryKey: labelKeys.all,
		queryFn: fetchAllLabels,
		select: sortLabelsAlphabetically,
		staleTime: 5 * 60 * 1000,
	})
}

export function ensureLabels(): Promise<Label[]> {
	return queryClient.ensureQueryData(labelsQuery())
}

export function refreshLabels(): Promise<Label[]> {
	return queryClient.fetchQuery({...labelsQuery(), staleTime: 0})
}

export function getLabelById(labels: Label[], id: number): Label | undefined {
	return labels.find(label => label.id === id)
}

export function getLabelsByIds(labels: Label[], ids: number[]): Label[] {
	return ids.map(id => getLabelById(labels, id)).filter((label): label is Label => Boolean(label))
}

export function getLabelByExactTitle(labels: Label[], title: string): Label | undefined {
	return labels.find(label => (label.title ?? '').toLowerCase() === title.toLowerCase())
}

export function getLabelsByExactTitles(labels: Label[], titles: string[]): Label[] {
	return labels.filter(label => titles.some(title => title.toLowerCase() === (label.title ?? '').toLowerCase()))
}

export function filterLabelsByQuery(labels: Label[], labelsToHide: Label[], query: string): Label[] {
	if (query === '') {
		return []
	}

	const hiddenIds = new Set(labelsToHide.map(label => label.id))
	const normalizedQuery = query.toLowerCase()
	return labels
		.filter(label => !hiddenIds.has(label.id))
		.filter(label => (label.title ?? '').toLowerCase().includes(normalizedQuery) ||
			(label.description ?? '').toLowerCase().includes(normalizedQuery))
}

function labelBody(label: CreateLabelInput): LabelWritable {
	return {
		title: label.title,
		description: label.description,
		hex_color: colorFromHex(label.hex_color ?? ''),
	}
}

export function createLabelMutationOptions() {
	return mutationOptions({
		mutationFn: async (label: CreateLabelInput) => {
			const {data} = await labelsCreate({body: labelBody(label)})
			return data
		},
		onSuccess: (created, _label, _context, {client}) => {
			client.setQueryData<Label[]>(labelKeys.all, current => current ? [...current, created] : current)
		},
		onSettled: (_data, _error, _label, _context, {client}) =>
			client.invalidateQueries({queryKey: labelKeys.all}),
	})
}

async function snapshotLabels(client: QueryClient): Promise<Label[] | undefined> {
	await client.cancelQueries({queryKey: labelKeys.all})
	return client.getQueryData<Label[]>(labelKeys.all)
}

function restoreLabels(client: QueryClient, previous: Label[] | undefined) {
	if (previous) {
		client.setQueryData<Label[]>(labelKeys.all, previous)
	}
}

export function updateLabelMutationOptions() {
	return mutationOptions({
		mutationFn: async ({id, ...label}: UpdateLabelInput) => {
			const {data} = await labelsUpdate({path: {id}, body: labelBody(label)})
			return data
		},
		onMutate: async ({id, ...label}, {client}) => {
			const previous = await snapshotLabels(client)
			client.setQueryData<Label[]>(labelKeys.all, current =>
				current?.map(existing => existing.id === id ? {...existing, ...labelBody(label)} : existing),
			)
			return {previous}
		},
		onError: (_error, _label, context, {client}) => restoreLabels(client, context?.previous),
		onSuccess: (updated, _label, _context, {client}) => {
			client.setQueryData<Label[]>(labelKeys.all, current =>
				current?.map(existing => existing.id === updated.id ? updated : existing),
			)
			success({message: i18n.global.t('label.edit.success')})
		},
		onSettled: (_data, _error, _label, _context, {client}) =>
			client.invalidateQueries({queryKey: labelKeys.all}),
	})
}

export function deleteLabelMutationOptions() {
	return mutationOptions({
		mutationFn: async (label: Label) => {
			if (typeof label.id === 'undefined') {
				throw new Error('Cannot delete a label without an id')
			}
			await labelsDelete({path: {id: label.id}})
		},
		onMutate: async (label, {client}) => {
			const previous = await snapshotLabels(client)
			client.setQueryData<Label[]>(labelKeys.all, current =>
				current?.filter(existing => existing.id !== label.id),
			)
			return {previous}
		},
		onError: (_error, _label, context, {client}) => restoreLabels(client, context?.previous),
		onSuccess: () => {
			success({message: i18n.global.t('label.deleteSuccess')})
		},
		onSettled: (_data, _error, _label, _context, {client}) =>
			client.invalidateQueries({queryKey: labelKeys.all}),
	})
}

export function useCreateLabelMutation() {
	return useMutation(createLabelMutationOptions())
}

export function useUpdateLabelMutation() {
	return useMutation(updateLabelMutationOptions())
}

export function useDeleteLabelMutation() {
	return useMutation(deleteLabelMutationOptions())
}
