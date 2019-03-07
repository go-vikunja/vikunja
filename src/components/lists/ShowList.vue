<template>
	<div class="loader-container" :class="{ 'is-loading': listService.loading}">
		<div class="content">
			<router-link :to="{ name: 'editList', params: { id: list.id } }" class="icon settings is-medium">
				<icon icon="cog" size="2x"/>
			</router-link>
			<h1>{{ list.title }}</h1>
		</div>
		<form @submit.prevent="addTask()">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" :class="{ 'is-loading': taskService.loading}">
					<input v-focus class="input" :class="{ 'disabled': taskService.loading}" v-model="newTask.text" type="text" placeholder="Add a new task...">
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
							<form @submit.prevent="editTaskSubmit()">
								<div class="field">
									<label class="label" for="tasktext">Task Text</label>
									<div class="control">
										<input v-focus :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="input" type="text" id="tasktext" placeholder="The task text is here..." v-model="taskEditTask.text">
									</div>
								</div>
								<div class="field">
									<label class="label" for="taskdescription">Description</label>
									<div class="control">
										<textarea :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="textarea" placeholder="The tasks description goes here..." id="taskdescription" v-model="taskEditTask.description"></textarea>
									</div>
								</div>

								<b>Reminder Dates</b>
								<div class="reminder-input" :class="{ 'overdue': (r < nowUnix && index !== (taskEditTask.reminderDates.length - 1))}" v-for="(r, index) in taskEditTask.reminderDates" :key="index">
									<flat-pickr
										:class="{ 'disabled': taskService.loading}"
										:disabled="taskService.loading"
										:v-model="taskEditTask.reminderDates"
										:config="flatPickerConfig"
										:id="'taskreminderdate' + index"
										:value="r"
										:data-index="index"
										placeholder="Add a new reminder...">
									</flat-pickr>
									<a v-if="index !== (taskEditTask.reminderDates.length - 1)" @click="removeReminderByIndex(index)"><icon icon="times"></icon></a>
								</div>

								<div class="field">
									<label class="label" for="taskduedate">Due Date</label>
									<div class="control">
										<flat-pickr
											:class="{ 'disabled': taskService.loading}"
											class="input"
											:disabled="taskService.loading"
											v-model="taskEditTask.dueDate"
											:config="flatPickerConfig"
											id="taskduedate"
											placeholder="The tasks due date is here...">
										</flat-pickr>
									</div>
								</div>

								<div class="field">
									<label class="label" for="">Duration</label>
									<div class="control columns">
										<div class="column">
											<flat-pickr
													:class="{ 'disabled': taskService.loading}"
													class="input"
													:disabled="taskService.loading"
													v-model="taskEditTask.startDate"
													:config="flatPickerConfig"
													id="taskduedate"
													placeholder="Start date">
											</flat-pickr>
										</div>
										<div class="column">
											<flat-pickr
													:class="{ 'disabled': taskService.loading}"
													class="input"
													:disabled="taskService.loading"
													v-model="taskEditTask.endDate"
													:config="flatPickerConfig"
													id="taskduedate"
													placeholder="End date">
											</flat-pickr>
										</div>
									</div>
								</div>

								<div class="field">
									<label class="label" for="">Repeat after</label>
									<div class="control repeat-after-input columns">
										<div class="column">
											<input class="input" placeholder="Specify an amount..." v-model="taskEditTask.repeatAfter.amount"/>
										</div>
										<div class="column is-3">
											<div class="select">
												<select v-model="taskEditTask.repeatAfter.type">
													<option value="hours">Hours</option>
													<option value="days">Days</option>
													<option value="weeks">Weeks</option>
													<option value="months">Months</option>
													<option value="years">Years</option>
												</select>
											</div>
										</div>
									</div>
								</div>

								<div class="field">
									<label class="label" for="">Priority</label>
									<div class="control priority-select">
										<div class="select">
											<select v-model="taskEditTask.priority">
												<option :value="priorities.UNSET">Unset</option>
												<option :value="priorities.LOW">Low</option>
												<option :value="priorities.MEDIUM">Medium</option>
												<option :value="priorities.HIGH">High</option>
												<option :value="priorities.URGENT">Urgent</option>
												<option :value="priorities.DO_NOW">DO NOW</option>
											</select>
										</div>
									</div>
								</div>

								<div class="field">
									<label class="label" for="">Assignees</label>
									<ul class="assingees">
										<li v-for="(a, index) in taskEditTask.assignees" :key="a.id">
											{{a.username}}
											<a @click="deleteAssigneeByIndex(index)"><icon icon="times"/></a>
										</li>
									</ul>
								</div>

								<div class="field has-addons">
									<div class="control is-expanded">
										<multiselect
												v-model="newAssignee"
												:options="foundUsers"
												:multiple="false"
												:searchable="true"
												:loading="userService.loading"
												:internal-search="true"
												@search-change="findUser"
												placeholder="Type to search"
												label="username"
												track-by="id">
											<template slot="clear" slot-scope="props">
												<div class="multiselect__clear" v-if="newAssignee !== null && newAssignee.id !== 0" @mousedown.prevent.stop="clearAllFoundUsers(props.search)"></div>
											</template>
											<span slot="noResult">Oops! No user found. Consider changing the search query.</span>
										</multiselect>
									</div>
									<div class="control">
										<a @click="addAssignee" class="button is-primary fullheight">
											<span class="icon is-small">
												<icon icon="plus"/>
											</span>
										</a>
									</div>
								</div>

								<div class="field">
									<label class="label">Labels</label>
									<div class="control">
										<multiselect
												:multiple="true"
												:close-on-select="false"
												:clear-on-select="true"
												:options-limit="300"
												:hide-selected="true"
												v-model="taskEditTask.labels"
												:options="foundLabels"
												:searchable="true"
												:loading="labelService.loading || labelTaskService.loading"
												:internal-search="true"
												@search-change="findLabel"
												@select="addLabel"
												placeholder="Type to search"
												label="title"
												track-by="id"
												:taggable="true"
												@tag="createAndAddLabel"
												tag-placeholder="Add this as new label"
										>
											<template slot="tag" slot-scope="{ option, remove }">
												<span class="tag" :style="{'background': option.hex_color, 'color': option.textColor}">
													<span>{{ option.title }}</span>
													<a class="delete is-small" @click="removeLabel(option)"></a>
												</span>
											</template>
											<template slot="clear" slot-scope="props">
												<div class="multiselect__clear" v-if="taskEditTask.labels.length" @mousedown.prevent.stop="clearAllLabels(props.search)"></div>
											</template>
										</multiselect>
									</div>
								</div>

								<div class="field">
									<label class="label" for="subtasks">Subtasks</label>
									<div class="tasks noborder" v-if="taskEditTask.subtasks && taskEditTask.subtasks.length > 0">
										<div class="task" v-for="s in taskEditTask.subtasks" :key="s.id">
											<label :for="s.id">
												<div class="fancycheckbox">
													<input @change="markAsDone" type="checkbox" :id="s.id" :checked="s.done" style="display: none;">
													<label  :for="s.id" class="check">
														<svg width="18px" height="18px" viewBox="0 0 18 18">
															<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
															<polyline points="1 9 7 14 15 4"></polyline>
														</svg>
													</label>
												</div>
												<span class="tasktext" :class="{ 'done': s.done}">
													{{s.text}}
												</span>
											</label>
										</div>
									</div>
								</div>
								<div class="field has-addons">
									<div class="control is-expanded">
										<input @keyup.enter="addSubtask()" :class="{ 'disabled': taskService.loading}" :disabled="taskService.loading" class="input" type="text" id="tasktext" placeholder="New subtask" v-model="newTask.text"/>
									</div>
									<div class="control">
										<a class="button is-primary" @click="addSubtask()"><icon icon="plus"></icon></a>
									</div>
								</div>

								<button type="submit" class="button is-success is-fullwidth" :class="{ 'is-loading': taskService.loading}">
									Save
								</button>

							</form>
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
	import flatPickr from 'vue-flatpickr-component'
	import 'flatpickr/dist/flatpickr.css'
	import multiselect from 'vue-multiselect'
	import {differenceWith} from 'lodash'

	import ListService from '../../services/list'
	import TaskService from '../../services/task'
	import TaskModel from '../../models/task'
	import ListModel from '../../models/list'
	import UserModel from '../../models/user'
	import UserService from '../../services/user'
	import priorities from '../../models/priorities'
	import LabelTaskService from '../../services/labelTask'
	import LabelService from '../../services/label'
	import LabelTaskModel from '../../models/labelTask'
	import LabelModel from '../../models/label'

	export default {
		data() {
			return {
				listID: this.$route.params.id,
				listService: ListService,
				taskService: TaskService,

				priorities: priorities,
				list: {},
				newTask: TaskModel,
				isTaskEdit: false,
				taskEditTask: {
					subtasks: [],
				},
				lastReminder: 0,
				nowUnix: new Date(),
				flatPickerConfig:{
					altFormat: 'j M Y H:i',
					altInput: true,
					dateFormat: 'Y-m-d H:i',
					enableTime: true,
					onOpen: this.updateLastReminderDate,
					onClose: this.addReminderDate,
				},

				newAssignee: UserModel,
				userService: UserService,
				foundUsers: [],

				labelService: LabelService,
				labelTaskService: LabelTaskService,
				foundLabels: [],
				labelTimeout: null,
			}
		},
		components: {
			flatPickr,
			multiselect,
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
			this.newTask = new TaskModel()
			this.userService = new UserService()
			this.newAssignee = new UserModel()
			this.labelService = new LabelService()
			this.labelTaskService = new LabelTaskService()
			this.loadList()
		},
		watch: {
			// call again the method if the route changes
			'$route': 'loadList'
		},
		methods: {
			loadList() {
				this.isTaskEdit = false
				// We create an extra list object instead of creating it in this.list because that would trigger a ui update which would result in bad ux.
				let list = new ListModel({id: this.$route.params.id})
				this.listService.get(list)
					.then(r => {
						this.$set(this, 'list', r)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			addTask() {
				this.newTask.listID = this.$route.params.id
				this.taskService.create(this.newTask)
					.then(r => {
						this.list.addTaskToList(r)
						message.success({message: 'The task was successfully created.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})

				this.newTask = {}
			},
			markAsDone(e) {
				let updateFunc = () => {
					// We get the task, update the 'done' property and then push it to the api.
					let task = this.list.getTaskByID(e.target.id)
					task.done = e.target.checked
					this.taskService.update(task)
						.then(r => {
							this.updateTaskInList(r)
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
			editTaskSubmit() {
				this.taskService.update(this.taskEditTask)
					.then(r => {
						this.updateTaskInList(r)
						this.$set(this, 'taskEditTask', r)
						message.success({message: 'The task was successfully updated.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			addSubtask() {
				this.newTask.parentTaskID = this.taskEditTask.id
				this.addTask()
			},
			updateTaskInList(task) {
				// We need to do the update here in the component, because there is no way of notifiying vue of the change otherwise.
				for (const t in this.list.tasks) {
					if (this.list.tasks[t].id === task.id) {
						this.$set(this.list.tasks, t, task)
						break
					}

					if (this.list.tasks[t].id === task.parentTaskID) {
						for (const s in this.list.tasks[t].subtasks) {
							if (this.list.tasks[t].subtasks[s].id === task.id) {
								this.$set(this.list.tasks[t].subtasks, s, task)
								break
							}
						}
					}
				}
				this.list.sortTasks()
			},
			updateLastReminderDate(selectedDates) {
				this.lastReminder = +new Date(selectedDates[0])
			},
			addReminderDate(selectedDates, dateStr, instance) {
				let newDate = +new Date(selectedDates[0])

				// Don't update if nothing changed
				if (newDate === this.lastReminder) {
					return
				}

				let index = parseInt(instance.input.dataset.index)
				this.taskEditTask.reminderDates[index] = newDate

				let lastIndex = this.taskEditTask.reminderDates.length - 1
				// put a new null at the end if we changed something
				if (lastIndex === index && !isNaN(newDate)) {
					this.taskEditTask.reminderDates.push(null)
				}
			},
			removeReminderByIndex(index) {
				this.taskEditTask.reminderDates.splice(index, 1)
				// Reset the last to 0 to have the "add reminder" button
				this.taskEditTask.reminderDates[this.taskEditTask.reminderDates.length - 1] = null
			},
			addAssignee() {
				this.taskEditTask.assignees.push(this.newAssignee)
			},
			deleteAssigneeByIndex(index) {
				this.taskEditTask.assignees.splice(index, 1)
			},
			findUser(query) {
				if(query === '') {
					this.clearAllFoundUsers()
					return
				}

				this.userService.getAll({}, {s: query})
					.then(response => {
						// Filter the results to not include users who are already assigned
						this.$set(this, 'foundUsers', differenceWith(response, this.taskEditTask.assignees, (first, second) => {
							return first.id === second.id
						}))
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			clearAllFoundUsers () {
				this.$set(this, 'foundUsers', [])
			},
			findLabel(query) {
				if(query === '') {
					this.clearAllLabels()
					return
				}

				if(this.labelTimeout !== null) {
					clearTimeout(this.labelTimeout)
				}

				// Delay the search 300ms to not send a request on every keystroke
				this.labelTimeout = setTimeout(() => {
					this.labelService.getAll({}, {s: query})
						.then(response => {
							this.$set(this, 'foundLabels', differenceWith(response, this.taskEditTask.labels, (first, second) => {
								return first.id === second.id
							}))
							this.labelTimeout = null
						})
						.catch(e => {
							message.error(e, this)
						})
				}, 300)
			},
			clearAllLabels () {
				this.$set(this, 'foundLabels', [])
			},
			addLabel(label) {
				let labelTask = new LabelTaskModel({taskID: this.taskEditTask.id, label_id: label.id})
				this.labelTaskService.create(labelTask)
					.then(() => {
						message.success({message: 'The label was successfully added.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			removeLabel(label) {
				let labelTask = new LabelTaskModel({taskID: this.taskEditTask.id, label_id: label.id})
				this.labelTaskService.delete(labelTask)
					.then(() => {
						// Remove the label from the list
						for (const l in this.taskEditTask.labels) {
							if (this.taskEditTask.labels[l].id === label.id) {
								this.taskEditTask.labels.splice(l, 1)
							}
						}
						message.success({message: 'The label was successfully removed.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			createAndAddLabel(title) {
				let newLabel = new LabelModel({title: title})
				this.labelService.create(newLabel)
					.then(r => {
						this.addLabel(r)
						this.taskEditTask.labels.push(r)
					})
					.catch(e => {
						message.error(e, this)
					})
			}
		}
	}
</script>