import {defineStore} from 'pinia'
import {useLocalStorage} from '@vueuse/core'
import type {LocationQueryRaw} from 'vue-router'
import type {IProjectView} from '@/modelTypes/IProjectView'

export const useViewFiltersStore = defineStore('viewFilters', () => {
	// localStorage so sort/filter survive reloads; URL query still wins when present.
	const viewQueries = useLocalStorage<Record<IProjectView['id'], LocationQueryRaw>>('viewFilters', {})

	function setViewQuery(viewId: IProjectView['id'], query: LocationQueryRaw) {
		viewQueries.value[viewId] = query
	}

	function getViewQuery(viewId: IProjectView['id']): LocationQueryRaw {
		return viewQueries.value[viewId] ?? {}
	}

	function clearViewQuery(viewId: IProjectView['id']) {
		delete viewQueries.value[viewId]
	}

	return {
		viewQueries,
		setViewQuery,
		getViewQuery,
		clearViewQuery,
	}
})
