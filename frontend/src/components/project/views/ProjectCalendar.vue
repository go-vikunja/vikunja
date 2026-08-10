<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue'
import {useRouter} from 'vue-router'
import TaskService from '@/services/task'
import type {ITask} from '@/modelTypes/ITask'
import {useAuthStore} from '@/stores/auth'
import {formatDate, useWeekDayFromDate} from '@/helpers/time/formatDate'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'

const props = defineProps<{
	projectId: number
	viewId: number
	isLoadingProject?: boolean
}>()

const router = useRouter()
const authStore = useAuthStore()
const weekDayFromDate = useWeekDayFromDate()
const {store: timeFormat} = useTimeFormat()
const tasks = ref<ITask[]>([])
const loading = ref(false)
const currentMonth = ref(new Date())
const viewMode = ref<'month' | 'week'>('month')

// 0 = Sunday … 6 = Saturday, matching the user setting. Monday is the fallback.
const weekStart = computed(() => authStore.settings.weekStart ?? 1)

function weekdayOffset(d: Date): number {
	return (d.getDay() - weekStart.value + 7) % 7
}

// toISOString() would shift the day for any positive UTC offset, so the range has
// to be formatted from the local calendar fields the grid is built from.
function toDateParam(d: Date): string {
	return formatDate(d, 'YYYY-MM-DD')
}

function startOfPeriod(d: Date): Date {
	if (viewMode.value === 'month') {
		return new Date(d.getFullYear(), d.getMonth(), 1)
	}
	return new Date(d.getFullYear(), d.getMonth(), d.getDate() - weekdayOffset(d))
}
function endOfPeriod(d: Date): Date {
	if (viewMode.value === 'month') {
		return new Date(d.getFullYear(), d.getMonth() + 1, 0)
	}
	const start = startOfPeriod(d)
	return new Date(start.getFullYear(), start.getMonth(), start.getDate() + 6)
}

async function loadTasks() {
	loading.value = true
	try {
		const service = new TaskService()
		const from = startOfPeriod(currentMonth.value)
		const to = endOfPeriod(currentMonth.value)
		const result = await service.getAll(
			{} as ITask,
			{
				projectId: props.projectId,
				filter: `due_date >= "${toDateParam(from)}" && due_date <= "${toDateParam(to)}"`,
				sort_by: ['due_date'],
				order_by: ['asc'],
			},
		)
		tasks.value = result as ITask[]
	} catch (e) {
		console.error('Failed to load calendar tasks', e)
		tasks.value = []
	} finally {
		loading.value = false
	}
}

// Build day grid (month: 6x7, week: 1x7)
const calendarDays = computed(() => {
	const days: Date[] = []
	if (viewMode.value === 'month') {
		const year = currentMonth.value.getFullYear()
		const month = currentMonth.value.getMonth()
		const first = new Date(year, month, 1)
		const pre = new Date(year, month, 1 - weekdayOffset(first))
		for (let i = 0; i < 42; i++) {
			days.push(new Date(pre.getFullYear(), pre.getMonth(), pre.getDate() + i))
		}
	} else {
		const start = startOfPeriod(currentMonth.value)
		for (let i = 0; i < 7; i++) {
			days.push(new Date(start.getFullYear(), start.getMonth(), start.getDate() + i))
		}
	}
	return days
})

function tasksForDay(day: Date): ITask[] {
	const y = day.getFullYear()
	const m = day.getMonth()
	const d = day.getDate()
	return tasks.value.filter(t => {
		if (!t.dueDate) return false
		const dt = new Date(t.dueDate)
		return dt.getFullYear() === y && dt.getMonth() === m && dt.getDate() === d
	})
}

function timeOf(t: ITask): string {
	if (!t.dueDate) return ''
	return formatDate(new Date(t.dueDate), timeFormat.value === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'hh:mm A')
}

function priorityClass(p?: number): string {
	switch (p) {
		case 3: return 'prio-high'
		case 2: return 'prio-medium'
		case 1: return 'prio-low'
		default: return 'prio-none'
	}
}

const monthLabel = computed(() => formatDate(currentMonth.value, 'MMMM YYYY'))
const weekLabel = computed(() => {
	const f = formatDate(startOfPeriod(currentMonth.value), 'DD MMM')
	const t = formatDate(endOfPeriod(currentMonth.value), 'DD MMM YYYY')
	return `${f} - ${t}`
})

// 2024-01-07 was a Sunday, so it anchors the getDay() index the setting uses.
const weekdayLabels = computed(() => Array.from(
	{length: 7},
	(_, i) => weekDayFromDate.value(new Date(2024, 0, 7 + ((weekStart.value + i) % 7))),
))

const isCurrentMonth = (day: Date) => day.getMonth() === currentMonth.value.getMonth()
const isToday = (day: Date) => day.toDateString() === new Date().toDateString()

function openTask(id: number) {
	router.push({name: 'task.detail', params: {id}})
}

function prev() {
	if (viewMode.value === 'month') {
		currentMonth.value = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth() - 1, 1)
	} else {
		currentMonth.value = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth(), currentMonth.value.getDate() - 7)
	}
}
function next() {
	if (viewMode.value === 'month') {
		currentMonth.value = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth() + 1, 1)
	} else {
		currentMonth.value = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth(), currentMonth.value.getDate() + 7)
	}
}
function today() {
	currentMonth.value = new Date()
}

watch(() => [props.projectId, props.viewId, currentMonth.value, viewMode.value], loadTasks, {immediate: false})
onMounted(loadTasks)
</script>

<template>
	<div class="project-calendar">
		<div class="calendar-toolbar">
			<h2>{{ viewMode === 'month' ? monthLabel : weekLabel }}</h2>
			<div class="calendar-nav">
				<button
					class="btn"
					:aria-label="$t('project.calendar.previousPeriod')"
					@click="prev"
				>
					‹
				</button>
				<button
					class="btn"
					@click="today"
				>
					{{ $t('project.calendar.today') }}
				</button>
				<button
					class="btn"
					:aria-label="$t('project.calendar.nextPeriod')"
					@click="next"
				>
					›
				</button>
				<div class="view-toggle">
					<button
						class="btn"
						:class="{active: viewMode === 'month'}"
						@click="viewMode = 'month'"
					>
						{{ $t('project.calendar.month') }}
					</button>
					<button
						class="btn"
						:class="{active: viewMode === 'week'}"
						@click="viewMode = 'week'"
					>
						{{ $t('project.calendar.week') }}
					</button>
				</div>
			</div>
		</div>

		<div
			class="calendar-grid"
			:class="viewMode"
		>
			<div
				v-for="wd in weekdayLabels"
				:key="wd"
				class="calendar-weekday"
			>
				{{ wd }}
			</div>
			<div
				v-for="day in calendarDays"
				:key="day.toISOString()"
				class="calendar-cell"
				:class="{ 'other-period': viewMode==='month' && !isCurrentMonth(day), 'is-today': isToday(day) }"
			>
				<div class="calendar-daynum">
					{{ day.getDate() }}
				</div>
				<div class="calendar-tasks">
					<button
						v-for="task in tasksForDay(day)"
						:key="task.id"
						class="calendar-task"
						:class="[priorityClass(task.priority), { done: task.done }]"
						:title="task.title"
						@click="openTask(task.id)"
					>
						<span
							v-if="!task.done"
							class="task-time"
						>{{ timeOf(task) }}</span>
						<span class="task-title">{{ task.title }}</span>
					</button>
				</div>
			</div>
		</div>

		<p
			v-if="loading"
			class="calendar-loading"
		>
			{{ $t('misc.loading') }}
		</p>
	</div>
</template>

<style scoped>
.project-calendar {
	padding: 1rem;
}
.calendar-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-block-end: 1rem;
	flex-wrap: wrap;
	gap: 0.5rem;
}
.calendar-toolbar h2 {
	font-size: 1.25rem;
	margin: 0;
}
.calendar-nav {
	display: flex;
	gap: 0.5rem;
	align-items: center;
}
.calendar-nav .btn {
	background: var(--grey-100);
	border: 1px solid var(--border);
	color: var(--text);
	border-radius: 8px;
	padding: 0.4rem 0.8rem;
	cursor: pointer;
	font-size: 0.85rem;
}
.calendar-nav .btn:hover {
	border-color: var(--primary);
}
.calendar-nav .btn.active {
	background: var(--primary);
	color: var(--primary-invert);
	border-color: var(--primary);
}
.view-toggle {
	display: flex;
	gap: 0.25rem;
	margin-inline-start: 0.5rem;
}
.calendar-grid {
	display: grid;
	grid-template-columns: repeat(7, 1fr);
	gap: 4px;
}
.calendar-grid.week .calendar-cell {
	min-block-size: 240px;
}
.calendar-weekday {
	text-align: center;
	font-size: 0.75rem;
	color: var(--text-muted);
	padding: 0.5rem 0;
	font-weight: 600;
}
.calendar-cell {
	min-block-size: 96px;
	background: var(--white);
	border: 1px solid var(--border);
	border-radius: 8px;
	padding: 0.35rem;
	overflow: hidden;
	display: flex;
	flex-direction: column;
	gap: 3px;
}
.calendar-cell.other-period {
	opacity: 0.4;
}
.calendar-cell.is-today {
	border-color: var(--primary);
	box-shadow: inset 0 0 0 1px var(--primary);
}
.calendar-daynum {
	font-size: 0.8rem;
	color: var(--text-muted);
	margin-block-end: 2px;
	font-variant-numeric: tabular-nums;
}
.calendar-tasks {
	display: flex;
	flex-direction: column;
	gap: 3px;
}
.calendar-task {
	display: flex;
	align-items: baseline;
	gap: 0.35rem;
	font-size: 0.72rem;
	text-align: start;
	border: none;
	border-inline-start: 3px solid var(--grey-400);
	background: var(--grey-100);
	color: var(--text);
	border-radius: 4px;
	padding: 0.2rem 0.4rem;
	cursor: pointer;
	font-family: inherit;
	inline-size: 100%;
}
.calendar-task:hover {
	background: var(--grey-200);
}
.calendar-task.done {
	text-decoration: line-through;
	opacity: 0.6;
	border-inline-start-color: var(--success);
}
.calendar-task.prio-high {
	border-inline-start-color: var(--danger);
}
.calendar-task.prio-medium {
	border-inline-start-color: var(--warning);
}
.calendar-task.prio-low {
	border-inline-start-color: var(--link);
}
.calendar-task.prio-none {
	border-inline-start-color: var(--grey-400);
}
.task-time {
	font-variant-numeric: tabular-nums;
	font-weight: 600;
	color: var(--text-muted);
	flex-shrink: 0;
}
.task-title {
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}
.calendar-loading {
	color: var(--text-muted);
	font-size: 0.85rem;
}
</style>
