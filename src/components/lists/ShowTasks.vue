<template>
	<div>
		<h3 v-if="showAll">Current tasks</h3>
		<h3 v-else>Tasks from {{startDate.toLocaleDateString()}} until {{endDate.toLocaleDateString()}}</h3>
		<template v-if="!loading && (!hasUndoneTasks || !tasks)">
			<h3 class="nothing">Nothing to to - Have a nice day!</h3>
			<img src="/images/cool.svg" alt=""/>
		</template>
		<div class="spinner" :class="{ 'is-loading': loading}"></div>
		<div class="tasks" v-if="tasks && tasks.length > 0">
			<div @click="gotoList(l.listID)" class="task" v-for="l in tasks" :key="l.id" v-if="!l.done">
				<label :for="l.id">
					<div class="fancycheckbox">
						<input type="checkbox" :id="l.id" :checked="l.done" style="display: none;" disabled>
						<label  :for="l.id" class="check">
							<svg width="18px" height="18px" viewBox="0 0 18 18">
								<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
								<polyline points="1 9 7 14 15 4"></polyline>
							</svg>
						</label>
					</div>
					<span class="tasktext">
							{{l.text}}
							<i v-if="l.dueDate > 0" :class="{'overdue': (new Date(l.dueDate * 1000) <= new Date())}"> - Due on {{formatUnixDate(l.dueDate)}}</i>
						</span>
				</label>
			</div>
		</div>
	</div>
</template>
<script>
	import router from '../../router'
	import {HTTP} from '../../http-common'
	import message from '../../message'
	import TaskService from '../../services/task'

	export default {
		name: "ShowTasks",
		data() {
			return {
				loading: true,
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
		methods: {
			loadPendingTasks() {
				// We can't really make this code work until 0.6 is released which will make this exact thing a lot easier.
				// Because the feature we need here (specifying sort order and start/end date via query parameters) is already in master, we'll just wait and use the legacy method until then.
				/*
				let taskDummy = new TaskModel() // Used to specify options for the request
				this.taskService.getAll(taskDummy)
					.then(r => {
						this.tasks = r
					})
					.catch(e => {
						message.error(e, this)
					})*/
				const cancel = message.setLoading(this)

				let url = `tasks/all/duedate`
				if (!this.showAll) {
					url += `/` + Math.round(+ this.startDate / 1000) + `/` + Math.round(+ this.endDate / 1000)
				}

				HTTP.get(url, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
					.then(response => {
						// Filter all done tasks
						if (response.data !== null) {
							for (const index in response.data) {
								if (response.data[index].done !== true) {
									this.hasUndoneTasks = true
								}
							}
							response.data.sort(this.sortyByDeadline)
						}
						this.$set(this, 'tasks', response.data)
						cancel()
					})
					.catch(e => {
						cancel()
						this.handleError(e)
					})
			},
			formatUnixDate(dateUnix) {
				return (new Date(dateUnix * 1000)).toLocaleString()
			},
			sortyByDeadline(a, b) {
				return ((a.dueDate > b.dueDate) ? -1 : ((a.dueDate < b.dueDate) ? 1 : 0));
			},
			gotoList(lid) {
				router.push({name: 'showList', params: {id: lid}})
			},
			handleError(e) {
				message.error(e, this)
			}
		},
	}
</script>

<style scoped>
	h3{
		text-align: left;
	}

	h3.nothing{
		text-align: center;
		margin-top: 3em;
	}

	img{
		margin-top: 2em;
	}

	.spinner.is-loading:after {
		margin-left: calc(40% - 1em);
	}
</style>