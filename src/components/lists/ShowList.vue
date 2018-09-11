<template>
	<div>
		<div class="full-loader-wrapper" v-if="loading">
			<div class="half-circle-spinner">
				<div class="circle circle-1"></div>
				<div class="circle circle-2"></div>
			</div>
		</div>
		<div class="content">
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
							<input @change="markAsDone" type="checkbox" v-bind:id="l.id" v-bind:checked="l.done">
							{{l.text}}
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
										<input class="input" type="text" id="tasktext" placeholder="The task text is here..." v-model="taskEditTask.text">
									</div>
								</div>
								<div class="field">
									<label class="label" for="taskdescription">Description</label>
									<div class="control">
										<textarea class="textarea" placeholder="The tasks description goes here..." id="taskdescription" v-model="taskEditTask.description"></textarea>
									</div>
								</div>

								<div class="columns">
									<div class="column">
										<div class="field">
											<label class="label" for="taskduedate">Due Date</label>
											<div class="control">
												<input type="date" class="input" id="taskduedate" placeholder="The tasks due date is here..." v-model="taskEditTask.dueDate">
											</div>
										</div>
									</div>
									<div class="column">
										<div class="field">
											<label class="label" for="taskreminderdate">Reminder Date</label>
											<div class="control">
												<input type="date" class="input" id="taskreminderdate" placeholder="The tasks reminder date is here..." v-model="taskEditTask.reminderDate">
											</div>
										</div>
									</div>
								</div>

								<button type="submit" class="button is-success is-fullwidth">
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
            }
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
                // Find the slected task and set it to the current object
                for (const t in this.list.tasks) {
                    if (this.list.tasks[t].id === id) {
                        this.taskEditTask = this.list.tasks[t]
                        break
                    }
                }

				this.isTaskEdit = true
			},
			editTaskSubmit() {
                this.loading = true

				// Convert the date in a unix timestamp
				let duedate = (+ new Date(this.taskEditTask.dueDate)) / 1000
				let reminderdate = (+ new Date(this.taskEditTask.reminderDate)) / 1000
				this.taskEditTask.dueDate = duedate
				this.taskEditTask.reminderDate = reminderdate

                HTTP.post(`tasks/` + this.taskEditTask.id, this.taskEditTask, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.updateTaskByID(this.taskEditTask.id, response.data)
                        this.handleSuccess({message: 'The task was successfully updated.'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
			},
			updateTaskByID(id, updatedTask) {
                for (const t in this.list.tasks) {
                    if (this.list.tasks[t].id === id) {
                        this.$set(this.list.tasks, t, updatedTask)
                        break
                    }
                }
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

<style scoped lang="scss">
	.tasks {
		margin-top: 1rem;
		padding: 0;

		.task {
			display: block;
			padding: 0.5rem 1rem;
			border-bottom: 1px solid darken(#fff, 10%);

			label{
				width: 96%;
				display: inline-block;
				cursor: pointer;
			}

			input[type="checkbox"] {
				vertical-align: middle;
			}

			.settings{
				float: right;
				width: 4%;
				cursor: pointer;
			}
		}

		.task:last-child {
			border-bottom: none;
		}
	}

	.taskedit{
		min-height: calc(100% - 1rem);
		margin-top: 1rem;
	}
</style>