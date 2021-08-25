import TaskCollectionService from '@/services/taskCollection'
import cloneDeep from 'lodash/cloneDeep'
import {calculateItemPosition} from '../../../helpers/calculateItemPosition'

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

			showTaskSearch: false,
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
				params: params,
				search: search,
				page: page,
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
		sortTasks() {
			if (this.tasks === null || this.tasks === []) {
				return
			}
			return this.tasks.sort(function (a, b) {
				if (a.done < b.done)
					return -1
				if (a.done > b.done)
					return 1

				if (a.position < b.position)
					return -1
				if (a.position > b.position)
					return 1
				return 0
			})
		},
		searchTasks() {
			// Only search if the search term changed
			if (this.$route.query === this.searchTerm) {
				return
			}

			this.$router.push({
				name: 'list.list',
				query: {search: this.searchTerm},
			})
		},
		hideSearchBar() {
			// This is a workaround.
			// When clicking on the search button, @blur from the input is fired. If we
			// would then directly hide the whole search bar directly, no click event
			// from the button gets fired. To prevent this, we wait 200ms until we hide
			// everything so the button has a chance of firering the search event.
			setTimeout(() => {
				this.showTaskSearch = false
			}, 200)
		},
		saveTaskPosition(e) {
			this.drag = false
			
			const task = this.tasks[e.newIndex]
			const taskBefore = this.tasks[e.newIndex - 1] ?? null
			const taskAfter = this.tasks[e.newIndex + 1] ??  null
			
			task.position = calculateItemPosition(taskBefore !== null ? taskBefore.position : null, taskAfter !== null ? taskAfter.position : null)

			this.$store.dispatch('tasks/update', task)
				.then(r => {
					this.$set(this.tasks, e.newIndex, r)
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
	},
}