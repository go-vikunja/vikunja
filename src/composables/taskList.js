import { ref, watch, computed } from 'vue'
import { useRoute } from 'vue-router'

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

/**
 * This mixin provides a base set of methods and properties to get tasks on a list.
 */
export function createTaskList(initTasks) {
	const taskCollectionService = ref(new TaskCollectionService())
	const loading = computed(() => taskCollectionService.value.loading)
	const totalPages = computed(() => taskCollectionService.value.totalPages)

    const tasks = ref([])
	const currentPage = ref(0)
	const loadedList = ref(null)
	const searchTerm = ref('')
	const showTaskFilter = ref(false)
	const params = ref({...getDefaultParams()})

	const route = useRoute()

	async function loadTasks(
		page = 1,
		search = '',
		loadParams = { ...params.value },
		forceLoading = false,
	) {

		// Because this function is triggered every time on topNavigation, we're putting a condition here to only load it when we actually want to show tasks
		// FIXME: This is a bit hacky -> Cleanup.
		if (
			route.name !== 'list.list' &&
			route.name !== 'list.table' &&
			!forceLoading
		) {
			return
		}

		if (search !== '') {
			loadParams.s = search
		}

		const list = {listId: parseInt(route.params.listId)}

		const currentList = {
			id: list.listId,
			params: loadParams,
			search,
			page,
		}
		if (
			JSON.stringify(currentList) === JSON.stringify(loadedList.value) &&
			!forceLoading
		) {
			return
		}

		tasks.value = []
		tasks.value = await taskCollectionService.value.getAll(list, loadParams, page)
		currentPage.value = page
		loadedList.value = JSON.parse(JSON.stringify(currentList))
	}

	async function loadTasksForPage(query) {
		const { page, search } = query
		initTasks(params)
		await loadTasks(
			// The page parameter can be undefined, in the case where the user loads a new list from the side bar menu
			typeof page === 'undefined' ? 1 : Number(page),
			search,
			params.value,
		)
	}
		
	async function loadTasksOnSavedFilter() {
		if (
			typeof route.params.listId !== 'undefined' &&
			parseInt(route.params.listId) < 0
		) {
			await loadTasks(1, '', null, true)
		}
	}

	function initTaskList() {
		// Only listen for query path changes
		watch(() => route.query, loadTasksForPage, { immediate: true })
		watch(() => route.path, loadTasksOnSavedFilter)
	}

	return {
		tasks,
		initTaskList,
		loading,
		totalPages,
		currentPage,
		showTaskFilter,
		loadTasks,
		searchTerm,
		params,
	}
}