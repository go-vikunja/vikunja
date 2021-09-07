import TaskCollectionService from '@/services/taskCollection'
import cloneDeep from 'lodash/cloneDeep'

// FIXME: merge with DEFAULT_PARAMS in filters.vue
const DEFAULT_PARAMS = {
	sort_by: ['position', 'id'],
	order_by: ['asc', 'desc'],
	filter_by: ['done'],
	filter_value: ['false'],
	filter_comparator: ['equals'],
	filter_concat: 'and',
}

/**
 * This mixin provides a base set of methods and properties to get tasks on a list.
 */
export default {
	data() {
		return {
			taskCollectionService: new TaskCollectionService(),
			tasks: [],

			currentPage: 0,

			loadedList: null,

			searchTerm: '',

			showTaskFilter: false,
			params: DEFAULT_PARAMS,
		}
	},
	watch: {
		// Only listen for query path changes
		'$route.query': {
			handler: 'loadTasksForPage',
			immediate: true,
		},
		'$route.path': 'loadTasksOnSavedFilter',
	},
	methods: {
		loadTasks(
			page,
			search = '',
			params = null,
			forceLoading = false,
		) {

			// Because this function is triggered every time on topNavigation, we're putting a condition here to only load it when we actually want to show tasks
			// FIXME: This is a bit hacky -> Cleanup.
			if (
				this.$route.name !== 'list.list' &&
				this.$route.name !== 'list.table' &&
				!forceLoading
			) {
				return
			}

			if (params === null) {
				params = this.params
			}

			if (search !== '') {
				params.s = search
			}

			const list = {listId: parseInt(this.$route.params.listId)}

			const currentList = {
				id: list.listId,
				params,
				search,
				page,
			}
			if (JSON.stringify(currentList) === JSON.stringify(this.loadedList) && !forceLoading) {
				return
			}

			this.tasks = []

			this.taskCollectionService.getAll(list, params, page)
				.then(r => {
					this.tasks = r
					this.currentPage = page

					this.loadedList = cloneDeep(currentList)
				})
				.catch(e => {
					this.$message.error(e)
				})
		},

		loadTasksForPage(e) {
			// The page parameter can be undefined, in the case where the user loads a new list from the side bar menu
			let page = Number(e.page)
			if (typeof e.page === 'undefined') {
				page = 1
			}
			let search = e.search
			if (typeof e.search === 'undefined') {
				search = ''
			}
			this.initTasks(page, search)
		},
		loadTasksOnSavedFilter() {
			if(typeof this.$route.params.listId !== 'undefined' && parseInt(this.$route.params.listId) < 0) {
				this.loadTasks(1, '', null, true)
			}
		},
		getRouteForPagination,
	},
}