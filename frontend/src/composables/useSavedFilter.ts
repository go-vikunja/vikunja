import {computed, ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import {
	createSavedFilter,
	createSavedFilterDraft,
	deleteSavedFilter,
	getProjectIdFromSavedFilterId,
	getSavedFilterIdFromProjectId,
	savedFilterQuery,
	updateSavedFilter,
} from '@/client/queries/savedFilters'
import type {SavedFilterDraft, SavedFilterResponse} from '@/client/queries/savedFilters'
import {success} from '@/message'

type SavedFilterForm = SavedFilterDraft & Pick<SavedFilterResponse, 'id'>

export function useSavedFilter(projectId?: MaybeRefOrGetter<number | undefined>) {
	const router = useRouter()
	const {t} = useI18n({useScope: 'global'})
	const savedFilterId = computed(() => getSavedFilterIdFromProjectId(toValue(projectId) ?? 0))
	const query = useQuery(computed(() => ({
		...savedFilterQuery(savedFilterId.value),
		enabled: savedFilterId.value > 0,
	})))
	const filter = ref<SavedFilterForm>({id: 0, ...createSavedFilterDraft()})
	const loadedSavedFilterId = ref(0)
	const isSaving = ref(false)
	const titleValid = ref(true)

	watch([
		savedFilterId,
		query.data,
		query.isFetching,
		query.isError,
	], ([id, value, isFetching]) => {
		if (loadedSavedFilterId.value !== id) {
			filter.value = {id: 0, ...createSavedFilterDraft()}
		}
		if (!isFetching && value && loadedSavedFilterId.value !== id) {
			filter.value = {
				...value,
				filters: {
					...value.filters,
					sort_by: [...value.filters.sort_by],
					order_by: [...value.filters.order_by],
				},
			}
			titleValid.value = value.title !== ''
			loadedSavedFilterId.value = id
		}
	}, {immediate: true})

	const filters = computed({
		get: () => filter.value.filters,
		set(value) {
			filter.value.filters = value
		},
	})

	const validateTitleField = useDebounceFn(() => {
		titleValid.value = filter.value.title !== ''
	}, 100)

	function writableFilter(): SavedFilterDraft {
		return {
			title: filter.value.title,
			description: filter.value.description,
			filters: filter.value.filters,
			is_favorite: filter.value.is_favorite,
		}
	}

	async function createFilter() {
		isSaving.value = true
		try {
			filter.value = await createSavedFilter(writableFilter())
			await router.push({
				name: 'project.index',
				params: {projectId: getProjectIdFromSavedFilterId(filter.value.id)},
			})
		} finally {
			isSaving.value = false
		}
	}

	async function saveFilter() {
		if (loadedSavedFilterId.value !== savedFilterId.value || filter.value.id <= 0) {
			throw new Error('Saved filter details are not loaded')
		}

		isSaving.value = true
		try {
			filter.value = await updateSavedFilter({
				id: filter.value.id,
				...writableFilter(),
			})
			success({message: t('filters.edit.success')})
			router.back()
		} finally {
			isSaving.value = false
		}
	}

	async function deleteFilter() {
		isSaving.value = true
		try {
			await deleteSavedFilter(savedFilterId.value)
			success({message: t('filters.delete.success')})
			await router.push({name: 'projects.index'})
		} finally {
			isSaving.value = false
		}
	}

	async function createFilterWithValidation() {
		if (titleValid.value) {
			return createFilter()
		}
	}

	async function saveFilterWithValidation() {
		if (titleValid.value) {
			return saveFilter()
		}
	}

	return {
		createFilter,
		createFilterWithValidation,
		saveFilter,
		saveFilterWithValidation,
		deleteFilter,
		filter,
		filters,
		isLoading: computed(() =>
			(savedFilterId.value > 0 && (
				query.isFetching.value ||
				(!query.isError.value && loadedSavedFilterId.value !== savedFilterId.value)
			)) || isSaving.value,
		),
		error: query.error,
		titleValid,
		validateTitleField,
	}
}
