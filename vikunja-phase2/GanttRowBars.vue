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
			<!-- Gradient definitions for partial-date bars -->
			<defs v-if="bar.meta?.dateType === 'startOnly' || bar.meta?.dateType === 'endOnly'">
				<linearGradient
					:id="`gradient-${bar.id}`"
					x1="0"
					y1="0"
					x2="1"
					y2="0"
				>
					<stop
						v-if="bar.meta?.dateType === 'endOnly'"
						offset="0%"
						:stop-color="getBarFill(bar)"
						stop-opacity="0"
					/>
					<stop
						v-if="bar.meta?.dateType === 'endOnly'"
						offset="40%"
						:stop-color="getBarFill(bar)"
						stop-opacity="1"
					/>
					<stop
						v-if="bar.meta?.dateType === 'startOnly'"
						offset="60%"
						:stop-color="getBarFill(bar)"
						stop-opacity="1"
					/>
					<stop
						v-if="bar.meta?.dateType === 'startOnly'"
						offset="100%"
						:stop-color="getBarFill(bar)"
						stop-opacity="0"
					/>
				</linearGradient>
			</defs>

			<!-- Overdue stripe pattern -->
			<defs v-if="bar.meta?.isOverdue">
				<pattern
					:id="`overdue-stripe-${bar.id}`"
					width="8"
					height="8"
					patternUnits="userSpaceOnUse"
					patternTransform="rotate(45)"
				>
					<rect width="8" height="8" :fill="getBarFill(bar)" />
					<rect width="3" height="8" fill="rgba(0,0,0,0.2)" />
				</pattern>
			</defs>

			<!-- Main bar -->
			<rect
				:x="getBarX(bar)"
				:y="4"
				:width="getBarWidth(bar)"
				:height="32"
				:rx="4"
				:fill="bar.meta?.isOverdue ? `url(#overdue-stripe-${bar.id})` : getBarFillAttr(bar)"
				:opacity="bar.meta?.isDone ? 0.5 : 1"
				:stroke="bar.meta?.isOverdue ? '#e74c3c' : getBarStroke(bar)"
				:stroke-width="bar.meta?.isOverdue ? '2' : getBarStrokeWidth(bar)"
				:stroke-dasharray="isDateless(bar) ? '5,5' : 'none'"
				class="gantt-bar"
				role="button"
				:aria-label="getBarAriaLabel(bar)"
				:aria-pressed="isRowFocused"
				@pointerdown="handleBarPointerDown(bar, $event)"
			>
				<title v-if="bar.meta?.isOverdue">
					{{ getOverdueTooltip(bar) }}
				</title>
			</rect>

			<!-- Overdue left-arrow indicator -->
			<g v-if="bar.meta?.isOverdue">
				<polygon
					:points="getOverdueArrowPoints(bar)"
					fill="#e74c3c"
					class="overdue-arrow"
				/>
				<text
					:x="getBarX(bar) + 16"
					:y="24"
					fill="#e74c3c"
					font-size="11"
					font-weight="bold"
					class="overdue-label"
				>
					⏰ {{ getOverdueText(bar) }}
				</text>
			</g>

			<!-- Left resize handle (hidden for endOnly and overdue bars) -->
			<rect
				v-if="bar.meta?.dateType !== 'endOnly' && !bar.meta?.isOverdue"
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

			<!-- Right resize handle (hidden for startOnly and overdue bars) -->
			<rect
				v-if="bar.meta?.dateType !== 'startOnly' && !bar.meta?.isOverdue"
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
				v-if="!bar.meta?.isOverdue"
				:x="getBarTextX(bar)"
				:y="24"
				:text-anchor="bar.meta?.dateType === 'endOnly' ? 'end' : 'start'"
				class="gantt-bar-text"
				:fill="getBarTextColor(bar)"
				:text-decoration="bar.meta?.isDone ? 'line-through' : 'none'"
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
import {useI18n} from 'vue-i18n'

import type {GanttBarModel} from '@/composables/useGanttBar'
import {getTextColor, LIGHT} from '@/helpers/color/getTextColor'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import {roundToNaturalDayBoundary} from '@/helpers/time/roundToNaturalDayBoundary'

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

const {t} = useI18n({useScope: 'global'})

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
	return Math.ceil(
		(roundToNaturalDayBoundary(endDate).getTime() - roundToNaturalDayBoundary(startDate, true).getTime()) /
MILLISECONDS_A_DAY,
	)
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
	if (bar.meta?.dateType === 'endOnly') {
		return getBarX.value(bar) + getBarWidth.value(bar) - 8
	}
	// When the bar starts before the visible range, clamp text to the left edge
	// so the title remains visible within the visible portion of the bar.
	return Math.max(getBarX.value(bar) + 8, 8)
})

function isPartialDate(bar: GanttBarModel) {
	return bar.meta?.dateType === 'startOnly' || bar.meta?.dateType === 'endOnly'
}

function isDateless(bar: GanttBarModel) {
	return !bar.meta?.hasActualDates && !isPartialDate(bar)
}

function getBarFill(bar: GanttBarModel) {
	// Partial dates still have "actual" dates on one side — use the task color
	if (isPartialDate(bar)) {
		if (bar.meta?.color) {
			return bar.meta.color
		}
		return 'var(--primary)'
	}

	if (bar.meta?.hasActualDates) {
		if (bar.meta?.color) {
			return bar.meta.color
		}
		return 'var(--primary)'
	}

	return 'var(--grey-100)'
}

function getBarFillAttr(bar: GanttBarModel): string {
	if (isPartialDate(bar)) {
		return `url(#gradient-${bar.id})`
	}
	return getBarFill(bar)
}

function getBarStroke(bar: GanttBarModel) {
	if (isDateless(bar)) {
		return 'var(--grey-300)' // Gray for dashed border
	}
	return 'none'
}

function getBarStrokeWidth(bar: GanttBarModel) {
	if (isDateless(bar)) {
		return '2'
	}
	return '0'
}

function getBarTextColor(bar: GanttBarModel) {
	if (isDateless(bar)) {
		return 'var(--grey-800)'
	}

	if (bar.meta?.color) {
		return getTextColor(bar.meta.color)
	}

	return LIGHT
}

function getBarAriaLabel(bar: GanttBarModel): string {
	const task = bar.meta?.label || bar.id
	const startDate = bar.start.toLocaleDateString()
	const endDate = bar.end.toLocaleDateString()

	let dateType: string
	if (bar.meta?.dateType === 'startOnly') {
		dateType = t('project.gantt.partialDatesStart')
	} else if (bar.meta?.dateType === 'endOnly') {
		dateType = t('project.gantt.partialDatesEnd')
	} else if (bar.meta?.hasActualDates) {
		dateType = t('project.gantt.scheduledDates')
	} else {
		dateType = t('project.gantt.estimatedDates')
	}

	return t('project.gantt.taskBarLabel', {task, startDate, endDate, dateType})
}

function handleBarPointerDown(bar: GanttBarModel, event: PointerEvent) {
	emit('barPointerDown', bar, event)
}

function startResize(bar: GanttBarModel, edge: 'start' | 'end', event: PointerEvent) {
	emit('startResize', bar, edge, event)
}

function getOverdueArrowPoints(bar: GanttBarModel): string {
	const x = getBarX.value(bar)
	// Left-pointing chevron
	return `${x},10 ${x - 8},20 ${x},30`
}

function getOverdueText(bar: GanttBarModel): string {
	const originalEnd = bar.meta?.originalEnd as Date | undefined
	if (!originalEnd) return 'Overdue'
	const now = new Date()
	const diffMs = now.getTime() - originalEnd.getTime()
	const diffDays = Math.floor(diffMs / MILLISECONDS_A_DAY)
	if (diffDays <= 0) return bar.meta?.label || 'Overdue'
	if (diffDays === 1) return `${bar.meta?.label} (1 day overdue)`
	if (diffDays < 30) return `${bar.meta?.label} (${diffDays}d overdue)`
	const weeks = Math.floor(diffDays / 7)
	if (diffDays < 60) return `${bar.meta?.label} (${weeks}w overdue)`
	const months = Math.floor(diffDays / 30)
	return `${bar.meta?.label} (${months}mo overdue)`
}

function getOverdueTooltip(bar: GanttBarModel): string {
	const origStart = bar.meta?.originalStart as Date | undefined
	const origEnd = bar.meta?.originalEnd as Date | undefined
	if (!origStart || !origEnd) return 'Overdue task'
	return `Overdue: was ${origStart.toLocaleDateString()} – ${origEnd.toLocaleDateString()}`
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

.overdue-arrow {
	pointer-events: none;
	animation: overdue-pulse 2s ease-in-out infinite;
}

.overdue-label {
	pointer-events: none;
	user-select: none;
}

@keyframes overdue-pulse {
	0%, 100% { opacity: 1; }
	50% { opacity: 0.5; }
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
