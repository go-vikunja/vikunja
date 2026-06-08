<template>
	<div class="time-tracking">
		<div class="time-tracking__actions">
			<span class="time-tracking__range">
				{{ rangeLabel }}
			</span>
			<div class="time-tracking__buttons">
				<XButton
					v-cy="'addTimeEntry'"
					variant="secondary"
					icon="plus"
					:class="{'is-active': showForm}"
					@click="showForm = !showForm"
				>
					{{ $t('timeTracking.logTime') }}
				</XButton>
				<XButton
					v-cy="'openTimeTrackingFilters'"
					variant="secondary"
					icon="filter"
					:class="{'has-filters': hasFilters}"
					@click="filterModalOpen = true"
				>
					{{ $t('filters.title') }}
				</XButton>
			</div>
		</div>

		<Card
			v-if="formVisible"
			:title="$t(editingEntry ? 'timeTracking.editEntry' : 'timeTracking.logTime')"
		>
			<TimeEntryForm
				:entry="editingEntry"
				:recent-entries="timeTrackingStore.browsedEntries"
				@saved="onSaved"
				@cancel="editingEntry = null"
			/>
		</Card>

		<TimeEntryList
			:entries="timeTrackingStore.browsedEntries"
			:empty-text="$t('timeTracking.list.emptyFiltered')"
			@edit="editingEntry = $event"
			@delete="onDelete"
		/>

		<Modal
			:enabled="filterModalOpen"
			:overflow="true"
			variant="hint-modal"
			@close="filterModalOpen = false"
		>
			<Card
				class="has-overflow"
				:title="$t('filters.title')"
				show-close
				@close="filterModalOpen = false"
			>
				<div class="field">
					<label class="label">{{ $t('misc.dateRange') }}</label>
					<DatepickerWithRange v-model="dateRange">
						<template #trigger="{toggle, buttonText}">
							<XButton
								variant="secondary"
								:shadow="false"
								@click.prevent.stop="toggle()"
							>
								{{ buttonText || $t('timeTracking.browse.selectRange') }}
							</XButton>
						</template>
					</DatepickerWithRange>
				</div>
				<div class="filter-columns">
					<div class="field">
						<label class="label">{{ $t('task.attributes.project') }}</label>
						<ProjectSearch v-model="selectedProject" />
					</div>
					<div class="field">
						<label class="label">{{ $t('timeTracking.form.task') }}</label>
						<Multiselect
							v-model="selectedTask"
							:placeholder="$t('timeTracking.form.taskSearch')"
							:loading="taskService.loading"
							:search-results="foundTasks"
							label="title"
							@search="findTasks"
						>
							<template #searchResult="{option}">
								{{ option.title }}
							</template>
						</Multiselect>
					</div>
				</div>
				<div class="field">
					<label class="label">{{ $t('misc.user') }}</label>
					<Multiselect
						v-model="selectedUser"
						:placeholder="$t('timeTracking.browse.userSearch')"
						:loading="userService.loading"
						:search-results="foundUsers"
						label="username"
						@search="findUsers"
					>
						<template #searchResult="{option}">
							{{ option.username }}
						</template>
					</Multiselect>
				</div>
			</Card>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, shallowReactive, watch, nextTick, onMounted} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'
import Card from '@/components/misc/Card.vue'
import DatepickerWithRange from '@/components/date/DatepickerWithRange.vue'
import {DATE_RANGES} from '@/components/date/dateRanges'
import Multiselect from '@/components/input/Multiselect.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import TimeEntryForm from '@/components/time-tracking/TimeEntryForm.vue'
import TimeEntryList from '@/components/time-tracking/TimeEntryList.vue'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import UserService from '@/services/user'
import {useTitle} from '@/composables/useTitle'
import {useTimeTrackingStore} from '@/stores/timeTracking'
import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

import type {IProject} from '@/modelTypes/IProject'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

const {t} = useI18n()
const route = useRoute()
const router = useRouter()
const timeTrackingStore = useTimeTrackingStore()
const baseStore = useBaseStore()
const projectStore = useProjectStore()

useTitle(() => t('timeTracking.title'))

const showForm = ref(false)
const editingEntry = ref<ITimeEntry | null>(null)
const formVisible = computed(() => showForm.value || editingEntry.value !== null)

function onSaved() {
	editingEntry.value = null
	showForm.value = false
	timeTrackingStore.browseEntries(filter.value)
}

function onDelete(id: number) {
	timeTrackingStore.removeEntry(id)
}

// --- Filter ---------------------------------------------------------------

// DatepickerWithRange emits null for a side when the range is cleared (Custom).
const dateRange = ref<{dateFrom: Date | string | null, dateTo: Date | string | null}>({
	dateFrom: 'now/d',
	dateTo: 'now/d+1d',
})
const selectedProject = ref<IProject | null>(null)
const selectedTask = ref<ITask | null>(null)
const selectedUser = ref<IUser | null>(null)
const filterModalOpen = ref(false)

const hasFilters = computed(() =>
	selectedProject.value !== null ||
	selectedTask.value !== null ||
	selectedUser.value !== null ||
	dateRange.value.dateFrom !== 'now/d' ||
	dateRange.value.dateTo !== 'now/d+1d',
)

// The active range as a label (the preset name when it matches, else the dates).
const rangeLabel = computed(() => {
	const {dateFrom, dateTo} = dateRange.value
	if (!dateFrom || !dateTo) {
		return t('timeTracking.browse.selectRange')
	}
	const preset = Object.entries(DATE_RANGES).find(
		([, range]) => dateFrom === range[0] && dateTo === range[1],
	)
	if (preset) {
		return t(`input.datepickerRange.ranges.${preset[0]}`)
	}
	return t('input.datepickerRange.fromto', {from: dateValue(dateFrom), to: dateValue(dateTo)})
})

const taskService = shallowReactive(new TaskService())
const foundTasks = ref<ITask[]>([])
async function findTasks(query: string) {
	if (query === '') {
		foundTasks.value = []
		return
	}
	foundTasks.value = await taskService.getAll({}, {s: query, sort_by: 'done'}) as ITask[]
}

const userService = shallowReactive(new UserService())
const foundUsers = ref<IUser[]>([])
async function findUsers(query: string) {
	if (query === '') {
		foundUsers.value = []
		return
	}
	foundUsers.value = await userService.getAll({}, {s: query}) as IUser[]
}

// Datemath preset strings (now/M) pass through unchanged; a custom Date becomes
// YYYY-MM-DD — both avoid the ':' the filter grammar tokenises on.
function dateValue(value: Date | string): string {
	if (typeof value === 'string') {
		return value
	}
	const year = value.getFullYear()
	const month = String(value.getMonth() + 1).padStart(2, '0')
	const day = String(value.getDate()).padStart(2, '0')
	return `${year}-${month}-${day}`
}

const filter = computed(() => {
	const parts: string[] = []
	if (dateRange.value.dateFrom) {
		parts.push(`start_time > ${dateValue(dateRange.value.dateFrom)}`)
	}
	if (dateRange.value.dateTo) {
		parts.push(`start_time < ${dateValue(dateRange.value.dateTo)}`)
	}
	if (selectedUser.value !== null) {
		parts.push(`user_id = ${selectedUser.value.id}`)
	}
	if (selectedTask.value !== null) {
		parts.push(`task_id = ${selectedTask.value.id}`)
	}
	if (selectedProject.value !== null) {
		parts.push(`project_id = ${selectedProject.value.id}`)
	}
	return parts.join(' && ')
})

// Persist the active filter to the URL so it's shareable and survives reloads.
const filterQuery = computed(() => {
	const q: Record<string, string> = {}
	if (dateRange.value.dateFrom && dateRange.value.dateFrom !== 'now/d') {
		q.from = dateValue(dateRange.value.dateFrom)
	}
	if (dateRange.value.dateTo && dateRange.value.dateTo !== 'now/d+1d') {
		q.to = dateValue(dateRange.value.dateTo)
	}
	if (selectedProject.value !== null) {
		q.project = String(selectedProject.value.id)
	}
	if (selectedTask.value !== null) {
		q.task = String(selectedTask.value.id)
	}
	if (selectedUser.value !== null) {
		q.user = selectedUser.value.username
	}
	return q
})

const ready = ref(false)

async function restoreFromQuery() {
	const q = route.query
	if (typeof q.from === 'string') {
		dateRange.value.dateFrom = q.from
	}
	if (typeof q.to === 'string') {
		dateRange.value.dateTo = q.to
	}
	// Resolve project/task by id and the user by username up front (the project
	// store may not be hydrated yet on a hard reload), so the first request
	// already carries the full filter — and the modal shows the real names.
	await Promise.all([
		typeof q.project === 'string'
			? projectStore.loadProject(Number(q.project))
				.then(p => { selectedProject.value = p as IProject })
				.catch(() => { /* project gone — drop the filter */ })
			: Promise.resolve(),
		typeof q.task === 'string'
			? taskService.get(new TaskModel({id: Number(q.task)}))
				.then(t => { selectedTask.value = t as ITask })
				.catch(() => { /* task gone — drop the filter */ })
			: Promise.resolve(),
		typeof q.user === 'string'
			? userService.getAll({}, {s: q.user})
				.then(users => {
					selectedUser.value = (users as IUser[]).find(u => u.username === q.user) ?? null
				})
				.catch(() => { /* user not found — drop the filter */ })
			: Promise.resolve(),
	])
}

onMounted(async () => {
	// Standalone page: drop any stale project so the app header shows this
	// page's title instead of the last visited project.
	baseStore.handleSetCurrentProject({project: null})
	await restoreFromQuery()
	ready.value = true
	// One request with the fully-restored filter — no flicker through partial filters.
	timeTrackingStore.browseEntries(filter.value)
})

// DatepickerWithRange only syncs its display from modelValue on change, and it
// remounts each time the modal opens — re-push the value so the range shows.
watch(filterModalOpen, open => {
	if (open) {
		nextTick(() => {
			dateRange.value = {...dateRange.value}
		})
	}
})

watch(filterQuery, q => {
	if (!ready.value) {
		return
	}
	router.replace({query: q}).catch(() => { /* ignore redundant navigation */ })
})

watch(filter, value => {
	if (!ready.value) {
		return
	}
	timeTrackingStore.browseEntries(value)
})
</script>

<style lang="scss" scoped>
.time-tracking__actions {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 1rem;
	margin-block-end: 1.5rem;
}

.time-tracking__range {
	color: var(--grey-500);
	font-size: .9rem;
}

.time-tracking__buttons {
	display: flex;
	gap: .5rem;
}

.filter-columns {
	display: flex;
	gap: 1rem;

	> .field {
		flex: 1;
		min-inline-size: 0;
	}
}

// The multiselect's per-row "click or press enter to select" hint is
// transparent but still reserves its (long) width, clipping the project/task
// title to a few characters in the narrow side-by-side columns. Drop it.
:deep(.hint-text) {
	display: none;
}

$filter-bubble-size: .75rem;

// Blue dot on the filter button when any filter is active (mirrors project views).
.has-filters {
	position: relative;

	&::after {
		content: '';
		position: absolute;
		inset-block-start: math.div($filter-bubble-size, -2);
		inset-inline-end: math.div($filter-bubble-size, -2);

		inline-size: $filter-bubble-size;
		block-size: $filter-bubble-size;
		border-radius: 100%;
		background: var(--primary);
	}
}
</style>
