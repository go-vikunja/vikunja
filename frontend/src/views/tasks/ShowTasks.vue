<template>
	<div
		v-cy="'showTasks'"
		class="content-widescreen"
	>
		<h2>{{ pageTitle }}</h2>
		<p class="has-text-grey">
			{{ $t('task.show.pageDescription') }}
		</p>

		<Message
			v-if="filteredLabels.length > 0"
			class="label-filter-info"
		>
			<i18n-t
				keypath="task.show.filterByLabel"
				tag="span"
				class="filter-label-text"
			>
				<template #label>
					<XLabel
						v-for="label in filteredLabels"
						:key="label.id"
						:label="label"
					/>
				</template>
			</i18n-t>
			<BaseButton
				v-tooltip="$t('task.show.clearLabelFilter')"
				class="clear-filter-button"
				@click="clearLabelFilter"
			>
				<Icon icon="times" />
			</BaseButton>
		</Message>

		<hr class="page-separator">

		<div class="options-bar">
			<DatepickerWithRange
				v-if="!showAll"
				@update:modelValue="setDate"
			>
				<template #trigger="{toggle}">
					<XButton
						variant="primary"
						:shadow="false"
						@click.prevent.stop="toggle()"
					>
						{{ $t('task.show.select') }}
					</XButton>
				</template>
			</DatepickerWithRange>
			<div class="options-checks">
				<FancyCheckbox
					:model-value="effectiveShowNulls"
					@update:modelValue="setShowNulls"
				>
					{{ $t('task.show.noDates') }}
				</FancyCheckbox>
				<FancyCheckbox
					:model-value="effectiveShowOverdue"
					@update:modelValue="setShowOverdue"
				>
					{{ $t('task.show.overdue') }}
				</FancyCheckbox>
				<FancyCheckbox
					:model-value="effectiveAssignedToMe"
					@update:modelValue="setAssignedToMe"
				>
					{{ $t('task.show.assignedToMe') }}
				</FancyCheckbox>
			</div>
		</div>

		<template v-if="!loading && (!tasks || tasks.length === 0) && showNothingToDo">
			<div class="has-text-centered p-4">
				<h3 class="has-text-grey">
					{{ $t('task.show.noTasks') }}
				</h3>
				<LlamaCool class="llama-cool" />
			</div>
		</template>

		<Card
			v-if="hasTasks"
			:padding="false"
			class="has-overflow task-card"
			:has-content="false"
			:loading="loading"
		>
			<div class="p-2">
				<SingleTaskInProject
					v-for="task in tasks"
					:key="task.id"
					:show-project="true"
					:the-task="task"
					@taskUpdated="updateTasks"
				/>
			</div>
		</Card>
		<div
			v-else-if="loading"
			class="has-text-centered p-4"
		>
			<span class="loader is-loading" />
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import {formatDate} from '@/helpers/time/formatDate'
import {setTitle} from '@/helpers/setTitle'

import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import Message from '@/components/misc/Message.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import SingleTaskInProject from '@/components/tasks/partials/SingleTaskInProject.vue'
import DatepickerWithRange from '@/components/date/DatepickerWithRange.vue'
import XLabel from '@/components/tasks/partials/Label.vue'
import {DATE_RANGES} from '@/components/date/dateRanges'
import LlamaCool from '@/assets/llama-cool.svg?component'
import type {ITask} from '@/modelTypes/ITask'
import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import type {TaskFilterParams} from '@/services/taskCollection'
import TaskCollectionService from '@/services/taskCollection'
import {useStorage} from '@vueuse/core'

const props = withDefaults(defineProps<{
	dateFrom?: Date | string,
	dateTo?: Date | string,
	showNulls?: boolean,
	showOverdue?: boolean,
	labelIds?: string[],
}>(), {
	showNulls: false,
	showOverdue: false,
	dateFrom: undefined,
	dateTo: undefined,
	labelIds: undefined,
})

const emit = defineEmits<{
	'tasksLoaded': true,
	'clearLabelFilter': void,
}>()

const authStore = useAuthStore()
const taskStore = useTaskStore()
const projectStore = useProjectStore()
const labelStore = useLabelStore()

const route = useRoute()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const tasks = ref<ITask[]>([])
const showNothingToDo = ref<boolean>(false)
const taskCollectionService = ref(new TaskCollectionService())

// Persist checkbox state so it survives navigation away and back
const storedShowNulls = useStorage('upcomingShowNulls', false)
const storedShowOverdue = useStorage('upcomingShowOverdue', false)
const storedAssignedToMe = useStorage('upcomingAssignedToMe', false)

// Effective values: use prop (from query param) if explicitly set, otherwise use stored
const effectiveShowNulls = computed(() => {
	// If the query param is present, use it and sync to storage
	if (route.query.showNulls !== undefined) {
		const val = props.showNulls
		storedShowNulls.value = val
		return val
	}
	return storedShowNulls.value
})

const effectiveShowOverdue = computed(() => {
	if (route.query.showOverdue !== undefined) {
		const val = props.showOverdue
		storedShowOverdue.value = val
		return val
	}
	return storedShowOverdue.value
})

const effectiveAssignedToMe = computed(() => {
	if (route.query.assignedToMe !== undefined) {
		const val = route.query.assignedToMe === 'true'
		storedAssignedToMe.value = val
		return val
	}
	return storedAssignedToMe.value
})

setTimeout(() => showNothingToDo.value = true, 100)

const showAll = computed(() => typeof props.dateFrom === 'undefined' || typeof props.dateTo === 'undefined')

const filteredLabels = computed(() => {
	if (!props.labelIds || props.labelIds.length === 0) {
		return []
	}
	return props.labelIds
		.map(id => labelStore.getLabelById(Number(id)))
		.filter(label => label !== null && label !== undefined)
})

const pageTitle = computed(() => {
	// We need to define "key" because it is the first parameter in the array and we need the second
	const predefinedRange = Object.entries(DATE_RANGES)
		.find(([, value]) => props.dateFrom === value[0] && props.dateTo === value[1])
		?.[0]
	if (typeof predefinedRange !== 'undefined') {
		return t(`input.datepickerRange.ranges.${predefinedRange}`)
	}

	return showAll.value
		? t('task.show.titleCurrent')
		: t('task.show.fromuntil', {
			from: formatDate(props.dateFrom, 'LL'),
			until: formatDate(props.dateTo, 'LL'),
		})
})
const hasTasks = computed(() => tasks.value && tasks.value.length > 0)
const userAuthenticated = computed(() => authStore.authenticated)
const loading = computed(() => taskStore.isLoading || taskCollectionService.value.loading)
const filterIdUsedOnOverview = computed(() => authStore.settings?.frontendSettings?.filterIdUsedOnOverview)

interface dateStrings {
	dateFrom: string,
	dateTo: string,
}

function setDate(dates: dateStrings) {
	router.push({
		name: route.name as string,
		query: {
			from: dates.dateFrom ?? props.dateFrom,
			to: dates.dateTo ?? props.dateTo,
			showOverdue: effectiveShowOverdue.value ? 'true' : 'false',
			showNulls: effectiveShowNulls.value ? 'true' : 'false',
			assignedToMe: effectiveAssignedToMe.value ? 'true' : 'false',
		},
	})
}

function setShowOverdue(show: boolean) {
	storedShowOverdue.value = show
	router.push({
		name: route.name as string,
		query: {
			...route.query,
			showOverdue: show ? 'true' : 'false',
		},
	})
}

function setShowNulls(show: boolean) {
	storedShowNulls.value = show
	router.push({
		name: route.name as string,
		query: {
			...route.query,
			showNulls: show ? 'true' : 'false',
		},
	})
}

function setAssignedToMe(show: boolean) {
	storedAssignedToMe.value = show
	router.push({
		name: route.name as string,
		query: {
			...route.query,
			assignedToMe: show ? 'true' : 'false',
		},
	})
}

function clearLabelFilter() {
	emit('clearLabelFilter')
}

async function loadPendingTasks(from: Date|string, to: Date|string, filterId: number | null | undefined) {
	// FIXME: HACK! This should never happen.
	// Since this route is authentication only, users would get an error message if they access the page unauthenticated.
	// Since this component is mounted as the home page before unauthenticated users get redirected
	// to the login page, they will almost always see the error message.
	if (!userAuthenticated.value) {
		return
	}

	const params: TaskFilterParams = {
		sort_by: ['due_date', 'id'],
		order_by: ['asc', 'desc'],
		filter: 'done = false',
		filter_include_nulls: effectiveShowNulls.value,
		s: '',
		expand: ['comment_count', 'is_unread'],
	}

	if (!showAll.value) {

		params.filter += ` && due_date < '${to instanceof Date ? to.toISOString() : to}'`

		// NOTE: Ideally we could also show tasks with a start or end date in the specified range, but the api
		//       is not capable (yet) of combining multiple filters with 'and' and 'or'.

		if (!effectiveShowOverdue.value) {
			params.filter += ` && due_date > '${from instanceof Date ? from.toISOString() : from}'`
		}
	} else {
		// In showAll mode, if overdue is unchecked, hide tasks with due_date in the past
		if (!effectiveShowOverdue.value) {
			params.filter += ` && due_date > '${new Date().toISOString()}'`
		}
	}

	// Add label filtering
	if (props.labelIds && props.labelIds.length > 0) {
		const labelFilter = `labels in ${props.labelIds.join(', ')}`
		params.filter += params.filter ? ` && ${labelFilter}` : labelFilter
	}

	// Add "assigned to me" filtering
	if (effectiveAssignedToMe.value) {
		const username = authStore.info?.username
		if (username) {
			params.filter += params.filter
				? ` && assignees = '${username}'`
				: `assignees = '${username}'`
		}
	}

	let projectId = null
	if (showAll.value && filterId && typeof projectStore.projects[filterId] !== 'undefined') {
		projectId = filterId
	}

	tasks.value = await taskStore.loadTasks(params, projectId)
	emit('tasksLoaded', true)
}

// FIXME: this modification should happen in the store
function updateTasks(updatedTask: ITask) {
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

// Use watch instead of watchEffect to prevent reloading tasks when unrelated settings change.
// watchEffect would track all reactive dependencies accessed inside loadPendingTasks,
// which includes the entire settings object. When sidebarWidth changes, the settings
// object is replaced, triggering the watchEffect even though filterIdUsedOnOverview
// hasn't changed. Using watch with explicit dependencies and immediate:true gives us
// the same behavior but only triggers when these specific values actually change.
watch(
	[() => props.dateFrom, () => props.dateTo, filterIdUsedOnOverview, effectiveShowNulls, effectiveShowOverdue, effectiveAssignedToMe],
	([from, to, filterId]) => loadPendingTasks(from, to, filterId),
	{immediate: true},
)
watchEffect(() => setTitle(pageTitle.value))
</script>

<style lang="scss" scoped>
.content-widescreen {
	max-inline-size: 900px;
	margin: 0 auto;
	padding: 1.5rem 1rem;
}

.options-bar {
	display: flex;
	align-items: center;
	gap: 1rem;
	flex-wrap: wrap;
	margin-block-end: 1.5rem;
}

.options-checks {
	display: flex;
	align-items: center;
	gap: 1rem;
}

.page-separator {
	border: none;
	border-block-start: 2px solid var(--grey-200);
	margin-block: 1rem 1.5rem;
}

.task-card {
	border-radius: $radius;
}

.llama-cool {
	margin: 1.5rem auto 0;
	display: block;
	max-block-size: 250px;
	overflow: visible;
}

.label-filter-info {
	margin-block-end: 1rem;
	
	.clear-filter-button {
		margin-inline-start: auto;
		padding: 0.25rem 0.5rem;
		
		&:hover {
			color: var(--danger);
		}
	}

	:deep(.message.info) {
		inline-size: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}
}
</style>
