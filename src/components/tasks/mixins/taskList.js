import TaskCollectionService from '../../../services/taskCollection'
import cloneDeep from 'lodash/cloneDeep'

/**
 * This mixin provides a base set of methods and properties to get tasks on a list.
 */
export default {
	data() {
		return {
			taskCollectionService: TaskCollectionService,
			tasks: [],

			pages: [],
			currentPage: 0,

			loadedList: null,

			showTaskSearch: false,
			searchTerm: '',

			showTaskFilter: false,
			params: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter_by: ['done'],
				filter_value: ['false'],
				filter_comparator: ['equals'],
				filter_concat: 'and',
			},
		}
	},
	watch: {
		'$route.query': 'loadTasksForPage', // Only listen for query path changes
		'$route.path': 'loadTasksOnSavedFilter',
	},
	beforeMount() {
		// Triggering loading the tasks in beforeMount lets the component maintain the current page, therefore the page
		// is not lost after navigating back from a task detail page for example.
		this.loadTasksForPage(this.$route.query)
	},
	created() {
		this.taskCollectionService = new TaskCollectionService()
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

			this.$set(this, 'tasks', [])

			this.taskCollectionService.getAll(list, params, page)
				.then(r => {
					this.$set(this, 'tasks', r)
					this.$set(this, 'pages', [])
					this.currentPage = page

					for (let i = 0; i < this.taskCollectionService.totalPages; i++) {

						// Show ellipsis instead of all pages
						if (
							i > 0 && // Always at least the first page
							(i + 1) < this.taskCollectionService.totalPages && // And the last page
							(
								// And the current with current + 1 and current - 1
								(i + 1) > this.currentPage + 1 ||
								(i + 1) < this.currentPage - 1
							)
						) {
							// Only add an ellipsis if the last page isn't already one
							if (this.pages[i - 1] && !this.pages[i - 1].isEllipsis) {
								this.pages.push({
									number: 0,
									isEllipsis: true,
								})
							}
							continue
						}

						this.pages.push({
							number: i + 1,
							isEllipsis: false,
						})
					}

					this.loadedList = cloneDeep(currentList)
				})
				.catch(e => {
					this.error(e)
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

				if (a.id > b.id)
					return -1
				if (a.id < b.id)
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
		getRouteForPagination(page = 1, type = 'list') {
			return {
				name: 'list.' + type,
				params: {
					type: type,
				},
				query: {
					page: page,
				},
			}
		},
	},
}