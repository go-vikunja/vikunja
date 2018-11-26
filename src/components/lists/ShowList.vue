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
					<input class="input" v-bind:class="{ 'disabled': loading}" v-model="newTask" type="text" placeholder="Add a new task...">
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
							<form  @submit.prevent="editTaskSubmit()">
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
                newTask: '',
                error: '',
                loading: false,
				isTaskEdit: false,
				taskEditTask: {},
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
                this.loading = true

                HTTP.get(`lists/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.loading = false

						// Make date objects from timestamps
                        for (const t in response.data.tasks) {
                            let dueDate = new Date(response.data.tasks[t].dueDate * 1000)
							if (dueDate === 0) {
								response.data.tasks[t].dueDate = null
							} else {
								response.data.tasks[t].dueDate = dueDate
							}

							for (const rd in response.data.tasks[t].reminderDates) {
								response.data.tasks[t].reminderDates[rd] = new Date(response.data.tasks[t].reminderDates[rd] * 1000)
							}
                        }

                        // This adds a new elemednt "list" to our object which contains all lists
                        this.$set(this, 'list', response.data)
                        if (this.list.tasks === null) {
                            this.list.tasks = []
                        }
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            addTask() {
                this.loading = true

                HTTP.put(`lists/` + this.$route.params.id, {text: this.newTask}, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.list.tasks.push(response.data)
                        this.handleSuccess({message: 'The task was successfully created.'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })

                this.newTask = ''
            },
			markAsDone(e) {

                this.loading = true

                HTTP.post(`tasks/` + e.target.id, {done: e.target.checked}, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.updateTaskByID(parseInt(e.target.id), response.data)
                        this.handleSuccess({message: 'The task was successfully ' + (e.target.checked ? 'un-' :'') + 'marked as done.'})
                    })
                    .catch(e => {
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
				this.isTaskEdit = true
			},
			editTaskSubmit() {
                this.loading = true

				// Convert the date in a unix timestamp
				let duedate = (+ new Date(this.taskEditTask.dueDate)) / 1000
				this.taskEditTask.dueDate = duedate

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
						this.$set(this, 'taskEditTask', response.data)
                        this.handleSuccess({message: 'The task was successfully updated.'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
			},
			updateTaskByID(id, updatedTask) {
                for (const t in this.list.tasks) {
                    if (this.list.tasks[t].id === id) {
						//updatedTask.reminderDates = this.makeJSReminderDatesAfterUpdate(updatedTask.reminderDates)
                        this.$set(this.list.tasks, t, updatedTask)
                        break
                    }
                }
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
                this.loading = false
                message.error(e, this)
            },
            handleSuccess(e) {
                this.loading = false
                message.success(e, this)
            }
        }
    }
</script>