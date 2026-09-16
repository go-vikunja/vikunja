import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import {
	filterLabelsByQuery as filterLabels,
	getLabelByExactTitle as findLabelByExactTitle,
	getLabelsByExactTitles as findLabelsByExactTitles,
	getLabelById as findLabelById,
	getLabelsByIds as findLabelsByIds,
	labelsQuery,
} from '@/client/queries/labels'
import type {Label} from '@/client/generated'

export function useLabels() {
	const query = useQuery(labelsQuery())
	const labels = computed(() => query.data.value ?? [])

	return {
		labels,
		isPending: query.isPending,
		filterLabelsByQuery: (hidden: Label[], value: string) => filterLabels(labels.value, hidden, value),
		getLabelById: (id: number) => findLabelById(labels.value, id),
		getLabelsByIds: (ids: number[]) => findLabelsByIds(labels.value, ids),
		getLabelByExactTitle: (title: string) => findLabelByExactTitle(labels.value, title),
		getLabelsByExactTitles: (titles: string[]) => findLabelsByExactTitles(labels.value, titles),
	}
}
