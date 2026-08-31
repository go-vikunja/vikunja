import {beforeEach, describe, expect, it, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'

const sdk = vi.hoisted(() => ({
	labelsList: vi.fn(),
	labelsCreate: vi.fn(),
	labelsUpdate: vi.fn(),
	labelsDelete: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)

import {
	createLabel,
	deleteLabel,
	ensureLabels,
	filterLabelsByQuery,
	getLabelByExactTitle,
	getLabelsByExactTitles,
	getLabelById,
	getLabelsByIds,
	labelKeys,
	labelsQuery,
	refreshLabels,
	sortLabelsAlphabetically,
	updateLabel,
} from './labels'

import type {Label} from '@/client/generated'

const labels: Label[] = [
	{id: 3, title: 'Zulu', description: 'last'},
	{id: 1, title: 'Alpha', description: 'first'},
	{id: 2, title: 'Bravo', description: 'middle'},
]

describe('labels query', () => {
	beforeEach(() => {
		queryClient.clear()
		sdk.labelsList.mockReset()
		sdk.labelsCreate.mockReset()
		sdk.labelsUpdate.mockReset()
		sdk.labelsDelete.mockReset()
	})

	it('loads every page with the maximum page size', async () => {
		sdk.labelsList
			.mockResolvedValueOnce({data: {items: labels.slice(0, 2), total_pages: 2}})
			.mockResolvedValueOnce({data: {items: labels.slice(2), total_pages: 2}})

		const result = await queryClient.fetchQuery(labelsQuery())

		expect(result).toEqual(labels)
		expect(sdk.labelsList).toHaveBeenNthCalledWith(1, {query: {page: 1, per_page: 1000}})
		expect(sdk.labelsList).toHaveBeenNthCalledWith(2, {query: {page: 2, per_page: 1000}})
	})

	it('sorts labels by title without changing the cached array', () => {
		const sorted = sortLabelsAlphabetically(labels, 'en')

		expect(sorted.map(label => label.id)).toEqual([1, 2, 3])
		expect(labels.map(label => label.id)).toEqual([3, 1, 2])
	})

	it('uses a five minute stale time', () => {
		expect(labelsQuery().staleTime).toBe(5 * 60 * 1000)
	})

	it('deduplicates concurrent imperative loads through the query cache', async () => {
		let resolveList: (value: unknown) => void = () => {}
		sdk.labelsList.mockReturnValue(new Promise(resolve => {
			resolveList = resolve
		}))

		const first = ensureLabels()
		const second = ensureLabels()
		resolveList({data: {items: labels, total_pages: 1}})

		await expect(Promise.all([first, second])).resolves.toEqual([labels, labels])
		expect(sdk.labelsList).toHaveBeenCalledOnce()
	})

	it('refreshes labels even while cached data is fresh', async () => {
		queryClient.setQueryData(labelKeys.all, [{id: 1, title: 'cached'}])
		const remoteLabels = [{id: 2, title: 'remote'}]
		sdk.labelsList.mockResolvedValue({data: {items: remoteLabels, total_pages: 1}})

		await expect(refreshLabels()).resolves.toEqual(remoteLabels)
		expect(sdk.labelsList).toHaveBeenCalledOnce()
	})
})

describe('label derivations', () => {
	it('looks labels up by id', () => {
		expect(getLabelById(labels, 2)?.title).toBe('Bravo')
		expect(getLabelsByIds(labels, [3, 1]).map(label => label.id)).toEqual([3, 1])
	})

	it('looks labels up by title case-insensitively', () => {
		expect(getLabelByExactTitle(labels, 'alpha')?.id).toBe(1)
		expect(getLabelsByExactTitles(labels, ['bravo', 'ZULU']).map(label => label.id)).toEqual([3, 2])
	})

	it('filters titles and descriptions while excluding hidden labels', () => {
		expect(filterLabelsByQuery(labels, [labels[0]], 'last')).toEqual([])
		expect(filterLabelsByQuery(labels, [], 'MID')).toEqual([labels[2]])
		expect(filterLabelsByQuery(labels, [], '')).toEqual([])
	})
})

describe('label cache mutations', () => {
	beforeEach(() => {
		queryClient.clear()
		sdk.labelsList.mockReset()
		sdk.labelsCreate.mockReset()
		sdk.labelsUpdate.mockReset()
		sdk.labelsDelete.mockReset()
		queryClient.setQueryData(labelKeys.all, labels)
	})

	it('adds a created label to the cache', async () => {
		const created = {id: 4, title: 'Created'}
		sdk.labelsCreate.mockResolvedValue({data: created})

		await expect(createLabel({title: 'Created'})).resolves.toEqual(created)
		expect(queryClient.getQueryData<Label[]>(labelKeys.all)).toContainEqual(created)
	})

	it('keeps a created label when an older list request finishes later', async () => {
		let resolveList: (value: unknown) => void = () => {}
		sdk.labelsList.mockReturnValue(new Promise(resolve => {
			resolveList = resolve
		}))
		const list = queryClient.fetchQuery({...labelsQuery(), staleTime: 0})
		const listSettled = list.catch(() => undefined)
		expect(sdk.labelsList).toHaveBeenCalledOnce()
		const created = {id: 4, title: 'Created'}
		sdk.labelsCreate.mockResolvedValue({data: created})

		await createLabel({title: 'Created'})
		resolveList({data: {items: labels, total_pages: 1}})
		await listSettled

		expect(queryClient.getQueryData<Label[]>(labelKeys.all)).toContainEqual(created)
	})

	it('replaces an updated label in the cache', async () => {
		const updated = {...labels[1], title: 'Updated'}
		sdk.labelsUpdate.mockResolvedValue({data: updated})

		await expect(updateLabel({id: 1, title: 'Updated'})).resolves.toEqual(updated)
		expect(getLabelById(queryClient.getQueryData<Label[]>(labelKeys.all) ?? [], 1)).toEqual(updated)
	})

	it('keeps an updated label when an older list request finishes later', async () => {
		let resolveList: (value: unknown) => void = () => {}
		sdk.labelsList.mockReturnValue(new Promise(resolve => {
			resolveList = resolve
		}))
		const list = queryClient.fetchQuery({...labelsQuery(), staleTime: 0})
		const listSettled = list.catch(() => undefined)
		expect(sdk.labelsList).toHaveBeenCalledOnce()
		const updated = {...labels[1], title: 'Updated'}
		sdk.labelsUpdate.mockResolvedValue({data: updated})

		await updateLabel({id: 1, title: 'Updated'})
		resolveList({data: {items: labels, total_pages: 1}})
		await listSettled

		expect(getLabelById(queryClient.getQueryData<Label[]>(labelKeys.all) ?? [], 1)).toEqual(updated)
	})

	it('removes a deleted label from the cache', async () => {
		sdk.labelsDelete.mockResolvedValue({data: undefined})

		await deleteLabel(labels[1])

		expect(getLabelById(queryClient.getQueryData<Label[]>(labelKeys.all) ?? [], 1)).toBeUndefined()
	})

	it('keeps a label deleted when an older list request finishes later', async () => {
		let resolveList: (value: unknown) => void = () => {}
		sdk.labelsList.mockReturnValue(new Promise(resolve => {
			resolveList = resolve
		}))
		const list = queryClient.fetchQuery({...labelsQuery(), staleTime: 0})
		const listSettled = list.catch(() => undefined)
		expect(sdk.labelsList).toHaveBeenCalledOnce()
		sdk.labelsDelete.mockResolvedValue({data: undefined})

		await deleteLabel(labels[1])
		resolveList({data: {items: labels, total_pages: 1}})
		await listSettled

		expect(getLabelById(queryClient.getQueryData<Label[]>(labelKeys.all) ?? [], 1)).toBeUndefined()
	})
})
