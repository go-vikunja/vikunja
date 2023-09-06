import {ref, shallowReactive, watch, computed, type ComputedGetter} from 'vue'
import {useRoute} from 'vue-router'
import {useRouteQuery} from '@vueuse/router'

import TaskCollectionService from '@/services/taskCollection'
import type {ITask} from '@/modelTypes/ITask'
import {error} from '@/message'
import type {IProject} from '@/modelTypes/IProject'

export type Order = 'asc' | 'desc' | 'none'

export interface SortBy {
	id?: Order
	index?: Order
	done?: Order
	title?: Order
	priority?: Order
	due_date?: Order
	start_date?: Order
	end_date?: Order
	percent_done?: Order
	created?: Order
	updated?: Order
}

// FIXME: merge with DEFAULT_PARAMS in filters.vue
export const getDefaultParams = () => ({
	sort_by: ['position', 'id'],
	order_by: ['asc', 'desc'],
	filter_by: ['done'],
	filter_value: ['false'],
	filter_comparator: ['equals'],
	filter_concat: 'and',
})

const SORT_BY_DEFAULT: SortBy = {
	id: 'desc',
}

	// This makes sure an id sort order is always sorted last.
	// When tasks would be sorted first by id and then by whatever else was specified, the id sort takes
	// precedence over everything else, making any other sort columns pretty useless.
	function formatSortOrder(sortBy, params) {
		let hasIdFilter = false
		const sortKeys = Object.keys(sortBy)
		for (const s of sortKeys) {
			if (s === 'id') {
				sortKeys.splice(s, 1)
				hasIdFilter = true
				break
			}
		}
		if (hasIdFilter) {
			sortKeys.push('id')
		}
		params.sort_by = sortKeys
		params.order_by = sortKeys.map(s => sortBy[s])

		return params
	}

/**
 * This mixin provides a base set of methods and properties to get tasks.
 */
export function useTaskList(projectIdGetter: ComputedGetter<IProject['id']>, sortByDefault: SortBy = SORT_BY_DEFAULT) {
	
	const projectId = computed(() => projectIdGetter())
	
	const params = ref({...getDefaultParams()})
	
	const search = ref('')
	const page = useRouteQuery('page', '1', { transform: Number })

	const sortBy = ref({ ...sortByDefault })
	
	const allParams = computed(() => {
		const loadParams = {...params.value}

		if (search.value !== '') {
			loadParams.s = search.value
		}

		return formatSortOrder(sortBy.value, loadParams)
	})
	
	watch(
		() => allParams.value,
		() => {
			// When parameters change, the page should always be the first
			page.value = 1
		},
	)

	const getAllTasksParams = computed(() => {
		return [
			{projectId: projectId.value},
			allParams.value,
			page.value,
		]
	})

	const taskCollectionService = shallowReactive(new TaskCollectionService())
	const loading = computed(() => taskCollectionService.loading)
	const totalPages = computed(() => taskCollectionService.totalPages)

	const tasks = ref<ITask[]>([])
	async function loadTasks() {
		tasks.value = []
		try {
			tasks.value = await taskCollectionService.getAll(...getAllTasksParams.value)
		} catch (e) {
			error(e)
		}
		return tasks.value
	}

	const route = useRoute()
	watch(() => route.query, (query) => {
		const { page: pageQueryValue, search: searchQuery } = query
		if (searchQuery !== undefined) {
			search.value = searchQuery as string
		}
		if (pageQueryValue !== undefined) {
			page.value = Number(pageQueryValue)
		}

	}, { immediate: true })


	// Only listen for query path changes
	watch(() => JSON.stringify(getAllTasksParams.value), (newParams, oldParams) => {
		if (oldParams === newParams) {
			return
		}

		loadTasks()
	}, { immediate: true })

	return {
		tasks,
		loading,
		totalPages,
		currentPage: page,
		loadTasks,
		searchTerm: search,
		params,
		sortByParam: sortBy,
	}
}