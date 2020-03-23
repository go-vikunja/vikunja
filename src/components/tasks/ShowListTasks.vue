<template>
	<div class="loader-container" :class="{ 'is-loading': listService.loading || taskCollectionService.loading}">
		<div class="search">
			<div class="field has-addons" :class="{ 'hidden': !showTaskSearch }">
				<div class="control has-icons-left has-icons-right">
					<input
							class="input"
							type="text"
							placeholder="Search"
							v-focus
							v-model="searchTerm"
							@keyup.enter="searchTasks"
							@blur="hideSearchBar()"/>
					<span class="icon is-left">
						<icon icon="search"/>
					</span>
				</div>
				<div class="control">
					<button
							class="button noshadow is-primary"
							@click="searchTasks"
							:class="{'is-loading': taskCollectionService.loading}"
							:disabled="searchTerm === ''">
						Search
					</button>
				</div>
			</div>
			<button class="button" @click="showTaskSearch = !showTaskSearch" v-if="!showTaskSearch">
				<span class="icon">
					<icon icon="search"/>
				</span>
			</button>
		</div>

		<div class="field task-add" v-if="!list.is_archived">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" :class="{ 'is-loading': taskService.loading}">
					<input v-focus class="input" :class="{ 'disabled': taskService.loading}" v-model="newTaskText" type="text" placeholder="Add a new task..." @keyup.enter="addTask()"/>
					<span class="icon is-small is-left">
						<icon icon="tasks"/>
					</span>
				</p>
				<p class="control">
					<button class="button is-success" :disabled="newTaskText.length < 3" @click="addTask()">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
						Add
					</button>
				</p>
			</div>
			<p class="help is-danger" v-if="showError && newTaskText.length < 3">
				Please specify at least three characters.
			</p>
		</div>

		<div class="columns">
			<div class="column">
				<div class="tasks" v-if="tasks && tasks.length > 0" :class="{'short': isTaskEdit}">
					<div class="task" v-for="t in tasks" :key="t.id">
						<single-task-in-list :the-task="t" @taskUpdated="updateTasks"/>
						<div @click="editTask(t.id)" class="icon settings" v-if="!list.is_archived">
							<icon icon="pencil-alt"/>
						</div>
					</div>
				</div>
			</div>
			<div class="column is-4" v-if="isTaskEdit">
				<div class="card taskedit">
					<header class="card-header">
						<p class="card-header-title">
							Edit Task
						</p>
						<a class="card-header-icon" @click="isTaskEdit = false">
							<span class="icon">
								<icon icon="angle-right"/>
							</span>
						</a>
					</header>
					<div class="card-content">
						<div class="content">
							<edit-task :task="taskEditTask"/>
						</div>
					</div>
				</div>
			</div>
		</div>

		<nav class="pagination is-centered" role="navigation" aria-label="pagination" v-if="taskCollectionService.totalPages > 1">
			<router-link class="pagination-previous" :to="{name: 'showList', query: { page: currentPage - 1 }}" tag="button" :disabled="currentPage === 1">Previous</router-link>
			<router-link class="pagination-next" :to="{name: 'showList', query: { page: currentPage + 1 }}" tag="button" :disabled="currentPage === taskCollectionService.totalPages">Next page</router-link>
			<ul class="pagination-list">
				<template v-for="(p, i) in pages">
					<li :key="'page'+i" v-if="p.isEllipsis"><span class="pagination-ellipsis">&hellip;</span></li>
					<li :key="'page'+i" v-else>
						<router-link :to="{name: 'showList', query: { page: p.number }}" :class="{'is-current': p.number === currentPage}" class="pagination-link" :aria-label="'Goto page ' + p.number">{{ p.number }}</router-link>
					</li>
				</template>
			</ul>
		</nav>
	</div>
</template>

<script>
	import ListService from '../../services/list'
	import TaskService from '../../services/task'
	import ListModel from '../../models/list'
	import EditTask from './edit-task'
	import TaskModel from '../../models/task'
	import TaskCollectionService from '../../services/taskCollection'
	import SingleTaskInList from './reusable/singleTaskInList'

	export default {
		data() {
			return {
				listID: this.$route.params.id,
				listService: ListService,
				taskService: TaskService,
				taskCollectionService: TaskCollectionService,
				pages: [],
				currentPage: 0,
				list: {},
				tasks: [],
				isTaskEdit: false,
				taskEditTask: TaskModel,
				newTaskText: '',

				showError: false,

				showTaskSearch: false,
				searchTerm: '',
			}
		},
		components: {
			SingleTaskInList,
			EditTask,
		},
		props: {
			theList: {
				type: ListModel,
				required: true,
			}
		},
		watch: {
			theList() {
				this.list = this.theList
			},
			'$route.query': 'loadTasksForPage', // Only listen for query path changes
		},
		created() {
			this.listService = new ListService()
			this.taskService = new TaskService()
			this.taskCollectionService = new TaskCollectionService()
			this.initTasks(1)
		},
		methods: {
			// This function initializes the tasks page and loads the first page of tasks
			initTasks(page, search = '') {
				this.taskEditTask = null
				this.isTaskEdit = false
				this.loadTasks(page, search)
			},
			addTask() {
				if (this.newTaskText.length < 3) {
					this.showError = true
					return
				}
				this.showError = false

				let task = new TaskModel({text: this.newTaskText, listID: this.$route.params.id})
				this.taskService.create(task)
					.then(r => {
						this.tasks.push(r)
						this.sortTasks()
						this.newTaskText = ''
						this.success({message: 'The task was successfully created.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			loadTasks(page, search = '') {
				const params = {sort_by: ['done', 'id'], order_by: ['asc', 'desc']}
				if (search !== '') {
					params.s = search
				}
				this.taskCollectionService.getAll({listID: this.$route.params.id}, params, page)
					.then(r => {
						this.$set(this, 'tasks', r)
						this.$set(this, 'pages', [])
						this.currentPage = page

						for (let i = 0; i < this.taskCollectionService.totalPages; i++)  {

							// Show ellipsis instead of all pages
							if(
								i > 0 && // Always at least the first page
								(i + 1) < this.taskCollectionService.totalPages && // And the last page
								(
									// And the current with current + 1 and current - 1
									(i + 1) > this.currentPage + 1 ||
									(i + 1) < this.currentPage - 1
								)
							) {
								// Only add an ellipsis if the last page isn't already one
								if(this.pages[i - 1] && !this.pages[i - 1].isEllipsis) {
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
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			loadTasksForPage(e) {
				// The page parameter can be undefined, in the case where the user loads a new list from the side bar menu
				let page = e.page
				if (typeof e.page === 'undefined') {
					page = 1
				}
				let search = e.search
				if (typeof e.search === 'undefined') {
					search = ''
				}
				this.initTasks(page, search)
			},
			editTask(id) {
				// Find the selected task and set it to the current object
				let theTask = this.getTaskByID(id) // Somehow this does not work if we directly assign this to this.taskEditTask
				this.taskEditTask = theTask
				this.isTaskEdit = true
			},
			getTaskByID(id) {
				for (const t in this.tasks) {
					if (this.tasks[t].id === parseInt(id)) {
						return this.tasks[t]
					}
				}
				return {} // FIXME: This should probably throw something to make it clear to the user noting was found
			},
			sortTasks() {
				if (this.tasks === null || this.tasks === []) {
					return
				}
				return this.tasks.sort(function(a,b) {
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
			updateTasks(updatedTask) {
				for (const t in this.tasks) {
					if (this.tasks[t].id === updatedTask.id) {
						this.$set(this.tasks, t, updatedTask)
						break
					}
				}
				this.sortTasks()
			},
			searchTasks() {
				if (this.searchTerm === '') {
					return
				}
				this.$router.push({
					name: 'showList',
					query: {search: this.searchTerm}
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
		}
	}
</script>