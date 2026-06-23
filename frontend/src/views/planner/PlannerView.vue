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
				@update:pxPerHour="value => pxPerHour = value"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watchEffect} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'

import BaseButton from '@/components/base/BaseButton.vue'
import PlannerSidebar from './PlannerSidebar.vue'
import PlannerSettings from './PlannerSettings.vue'
import CalendarGrid from './grid/CalendarGrid.vue'

import {setTitle} from '@/helpers/setTitle'
import {useAuthStore} from '@/stores/auth'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useCalendarSettings} from './helpers/useCalendarSettings'
import {usePlannerTasks, type PlannerRange} from './helpers/usePlannerTasks'

const router = useRouter()
const {t} = useI18n({useScope: 'global'})
const {settings} = useCalendarSettings()
const authStore = useAuthStore()

const viewMode = ref<'week' | 'day'>('week')
const anchor = ref(new Date())
const sidebarFilter = ref<TaskFilterParams>({filter: '', s: ''} as TaskFilterParams)

const pxPerHour = ref(48)
const userZoomed = ref(false)

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
	const start = settings.value.fullWeek ? startOfWeek(anchor.value) : dayjs(anchor.value).startOf('day')
	return Array.from({length: 7}, (_, i) => start.add(i, 'day').toDate())
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

const {sidebarTasks, gridTasks, updateTask} = usePlannerTasks(range, sidebarFilter)

const visibleGridTasks = computed(() =>
	[...gridTasks.value.values()].filter(task => settings.value.showDone || !task.done),
)

function goPrev() {
	anchor.value = dayjs(anchor.value).subtract(1, viewMode.value).toDate()
}

function goNext() {
	anchor.value = dayjs(anchor.value).add(1, viewMode.value).toDate()
}

function goToday() {
	anchor.value = new Date()
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

function openTask(taskId: number) {
	router.push({
		name: 'task.detail',
		params: {id: taskId},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

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
