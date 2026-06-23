<template>
	<div
		ref="columnEl"
		class="day-column"
		:class="{'is-drop-target': isDropTarget}"
		:data-day="dayKey"
		@dragover.prevent="isDropTarget = true"
		@dragleave="isDropTarget = false"
		@drop="onDrop"
	>
		<div
			v-for="hour in 24"
			:key="hour"
			class="hour-slot"
			:style="{height: `${pxPerMinute * 60}px`}"
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
}>()

const columnEl = ref<HTMLElement | null>(null)
const isDropTarget = ref(false)
const now = ref(new Date())

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
onBeforeUnmount(() => clearInterval(timer))

function onDrop(event: DragEvent) {
	isDropTarget.value = false
	const taskId = Number(event.dataTransfer?.getData('text/plain'))
	if (!taskId || !columnEl.value) {
		return
	}

	const rect = columnEl.value.getBoundingClientRect()
	const offsetY = event.clientY - rect.top
	const rawMinutes = offsetY / props.pxPerMinute
	const snapped = Math.round(rawMinutes / props.slotMinutes) * props.slotMinutes
	const minutes = Math.min(Math.max(snapped, 0), 24 * 60 - props.slotMinutes)

	emit('dropTask', {taskId, minutes})
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
