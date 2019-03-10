<template>
	<div>
		<h3 v-if="showAll">Current tasks</h3>
		<h3 v-else>Tasks from {{startDate.toLocaleDateString()}} until {{endDate.toLocaleDateString()}}</h3>
		<template v-if="!taskService.loading && (!hasUndoneTasks || !tasks)">
			<h3 class="nothing">Nothing to to - Have a nice day!</h3>
			<img src="/images/cool.svg" alt=""/>
		</template>
		<div class="spinner" :class="{ 'is-loading': taskService.loading}"></div>
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
			</div>
		</div>
	</div>
</template>
<script>
	import router from '../../router'
	import message from '../../message'
	import TaskService from '../../services/task'
	import priorities from '../../models/priorities'

	export default {
		name: "ShowTasks",
		data() {
			return {
				tasks: [],
				hasUndoneTasks: false,
				taskService: TaskService,
				priorities: priorities,
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
				let params = {'sort': 'duedate'}
				if (!this.showAll) {
					params.startdate = Math.round(+ this.startDate / 1000)
					params.enddate = Math.round(+ this.endDate / 1000)
				}

				this.taskService.getAll({}, params)
					.then(r => {
						if (r.length > 0) {
							for (const index in r) {
								if (r[index].done !== true) {
									this.hasUndoneTasks = true
								}
							}
							r.sort(this.sortyByDeadline)
						}
						this.$set(this, 'tasks', r)
					})
					.catch(e => {
						message.error(e, this)
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