import {computed, ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import {getSavedFilterIdFromProjectId} from '@/client/queries/projects'
import {
	newSavedFilterDraft,
	savedFilterQuery,
} from '@/client/queries/savedFilters'
import type {SavedFilterDraft, SavedFilterResponse, UpdateSavedFilterInput} from '@/client/queries/savedFilters'

type SavedFilterForm = SavedFilterDraft & Pick<SavedFilterResponse, 'id'>

function newDraft(): SavedFilterForm {
	return {id: 0, ...newSavedFilterDraft()}
}

function toDraft(value: SavedFilterForm): SavedFilterForm {
	return {
		id: value.id,
		title: value.title,
		description: value.description,
		filters: {
			...value.filters,
			sort_by: [...value.filters.sort_by],
			order_by: [...value.filters.order_by],
		},
	}
}

export function useSavedFilterDraft() {
	const filter = ref<SavedFilterForm>(newDraft())
	const titleTouched = ref(false)
	const titleValid = computed(() => !titleTouched.value || filter.value.title !== '')

	function markTitleTouched() {
		titleTouched.value = true
	}

	function validate(): SavedFilterDraft | undefined {
		titleTouched.value = true
		if (!titleValid.value) {
			return
		}

		const {id: _id, ...payload} = toDraft(filter.value)
		return payload
	}

	return {filter, titleValid, markTitleTouched, validate}
}

export function useSavedFilter(projectId: MaybeRefOrGetter<number>) {
	const savedFilterId = computed(() => getSavedFilterIdFromProjectId(toValue(projectId)))
	const query = useQuery(computed(() => savedFilterQuery(savedFilterId.value)))
	const draft = useSavedFilterDraft()
	const {filter} = draft
	const isLoaded = computed(() => savedFilterId.value > 0 && filter.value.id === savedFilterId.value)

	watch([savedFilterId, query.data], ([id, value], [previousId]) => {
		if (id !== previousId) {
			filter.value = newDraft()
		}
		if (value && value.id === id && value.id !== filter.value.id) {
			filter.value = toDraft(value)
		}
	}, {immediate: true})

	return {
		...draft,
		isLoaded,
		isLoading: computed(() => savedFilterId.value > 0 && query.isPending.value),
		error: query.error,
		validate: (): Omit<UpdateSavedFilterInput, 'id'> | undefined => {
			const current = query.data.value
			if (!isLoaded.value || current?.id !== savedFilterId.value) {
				return
			}
			const payload = draft.validate()
			return payload && {...payload, is_favorite: current.is_favorite}
		},
	}
}
