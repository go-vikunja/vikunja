import {defineStore} from 'pinia'
import {ref} from 'vue'
import type {LocationQueryRaw} from 'vue-router'
import type {IProjectView} from '@/modelTypes/IProjectView'

export const useViewFiltersStore = defineStore('viewFilters', () => {
	const viewQueries = ref<Record<IProjectView['id'], LocationQueryRaw>>({})

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
