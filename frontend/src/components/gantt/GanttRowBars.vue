<template>
	<svg
		class="gantt-row-bars"
		:width="totalWidth"
		height="40"
		xmlns="http://www.w3.org/2000/svg"
		role="img"
		:aria-label="$t('project.gantt.taskBarsForRow', { rowId })"
		:data-row-id="rowId"
	>
		<GanttBarPrimitive
			v-for="bar in bars"
			:key="bar.id"
			:model="bar"
			:timeline-start="dateFromDate"
			:timeline-end="dateToDate"
			:on-update="(id, start, end) => emit('updateTask', id, start, end)"
		>
			<!-- Main bar -->
			<rect
				:x="getBarX(bar)"
				:y="4"
				:width="getBarWidth(bar)"
				:height="32"
				:rx="4"
				:fill="getBarFill(bar)"
				:stroke="getBarStroke(bar)"
				:stroke-width="getBarStrokeWidth(bar)"
				:stroke-dasharray="!bar.meta?.hasActualDates ? '5,5' : 'none'"
				class="gantt-bar"
				role="button"
				:aria-label="$t('project.gantt.taskBarLabel', {
					task: bar.meta?.label || bar.id,
					startDate: bar.start.toLocaleDateString(),
					endDate: bar.end.toLocaleDateString(),
					dateType: bar.meta?.hasActualDates ? $t('project.gantt.scheduledDates') : $t('project.gantt.estimatedDates')
				})"
				:aria-pressed="isRowFocused"
				@pointerdown="handleBarPointerDown(bar, $event)"
			/>

			<!-- Left resize handle -->
			<rect
				:x="getBarX(bar) - RESIZE_HANDLE_OFFSET"
				:y="4"
				:width="6"
				:height="32"
				:rx="3"
				fill="var(--white)"
				stroke="var(--primary)"
				stroke-width="1"
				class="gantt-resize-handle gantt-resize-left"
				role="button"
				:aria-label="$t('project.gantt.resizeStartDate', { task: bar.meta?.label || bar.id })"
				@pointerdown="startResize(bar, 'start', $event)"
			/>

			<!-- Right resize handle -->
			<rect
				:x="getBarX(bar) + getBarWidth(bar) - RESIZE_HANDLE_OFFSET"
				:y="4"
				:width="6"
				:height="32"
				:rx="3"
				fill="var(--white)"
				stroke="var(--primary)"
				stroke-width="1"
				class="gantt-resize-handle gantt-resize-right"
				role="button"
				:aria-label="$t('project.gantt.resizeEndDate', { task: bar.meta?.label || bar.id })"
				@pointerdown="startResize(bar, 'end', $event)"
			/>

			<!-- Task label with clipping -->
			<defs>
				<clipPath :id="`clip-${bar.id}`">
					<rect
						:x="getBarX(bar) + 2"
						:y="4"
						:width="getBarWidth(bar) - 4"
						:height="32"
						:rx="4"
					/>
				</clipPath>
			</defs>
			<text
				:x="getBarTextX(bar)"
				:y="24"
				class="gantt-bar-text"
				:fill="getBarTextColor(bar)"
				:clip-path="`url(#clip-${bar.id})`"
				aria-hidden="true"
			>
				{{ bar.meta?.label || bar.id }}
			</text>
		</GanttBarPrimitive>
	</svg>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import dayjs from 'dayjs'

import type {GanttBarModel} from '@/composables/useGanttBar'
import {getTextColor, LIGHT} from '@/helpers/color/getTextColor'
import {MILLISECONDS_A_DAY} from '@/constants/date'

import GanttBarPrimitive from './primitives/GanttBarPrimitive.vue'

const props = defineProps<{
	bars: GanttBarModel[]
	totalWidth: number
	dateFromDate: Date
	dateToDate: Date
	dayWidthPixels: number
	isDragging: boolean
	isResizing: boolean
	dragState: {
		barId: string
		startX: number
		originalStart: Date
		originalEnd: Date
		currentDays: number
		edge?: 'start' | 'end'
	} | null
	focusedRow: string | null
	focusedCell: number | null
	rowId: string
}>()

const emit = defineEmits<{
	(e: 'barPointerDown', bar: GanttBarModel, event: PointerEvent): void
	(e: 'startResize', bar: GanttBarModel, edge: 'start' | 'end', event: PointerEvent): void
	(e: 'updateTask', id: string, newStart: Date, newEnd: Date): void
}>()

const RESIZE_HANDLE_OFFSET = 3

function addDays(dateOrValue: Date | string | number, days: number): Date {
	const date = new Date(dateOrValue)
	const newDate = new Date(date)
	newDate.setDate(newDate.getDate() + days)
	return newDate
}

const isRowFocused = computed(() => props.focusedRow === props.rowId)

function computeBarX(startDate: Date) {
	const daysDiff = dayjs(startDate).diff(dayjs(props.dateFromDate), 'day')
	const x = daysDiff * props.dayWidthPixels
	return x
}

function getDaysDifference(startDate: Date, endDate: Date): number {
	return Math.floor((endDate.getTime() - startDate.getTime()) / MILLISECONDS_A_DAY)
}

function computeBarWidth(bar: GanttBarModel) {
	const diff = getDaysDifference(bar.start, bar.end)
	const width = diff * props.dayWidthPixels
	return width
}

const originalEndX = computed(() => props.dragState?.originalEnd 
	? computeBarX(props.dragState.originalEnd) 
	: 0)
const originalStartX = computed(() => props.dragState?.originalStart 
	? computeBarX(props.dragState.originalStart) 
	: 0)

const getBarX = computed(() => (bar: GanttBarModel) => {
	if (props.isDragging && props.dragState?.barId === bar.id) {
		const offset = props.dragState.currentDays * props.dayWidthPixels
		return originalStartX.value + offset
	}

	if (props.isResizing && props.dragState?.barId === bar.id && props.dragState.edge === 'start') {
		const newStart = addDays(props.dragState.originalStart, props.dragState.currentDays)
		return computeBarX(newStart)
	}
	return computeBarX(bar.start)
})

const getBarWidth = computed(() => (bar: GanttBarModel) => {
	if (props.isResizing && props.dragState?.barId === bar.id) {
		if (props.dragState.edge === 'start') {
			const newStart = addDays(props.dragState.originalStart, props.dragState.currentDays)
			const newStartX = computeBarX(newStart)
			return Math.max(0, originalEndX.value - newStartX)
		} else {
			const newEnd = addDays(props.dragState.originalEnd, props.dragState.currentDays)
			const newEndX = computeBarX(newEnd)
			return Math.max(0, newEndX - originalStartX.value)
		}
	}
	return computeBarWidth(bar)
})

const getBarTextX = computed(() => (bar: GanttBarModel) => {
	return getBarX.value(bar) + 8
})

function getBarFill(bar: GanttBarModel) {
	if (bar.meta?.hasActualDates) {
		if (bar.meta?.color) {
			return bar.meta.color
		}
		return 'var(--primary)'
	}

	return 'var(--grey-100)'
}

function getBarStroke(bar: GanttBarModel) {
	if (!bar.meta?.hasActualDates) {
		return 'var(--grey-300)' // Gray for dashed border
	}
	return 'none'
}

function getBarStrokeWidth(bar: GanttBarModel) {
	if (!bar.meta?.hasActualDates) {
		return '2'
	}
	return '0'
}

function getBarTextColor(bar: GanttBarModel) {
	if (!bar.meta?.hasActualDates) {
		return 'var(--grey-800)'
	}

	if (bar.meta?.color) {
		return getTextColor(bar.meta.color)
	}

	return LIGHT
}

function handleBarPointerDown(bar: GanttBarModel, event: PointerEvent) {
	emit('barPointerDown', bar, event)
}

function startResize(bar: GanttBarModel, edge: 'start' | 'end', event: PointerEvent) {
	emit('startResize', bar, edge, event)
}
</script>

<style scoped lang="scss">
.gantt-row-bars {
	position: absolute;
	inset-block-start: 0;
	inset-inline-start: 0;
	pointer-events: none;
	z-index: 4;

	.gantt-bar {
		cursor: grab;
		pointer-events: all;

		&:hover {
			opacity: 0.8;
		}

		&:active {
			cursor: grabbing;
		}
	}

	:deep(text) {
		pointer-events: none;
		user-select: none;
	}
}

.gantt-bar-text {
	font-size: .85rem;
	pointer-events: none;
	user-select: none;
}

:deep(.gantt-resize-handle) {
	cursor: col-resize !important;
	opacity: 0;
	transition: opacity 0.2s ease;
	pointer-events: all; // Ensure they receive pointer events
}

// Show resize handles on bar hover
:deep(g:hover) .gantt-resize-handle {
	opacity: 0.8;

	&:hover {
		opacity: 1;
		cursor: inherit; // Use the specific cursor defined above
	}
}

// Focus styles for task bars
:deep(g[role="slider"]:focus) {
	outline: none; // Remove default browser outline
	
	.gantt-bar {
		stroke: var(--primary) !important;
		stroke-width: 3 !important;
	}
}
</style>
