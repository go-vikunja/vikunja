import { ref, shallowReactive, watch, computed } from 'vue'
import {useRoute} from 'vue-router'

import TaskCollectionService from '@/services/taskCollection'

// FIXME: merge with DEFAULT_PARAMS in filters.vue
export const getDefaultParams = () => ({
	sort_by: ['position', 'id'],
	order_by: ['asc', 'desc'],
	filter_by: ['done'],
	filter_value: ['false'],
	filter_comparator: ['equals'],
	filter_concat: 'and',
})


const filters = {
	done: {
		value: false,
		comparator: 'equals',
		concat: 'and',
	},
}

const SORT_BY_DEFAULT = {
	id: 'desc',
}

/**
 * This mixin provides a base set of methods and properties to get tasks on a list.
 */
export function useTaskList(listId) {
	const params = ref({...getDefaultParams()})
	
	const search = ref('')
	const page = ref(1)

	const sortBy = ref({ ...SORT_BY_DEFAULT })


	// This makes sure an id sort order is always sorted last.
	// When tasks would be sorted first by id and then by whatever else was specified, the id sort takes
	// precedence over everything else, making any other sort columns pretty useless.
	function formatSortOrder(params) {
		let hasIdFilter = false
		const sortKeys = Object.keys(sortBy.value)
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
		params.order_by = sortKeys.map(s => sortBy.value[s])

		return params
	}

	const getAllTasksParams = computed(() => {
		let loadParams = {...params.value}

		if (search.value !== '') {
			loadParams.s = search.value
		}

		loadParams = formatSortOrder(loadParams)

		return [
			{listId: listId.value},
			loadParams,
			page.value || 1,
		]
	})

	const taskCollectionService = shallowReactive(new TaskCollectionService())
	const loading = computed(() => taskCollectionService.loading)
	const totalPages = computed(() => taskCollectionService.totalPages)

	const tasks = ref([])
	async function loadTasks() {
		tasks.value = await taskCollectionService.getAll(...getAllTasksParams.value)
		return tasks.value
	}

	const route = useRoute()
	watch(() => route.query, (query) => {
		const { page: pageQuery, search: searchQuery } = query
		search.value = searchQuery
		page.value = pageQuery

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
	}
}