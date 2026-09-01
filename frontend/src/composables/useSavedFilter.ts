import {computed, ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import {getSavedFilterIdFromProjectId} from '@/client/queries/projects'
import {
	createSavedFilter,
	newSavedFilterDraft,
	savedFilterQuery,
	updateSavedFilter,
} from '@/client/queries/savedFilters'
import type {SavedFilterDraft, SavedFilterResponse} from '@/client/queries/savedFilters'

type SavedFilterForm = SavedFilterDraft & Pick<SavedFilterResponse, 'id'>

function newDraft(): SavedFilterForm {
	return {id: 0, ...newSavedFilterDraft()}
}

// Deep enough clone that v-model edits never write into the query cache entry.
function toDraft(value: SavedFilterResponse): SavedFilterForm {
	return {
		id: value.id,
		title: value.title,
		description: value.description,
		is_favorite: value.is_favorite,
		filters: {
			...value.filters,
			sort_by: [...value.filters.sort_by],
			order_by: [...value.filters.order_by],
		},
	}
}

export function useSavedFilter(projectId?: MaybeRefOrGetter<number | undefined>) {
	const isCreate = computed(() => toValue(projectId) === undefined)
	const savedFilterId = computed(() => getSavedFilterIdFromProjectId(toValue(projectId) ?? 0))
	const query = useQuery(computed(() => savedFilterQuery(savedFilterId.value)))

	const filter = ref<SavedFilterForm>(newDraft())
	const isSaving = ref(false)
	const titleTouched = ref(false)

	watch([savedFilterId, query.data], ([id, value], [previousId]) => {
		if (id !== previousId) {
			filter.value = newDraft()
		}
		// Seed once per id; same-id refetches never overwrite the draft.
		if (value && value.id === id && value.id !== filter.value.id) {
			filter.value = toDraft(value)
		}
	}, {immediate: true})

	const titleValid = computed(() => !titleTouched.value || filter.value.title !== '')

	function markTitleTouched() {
		titleTouched.value = true
	}

	function writableFilter(): SavedFilterDraft {
		return {
			title: filter.value.title,
			description: filter.value.description,
			filters: filter.value.filters,
			is_favorite: filter.value.is_favorite,
		}
	}

	async function submit(): Promise<SavedFilterResponse | undefined> {
		titleTouched.value = true
		if (!titleValid.value) {
			return
		}

		// Updates need a real filter id and the draft seeded from it; an unseeded draft still carries id 0.
		if (!isCreate.value && (!(savedFilterId.value > 0) || filter.value.id !== savedFilterId.value)) {
			return
		}

		const id = savedFilterId.value
		isSaving.value = true
		try {
			const saved = isCreate.value
				? await createSavedFilter(writableFilter())
				: await updateSavedFilter({id: filter.value.id, ...writableFilter()})
			// The route id can change mid-flight; the watcher already reset the draft for it.
			if (savedFilterId.value === id) {
				filter.value = toDraft(saved)
			}
			return saved
		} finally {
			isSaving.value = false
		}
	}

	return {
		filter,
		// Disabled queries stay pending forever; refetches must not lock the form.
		isLoading: computed(() =>
			(savedFilterId.value > 0 && query.isPending.value) || isSaving.value,
		),
		error: query.error,
		titleValid,
		markTitleTouched,
		submit,
	}
}
