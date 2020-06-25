<template>
	<div class="loader-container" :class="{ 'is-loading': taskCollectionService.loading}">
		<div class="filter-container">
			<div class="items">
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
									:class="{'is-loading': taskCollectionService.loading}">
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
				<button class="button" @click="showTaskFilter = !showTaskFilter">
					<span class="icon is-small">
						<icon icon="filter"/>
					</span>
					Filters
				</button>
			</div>
			<transition name="fade">
				<filters
						v-if="showTaskFilter"
						v-model="params"
						@change="loadTasks(1)"
				/>
			</transition>
		</div>

		<div class="field task-add" v-if="!list.isArchived">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" :class="{ 'is-loading': taskService.loading}">
					<input
							v-focus
							class="input"
							:class="{ 'disabled': taskService.loading}"
							v-model="newTaskText"
							type="text"
							placeholder="Add a new task..."
							@keyup.enter="addTask()"/>
					<span class="icon is-small is-left">
						<icon icon="tasks"/>
					</span>
				</p>
				<p class="control">
					<button class="button is-success" :disabled="newTaskText.length === 0" @click="addTask()">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
						Add
					</button>
				</p>
			</div>
			<p class="help is-danger" v-if="showError && newTaskText === ''">
				Please specify a list title.
			</p>
		</div>

		<div class="columns">
			<div class="column">
				<div class="tasks" v-if="tasks && tasks.length > 0" :class="{'short': isTaskEdit}">
					<div class="task" v-for="t in tasks" :key="t.id">
						<single-task-in-list :the-task="t" @taskUpdated="updateTasks" task-detail-route="task.detail"/>
						<div @click="editTask(t.id)" class="icon settings" v-if="!list.isArchived">
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

		<nav
				class="pagination is-centered"
				role="navigation"
				aria-label="pagination"
				v-if="taskCollectionService.totalPages > 1">
			<router-link
					class="pagination-previous"
					:to="getRouteForPagination(currentPage - 1)"
					tag="button"
					:disabled="currentPage === 1">
				Previous
			</router-link>
			<router-link
					class="pagination-next"
					:to="getRouteForPagination(currentPage + 1)"
					tag="button"
					:disabled="currentPage === taskCollectionService.totalPages">
				Next page
			</router-link>
			<ul class="pagination-list">
				<template v-for="(p, i) in pages">
					<li :key="'page'+i" v-if="p.isEllipsis"><span class="pagination-ellipsis">&hellip;</span></li>
					<li :key="'page'+i" v-else>
						<router-link
								:to="getRouteForPagination(p.number)"
								:class="{'is-current': p.number === currentPage}"
								class="pagination-link"
								:aria-label="'Goto page ' + p.number">
							{{ p.number }}
						</router-link>
					</li>
				</template>
			</ul>
		</nav>

		<!-- This router view is used to show the task popup while keeping the kanban board itself -->
		<transition name="modal">
			<router-view/>
		</transition>

	</div>
</template>

<script>
	import TaskService from '../../../services/task'
	import EditTask from '../../../components/tasks/edit-task'
	import TaskModel from '../../../models/task'
	import SingleTaskInList from '../../../components/tasks/partials/singleTaskInList'
	import taskList from '../../../components/tasks/mixins/taskList'
	import {saveListView} from '../../../helpers/saveListView'
	import Filters from '../../../components/list/partials/filters'

	export default {
		name: 'List',
		data() {
			return {
				taskService: TaskService,
				list: {},
				isTaskEdit: false,
				taskEditTask: TaskModel,
				newTaskText: '',

				showError: false,
			}
		},
		mixins: [
			taskList,
		],
		components: {
			Filters,
			SingleTaskInList,
			EditTask,
		},
		created() {
			this.taskService = new TaskService()

			// Save the current list view to local storage
			// We use local storage and not vuex here to make it persistent across reloads.
			saveListView(this.$route.params.listId, this.$route.name)
		},
		methods: {
			// This function initializes the tasks page and loads the first page of tasks
			initTasks(page, search = '') {
				this.taskEditTask = null
				this.isTaskEdit = false
				this.loadTasks(page, search)
			},
			addTask() {
				if (this.newTaskText === '') {
					this.showError = true
					return
				}
				this.showError = false

				let task = new TaskModel({title: this.newTaskText, listId: this.$route.params.listId})
				this.taskService.create(task)
					.then(r => {
						this.tasks.push(r)
						this.sortTasks()
						this.newTaskText = ''
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			editTask(id) {
				// Find the selected task and set it to the current object
				let theTask = this.getTaskById(id) // Somehow this does not work if we directly assign this to this.taskEditTask
				this.taskEditTask = theTask
				this.isTaskEdit = true
			},
			getTaskById(id) {
				for (const t in this.tasks) {
					if (this.tasks[t].id === parseInt(id)) {
						return this.tasks[t]
					}
				}
				return {} // FIXME: This should probably throw something to make it clear to the user noting was found
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
		}
	}
</script>