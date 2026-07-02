<template>
	<div
		ref="columnEl"
		class="day-column"
		:class="{'is-drop-target': isDropTarget}"
		:data-day="dayKey"
		@dragover.prevent="isDropTarget = true"
		@dragleave="isDropTarget = false"
		@drop="onDrop"
		@dblclick="onDblClick"
		@pointerdown="onCreatePointerDown"
	>
		<div
			v-for="hour in 24"
			:key="hour"
			class="hour-slot"
			:style="{height: `${pxPerMinute * 60}px`}"
		/>

		<div
			v-if="selStart !== null && selEnd !== null"
			class="paint-selection"
			:style="{
				top: `${selStart * pxPerMinute}px`,
				height: `${Math.max(selEnd - selStart, slotMinutes) * pxPerMinute}px`,
			}"
		/>

		<div
			v-if="isToday"
			class="now-line"
			:style="{top: `${nowMinutes * pxPerMinute}px`}"
		>
			<span class="now-dot" />
			<span class="now-bar" />
		</div>

		<CalendarBlock
			v-for="block in blocks"
			:key="block.occurrence.key"
			:occurrence="block.occurrence"
			:col="block.col"
			:cols="block.cols"
			:top-minutes="block.topMinutes"
			:duration-minutes="block.durationMinutes"
			:px-per-minute="pxPerMinute"
			:slot-minutes="slotMinutes"
			@open="taskId => emit('openTask', taskId)"
			@update="payload => emit('updateBlock', payload)"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, onMounted, ref} from 'vue'
import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import CalendarBlock from './CalendarBlock.vue'
import {timedBlocksForDay} from '../helpers/dayLayout'

const props = defineProps<{
	day: Date
	tasks: ITask[]
	pxPerMinute: number
	slotMinutes: number
}>()

const emit = defineEmits<{
	openTask: [taskId: number]
	updateBlock: [payload: {taskId: number, start: Date | null, end: Date | null}]
	dropTask: [payload: {taskId: number, minutes: number}]
	createTask: [payload: {startMinutes: number, endMinutes: number | null}]
}>()

const columnEl = ref<HTMLElement | null>(null)
const isDropTarget = ref(false)
const now = ref(new Date())
const selStart = ref<number | null>(null)
const selEnd = ref<number | null>(null)

const dayKey = computed(() => dayjs(props.day).format('YYYY-MM-DD'))
const blocks = computed(() => timedBlocksForDay(props.tasks, props.day))
const isToday = computed(() => dayjs(props.day).isSame(now.value, 'day'))
const nowMinutes = computed(() => dayjs(now.value).diff(dayjs(now.value).startOf('day'), 'minute'))

// Keep the current-time marker fresh. Only today's column needs the ticker.
let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
	if (dayjs(props.day).isSame(new Date(), 'day')) {
		timer = setInterval(() => now.value = new Date(), 60_000)
	}
})

// Listeners for an in-flight create gesture, torn down on unmount so a mid-drag
// re-render can't leak them onto document.
let longPressTimer: ReturnType<typeof setTimeout> | undefined
let activeMove: ((e: PointerEvent) => void) | null = null
let activeEnd: ((e: PointerEvent) => void) | null = null
function detachCreate() {
	clearTimeout(longPressTimer)
	if (activeMove) {
		document.removeEventListener('pointermove', activeMove)
	}
	if (activeEnd) {
		document.removeEventListener('pointerup', activeEnd)
	}
	activeMove = null
	activeEnd = null
}
onBeforeUnmount(() => {
	clearInterval(timer)
	detachCreate()
})

function onDrop(event: DragEvent) {
	isDropTarget.value = false
	const taskId = Number(event.dataTransfer?.getData('text/plain'))
	if (!taskId || !columnEl.value) {
		return
	}

	emit('dropTask', {taskId, minutes: minutesAt(event.clientY)})
}

// Pixel position within the column → minute-of-day, snapped to the slot grid.
function minutesAt(clientY: number): number {
	if (!columnEl.value) {
		return 0
	}
	const raw = (clientY - columnEl.value.getBoundingClientRect().top) / props.pxPerMinute
	const snapped = Math.round(raw / props.slotMinutes) * props.slotMinutes
	return Math.min(Math.max(snapped, 0), 24 * 60 - props.slotMinutes)
}

function onEmptyArea(target: EventTarget | null): boolean {
	return !(target as HTMLElement)?.closest?.('.calendar-block')
}

// Desktop: double-click an empty slot to create with the default duration.
function onDblClick(event: MouseEvent) {
	if (!onEmptyArea(event.target)) {
		return
	}
	emit('createTask', {startMinutes: minutesAt(event.clientY), endMinutes: null})
}

function onCreatePointerDown(event: PointerEvent) {
	if (!onEmptyArea(event.target) || (event.pointerType === 'mouse' && event.button !== 0)) {
		return
	}
	const startY = event.clientY
	const startMinutes = minutesAt(startY)

	if (event.pointerType === 'mouse') {
		// Click-drag paints a range; a plain click does nothing (dblclick handles it).
		let painting = false
		const onMove = (e: PointerEvent) => {
			if (!painting && Math.abs(e.clientY - startY) > 4) {
				painting = true
			}
			if (painting) {
				const m = minutesAt(e.clientY)
				selStart.value = Math.min(startMinutes, m)
				selEnd.value = Math.max(startMinutes, m)
			}
		}
		const onUp = () => {
			detachCreate()
			if (painting && selStart.value !== null && selEnd.value !== null) {
				const end = Math.max(selEnd.value, selStart.value + props.slotMinutes)
				emit('createTask', {startMinutes: selStart.value, endMinutes: end})
			}
			selStart.value = null
			selEnd.value = null
		}
		activeMove = onMove
		activeEnd = onUp
		document.addEventListener('pointermove', onMove)
		document.addEventListener('pointerup', onUp)
		return
	}

	// Touch/pen: long-press creates at the slot. Bail out if the finger moves
	// first, so the gesture doesn't hijack vertical scrolling of the grid.
	let moved = false
	const onMove = (e: PointerEvent) => {
		if (Math.abs(e.clientY - startY) > 10) {
			moved = true
			detachCreate()
		}
	}
	activeMove = onMove
	activeEnd = detachCreate
	document.addEventListener('pointermove', onMove)
	document.addEventListener('pointerup', detachCreate)
	longPressTimer = setTimeout(() => {
		detachCreate()
		if (!moved) {
			emit('createTask', {startMinutes, endMinutes: null})
		}
	}, 500)
}
</script>

<style lang="scss" scoped>
.day-column {
	position: relative;
	flex: 1 1 0;
	min-inline-size: 0;
	border-inline-start: 1px solid var(--grey-200);

	&.is-drop-target {
		background-color: var(--grey-100);
	}
}

.hour-slot {
	border-block-end: 1px solid var(--grey-200);
	box-sizing: border-box;
}

.paint-selection {
	position: absolute;
	inset-inline: 2px;
	z-index: 14;
	border-radius: 4px;
	background-color: var(--primary);
	opacity: .25;
	pointer-events: none;
}

.now-line {
	position: absolute;
	inset-inline: 0;
	display: flex;
	align-items: center;
	transform: translateY(-50%);
	z-index: 15;
	pointer-events: none;
}

.now-dot {
	flex: 0 0 auto;
	inline-size: 9px;
	block-size: 9px;
	margin-inline-start: -4px;
	border-radius: 50%;
	background-color: var(--danger);
}

.now-bar {
	flex: 1 1 auto;
	block-size: 2px;
	background-color: var(--danger);
}
</style>
