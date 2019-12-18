<template>
	<div class="loader-container" :class="{ 'is-loading': listService.loading || taskCollectionService.loading}">
		<form @submit.prevent="addTask()">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" :class="{ 'is-loading': taskService.loading}">
					<input v-focus class="input" :class="{ 'disabled': taskService.loading}" v-model="newTaskText" type="text" placeholder="Add a new task...">
					<span class="icon is-small is-left">
						<icon icon="tasks"/>
					</span>
				</p>
				<p class="control">
					<button type="submit" class="button is-success">
					<span class="icon is-small">
						<icon icon="plus"/>
					</span>
						Add
					</button>
				</p>
			</div>
		</form>

		<div class="columns">
			<div class="column">
				<div class="tasks" v-if="tasks && tasks.length > 0" :class="{'short': isTaskEdit}">
					<div class="task" v-for="l in tasks" :key="l.id">
						<span>
							<div class="fancycheckbox">
								<input @change="markAsDone" type="checkbox" :id="l.id" :checked="l.done" style="display: none;">
								<label :for="l.id" class="check">
									<svg width="18px" height="18px" viewBox="0 0 18 18">
										<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
										<polyline points="1 9 7 14 15 4"></polyline>
									</svg>
								</label>
							</div>
							<router-link :to="{ name: 'taskDetailView', params: { id: l.id } }" class="tasktext" :class="{ 'done': l.done}">
								<!-- Show any parent tasks to make it clear this task is a sub task of something -->
								<span class="parent-tasks" v-if="typeof l.related_tasks.parenttask !== 'undefined'">
									<template v-for="(pt, i) in l.related_tasks.parenttask">
										{{ pt.text }}<template v-if="(i + 1) < l.related_tasks.parenttask.length">,&nbsp;</template>
									</template>
									>
								</span>
								{{l.text}}
								<span class="tag" v-for="label in l.labels" :style="{'background': label.hex_color, 'color': label.textColor}" :key="label.id">
									<span>{{ label.title }}</span>
								</span>
								<img :src="gravatar(a)" :alt="a.username" v-for="(a, i) in l.assignees" class="avatar" :key="l.id + 'assignee' + a.id + i"/>
								<i v-if="l.dueDate > 0" :class="{'overdue': (l.dueDate <= new Date())}"> - Due on {{new Date(l.dueDate).toLocaleString()}}</i>
								<priority-label :priority="l.priority"/>
							</router-link>
						</span>
						<div @click="editTask(l.id)" class="icon settings">
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
	import message from '../../message'

	import ListService from '../../services/list'
	import TaskService from '../../services/task'
	import ListModel from '../../models/list'
	import EditTask from './edit-task'
	import TaskModel from '../../models/task'
	import PriorityLabel from './reusable/priorityLabel'
	import TaskCollectionService from '../../services/taskCollection'

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
			}
		},
		components: {
			PriorityLabel,
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
			initTasks(page) {
				this.taskEditTask = null
				this.isTaskEdit = false
				this.loadTasks(page)
			},
			addTask() {
				let task = new TaskModel({text: this.newTaskText, listID: this.$route.params.id})
				this.taskService.create(task)
					.then(r => {
						this.tasks.push(r)
						this.sortTasks()
						this.newTaskText = ''
						message.success({message: 'The task was successfully created.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			loadTasks(page) {
				const params = {sort_by: ['done', 'id'], order_by: ['asc', 'desc']}
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
						message.error(e, this)
					})
			},
			loadTasksForPage(e) {
				// The page parameter can be undefined, in the case where the user loads a new list from the side bar menu
				let page = e.page
				if (typeof e.page === 'undefined') {
					page = 1
				}
				this.initTasks(page)
			},
			markAsDone(e) {
				let updateFunc = () => {
					// We get the task, update the 'done' property and then push it to the api.
					let task = this.getTaskByID(e.target.id)
					task.done = e.target.checked
					this.taskService.update(task)
						.then(() => {
							this.sortTasks()
							message.success({message: 'The task was successfully ' + (task.done ? '' : 'un-') + 'marked as done.'}, this)
						})
						.catch(e => {
							message.error(e, this)
						})
				}

				if (e.target.checked) {
					setTimeout(updateFunc(), 300); // Delay it to show the animation when marking a task as done
				} else {
					updateFunc() // Don't delay it when un-marking it as it doesn't have an animation the other way around
				}
			},
			editTask(id) {
				// Find the selected task and set it to the current object
				let theTask = this.getTaskByID(id) // Somehow this does not work if we directly assign this to this.taskEditTask
				this.taskEditTask = theTask
				this.isTaskEdit = true
			},
			gravatar(user) {
				return 'https://www.gravatar.com/avatar/' + user.avatarUrl + '?s=27'
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
		}
	}
</script>