<template>
	<div class="calendar-grid">
		<div
			class="grid-head"
			:style="headerStyle"
		>
			<div class="axis-gutter" />
			<div
				v-for="day in days"
				:key="day.toISOString()"
				class="day-head"
				:class="{'is-today': isToday(day)}"
			>
				<span class="day-name">{{ formatWeekday(day) }}</span>
				<span class="day-number">{{ day.getDate() }}</span>
			</div>
		</div>

		<div
			class="all-day-row"
			:style="headerStyle"
		>
			<div class="axis-gutter all-day-label">
				{{ $t('planner.allDay') }}
			</div>
			<div
				v-for="day in days"
				:key="day.toISOString()"
				class="all-day-cell"
				:class="{'is-drop-target': allDayDropDay === formatDayKey(day)}"
				:data-day="formatDayKey(day)"
				@dragover.prevent="allDayDropDay = formatDayKey(day)"
				@dragleave="allDayDropDay = null"
				@drop="onAllDayDrop($event, day)"
			>
				<button
					v-for="task in allDayTasksForDay(tasks, day)"
					:key="task.id"
					class="all-day-chip"
					:class="{'is-done': task.done}"
					:style="{'--chip-color': taskColor(task)}"
					draggable="true"
					@dragstart="onChipDragStart($event, task)"
					@click="emit('openTask', task.id)"
				>
					{{ task.title }}
				</button>
			</div>
		</div>

		<div
			ref="bodyEl"
			class="grid-body"
		>
			<div
				class="grid-content"
				:style="{blockSize: `${pxPerHour * 24}px`}"
			>
				<div class="time-axis">
					<div
						v-for="hour in 24"
						:key="hour"
						class="axis-hour"
						:style="{height: `${pxPerMinute * 60}px`}"
					>
						<span>{{ formatHour(hour - 1) }}</span>
					</div>
				</div>
				<div class="day-columns">
					<CalendarDayColumn
						v-for="day in days"
						:key="day.toISOString()"
						:day="day"
						:tasks="tasks"
						:px-per-minute="pxPerMinute"
						:slot-minutes="slotMinutes"
						@openTask="taskId => emit('openTask', taskId)"
						@updateBlock="payload => emit('updateBlock', payload)"
						@dropTask="payload => emit('dropTask', {...payload, day})"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from 'vue'
import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import {useProjectStore} from '@/stores/projects'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'
import CalendarDayColumn from './CalendarDayColumn.vue'
import {allDayTasksForDay} from '../helpers/dayLayout'

const props = defineProps<{
	days: Date[]
	tasks: ITask[]
	slotMinutes: number
	dayStartHour: number
	dayEndHour: number
	pxPerHour: number
	autoFit: boolean
}>()

const emit = defineEmits<{
	openTask: [taskId: number]
	updateBlock: [payload: {taskId: number, start: Date | null, end: Date | null}]
	dropTask: [payload: {taskId: number, minutes: number, day: Date}]
	dropAllDay: [payload: {taskId: number, day: Date}]
	'update:pxPerHour': [value: number]
}>()

const projectStore = useProjectStore()
const {store: timeFormat} = useTimeFormat()
const bodyEl = ref<HTMLElement | null>(null)
const scrollbarWidth = ref(0)
const allDayDropDay = ref<string | null>(null)

function onAllDayDrop(event: DragEvent, day: Date) {
	allDayDropDay.value = null
	const taskId = Number(event.dataTransfer?.getData('text/plain'))
	if (taskId) {
		emit('dropAllDay', {taskId, day})
	}
}

function onChipDragStart(event: DragEvent, task: ITask) {
	event.dataTransfer?.setData('text/plain', String(task.id))
	if (event.dataTransfer) {
		event.dataTransfer.effectAllowed = 'move'
	}
}

const pxPerMinute = computed(() => props.pxPerHour / 60)

// The body has a vertical scrollbar but the header/all-day rows don't; reserve
// the same width on them so the day-column verticals line up.
const headerStyle = computed(() => ({paddingInlineEnd: `${scrollbarWidth.value}px`}))

function measureScrollbar() {
	if (bodyEl.value) {
		scrollbarWidth.value = bodyEl.value.offsetWidth - bodyEl.value.clientWidth
	}
}

function taskColor(task: ITask): string {
	const hex = projectStore.projects[task.projectId]?.hexColor || task.hexColor
	if (!hex) {
		return 'var(--primary)'
	}
	return hex.startsWith('#') ? hex : `#${hex}`
}

// Choose a zoom level so the working-hours window fills the visible grid, then
// scroll to the day start. Off-hours stay reachable by scrolling.
function fitToWorkingHours() {
	if (!props.autoFit || !bodyEl.value) {
		return
	}
	const workingHours = Math.max(props.dayEndHour - props.dayStartHour, 1)
	const height = bodyEl.value.clientHeight
	if (height > 0) {
		emit('update:pxPerHour', Math.min(Math.max(Math.round(height / workingHours), 16), 200))
	}
}

function scrollToDayStart() {
	if (bodyEl.value) {
		bodyEl.value.scrollTop = props.dayStartHour * props.pxPerHour
	}
}

onMounted(() => {
	fitToWorkingHours()
	nextTick(() => {
		scrollToDayStart()
		measureScrollbar()
	})
	window.addEventListener('resize', measureScrollbar)
})

onBeforeUnmount(() => window.removeEventListener('resize', measureScrollbar))

watch(() => [props.dayStartHour, props.dayEndHour, props.days, props.autoFit], () => {
	fitToWorkingHours()
	nextTick(() => {
		scrollToDayStart()
		measureScrollbar()
	})
})

watch(() => props.pxPerHour, () => nextTick(() => {
	scrollToDayStart()
	measureScrollbar()
}))

function isToday(day: Date): boolean {
	return dayjs(day).isSame(dayjs(), 'day')
}

function formatWeekday(day: Date): string {
	return dayjs(day).format('ddd')
}

function formatDayKey(day: Date): string {
	return dayjs(day).format('YYYY-MM-DD')
}

function formatHour(hour: number): string {
	return dayjs().hour(hour).minute(0)
		.format(timeFormat.value === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'h A')
}
</script>

<style lang="scss" scoped>
$gutter-width: 3.5rem;

.calendar-grid {
	display: flex;
	flex-direction: column;
	flex: 1 1 auto;
	min-block-size: 0;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	overflow: hidden;
	background: var(--white);
}

.grid-head,
.all-day-row {
	display: flex;
	border-block-end: 1px solid var(--grey-200);
}

.day-head {
	flex: 1 1 0;
	min-inline-size: 0;
	text-align: center;
	padding: .25rem 0;
	border-inline-start: 1px solid var(--grey-200);

	&.is-today {
		color: var(--primary);
		font-weight: 700;
	}

	.day-name {
		display: block;
		font-size: .75rem;
		text-transform: uppercase;
	}

	.day-number {
		font-size: 1.1rem;
	}
}

.axis-gutter {
	flex: 0 0 $gutter-width;
	inline-size: $gutter-width;
}

.all-day-label {
	font-size: .7rem;
	color: var(--grey-500);
	display: flex;
	align-items: center;
	justify-content: flex-end;
	padding-inline-end: .35rem;
}

.all-day-cell {
	flex: 1 1 0;
	min-inline-size: 0;
	min-block-size: 1.5rem;
	border-inline-start: 1px solid var(--grey-200);
	padding: 2px;
	display: flex;
	flex-direction: column;
	gap: 2px;

	&.is-drop-target {
		background-color: var(--grey-100);
	}
}

.all-day-chip {
	display: block;
	inline-size: 100%;
	text-align: start;
	border: none;
	cursor: grab;
	border-radius: 3px;
	padding: 1px 5px;
	font-size: .72rem;
	color: var(--white);
	background-color: var(--chip-color);
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;

	&.is-done {
		opacity: .6;
		text-decoration: line-through;
	}
}

.grid-body {
	flex: 1 1 auto;
	min-block-size: 0;
	overflow-y: auto;
}

.grid-content {
	display: flex;
}

.time-axis {
	flex: 0 0 $gutter-width;
	inline-size: $gutter-width;
}

.axis-hour {
	position: relative;
	text-align: end;
	padding-inline-end: .35rem;
	box-sizing: border-box;

	span {
		position: relative;
		inset-block-start: -.5em;
		font-size: .7rem;
		color: var(--grey-500);
	}
}

.day-columns {
	display: flex;
	flex: 1 1 auto;
	min-inline-size: 0;
}
</style>
