<template>
	<div class="is-max-width-desktop show-tasks">
		<h3 class="mb-2">
			{{ pageTitle }}
		</h3>
		<p v-if="!showAll">
			{{ $t('task.show.select') }}
			<datepicker-with-range @dateChanged="setDate"/>
		</p>
		<template v-if="!loading && (!tasks || tasks.length === 0) && showNothingToDo">
			<h3 class="nothing">{{ $t('task.show.noTasks') }}</h3>
			<LlamaCool class="llama-cool"/>
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
import SingleTaskInList from '@/components/tasks/partials/singleTaskInList'
import {mapState} from 'vuex'

import Fancycheckbox from '@/components/input/fancycheckbox'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'

import LlamaCool from '@/assets/llama-cool.svg?component'
import DatepickerWithRange from '@/components/date/datepickerWithRange'

function formatDate(date) {
	return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes()}`
}

export default {
	name: 'ShowTasks',
	components: {
		DatepickerWithRange,
		Fancycheckbox,
		SingleTaskInList,
		LlamaCool,
	},
	data() {
		return {
			tasks: [],
			showNulls: true,
			showOverdue: false,

			// TODO: Set the date range based on the default (to make sure it shows up in the picker)  -> maybe also use a computed which depends on dateFrom and dateTo?
			dateRange: null,

			showNothingToDo: false,
		}
	},
	props: {
		startDate: Date,
		endDate: Date,
		showAll: Boolean,
	},
	created() {
		this.loadPendingTasks()
	},
	mounted() {
		// FIXME
		setTimeout(() => this.showNothingToDo = true, 100)
	},
	watch: {
		'$route': {
			handler: 'loadPendingTasks',
			deep: true,
		},
	},
	computed: {
		dateFrom() {
			const d = new Date(Number(this.$route.query.from))

			return !isNaN(d)
				? d
				: this.startDate
		},
		dateTo() {
			const d = new Date(Number(this.$route.query.to))

			return !isNaN(d)
				? d
				: this.endDate
		},
		pageTitle() {
			const title = this.showAll
				? this.$t('task.show.titleCurrent')
				: this.$t('task.show.fromuntil', {
					from: this.formatDateShort(this.dateFrom),
					until: this.formatDateShort(this.dateTo)
				})

			this.setTitle(title)

			return title
		},
		...mapState({
			userAuthenticated: state => state.auth.authenticated,
			loading: state => state[LOADING] && state[LOADING_MODULE] === 'tasks',
		}),
	},
	methods: {
		setDate({dateFrom, dateTo}) {
			this.$router.push({
				name: this.$route.name,
				query: {
					from: +new Date(dateFrom),
					to: +new Date(dateTo),
					showOverdue: this.showOverdue,
					showNulls: this.showNulls,
				},
			})
		},
		async loadPendingTasks() {
			// Since this route is authentication only, users would get an error message if they access the page unauthenticated.
			// Since this component is mounted as the home page before unauthenticated users get redirected
			// to the login page, they will almost always see the error message.
			if (!this.userAuthenticated) {
				return
			}

			this.showOverdue = this.$route.query.showOverdue
			this.showNulls = this.$route.query.showNulls

			const params = {
				sort_by: ['due_date', 'id'],
				order_by: ['desc', 'desc'],
				filter_by: ['done'],
				filter_value: [false],
				filter_comparator: ['equals'],
				filter_concat: 'and',
				filter_include_nulls: this.showNulls,
			}
			
			// FIXME: Add button to show / hide overdue
			
			if (!this.showAll) {
				if (this.showNulls) {
					params.filter_by.push('start_date')
					params.filter_value.push(this.dateFrom)
					params.filter_comparator.push('greater')

					params.filter_by.push('end_date')
					params.filter_value.push(this.dateTo)
					params.filter_comparator.push('less')
				}

				params.filter_by.push('due_date')
				params.filter_value.push(this.dateFrom)
				params.filter_comparator.push('less')

				if (!this.showOverdue) {
					params.filter_by.push('due_date')
					params.filter_value.push(this.dateTo)
					params.filter_comparator.push('greater')
				}
			}

			const tasks = await this.$store.dispatch('tasks/loadTasks', params)

			// FIXME: sort tasks in computed
			// Sort all tasks to put those with a due date before the ones without a due date, the
			// soonest before the later ones.
			// We can't use the api sorting here because that sorts tasks with a due date after
			// ones without a due date.
			this.tasks = tasks.sort((a, b) => {
				const sortByDueDate = b.dueDate - a.dueDate
				return sortByDueDate === 0
					? b.id - a.id
					: sortByDueDate
			})
		},

		// FIXME: this modification should happen in the store
		updateTasks(updatedTask) {
			for (const t in this.tasks) {
				if (this.tasks[t].id === updatedTask.id) {
					this.tasks[t] = updatedTask
					// Move the task to the end of the done tasks if it is now done
					if (updatedTask.done) {
						this.tasks.splice(t, 1)
						this.tasks.push(updatedTask)
					}
					break
				}
			}
		},
	},
}
</script>

<style lang="scss" scoped>
h3 {
	text-align: left;

	&.nothing {
		text-align: center;
		margin-top: 3rem;
	}

	:deep(.input) {
		width: 190px;
		vertical-align: middle;
		margin: .5rem 0;
	}
}

.tasks {
	padding: .5rem;
}

.llama-cool {
	margin-top: 2rem;
}
</style>