<template>
	<div
		class="calendar-grid"
		@wheel="onWheel"
	>
		<div
			class="grid-head"
			:style="headerStyle"
			@touchstart.passive="onTouchStart"
			@touchend.passive="onTouchEnd"
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
				@dblclick="onAllDayDblClick($event, day)"
				@pointerdown="onAllDayPointerDown($event, day)"
			>
				<button
					v-for="item in allDayItemsByDay.get(formatDayKey(day)) ?? []"
					:key="item.task.id"
					class="all-day-chip"
					:class="{'is-done': item.task.done, 'is-ghost': item.isGhost}"
					:style="{'--chip-color': taskColor(item.task), '--chip-text': getTextColor(taskColor(item.task))}"
					:title="chipTitle(item.task)"
					:draggable="!item.isGhost"
					@dragstart="onChipDragStart($event, item.task)"
					@click="emit('openTask', item.task.id)"
				>
					{{ item.task.title }}
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
						@createTask="payload => emit('createTask', {...payload, day})"
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
import {getTextColor} from '@/helpers/color/getTextColor'
import CalendarDayColumn from './CalendarDayColumn.vue'
import {allDayTasksForDay, type AllDayItem} from '../helpers/dayLayout'
import {plannerTaskColor} from '../helpers/taskColor'

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
	createTask: [payload: {day: Date, startMinutes: number, endMinutes: number | null}]
	createAllDay: [payload: {day: Date}]
	navigate: [delta: number]
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

function onAllDayCell(target: EventTarget | null): boolean {
	return !(target as HTMLElement)?.closest?.('.all-day-chip')
}

// Desktop: double-click an empty all-day cell to create an all-day task.
function onAllDayDblClick(event: MouseEvent, day: Date) {
	if (onAllDayCell(event.target)) {
		emit('createAllDay', {day})
	}
}

// Touch/pen: long-press an empty all-day cell to create.
let allDayTimer: ReturnType<typeof setTimeout> | undefined
let allDayMove: ((e: PointerEvent) => void) | null = null
let allDayEnd: ((e: PointerEvent) => void) | null = null
function detachAllDay() {
	clearTimeout(allDayTimer)
	if (allDayMove) {
		document.removeEventListener('pointermove', allDayMove)
	}
	if (allDayEnd) {
		document.removeEventListener('pointerup', allDayEnd)
	}
	allDayMove = null
	allDayEnd = null
}
function onAllDayPointerDown(event: PointerEvent, day: Date) {
	if (event.pointerType === 'mouse' || !onAllDayCell(event.target)) {
		return
	}
	const startX = event.clientX
	const startY = event.clientY
	let moved = false
	const onMove = (e: PointerEvent) => {
		if (Math.abs(e.clientX - startX) > 10 || Math.abs(e.clientY - startY) > 10) {
			moved = true
			detachAllDay()
		}
	}
	allDayMove = onMove
	allDayEnd = detachAllDay
	document.addEventListener('pointermove', onMove)
	document.addEventListener('pointerup', detachAllDay)
	allDayTimer = setTimeout(() => {
		detachAllDay()
		if (!moved) {
			emit('createAllDay', {day})
		}
	}, 500)
}

// Horizontal wheel/trackpad scroll slides the window a day at a time, with a
// short cooldown so one flick doesn't skip several days.
let navCooldown = false
function navigate(delta: number) {
	if (navCooldown) {
		return
	}
	emit('navigate', delta)
	navCooldown = true
	setTimeout(() => navCooldown = false, 250)
}

function onWheel(event: WheelEvent) {
	if (Math.abs(event.deltaX) <= Math.abs(event.deltaY) || Math.abs(event.deltaX) < 30) {
		return
	}
	event.preventDefault()
	navigate(event.deltaX > 0 ? 1 : -1)
}

// Touch swipe on the day header navigates without colliding with the grid's
// create/paint gestures, which live in the columns below.
let touchStartX = 0
let touchStartY = 0
function onTouchStart(event: TouchEvent) {
	touchStartX = event.touches[0].clientX
	touchStartY = event.touches[0].clientY
}

function onTouchEnd(event: TouchEvent) {
	const dx = event.changedTouches[0].clientX - touchStartX
	const dy = event.changedTouches[0].clientY - touchStartY
	if (Math.abs(dx) > 50 && Math.abs(dx) > Math.abs(dy)) {
		emit('navigate', dx < 0 ? 1 : -1)
	}
}

const pxPerMinute = computed(() => props.pxPerHour / 60)

// Resolve the all-day items per day once per render instead of re-filtering all
// tasks inside the template v-for (each lookup walks recurrences).
const allDayItemsByDay = computed(() => {
	const map = new Map<string, AllDayItem[]>()
	for (const day of props.days) {
		map.set(formatDayKey(day), allDayTasksForDay(props.tasks, day))
	}
	return map
})

// The body has a vertical scrollbar but the header/all-day rows don't; reserve
// the same width on them so the day-column verticals line up.
const headerStyle = computed(() => ({paddingInlineEnd: `${scrollbarWidth.value}px`}))

function measureScrollbar() {
	if (bodyEl.value) {
		scrollbarWidth.value = bodyEl.value.offsetWidth - bodyEl.value.clientWidth
	}
}

function taskColor(task: ITask): string {
	return plannerTaskColor(task.hexColor, projectStore.projects[task.projectId]?.hexColor)
}

// All-day chips are single-line; surface the project name via the tooltip.
function chipTitle(task: ITask): string {
	const project = projectStore.projects[task.projectId]?.title
	return project ? `${task.title} · ${project}` : task.title
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

onBeforeUnmount(() => {
	window.removeEventListener('resize', measureScrollbar)
	detachAllDay()
})

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
	font-size: .82rem;
	color: var(--chip-text);
	background-color: var(--chip-color);
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;

	&.is-done {
		opacity: .6;
		text-decoration: line-through;
	}

	// Projected recurrence: read-only, visually dimmed like timed ghosts.
	&.is-ghost {
		cursor: pointer;
		opacity: .55;
		background-image: repeating-linear-gradient(
			45deg,
			hsla(0, 0%, 100%, .15),
			hsla(0, 0%, 100%, .15) 4px,
			transparent 4px,
			transparent 8px
		);
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
