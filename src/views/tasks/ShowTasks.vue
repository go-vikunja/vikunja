<template>
	<div class="is-max-width-desktop show-tasks">
		<fancycheckbox
			@change="loadPendingTasks"
			class="is-pulled-right"
			v-if="!showAll"
			v-model="showNulls"
		>
			Show tasks without dates
		</fancycheckbox>
		<h3 v-if="showAll">Current tasks</h3>
		<h3 v-else>
			Tasks from
			<flat-pickr
				:class="{ 'disabled': taskService.loading}"
				:config="flatPickerConfig"
				:disabled="taskService.loading"
				@on-close="loadPendingTasks"
				class="input"
				v-model="cStartDate"
			/>
			until
			<flat-pickr
				:class="{ 'disabled': taskService.loading}"
				:config="flatPickerConfig"
				:disabled="taskService.loading"
				@on-close="loadPendingTasks"
				class="input"
				v-model="cEndDate"
			/>
		</h3>
		<template v-if="!taskService.loading && (!hasUndoneTasks || !tasks || tasks.length === 0)">
			<h3 class="nothing">Nothing to do - Have a nice day!</h3>
			<img alt="" src="/images/cool.svg"/>
		</template>
		<div :class="{ 'is-loading': taskService.loading}" class="spinner"></div>
		<div class="tasks" v-if="tasks && tasks.length > 0">
			<div :key="t.id" class="task" v-for="t in tasks">
				<single-task-in-list :show-list="true" :the-task="t" @taskUpdated="updateTasks"/>
			</div>
		</div>
	</div>
</template>
<script>
import TaskService from '../../services/task'
import SingleTaskInList from '../../components/tasks/partials/singleTaskInList'
import {HAS_TASKS} from '@/store/mutation-types'
import {mapState} from 'vuex'

import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import Fancycheckbox from '../../components/input/fancycheckbox'

export default {
	name: 'ShowTasks',
	components: {
		Fancycheckbox,
		SingleTaskInList,
		flatPickr,
	},
	data() {
		return {
			tasks: [],
			hasUndoneTasks: false,
			taskService: TaskService,
			showNulls: true,

			cStartDate: null,
			cEndDate: null,

			flatPickerConfig: {
				altFormat: 'j M Y H:i',
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
			},
		}
	},
	props: {
		startDate: Date,
		endDate: Date,
		showAll: Boolean,
	},
	created() {
		this.taskService = new TaskService()
		this.cStartDate = this.startDate
		this.cEndDate = this.endDate
		this.loadPendingTasks()
	},
	watch: {
		'$route': 'loadPendingTasks',
		startDate(newVal) {
			this.cStartDate = newVal
		},
		endDate(newVal) {
			this.cEndDate = newVal
		},
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

			// Make sure all dates are date objects
			this.cStartDate = new Date(this.cStartDate)
			this.cEndDate = new Date(this.cEndDate)

			if (this.showAll) {
				this.setTitle('Current Tasks')
			} else {
				this.setTitle(`Tasks from ${this.cStartDate.toLocaleDateString()} until ${this.cEndDate.toLocaleDateString()}`)
			}

			const params = {
				sort_by: ['due_date', 'id'],
				order_by: ['desc', 'desc'],
				filter_by: ['done'],
				filter_value: [false],
				filter_comparator: ['equals'],
				filter_concat: 'and',
				filter_include_nulls: this.showNulls,
			}
			if (!this.showAll) {
				if (this.showNulls) {
					params.filter_by.push('start_date')
					params.filter_value.push(this.cStartDate)
					params.filter_comparator.push('greater')

					params.filter_by.push('end_date')
					params.filter_value.push(this.cEndDate)
					params.filter_comparator.push('less')
				}

				params.filter_by.push('due_date')
				params.filter_value.push(this.cEndDate)
				params.filter_comparator.push('less')

				params.filter_by.push('due_date')
				params.filter_value.push(this.cStartDate)
				params.filter_comparator.push('greater')
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
