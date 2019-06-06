<template>
	<div class="loader-container" :class="{ 'is-loading': listService.loading}">
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
				<div class="tasks" v-if="this.list.tasks && this.list.tasks.length > 0" :class="{'short': isTaskEdit}">
					<div class="task" v-for="l in list.tasks" :key="l.id">
						<label :for="l.id">
							<div class="fancycheckbox">
								<input @change="markAsDone" type="checkbox" :id="l.id" :checked="l.done" style="display: none;">
								<label :for="l.id" class="check">
									<svg width="18px" height="18px" viewBox="0 0 18 18">
										<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
										<polyline points="1 9 7 14 15 4"></polyline>
									</svg>
								</label>
							</div>
							<span class="tasktext" :class="{ 'done': l.done}">
								{{l.text}}
								<span class="tag" v-for="label in l.labels" :style="{'background': label.hex_color, 'color': label.textColor}" :key="label.id">
									<span>{{ label.title }}</span>
								</span>
                                <img :src="gravatar(a)" :alt="a.username" v-for="a in l.assignees" class="avatar" :key="l.id + 'assignee' + a.id"/>
								<i v-if="l.dueDate > 0" :class="{'overdue': (l.dueDate <= new Date())}"> - Due on {{new Date(l.dueDate).toLocaleString()}}</i>
								<span v-if="l.priority >= priorities.HIGH" class="high-priority" :class="{'not-so-high': l.priority === priorities.HIGH}">
									<span class="icon">
										<icon icon="exclamation"/>
									</span>
									<template v-if="l.priority === priorities.HIGH">High</template>
									<template v-if="l.priority === priorities.URGENT">Urgent</template>
									<template v-if="l.priority === priorities.DO_NOW">DO NOW</template>
									<span class="icon" v-if="l.priority === priorities.DO_NOW">
										<icon icon="exclamation"/>
									</span>
								</span>
							</span>
						</label>
						<div @click="editTask(l.id)" class="icon settings">
							<icon icon="cog"/>
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
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import message from '../../message'

	import ListService from '../../services/list'
	import TaskService from '../../services/task'
	import ListModel from '../../models/list'
    import EditTask from './edit-task'
    import TaskModel from '../../models/task'
    import priorities from '../../models/priorities'

	export default {
		data() {
			return {
				listID: this.$route.params.id,
				listService: ListService,
				taskService: TaskService,
                list: {},
                isTaskEdit: false,
				taskEditTask: TaskModel,
				newTaskText: '',
                priorities: {},
			}
		},
        components: {
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
			}
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}
		},
		created() {
			this.listService = new ListService()
			this.taskService = new TaskService()
            this.priorities = priorities
            this.taskEditTask = null
			this.isTaskEdit = false
		},
		methods: {
			addTask() {
				let task = new TaskModel({text: this.newTaskText, listID: this.$route.params.id})
				this.taskService.create(task)
					.then(r => {
						this.list.addTaskToList(r)
						this.newTaskText = ''
						message.success({message: 'The task was successfully created.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			markAsDone(e) {
				let updateFunc = () => {
					// We get the task, update the 'done' property and then push it to the api.
					let task = this.list.getTaskByID(e.target.id)
					task.done = e.target.checked
					this.taskService.update(task)
						.then(() => {
                            this.list.sortTasks()
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
				let theTask = this.list.getTaskByID(id) // Somehow this does not work if we directly assign this to this.taskEditTask
                this.taskEditTask = theTask
				this.isTaskEdit = true
			},
			gravatar(user) {
				return 'https://www.gravatar.com/avatar/' + user.avatarUrl + '?s=27'
			},
		}
	}
</script>