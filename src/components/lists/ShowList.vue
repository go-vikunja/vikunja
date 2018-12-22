<template>
	<div class="loader-container" v-bind:class="{ 'is-loading': loading}">
		<div class="content">
			<router-link :to="{ name: 'editList', params: { id: list.id } }" class="icon settings is-medium">
				<icon icon="cog" size="2x"/>
			</router-link>
			<h1>{{ list.title }}</h1>
		</div>
		<form @submit.prevent="addTask()">
			<div class="field is-grouped">
				<p class="control has-icons-left is-expanded" v-bind:class="{ 'is-loading': loading}">
					<input class="input" v-bind:class="{ 'disabled': loading}" v-model="newTask.text" type="text" placeholder="Add a new task...">
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
				<div class="box tasks" v-if="this.list.tasks && this.list.tasks.length > 0">
					<div class="task" v-for="l in list.tasks" v-bind:key="l.id">
						<label v-bind:for="l.id">
							<div class="fancycheckbox">
								<input @change="markAsDone" type="checkbox" v-bind:id="l.id" v-bind:checked="l.done" style="display: none;">
								<label  v-bind:for="l.id" class="check">
									<svg width="18px" height="18px" viewBox="0 0 18 18">
										<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
										<polyline points="1 9 7 14 15 4"></polyline>
									</svg>
								</label>
							</div>
							<span class="tasktext">
								{{l.text}}
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
										<input :class="{ 'disabled': loading}" :disabled="loading" class="input" type="text" id="tasktext" placeholder="The task text is here..." v-model="taskEditTask.text">
									</div>
								</div>
								<div class="field">
									<label class="label" for="taskdescription">Description</label>
									<div class="control">
										<textarea :class="{ 'disabled': loading}" :disabled="loading" class="textarea" placeholder="The tasks description goes here..." id="taskdescription" v-model="taskEditTask.description"></textarea>
									</div>
								</div>

								<b>Reminder Dates</b>
								<div class="reminder-input" :class="{ 'overdue': (r < nowUnix && index !== (taskEditTask.reminderDates.length - 1))}" v-for="(r, index) in taskEditTask.reminderDates" v-bind:key="index">
									<flat-pickr
										:class="{ 'disabled': loading}"
										:disabled="loading"
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
											:class="{ 'disabled': loading}"
											class="input"
											:disabled="loading"
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
													:class="{ 'disabled': loading}"
													class="input"
													:disabled="loading"
													v-model="taskEditTask.startDate"
													:config="flatPickerConfig"
													id="taskduedate"
													placeholder="Start date">
											</flat-pickr>
										</div>
										<div class="column">
											<flat-pickr
													:class="{ 'disabled': loading}"
													class="input"
													:disabled="loading"
													v-model="taskEditTask.endDate"
													:config="flatPickerConfig"
													id="taskduedate"
													placeholder="Start date">
											</flat-pickr>
										</div>
									</div>
								</div>

								<div class="field">
									<label class="label" for="">Repeat after</label>
									<div class="control repeat-after-input columns">
										<div class="column">
											<input class="input" placeholder="Specify an amount..." v-model="repeatAfter.amount"/>
										</div>
										<div class="column">
											<div class="select">
												<select v-model="repeatAfter.type">
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
									<label class="label" for="subtasks">Subtasks</label>
									<div class="control subtasks">

										<div class="tasks noborder" v-if="taskEditTask.subtasks && taskEditTask.subtasks.length > 0">
											<div class="task" v-for="s in taskEditTask.subtasks" v-bind:key="s.id">
												<label v-bind:for="s.id">
													<div class="fancycheckbox">
														<input @change="markAsDone" type="checkbox" v-bind:id="s.id" v-bind:checked="s.done" style="display: none;">
														<label  v-bind:for="s.id" class="check">
															<svg width="18px" height="18px" viewBox="0 0 18 18">
																<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
																<polyline points="1 9 7 14 15 4"></polyline>
															</svg>
														</label>
													</div>
													<span class="tasktext">
														{{s.text}}
													</span>
												</label>
											</div>
										</div>

										<input :class="{ 'disabled': loading}" :disabled="loading" class="input" type="text" id="tasktext" placeholder="New subtask" v-model="newTask.text"/>
										<a class="button" @click="addSubtask()"><icon icon="plus"></icon></a>

									</div>
								</div>

								<button type="submit" class="button is-success is-fullwidth" :class="{ 'is-loading': loading}">
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
    import {HTTP} from '../../http-common'
    import message from '../../message'
    import flatPickr from 'vue-flatpickr-component';
    import 'flatpickr/dist/flatpickr.css';

    export default {
        data() {
            return {
                listID: this.$route.params.id,
                list: {},
                newTask: {text: ''},
                error: '',
                loading: false,
				isTaskEdit: false,
				taskEditTask: {
					subtasks: [],
				},
				lastReminder: 0,
				nowUnix: new Date(),
				repeatAfter: {type: 'days', amount: null},
                flatPickerConfig:{
                    altFormat: 'j M Y H:i',
                    altInput: true,
                    dateFormat: 'Y-m-d H:i',
					enableTime: true,
					onOpen: this.updateLastReminderDate,
					onClose: this.addReminderDate,
				},
            }
        },
		components: {
			flatPickr
		},
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        created() {
            this.loadList()
        },
        watch: {
            // call again the method if the route changes
            '$route': 'loadList'
        },
        methods: {
            loadList() {
                this.isTaskEdit = false
				const cancel = message.setLoading(this)

                HTTP.get(`lists/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        for (const t in response.data.tasks) {
							response.data.tasks[t] = this.fixStuffComingFromAPI(response.data.tasks[t])
                        }

                        // This adds a new elemednt "list" to our object which contains all lists
                        this.$set(this, 'list', response.data)
                        if (this.list.tasks === null) {
                            this.list.tasks = []
                        }
						cancel() // cancel the timer
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
            },
            addTask() {
				const cancel = message.setLoading(this)

                HTTP.put(`lists/` + this.$route.params.id, this.newTask, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
						this.addTaskToList(response.data)
                        this.handleSuccess({message: 'The task was successfully created.'})
						cancel() // cancel the timer
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })

                this.newTask = {}
            },
			addTaskToList(task) {
				// If it's a subtask, add it to its parent, otherwise append it to the list of tasks
				if (task.parentTaskID === 0) {
					this.list.tasks.push(task)
				} else {
					for (const t in this.list.tasks) {
						if (this.list.tasks[t].id === task.parentTaskID) {
							this.list.tasks[t].subtasks.push(task)
							break
						}
					}
				}

				// Update the current edit task if needed
				if (task.ParentTask === this.taskEditTask.id) {
					this.taskEditTask.subtasks.push(task)
				}

			},
			markAsDone(e) {
				const cancel = message.setLoading(this)

                HTTP.post(`tasks/` + e.target.id, {done: e.target.checked}, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.updateTaskByID(parseInt(e.target.id), response.data)
                        this.handleSuccess({message: 'The task was successfully ' + (e.target.checked ? 'un-' :'') + 'marked as done.'})
						cancel() // To not set the spinner to loading when the request is made in less than 100ms, would lead to loading infinitly.
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
			},
			editTask(id) {
                // Find the selected task and set it to the current object
                for (const t in this.list.tasks) {
                    if (this.list.tasks[t].id === id) {
                        this.taskEditTask = this.list.tasks[t]
                        break
                    }
                }

                if (this.taskEditTask.reminderDates === null) {
					this.taskEditTask.reminderDates = []
				}
				this.taskEditTask.reminderDates = this.removeNullsFromArray(this.taskEditTask.reminderDates)
                this.taskEditTask.reminderDates.push(null)

				// Re-convert the the amount from seconds to be used with our form
				let repeatAfterHours = (this.taskEditTask.repeatAfter / 60) / 60
				// if its dividable by 24, its something with days
				if (repeatAfterHours % 24 === 0) {
					let repeatAfterDays = repeatAfterHours / 24
					if (repeatAfterDays % 7 === 0) {
						this.repeatAfter.type = 'weeks'
						this.repeatAfter.amount = repeatAfterDays / 7
					} else if (repeatAfterDays % 30 === 0) {
						this.repeatAfter.type = 'months'
						this.repeatAfter.amount = repeatAfterDays / 30
					} else if (repeatAfterDays % 365 === 0) {
						this.repeatAfter.type = 'years'
						this.repeatAfter.amount = repeatAfterDays / 365
					} else {
						this.repeatAfter.type = 'days'
						this.repeatAfter.amount = repeatAfterDays
					}
				} else {
					// otherwise hours
					this.repeatAfter.type = 'hours'
					this.repeatAfter.amount = repeatAfterHours
				}

				if(this.taskEditTask.subtasks === null) {
					this.taskEditTask.subtasks = [];
				}

				this.isTaskEdit = true
			},
			editTaskSubmit() {
				const cancel = message.setLoading(this)

				// Convert the date in a unix timestamp
				this.taskEditTask.dueDate = (+ new Date(this.taskEditTask.dueDate)) / 1000
				this.taskEditTask.startDate = (+ new Date(this.taskEditTask.startDate)) / 1000
				this.taskEditTask.endDate = (+ new Date(this.taskEditTask.endDate)) / 1000


				// remove all nulls
				this.taskEditTask.reminderDates = this.removeNullsFromArray(this.taskEditTask.reminderDates)
				// Make normal timestamps from js timestamps
				for (const t in this.taskEditTask.reminderDates) {
					this.taskEditTask.reminderDates[t] = Math.round(this.taskEditTask.reminderDates[t] / 1000)
				}

				// Make the repeating amount to seconds
				let repeatAfterSeconds = 0
				if (this.repeatAfter.amount !== null || this.repeatAfter.amount !== 0) {
					switch (this.repeatAfter.type) {
						case 'hours':
							repeatAfterSeconds = this.repeatAfter.amount * 60 * 60
							break;
						case 'days':
							repeatAfterSeconds = this.repeatAfter.amount * 60 * 60 * 24
							break;
						case 'weeks':
							repeatAfterSeconds = this.repeatAfter.amount * 60 * 60 * 24 * 7
							break;
						case 'months':
							repeatAfterSeconds = this.repeatAfter.amount * 60 * 60 * 24 * 30
							break;
						case 'years':
							repeatAfterSeconds = this.repeatAfter.amount * 60 * 60 * 24 * 365
							break;
					}
				}
				this.taskEditTask.repeatAfter = repeatAfterSeconds

                HTTP.post(`tasks/` + this.taskEditTask.id, this.taskEditTask, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        response.data.dueDate = new Date(response.data.dueDate * 1000)
						response.data.reminderDates = this.makeJSReminderDatesAfterUpdate(response.data.reminderDates)

						// Update the task in the list
                        this.updateTaskByID(this.taskEditTask.id, response.data)

						// Also update the current taskedit object so the ui changes
						this.$set(this, 'taskEditTask', this.fixStuffComingFromAPI(response.data))
                        this.handleSuccess({message: 'The task was successfully updated.'})
						cancel() // cancel the timers
                    })
                    .catch(e => {
                        cancel()
						this.handleError(e)
                    })
			},
			addSubtask() {
				this.newTask.parentTaskID = this.taskEditTask.id
				this.addTask()
			},
			updateTaskByID(id, updatedTask) {
                for (const t in this.list.tasks) {
                    if (this.list.tasks[t].id === id) {
                        this.$set(this.list.tasks, t, this.fixStuffComingFromAPI(updatedTask))
                        break
                    }

					if (this.list.tasks[t].id === updatedTask.parentTaskID) {
						for (const s in this.list.tasks[t].subtasks) {
							if (this.list.tasks[t].subtasks[s].id === updatedTask.id) {
								this.$set(this.list.tasks[t].subtasks, s, updatedTask)
								break
							}
						}
					}
                }
			},
			fixStuffComingFromAPI(task) {
				// Make date objects from timestamps
				task.dueDate = this.parseDateIfNessecary(task.dueDate)
				task.startDate = this.parseDateIfNessecary(task.startDate)
				task.endDate = this.parseDateIfNessecary(task.endDate)

				for (const rd in task.reminderDates) {
					task.reminderDates[rd] = this.parseDateIfNessecary(task.reminderDates[rd])
				}

				// Make subtasks into empty array if null
				if (task.subtasks === null) {
					task.subtasks = []
				}
				return task
			},
			parseDateIfNessecary(dateUnix) {
				let dateobj = (+new Date(dateUnix * 1000))
				if (dateobj === 0 || dateUnix === 0) {
					dateUnix = null
				} else {
					dateUnix = dateobj
				}
				return dateUnix
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
			removeNullsFromArray(array) {
				for (const index in array) {
					if (array[index] === null) {
						array.splice(index, 1)
					}
				}
				return array
			},
			makeJSReminderDatesAfterUpdate(dates) {
				// Make js timestamps from normal timestamps
				for (const rd in dates) {
					dates[rd] = +new Date(dates[rd] * 1000)
				}

				if (dates == null) {
					dates = []
				}
				dates.push(null)
				return dates
			},
            handleError(e) {
                message.error(e, this)
            },
            handleSuccess(e) {
                message.success(e, this)
            }
        }
    }
</script>