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
			<div @click="gotoTask(l)" class="task" v-for="l in undoneTasks" :key="l.id">
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
						<i v-if="l.dueDate > 0" :class="{'overdue': l.dueDate <= new Date()}" v-tooltip="formatDate(l.dueDate)"> - Due {{formatDateSince(l.dueDate)}}</i>
						<priority-label :priority="l.priority"/>
					</span>
				</label>
			</div>
		</div>
	</div>
</template>
<script>
	import router from '../../router'
	import TaskService from '../../services/task'
	import PriorityLabel from './reusable/priorityLabel'

	export default {
		name: 'ShowTasks',
		components: {
			PriorityLabel
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
		computed: {
			undoneTasks: function () {
				return this.tasks.filter(t => !t.done)
			}
		},
		methods: {
			loadPendingTasks() {
				let params = {sort_by: ['due_date_unix', 'id'], order_by: ['desc', 'desc']}
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
						}
						this.$set(this, 'tasks', r)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			gotoTask(task) {
				router.push({name: 'taskDetailView', params: {id: task.id}})
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