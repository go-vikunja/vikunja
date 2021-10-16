<template>
	<div class="is-max-width-desktop show-tasks">
		<fancycheckbox
			@change="setDate"
			class="is-pulled-right"
			v-if="!showAll"
			v-model="showNulls"
		>
			{{ $t('task.show.noDates') }}
		</fancycheckbox>
		<h3 v-if="showAll && tasks.length > 0">
			{{ $t('task.show.current') }}
		</h3>
		<h3 v-else-if="!showAll" class="mb-2">
			{{ $t('task.show.from') }}
			<flat-pickr
				:class="{ 'disabled': loading}"
				:config="flatPickerConfig"
				:disabled="loading"
				@on-close="setDate"
				class="input"
				v-model="cStartDate"
			/>
			{{ $t('task.show.until') }}
			<flat-pickr
				:class="{ 'disabled': loading}"
				:config="flatPickerConfig"
				:disabled="loading"
				@on-close="setDate"
				class="input"
				v-model="cEndDate"
			/>
		</h3>
		<div v-if="!showAll" class="mb-4">
			<x-button type="secondary" @click="showTodaysTasks()" class="mr-2">{{ $t('task.show.today') }}</x-button>
			<x-button type="secondary" @click="setDatesToNextWeek()" class="mr-2">{{
					$t('task.show.nextWeek')
				}}
			</x-button>
			<x-button type="secondary" @click="setDatesToNextMonth()">{{ $t('task.show.nextMonth') }}</x-button>
		</div>
		<template v-if="!loading && (!tasks || tasks.length === 0) && showNothingToDo">
			<h3 class="nothing">{{ $t('task.show.noTasks') }}</h3>
			<img alt="" :src="llamaCoolUrl"/>
		</template>
		<div :class="{ 'is-loading': loading}" class="spinner"></div>

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
import SingleTaskInList from '../../components/tasks/partials/singleTaskInList'
import {mapState} from 'vuex'

import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import Fancycheckbox from '../../components/input/fancycheckbox'
import {LOADING, LOADING_MODULE} from '../../store/mutation-types'

import llamaCoolUrl from '@/assets/llama-cool.svg'

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
			showNulls: true,
			showOverdue: false,

			cStartDate: null,
			cEndDate: null,

			showNothingToDo: false,
			llamaCoolUrl,
		}
	},
	props: {
		startDate: Date,
		endDate: Date,
		showAll: Boolean,
	},
	created() {
		this.cStartDate = this.startDate
		this.cEndDate = this.endDate
		this.loadPendingTasks()
	},
	mounted() {
		setTimeout(() => this.showNothingToDo = true, 100)
	},
	watch: {
		'$route': {
			handler: 'loadPendingTasks',
			deep: true,
		},
		startDate(newVal) {
			this.cStartDate = newVal
		},
		endDate(newVal) {
			this.cEndDate = newVal
		},
	},
	computed: {
		flatPickerConfig() {
			return {
				altFormat: this.$t('date.altFormatLong'),
				altInput: true,
				dateFormat: 'Y-m-d H:i',
				enableTime: true,
				time_24hr: true,
				locale: {
					firstDayOfWeek: this.$store.state.auth.settings.weekStart,
				},
			}
		},
		...mapState({
			userAuthenticated: state => state.auth.authenticated,
			loading: state => state[LOADING] && state[LOADING_MODULE] === 'tasks',
		}),
	},
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
				this.setTitle(this.$t('task.show.titleCurrent'))
			} else {
				this.setTitle(this.$t('task.show.titleDates', {
					from: this.cStartDate.toLocaleDateString(),
					to: this.cEndDate.toLocaleDateString(),
				}))
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

			this.$store.dispatch('tasks/loadTasks', params)
				.then(r => {

					// Sorting tasks with a due date so that the soonest or overdue are displayed at the top of the list.
					const tasksWithDueDates = r
						.filter(t => t.dueDate !== null)
						.sort((a, b) => a.dueDate > b.dueDate ? 1 : -1)

					const tasksWithoutDueDates = r.filter(t => t.dueDate === null)

					const tasks = [
						...tasksWithDueDates,
						...tasksWithoutDueDates,
					]

					this.$set(this, 'tasks', tasks)
				})
				.catch(e => {
					this.$message.error(e)
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
