import {beforeEach, describe, expect, it, vi} from 'vitest'
import type {MutationOptions} from '@tanstack/vue-query'

import {queryClient} from '@/client/queryClient'

const sdk = vi.hoisted(() => ({
	labelsList: vi.fn(),
	labelsCreate: vi.fn(),
	labelsUpdate: vi.fn(),
	labelsDelete: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn()}))

import {
	createLabelMutationOptions,
	deleteLabelMutationOptions,
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
	updateLabelMutationOptions,
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

	it('requests labels with the maximum page size', async () => {
		sdk.labelsList.mockResolvedValue({data: {items: labels, total_pages: 1}})

		const result = await queryClient.fetchQuery(labelsQuery())

		expect(result).toEqual(labels)
		expect(sdk.labelsList).toHaveBeenCalledExactlyOnceWith({query: {page: 1, per_page: 1000}})
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

function runMutation<TData, TVariables, TContext>(
	options: MutationOptions<TData, Error, TVariables, TContext>,
	variables: TVariables,
): Promise<TData> {
	return queryClient.getMutationCache().build(queryClient, options).execute(variables)
}

function cachedLabels(): Label[] {
	return queryClient.getQueryData<Label[]>(labelKeys.all) ?? []
}

describe('label cache mutations', () => {
	beforeEach(() => {
		queryClient.clear()
		sdk.labelsCreate.mockReset()
		sdk.labelsUpdate.mockReset()
		sdk.labelsDelete.mockReset()
		queryClient.setQueryData(labelKeys.all, labels)
	})

	it('adds a created label to the cache and invalidates the list', async () => {
		const created = {id: 4, title: 'Created'}
		sdk.labelsCreate.mockResolvedValue({data: created})
		const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

		await expect(runMutation(createLabelMutationOptions(), {title: 'Created'})).resolves.toEqual(created)

		expect(cachedLabels()).toContainEqual(created)
		expect(invalidate).toHaveBeenCalledWith({queryKey: labelKeys.all})
	})

	it('does not materialize a label list nobody loaded', async () => {
		queryClient.clear()
		sdk.labelsCreate.mockResolvedValue({data: {id: 4, title: 'Created'}})

		await runMutation(createLabelMutationOptions(), {title: 'Created'})

		expect(queryClient.getQueryData(labelKeys.all)).toBeUndefined()
	})

	it('replaces an updated label in the cache', async () => {
		const updated = {...labels[1], title: 'Updated'}
		sdk.labelsUpdate.mockResolvedValue({data: updated})

		await runMutation(updateLabelMutationOptions(), {id: 1, title: 'Updated'})

		expect(getLabelById(cachedLabels(), 1)).toEqual(updated)
	})

	it('applies an update optimistically and rolls back on failure', async () => {
		sdk.labelsUpdate.mockImplementation(async () => {
			expect(getLabelById(cachedLabels(), 1)?.title).toBe('Updated')
			throw new Error('nope')
		})

		await expect(runMutation(updateLabelMutationOptions(), {id: 1, title: 'Updated'})).rejects.toThrow('nope')

		expect(getLabelById(cachedLabels(), 1)).toEqual(labels[1])
	})

	it('refuses to delete a label without an id', async () => {
		await expect(runMutation(deleteLabelMutationOptions(), {title: 'no id'})).rejects.toThrow()
		expect(sdk.labelsDelete).not.toHaveBeenCalled()
	})

	it('removes a label optimistically and restores it on failure', async () => {
		sdk.labelsDelete.mockImplementation(async () => {
			expect(getLabelById(cachedLabels(), 1)).toBeUndefined()
			throw new Error('nope')
		})

		await expect(runMutation(deleteLabelMutationOptions(), labels[1])).rejects.toThrow('nope')

		expect(cachedLabels()).toEqual(labels)
	})

	it('does not restore a list that was never loaded', async () => {
		queryClient.clear()
		sdk.labelsDelete.mockRejectedValue(new Error('nope'))

		await runMutation(deleteLabelMutationOptions(), labels[1]).catch(() => undefined)

		expect(queryClient.getQueryData(labelKeys.all)).toBeUndefined()
	})
})
