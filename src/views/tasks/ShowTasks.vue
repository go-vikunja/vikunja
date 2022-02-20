<template>
	<div class="is-max-width-desktop has-text-left ">
		<h3 class="mb-2 title">
			{{ pageTitle }}
		</h3>
		<p v-if="!showAll" class="show-tasks-options">
			<datepicker-with-range @dateChanged="setDate">
				<template #trigger="{toggle}">
					<x-button @click.prevent.stop="toggle()" variant="primary" :shadow="false" class="mb-2">
						{{ $t('task.show.select') }}
					</x-button>
				</template>
			</datepicker-with-range>
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
import {computed, ref, watchEffect} from 'vue'
import {useStore} from 'vuex'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import TaskModel from '@/models/task'
import {formatDate} from '@/helpers/time/formatDate'
import {setTitle} from '@/helpers/setTitle'
import {objectToSnakeCase} from '@/helpers/case'

import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import SingleTaskInList from '@/components/tasks/partials/singleTaskInList.vue'
import DatepickerWithRange from '@/components/date/datepickerWithRange.vue'
import {DATE_RANGES} from '@/components/date/dateRanges'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'
import LlamaCool from '@/assets/llama-cool.svg?component'

const store = useStore()
const route = useRoute()
const router = useRouter()
const {t} = useI18n()

const tasks = ref<TaskModel[]>([])
const showNothingToDo = ref<boolean>(false)

setTimeout(() => showNothingToDo.value = true, 100)

// NOTE: You MUST provide either dateFrom and dateTo OR showAll for the component to actually show tasks.
// Linting disabled because we explicitely enabled destructuring in vite's config, this will work.
// eslint-disable-next-line vue/no-setup-props-destructure
const {
	dateFrom,
	dateTo,
	showAll = false,
	showNulls = false,
	showOverdue = false,
} = defineProps<{
	dateFrom?: Date | string,
	dateTo?: Date | string,
	showAll?: Boolean,
	showNulls?: Boolean,
	showOverdue?: Boolean,
}>()

const pageTitle = computed(() => {
	let title = ''

	// We need to define "key" because it is the first parameter in the array and we need the second
	// eslint-disable-next-line no-unused-vars
	const predefinedRange = Object.entries(DATE_RANGES).find(([key, value]) => dateFrom === value[0] && dateTo === value[1])
	if (typeof predefinedRange !== 'undefined') {
		title = t(`input.datepickerRange.ranges.${predefinedRange[0]}`)
	} else {
		title = showAll
			? t('task.show.titleCurrent')
			: t('task.show.fromuntil', {
				from: formatDate(dateFrom, 'PPP'),
				until: formatDate(dateTo, 'PPP'),
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
const hasTasks = computed(() => tasks.value && tasks.value.length > 0)
const userAuthenticated = computed(() => store.state.auth.authenticated)
const loading = computed(() => store.state[LOADING] && store.state[LOADING_MODULE] === 'tasks')

interface dateStrings {
	dateFrom: string,
	dateTo: string,
}

function setDate(dates: dateStrings) {
	router.push({
		name: route.name as string,
		query: {
			from: dates.dateFrom ?? dateFrom,
			to: dates.dateTo ?? dateTo,
			showOverdue: showOverdue ? 'true' : 'false',
			showNulls: showNulls ? 'true' : 'false',
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

async function loadPendingTasks(from: string, to: string) {
	// FIXME: HACK! This should never happen.
	// Since this route is authentication only, users would get an error message if they access the page unauthenticated.
	// Since this component is mounted as the home page before unauthenticated users get redirected
	// to the login page, they will almost always see the error message.
	if (!userAuthenticated.value) {
		return
	}

	const params = {
		sortBy: ['due_date', 'id'],
		orderBy: ['desc', 'desc'],
		filterBy: ['done'],
		filterValue: ['false'],
		filterComparator: ['equals'],
		filterConcat: 'and',
		filterIncludeNulls: showNulls,
	}

	if (!showAll) {
		params.filterBy.push('due_date')
		params.filterValue.push(to)
		params.filterComparator.push('less')

		// NOTE: Ideally we could also show tasks with a start or end date in the specified range, but the api
		//       is not capable (yet) of combining multiple filters with 'and' and 'or'.

		if (!showOverdue) {
			params.filterBy.push('due_date')
			params.filterValue.push(from)
			params.filterComparator.push('greater')
		}
	}

	tasks.value = await store.dispatch('tasks/loadTasks', objectToSnakeCase(params))
}

// FIXME: this modification should happen in the store
function updateTasks(updatedTask: TaskModel) {
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

watchEffect(() => loadPendingTasks(dateFrom as string, dateTo as string))
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