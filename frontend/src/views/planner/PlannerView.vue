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
			</div>
		</header>

		<div class="planner-body">
			<PlannerSidebar
				v-model:filter="sidebarFilter"
				v-model:sort="sidebarSort"
				:tasks="sidebarTasks"
				@openTask="openTask"
				@unschedule="taskId => updateTask({id: taskId, startDate: null, endDate: null})"
			/>
			<CalendarGrid
				:days="days"
				:tasks="visibleGridTasks"
				:slot-minutes="settings.slotMinutes"
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
import {computed, onMounted, ref, watchEffect} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useStorage} from '@vueuse/core'
import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import BaseButton from '@/components/base/BaseButton.vue'
import PlannerSidebar from './PlannerSidebar.vue'
import PlannerSettings from './PlannerSettings.vue'
import PlannerCreateTaskModal from './PlannerCreateTaskModal.vue'
import CalendarGrid from './grid/CalendarGrid.vue'

import {setTitle} from '@/helpers/setTitle'
import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useCalendarSettings} from './helpers/useCalendarSettings'
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
		return dayjs(anchor.value).format('LL')
	}
	const first = dayjs(days.value[0])
	const last = dayjs(days.value[days.value.length - 1])
	return `${first.format('ll')} – ${last.format('ll')}`
})

const {sidebarTasks, gridTasks, updateTask, scheduleTask} = usePlannerTasks(range, sidebarFilter, sidebarSort)

const visibleGridTasks = computed(() =>
	[...gridTasks.value.values()].filter(task => settings.value.showDone || !task.done),
)

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

function onDropTask({taskId, minutes, day}: {taskId: number, minutes: number, day: Date}) {
	const start = dayjs(day).startOf('day').add(minutes, 'minute')
	const end = start.add(settings.value.defaultDurationMinutes, 'minute')
	updateTask({id: taskId, startDate: start.toDate(), endDate: end.toDate()})
}

function onUpdateBlock({taskId, start, end}: {taskId: number, start: Date | null, end: Date | null}) {
	updateTask({id: taskId, startDate: start, endDate: end})
}

function onDropAllDay({taskId, day}: {taskId: number, day: Date}) {
	// All-day = start and end pinned to midnight of that day.
	const midnight = dayjs(day).startOf('day').toDate()
	updateTask({id: taskId, startDate: midnight, endDate: midnight})
}

const {store: timeFormat} = useTimeFormat()

function formatTime(date: dayjs.Dayjs): string {
	return date.format(timeFormat.value === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'h:mm A')
}

// The pending create gesture: target dates plus a label shown in the modal.
// null while no create is in flight (and drives whether the modal is mounted).
const createCtx = ref<{startDate: Date, endDate: Date, label: string} | null>(null)

function onCreateTask({day, startMinutes, endMinutes}: {day: Date, startMinutes: number, endMinutes: number | null}) {
	const base = dayjs(day).startOf('day')
	const start = base.add(startMinutes, 'minute')
	const end = endMinutes !== null
		? base.add(endMinutes, 'minute')
		: start.add(settings.value.defaultDurationMinutes, 'minute')
	createCtx.value = {
		startDate: start.toDate(),
		endDate: end.toDate(),
		label: `${start.format('ll')} · ${formatTime(start)} – ${formatTime(end)}`,
	}
}

function onCreateAllDay({day}: {day: Date}) {
	const midnight = dayjs(day).startOf('day').toDate()
	createCtx.value = {
		startDate: midnight,
		endDate: midnight,
		label: t('planner.createAllDay', {date: dayjs(day).format('ll')}),
	}
}

function onCreated(task: ITask) {
	if (!createCtx.value) {
		return
	}
	scheduleTask(task, {startDate: createCtx.value.startDate, endDate: createCtx.value.endDate})
	createCtx.value = null
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
