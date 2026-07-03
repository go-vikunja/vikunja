<template>
	<div class="planner-view">
		<header class="planner-toolbar">
			<h1 class="planner-heading">
				{{ $t('planner.title') }}
			</h1>

			<div class="toolbar-controls">
				<XButton
					variant="secondary"
					:shadow="false"
					@click="goToday"
				>
					{{ $t('planner.today') }}
				</XButton>
				<div class="nav-arrows">
					<BaseButton
						:aria-label="$t('planner.previous')"
						@click="goPrev"
					>
						<Icon icon="angle-left" />
					</BaseButton>
					<BaseButton
						:aria-label="$t('planner.next')"
						@click="goNext"
					>
						<Icon icon="angle-right" />
					</BaseButton>
				</div>
				<span class="range-label">{{ rangeLabel }}</span>

				<div class="mode-toggle">
					<XButton
						:variant="viewMode === 'week' ? 'primary' : 'secondary'"
						:shadow="false"
						@click="viewMode = 'week'"
					>
						{{ $t('planner.week') }}
					</XButton>
					<XButton
						:variant="viewMode === 'day' ? 'primary' : 'secondary'"
						:shadow="false"
						@click="viewMode = 'day'"
					>
						{{ $t('planner.day') }}
					</XButton>
				</div>

				<div class="zoom-controls">
					<BaseButton
						:aria-label="$t('planner.zoomOut')"
						@click="zoomOut"
					>
						<Icon icon="minus" />
					</BaseButton>
					<BaseButton
						:aria-label="$t('planner.zoomIn')"
						@click="zoomIn"
					>
						<Icon icon="plus" />
					</BaseButton>
				</div>

				<PlannerSettings />

				<Loading
					v-if="isLoading"
					class="planner-loading is-loading-small"
				/>
			</div>
		</header>

		<Message
			v-if="loadError"
			variant="danger"
			class="planner-error"
		>
			{{ $t('planner.loadError') }}
		</Message>

		<div class="planner-body">
			<PlannerSidebar
				v-model:filter="sidebarFilter"
				v-model:sort="sidebarSort"
				:tasks="sidebarTasks"
				:overdue-tasks="overdueTasks"
				@openTask="openTask"
				@unschedule="unscheduleTask"
			/>
			<CalendarGrid
				:days="days"
				:tasks="visibleGridTasks"
				:slot-minutes="slotMinutes"
				:day-start-hour="dayStartHour"
				:day-end-hour="dayEndHour"
				:px-per-hour="pxPerHour"
				:auto-fit="!userZoomed"
				@openTask="openTask"
				@updateBlock="onUpdateBlock"
				@dropTask="onDropTask"
				@dropAllDay="onDropAllDay"
				@navigate="slideDays"
				@createTask="onCreateTask"
				@createAllDay="onCreateAllDay"
				@update:pxPerHour="value => pxPerHour = value"
			/>
		</div>

		<PlannerCreateTaskModal
			v-if="createCtx"
			:context="createCtx.label"
			@created="onCreated"
			@close="createCtx = null"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, ref, watchEffect} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useStorage} from '@vueuse/core'
import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import BaseButton from '@/components/base/BaseButton.vue'
import Loading from '@/components/misc/Loading.vue'
import Message from '@/components/misc/Message.vue'
import PlannerSidebar from './PlannerSidebar.vue'
import PlannerSettings from './PlannerSettings.vue'
import PlannerCreateTaskModal from './PlannerCreateTaskModal.vue'
import CalendarGrid from './grid/CalendarGrid.vue'

import {setTitle} from '@/helpers/setTitle'
import {formatDate} from '@/helpers/time/formatDate'
import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useCalendarSettings} from './helpers/useCalendarSettings'
import {usePlannerTimeFormatter} from './helpers/usePlannerTimeFormatter'
import {usePlannerTasks, type PlannerRange, type PlannerSidebarSort, PLANNER_SIDEBAR_SORTS, DEFAULT_PLANNER_SIDEBAR_SORT} from './helpers/usePlannerTasks'

const router = useRouter()
const {t} = useI18n({useScope: 'global'})
const {settings} = useCalendarSettings()
const authStore = useAuthStore()
const baseStore = useBaseStore()

const viewMode = useStorage<'week' | 'day'>('planner-view-mode', 'week')
const anchor = ref(new Date())
// filter_include_nulls must be defined: Filters.vue binds it to a Boolean
// FancyCheckbox, which warns on undefined.
const sidebarFilter = ref<TaskFilterParams>({filter: '', s: '', filter_include_nulls: false} as TaskFilterParams)
const sidebarSort = useStorage<PlannerSidebarSort>('planner-sidebar-sort', DEFAULT_PLANNER_SIDEBAR_SORT)
// An earlier build stored this as an object; reset any value that isn't a known option.
if (!PLANNER_SIDEBAR_SORTS.includes(sidebarSort.value)) {
	sidebarSort.value = DEFAULT_PLANNER_SIDEBAR_SORT
}

const pxPerHour = useStorage('planner-px-per-hour', 48)
const userZoomed = useStorage('planner-user-zoomed', false)

// "HH:MM" working-hour strings → fractional hours for the grid's zoom/scroll.
function hoursFromTime(time: string): number {
	const [h, m] = (time || '0:0').split(':').map(Number)
	return (h || 0) + (m || 0) / 60
}
const dayStartHour = computed(() => hoursFromTime(settings.value.dayStart))
const dayEndHour = computed(() => hoursFromTime(settings.value.dayEnd))

// Respect the user's configured first day of the week (0 = Sunday … 6 = Saturday).
function startOfWeek(date: Date): dayjs.Dayjs {
	const weekStart = authStore.settings.weekStart ?? 0
	const day = dayjs(date).startOf('day')
	const diff = (day.day() - weekStart + 7) % 7
	return day.subtract(diff, 'day')
}

const days = computed<Date[]>(() => {
	if (viewMode.value === 'day') {
		return [dayjs(anchor.value).startOf('day').toDate()]
	}
	if (settings.value.fullWeek) {
		const start = startOfWeek(anchor.value)
		return Array.from({length: 7}, (_, i) => start.add(i, 'day').toDate())
	}
	const count = Math.min(Math.max(settings.value.daysToShow || 7, 1), 31)
	const start = dayjs(anchor.value).startOf('day')
	return Array.from({length: count}, (_, i) => start.add(i, 'day').toDate())
})

const range = computed<PlannerRange>(() => ({
	from: days.value[0],
	to: dayjs(days.value[days.value.length - 1]).endOf('day').toDate(),
}))

const rangeLabel = computed(() => {
	if (viewMode.value === 'day') {
		return formatDate(anchor.value, 'LL')
	}
	const first = days.value[0]
	const last = days.value[days.value.length - 1]
	return `${formatDate(first, 'll')} – ${formatDate(last, 'll')}`
})

const overdueEnabled = computed(() => settings.value.showOverdue)
const {sidebarTasks, gridTasks, overdueTasks, isLoading, loadError, updateTask, scheduleTask} = usePlannerTasks(range, sidebarFilter, sidebarSort, overdueEnabled)

function findTask(taskId: number): ITask | undefined {
	return gridTasks.value.get(taskId)
		?? sidebarTasks.value.find(t => t.id === taskId)
		?? overdueTasks.value.find(t => t.id === taskId)
}

const visibleGridTasks = computed(() =>
	[...gridTasks.value.values()].filter(task => settings.value.showDone || !task.done),
)

// Guard the duration/slot inputs: a stray 0 or blank would yield NaN positions
// and invalid dates downstream.
const slotMinutes = computed(() => Math.max(Math.round(settings.value.slotMinutes) || 0, 5))
const defaultDurationMinutes = computed(() => Math.max(Math.round(settings.value.defaultDurationMinutes) || 0, 5))

// Page by the visible window (day=1, full week=7, rolling=daysToShow).
function goPrev() {
	anchor.value = dayjs(anchor.value).subtract(days.value.length, 'day').toDate()
}

function goNext() {
	anchor.value = dayjs(anchor.value).add(days.value.length, 'day').toDate()
}

function goToday() {
	anchor.value = new Date()
}

// Horizontal swipe/scroll slides the window. In a date-aligned full week a
// one-day shift is invisible (the week snaps back), so page by a whole week
// there; rolling and day views slide a day at a time.
function slideDays(delta: number) {
	const unit = viewMode.value === 'week' && settings.value.fullWeek ? 'week' : 'day'
	anchor.value = dayjs(anchor.value).add(delta, unit).toDate()
}

function zoomIn() {
	userZoomed.value = true
	pxPerHour.value = Math.min(pxPerHour.value + 12, 200)
}

function zoomOut() {
	userZoomed.value = true
	pxPerHour.value = Math.max(pxPerHour.value - 12, 16)
}

function hasPastDue(task: ITask | undefined): boolean {
	return !!task?.dueDate && dayjs(task.dueDate).isBefore(dayjs().startOf('day'))
}

function onDropTask({taskId, minutes, day}: {taskId: number, minutes: number, day: Date}) {
	const start = dayjs(day).startOf('day').add(minutes, 'minute')
	const end = start.add(defaultDurationMinutes.value, 'minute')
	const partial: Parameters<typeof updateTask>[0] = {id: taskId, startDate: start.toDate(), endDate: end.toDate()}
	// Rescheduling an overdue task also defers its missed deadline — otherwise
	// it would stay "overdue" despite now being planned. Future deadlines are
	// left alone: timeboxing work doesn't move its due date.
	if (hasPastDue(findTask(taskId))) {
		partial.dueDate = end.toDate()
	}
	updateTask(partial)
}

// Unscheduling must clear the due date too: any remaining date re-files the
// task into the grid, so the drop onto the sidebar would silently no-op.
function unscheduleTask(taskId: number) {
	updateTask({id: taskId, startDate: null, endDate: null, dueDate: null})
}

function onUpdateBlock({taskId, start, end}: {taskId: number, start: Date | null, end: Date | null}) {
	// Blocks only emit null dates when dropped on the sidebar → unschedule.
	if (start === null && end === null) {
		unscheduleTask(taskId)
		return
	}
	updateTask({id: taskId, startDate: start, endDate: end})
}

function onDropAllDay({taskId, day}: {taskId: number, day: Date}) {
	const midnight = dayjs(day).startOf('day').toDate()
	const task = findTask(taskId)
	// A due-only task dropped on a day keeps its due-only nature: the drag
	// moves the deadline instead of converting it to an all-day span.
	if (task && !task.startDate && !task.endDate && task.dueDate) {
		updateTask({id: taskId, dueDate: midnight})
		return
	}
	// All-day = start and end pinned to midnight of that day.
	const partial: Parameters<typeof updateTask>[0] = {id: taskId, startDate: midnight, endDate: midnight}
	if (hasPastDue(task)) {
		partial.dueDate = midnight
	}
	updateTask(partial)
}

const formatTime = usePlannerTimeFormatter()

// The pending create gesture: target dates plus a label shown in the modal.
// null while no create is in flight (and drives whether the modal is mounted).
const createCtx = ref<{startDate: Date, endDate: Date, label: string} | null>(null)

function onCreateTask({day, startMinutes, endMinutes}: {day: Date, startMinutes: number, endMinutes: number | null}) {
	const base = dayjs(day).startOf('day')
	const start = base.add(startMinutes, 'minute')
	const end = endMinutes !== null
		? base.add(endMinutes, 'minute')
		: start.add(defaultDurationMinutes.value, 'minute')
	createCtx.value = {
		startDate: start.toDate(),
		endDate: end.toDate(),
		label: `${formatDate(start.toDate(), 'll')} · ${formatTime.value(start.toDate())} – ${formatTime.value(end.toDate())}`,
	}
}

function onCreateAllDay({day}: {day: Date}) {
	const midnight = dayjs(day).startOf('day').toDate()
	createCtx.value = {
		startDate: midnight,
		endDate: midnight,
		label: t('planner.createAllDay', {date: formatDate(day, 'll')}),
	}
}

// AddTask emits one `taskAdded` per line synchronously, so schedule each into
// the same painted slot and close once after the batch (nulling createCtx here
// would drop every task after the first).
function onCreated(task: ITask) {
	const ctx = createCtx.value
	if (!ctx) {
		return
	}
	scheduleTask(task, {startDate: ctx.startDate, endDate: ctx.endDate})
	nextTick(() => createCtx.value = null)
}

function openTask(taskId: number) {
	router.push({
		name: 'task.detail',
		params: {id: taskId},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

// Standalone page: drop any stale project so the app header shows the planner
// title instead of the last visited project.
onMounted(() => baseStore.handleSetCurrentProject({project: null}))

watchEffect(() => setTitle(t('planner.title')))
</script>

<style lang="scss" scoped>
.planner-view {
	display: flex;
	flex-direction: column;
	block-size: calc(100vh - #{$navbar-height} - 1.5rem);
}

.planner-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	flex-wrap: wrap;
	gap: .5rem;
	margin-block-end: .75rem;
}

.planner-heading {
	font-size: 1.4rem;
	font-weight: 700;
}

.toolbar-controls {
	display: flex;
	align-items: center;
	gap: .5rem;
	flex-wrap: wrap;
}

.nav-arrows {
	display: flex;
	gap: .25rem;
}

.range-label {
	font-weight: 600;
	min-inline-size: 11rem;
	text-align: center;
}

// A small inline refresh indicator in the toolbar; override the component's
// large default min sizes (meant for full-page use).
.planner-loading {
	min-block-size: 0 !important;
	min-inline-size: 0 !important;
	inline-size: 1.75rem;
	block-size: 1.75rem;
}

.planner-error {
	margin-block-end: .75rem;
}

.mode-toggle {
	display: flex;
	gap: .25rem;
}

.zoom-controls {
	display: flex;
	gap: .25rem;
}

.planner-body {
	display: flex;
	gap: .75rem;
	flex: 1 1 auto;
	min-block-size: 0;
}
</style>
