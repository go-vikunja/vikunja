<template>
	<div class="is-max-width-desktop show-tasks">
		<fancycheckbox
			@change="setDate"
			class="is-pulled-right"
			v-if="!showAll"
			v-model="showNulls"
		>
			Show tasks without dates
		</fancycheckbox>
		<h3 v-if="showAll && tasks.length > 0">Current tasks</h3>
		<h3 v-else-if="!showAll" class="mb-2">
			Tasks from
			<flat-pickr
				:class="{ 'disabled': taskService.loading}"
				:config="flatPickerConfig"
				:disabled="taskService.loading"
				@on-close="setDate"
				class="input"
				v-model="cStartDate"
			/>
			until
			<flat-pickr
				:class="{ 'disabled': taskService.loading}"
				:config="flatPickerConfig"
				:disabled="taskService.loading"
				@on-close="setDate"
				class="input"
				v-model="cEndDate"
			/>
		</h3>
		<div v-if="!showAll" class="mb-4">
			<x-button type="secondary" @click="showTodaysTasks()" class="mr-2">Today</x-button>
			<x-button type="secondary" @click="setDatesToNextWeek()" class="mr-2">Next Week</x-button>
			<x-button type="secondary" @click="setDatesToNextMonth()">Next Month</x-button>
		</div>
		<template v-if="!taskService.loading && (!tasks || tasks.length === 0) && showNothingToDo">
			<h3 class="nothing">Nothing to do - Have a nice day!</h3>
			<img alt="" src="/images/cool.svg"/>
		</template>
		<div :class="{ 'is-loading': taskService.loading}" class="spinner"></div>

		<card :padding="false" class="has-overflow" :has-content="false" v-if="tasks && tasks.length > 0">
			<div class="tasks">
				<single-task-in-list
					:key="t.id"
					class="task"
					v-for="t in tasks"
					:show-list="true"
					:the-task="t"
					@taskUpdated="updateTasks"/>
			</div>
		</card>
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
			taskService: TaskService,
			showNulls: true,
			showOverdue: false,

			cStartDate: null,
			cEndDate: null,

			showNothingToDo: false,
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
	mounted() {
		setTimeout(() => this.showNothingToDo = true, 100)
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
		flatPickerConfig: state => ({
			altFormat: 'j M Y H:i',
			altInput: true,
			dateFormat: 'Y-m-d H:i',
			enableTime: true,
			time_24hr: true,
			locale: {
				firstDayOfWeek: state.auth.settings.weekStart,
			},
		})
	}),
	methods: {
		setDate() {
			this.$router.push({
				name: this.$route.name,
				query: {
					from: +new Date(this.cStartDate),
					to: +new Date(this.cEndDate),
					showOverdue: this.showOverdue,
					showNulls: this.showNulls,
				},
			})
		},
		loadPendingTasks() {
			// Since this route is authentication only, users would get an error message if they access the page unauthenticated.
			// Since this component is mounted as the home page before unauthenticated users get redirected
			// to the login page, they will almost always see the error message.
			if (!this.userAuthenticated) {
				return
			}

			// Make sure all dates are date objects
			if (typeof this.$route.query.from !== 'undefined' && typeof this.$route.query.to !== 'undefined') {
				this.cStartDate = new Date(Number(this.$route.query.from))
				this.cEndDate = new Date(Number(this.$route.query.to))
			} else {
				this.cStartDate = new Date(this.cStartDate)
				this.cEndDate = new Date(this.cEndDate)
			}
			this.showOverdue = this.$route.query.showOverdue
			this.showNulls = this.$route.query.showNulls

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

				if (!this.showOverdue) {
					params.filter_by.push('due_date')
					params.filter_value.push(this.cStartDate)
					params.filter_comparator.push('greater')
				}
			}

			this.taskService.getAll({}, params)
				.then(r => {

					// Sort all tasks to put those with a due date before the ones without a due date, the
					// soonest before the later ones.
					// We can't use the api sorting here because that sorts tasks with a due date after
					// ones without a due date.
					r.sort((a, b) => {
						return a.dueDate === null && b.dueDate === null ? -1 : 1
					})
					const tasks = r.filter(t => t.dueDate !== null).concat(r.filter(t => t.dueDate === null))

					this.$set(this, 'tasks', tasks)
					this.$store.commit(HAS_TASKS, r.length > 0)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		updateTasks(updatedTask) {
			for (const t in this.tasks) {
				if (this.tasks[t].id === updatedTask.id) {
					this.$set(this.tasks, t, updatedTask)
					// Move the task to the end of the done tasks if it is now done
					if (updatedTask.done) {
						this.tasks.splice(t, 1)
						this.tasks.push(updatedTask)
					}
					break
				}
			}
		},
		setDatesToNextWeek() {
			this.cStartDate = new Date()
			this.cEndDate = new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
			this.showOverdue = false
			this.setDate()
		},
		setDatesToNextMonth() {
			this.cStartDate = new Date()
			this.cEndDate = new Date((new Date()).setMonth((new Date()).getMonth() + 1))
			this.showOverdue = false
			this.setDate()
		},
		showTodaysTasks() {
			const d = new Date()
			this.cStartDate = new Date()
			this.cEndDate = new Date(d.setDate(d.getDate() + 1))
			this.showOverdue = true
			this.setDate()
		},
	},
}
</script>
