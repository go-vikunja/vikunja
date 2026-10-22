import {defineStore} from 'pinia'
import {useLocalStorage} from '@vueuse/core'
import type {LocationQueryRaw} from 'vue-router'

export const useViewFiltersStore = defineStore('viewFilters', () => {
	// localStorage so sort/filter survive reloads; URL query still wins when present.
	const viewQueries = useLocalStorage<Record<number, LocationQueryRaw>>('viewFilters', {})

	function setViewQuery(viewId: number, query: LocationQueryRaw) {
		viewQueries.value[viewId] = query
	}

	function getViewQuery(viewId: number): LocationQueryRaw {
		return viewQueries.value[viewId] ?? {}
	}

	function clearViewQuery(viewId: number) {
		delete viewQueries.value[viewId]
	}

	return {
		viewQueries,
		setViewQuery,
		getViewQuery,
		clearViewQuery,
	}
})
