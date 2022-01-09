<template>
	<div class="is-max-width-desktop has-text-left ">
		<h3 class="mb-2 title">
			{{ pageTitle }}
		</h3>
		<p v-if="!showAll" class="show-tasks-options">
			<datepicker-with-range @dateChanged="setDate"/>
			<fancycheckbox @change="setShowNulls" class="mr-2">
				{{ $t('task.show.noDates') }}
			</fancycheckbox>
			<fancycheckbox @change="setShowOverdue">
				{{ $t('task.show.overdue') }}
			</fancycheckbox>
		</p>
		<template v-if="!loading && (!tasks || tasks.length === 0) && showNothingToDo">
			<h3 class="has-text-centered mt-6">{{ $t('task.show.noTasks') }}</h3>
			<LlamaCool class="llama-cool"/>
		</template>

		<card
			v-if="hasTasks"
			:padding="false"
			class="has-overflow"
			:has-content="false"
			:loading="loading"
		>
			<div class="p-2">
				<single-task-in-list
					v-for="t in tasksSorted"
					:key="t.id"
					class="task"
					:show-list="true"
					:the-task="t"
					@taskUpdated="updateTasks"/>
			</div>
		</card>
		<div v-else :class="{ 'is-loading': loading}" class="spinner"></div>
	</div>
</template>
<script>
import {dateRanges} from '@/components/date/dateRanges'
import SingleTaskInList from '@/components/tasks/partials/singleTaskInList'
import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {mapState} from 'vuex'

import Fancycheckbox from '@/components/input/fancycheckbox'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'

import LlamaCool from '@/assets/llama-cool.svg?component'
import DatepickerWithRange from '@/components/date/datepickerWithRange'

function getNextWeekDate() {
	return new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
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
			showNothingToDo: false,
		}
	},
	props: {
		showAll: Boolean,
	},
	created() {
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
	},
	computed: {
		dateFrom() {
			return parseDateOrString(this.$route.query.from, new Date())
		},
		dateTo() {
			return parseDateOrString(this.$route.query.to, getNextWeekDate())
		},
		showNulls() {
			return this.$route.query.showNulls === 'true'
		},
		showOverdue() {
			return this.$route.query.showOverdue === 'true'
		},
		pageTitle() {
			let title = ''

			// We need to define "key" because it is the first parameter in the array and we need the second
			// eslint-disable-next-line no-unused-vars
			const predefinedRange = Object.entries(dateRanges).find(([key, value]) => this.dateFrom === value[0] && this.dateTo === value[1])
			if (typeof predefinedRange !== 'undefined') {
				title = this.$t(`input.datepickerRange.ranges.${predefinedRange[0]}`)
			} else {
				title = this.showAll
					? this.$t('task.show.titleCurrent')
					: this.$t('task.show.fromuntil', {
						from: this.format(this.dateFrom, 'PPP'),
						until: this.format(this.dateTo, 'PPP'),
					})
			}

			this.setTitle(title)

			return title
		},
		tasksSorted() {
			// Sort all tasks to put those with a due date before the ones without a due date, the
			// soonest before the later ones.
			// We can't use the api sorting here because that sorts tasks with a due date after
			// ones without a due date.
			return [...this.tasks].sort((a, b) => {
				const sortByDueDate = b.dueDate - a.dueDate
				return sortByDueDate === 0
					? b.id - a.id
					: sortByDueDate
			})
		},
		hasTasks() {
			return this.tasks && this.tasks.length > 0
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
					from: dateFrom ?? this.dateFrom,
					to: dateTo ?? this.dateTo,
					showOverdue: this.showOverdue,
					showNulls: this.showNulls,
				},
			})
		},
		setShowOverdue(show) {
			this.$router.push({
				name: this.$route.name,
				query: {
					...this.$route.query,
					showOverdue: show,
				},
			})
		},
		setShowNulls(show) {
			this.$router.push({
				name: this.$route.name,
				query: {
					...this.$route.query,
					showNulls: show,
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
				params.filter_by.push('due_date')
				params.filter_value.push(this.dateTo)
				params.filter_comparator.push('less')

				// NOTE: Ideally we could also show tasks with a start or end date in the specified range, but the api
				//       is not capable (yet) of combining multiple filters with 'and' and 'or'.

				if (!this.showOverdue) {
					params.filter_by.push('due_date')
					params.filter_value.push(this.dateFrom)
					params.filter_comparator.push('greater')
				}
			}

			this.tasks = await this.$store.dispatch('tasks/loadTasks', params)
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
.show-tasks-options {
	display: flex;
	flex-direction: column;
}

.llama-cool {
	margin: 3rem auto 0;
	display: block;
}
</style>