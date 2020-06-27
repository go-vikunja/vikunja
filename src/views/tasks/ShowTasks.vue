<template>
	<div>
		<h3 v-if="showAll">Current tasks</h3>
		<h3 v-else>Tasks from {{startDate.toLocaleDateString()}} until {{endDate.toLocaleDateString()}}</h3>
		<template v-if="!taskService.loading && (!hasUndoneTasks || !tasks)">
			<h3 class="nothing">Nothing to do - Have a nice day!</h3>
			<img src="/images/cool.svg" alt=""/>
		</template>
		<div class="spinner" :class="{ 'is-loading': taskService.loading}"></div>
		<div class="tasks" v-if="tasks && tasks.length > 0">
			<div class="task" v-for="t in tasks" :key="t.id">
				<single-task-in-list :the-task="t" @taskUpdated="updateTasks" :show-list="true"/>
			</div>
		</div>
	</div>
</template>
<script>
	import TaskService from '../../services/task'
	import SingleTaskInList from '../../components/tasks/partials/singleTaskInList'
	import {HAS_TASKS} from '../../store/mutation-types'
	import {mapState} from 'vuex'

	export default {
		name: 'ShowTasks',
		components: {
			SingleTaskInList,
		},
		data() {
			return {
				tasks: [],
				hasUndoneTasks: false,
				taskService: TaskService,
			}
		},
		props: {
			startDate: Date,
			endDate: Date,
			showAll: Boolean,
		},
		created() {
			this.taskService = new TaskService()
			this.loadPendingTasks()
		},
		watch: {
			'$route': 'loadPendingTasks',
		},
		computed: mapState({
			userAuthenticated: state => state.auth.authenticated,
		}),
		methods: {
			loadPendingTasks() {
				// Since this route is authentication only, users would get an error message if they access the page unauthenticated.
				// Since this component is mounted as the home page before unauthenticated users get redirected
				// to the login page, they will almost always see the error message.
				if (!this.userAuthenticated) {
					return
				}

				const params = {
					sort_by: ['due_date', 'id'],
					order_by: ['desc', 'desc'],
					filter_by: ['done'],
					filter_value: [false],
					filter_comparator: ['equals'],
					filter_concat: 'and',
					filter_include_nulls: true,
				}
				if (!this.showAll) {
					params.filter_by.push('start_date')
					params.filter_value.push(this.startDate)
					params.filter_comparator.push('greater')

					params.filter_by.push('end_date')
					params.filter_value.push(this.endDate)
					params.filter_comparator.push('less')

					params.filter_by.push('due_date')
					params.filter_value.push(this.endDate)
					params.filter_comparator.push('less')
				}

				this.taskService.getAll({}, params)
					.then(r => {
						if (r.length > 0) {
							for (const index in r) {
								if (r[index].done !== true) {
									this.hasUndoneTasks = true
								}
							}
						}
						this.$set(this, 'tasks', r.filter(t => !t.done))
						this.$store.commit(HAS_TASKS, r.length > 0)
					})
					.catch(e => {
						this.error(e, this)
					})
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
			updateTasks(updatedTask) {
				for (const t in this.tasks) {
					if (this.tasks[t].id === updatedTask.id) {
						this.$set(this.tasks, t, updatedTask)
						break
					}
				}
				this.sortTasks()
			},
		},
	}
</script>

<style scoped>
	h3 {
		text-align: left;
	}

	h3.nothing {
		text-align: center;
		margin-top: 3em;
	}

	img {
		margin-top: 2em;
	}

	.spinner.is-loading:after {
		margin-left: calc(40% - 1em);
	}
</style>