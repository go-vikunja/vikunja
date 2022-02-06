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

<script setup lang="ts">
import {dateRanges} from '@/components/date/dateRanges'
import SingleTaskInList from '@/components/tasks/partials/singleTaskInList.vue'
import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {mapState, useStore} from 'vuex'
import {computed, ref, watchEffect} from 'vue'

import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'

import LlamaCool from '@/assets/llama-cool.svg?component'
import DatepickerWithRange from '@/components/date/datepickerWithRange.vue'
import TaskModel from '@/models/task'
import {useRoute, useRouter} from 'vue-router'
import {formatDate} from '@/helpers/time/formatDate'
import {useI18n} from 'vue-i18n'
import {setTitle} from '@/helpers/setTitle'

function getNextWeekDate() {
	return new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
}

const store = useStore()
const route = useRoute()
const router = useRouter()
const {t} = useI18n()

const tasks = ref<TaskModel[]>([])
const showNothingToDo = ref<boolean>(false)

setTimeout(() => showNothingToDo.value = true, 100)

const props = defineProps({
	showAll: Boolean,
})

const dateFrom = computed<Date | string>(() => parseDateOrString(route.query.from as string, new Date()))
const dateTo = computed<Date | string>(() => parseDateOrString(route.query.to as string, getNextWeekDate()))
const showNulls = computed(() => route.query.showNulls === 'true')
const showOverdue = computed(() => route.query.showOverdue === 'true')
const pageTitle = computed(() => {
	let title = ''

	// We need to define "key" because it is the first parameter in the array and we need the second
	// eslint-disable-next-line no-unused-vars
	const predefinedRange = Object.entries(dateRanges).find(([key, value]) => dateFrom.value === value[0] && dateTo.value === value[1])
	if (typeof predefinedRange !== 'undefined') {
		title = t(`input.datepickerRange.ranges.${predefinedRange[0]}`)
	} else {
		title = props.showAll
			? t('task.show.titleCurrent')
			: t('task.show.fromuntil', {
				from: formatDate(dateFrom.value, 'PPP'),
				until: formatDate(dateTo.value, 'PPP'),
			})
	}

	return title
})
const tasksSorted = computed(() => {
	// Sort all tasks to put those with a due date before the ones without a due date, the
	// soonest before the later ones.
	// We can't use the api sorting here because that sorts tasks with a due date after
	// ones without a due date.

	const tasksWithDueDate = [...tasks.value]
		.filter(t => t.dueDate !== null)
		.sort((a, b) => {
			const sortByDueDate = a.dueDate - b.dueDate
			return sortByDueDate === 0
				? b.id - a.id
				: sortByDueDate
		})
	const tasksWithoutDueDate = [...tasks.value]
		.filter(t => t.dueDate === null)

	return [
		...tasksWithDueDate,
		...tasksWithoutDueDate,
	]
})
const hasTasks = computed(() => tasks && tasks.value.length > 0)
const userAuthenticated = computed(() => store.state.auth.authenticated)
const loading = computed(() => store.state[LOADING] && store.state[LOADING_MODULE] === 'tasks')

interface dateStrings {
	from: string,
	to: string,
}

function setDate({from, to}: dateStrings) {
	router.push({
		name: route.name as string,
		query: {
			from: from ?? dateFrom,
			to: to ?? dateTo,
			showOverdue: showOverdue.value ? 'true' : 'false',
			showNulls: showNulls.value ? 'true' : 'false',
		},
	})
}

function setShowOverdue(show: boolean) {
	router.push({
		name: route.name as string,
		query: {
			...route.query,
			showOverdue: show ? 'true' : 'false',
		},
	})
}

function setShowNulls(show: boolean) {
	router.push({
		name: route.name as string,
		query: {
			...route.query,
			showNulls: show ? 'true' : 'false',
		},
	})
}

async function loadPendingTasks(from:string, to:string) {
	// Since this route is authentication only, users would get an error message if they access the page unauthenticated.
	// Since this component is mounted as the home page before unauthenticated users get redirected
	// to the login page, they will almost always see the error message.
	if (!userAuthenticated) {
		return
	}

	const params = {
		sort_by: ['due_date', 'id'],
		order_by: ['desc', 'desc'],
		filter_by: ['done'],
		filter_value: ['false'],
		filter_comparator: ['equals'],
		filter_concat: 'and',
		filter_include_nulls: showNulls.value,
	}

	if (!props.showAll) {
		params.filter_by.push('due_date')
		params.filter_value.push(to)
		params.filter_comparator.push('less')

		// NOTE: Ideally we could also show tasks with a start or end date in the specified range, but the api
		//       is not capable (yet) of combining multiple filters with 'and' and 'or'.

		if (!showOverdue.value) {
			params.filter_by.push('due_date')
			params.filter_value.push(from)
			params.filter_comparator.push('greater')
		}
	}

	tasks.value = await store.dispatch('tasks/loadTasks', params)
}

// FIXME: this modification should happen in the store
function updateTasks(updatedTask) {
	for (const t in tasks.value) {
		if (tasks.value[t].id === updatedTask.id) {
			tasks.value[t] = updatedTask
			// Move the task to the end of the done tasks if it is now done
			if (updatedTask.done) {
				tasks.value.splice(t, 1)
				tasks.value.push(updatedTask)
			}
			break
		}
	}
}

watchEffect(() => loadPendingTasks(dateFrom.value as string, dateTo.value as string))
watchEffect(() => setTitle(pageTitle.value))
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