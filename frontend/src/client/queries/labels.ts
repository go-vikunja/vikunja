import {queryOptions, useMutation} from '@tanstack/vue-query'

import {
	labelsCreate,
	labelsDelete,
	labelsList,
	labelsUpdate,
} from '@/client/generated'
import type {Label, LabelWritable} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
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

async function fetchAllLabels(): Promise<Label[]> {
	const result: Label[] = []
	let page = 1

	while (true) {
		const {data} = await labelsList({query: {page, per_page: 1000}})
		result.push(...(data.items ?? []))
		if (page >= (data.total_pages ?? 1)) {
			break
		}
		page++
	}

	return result
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

export async function createLabel(label: CreateLabelInput): Promise<Label> {
	await queryClient.cancelQueries({queryKey: labelKeys.all})
	const {data} = await labelsCreate({body: labelBody(label)})
	queryClient.setQueryData<Label[]>(labelKeys.all, current => current ? [...current, data] : current)
	return data
}

export async function updateLabel({id, ...label}: UpdateLabelInput): Promise<Label> {
	await queryClient.cancelQueries({queryKey: labelKeys.all})
	const {data} = await labelsUpdate({path: {id}, body: labelBody(label)})
	queryClient.setQueryData<Label[]>(labelKeys.all, current =>
		current?.map(existing => existing.id === data.id ? data : existing),
	)
	return data
}

export async function deleteLabel(label: Label): Promise<void> {
	if (typeof label.id === 'undefined') {
		throw new Error('Cannot delete a label without an id')
	}

	await queryClient.cancelQueries({queryKey: labelKeys.all})
	await labelsDelete({path: {id: label.id}})
	queryClient.setQueryData<Label[]>(labelKeys.all, current =>
		current?.filter(existing => existing.id !== label.id),
	)
}

export function useCreateLabelMutation() {
	return useMutation({mutationFn: createLabel})
}

export function useUpdateLabelMutation() {
	return useMutation({
		mutationFn: updateLabel,
		onSuccess: () => success({message: i18n.global.t('label.edit.success')}),
	})
}

export function useDeleteLabelMutation() {
	return useMutation({
		mutationFn: deleteLabel,
		onSuccess: () => success({message: i18n.global.t('label.deleteSuccess')}),
	})
}
