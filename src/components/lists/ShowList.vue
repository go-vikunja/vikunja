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

			<div class="box tasks" v-if="this.list.tasks.length > 0">
				<label class="task" v-for="l in list.tasks" v-bind:key="l.id" v-bind:for="l.id">
					<input @change="markAsDone" type="checkbox" v-bind:id="l.id" v-bind:checked="l.done">
					{{l.text}}
				</label>
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
                loading: false
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
                        for (const t in this.list.tasks) {
							if (this.list.tasks[t].id === response.data.id) {
								this.$set(this.list.tasks, t, response.data)
								break
							}
						}
                        this.handleSuccess({message: 'The task was successfully ' + (e.target.checked ? 'un-' :'') + 'marked as done.'})
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
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
			cursor: pointer;

			input[type="checkbox"] {
				vertical-align: middle;
			}
		}

		.task:last-child {
			border-bottom: none;
		}
	}
</style>