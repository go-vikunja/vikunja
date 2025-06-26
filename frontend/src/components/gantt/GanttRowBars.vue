<template>
	<svg
		class="gantt-row-bars"
		:width="totalWidth"
		height="40"
		xmlns="http://www.w3.org/2000/svg"
	>
		<g
			v-for="bar in bars"
			:key="bar.id"
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
				@pointerdown="handleBarPointerDown(bar, $event)"
			/>
			
			<!-- Left resize handle -->
			<rect
				:x="getBarX(bar) - 3"
				:y="4"
				:width="6"
				:height="32"
				:rx="3"
				fill="var(--white)"
				stroke="var(--primary)"
				stroke-width="1"
				class="gantt-resize-handle gantt-resize-left"
				@pointerdown="startResize(bar, 'start', $event)"
			/>
			
			<!-- Right resize handle -->
			<rect
				:x="getBarX(bar) + getBarWidth(bar) - 3"
				:y="4"
				:width="6"
				:height="32"
				:rx="3"
				fill="var(--white)"
				stroke="var(--primary)"
				stroke-width="1"
				class="gantt-resize-handle gantt-resize-right"
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
			>
				{{ bar.meta?.label || bar.id }}
			</text>
		</g>
	</svg>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import type {GanttBarModel} from '@/composables/useGanttBar'
import {colorIsDark} from '@/helpers/color/colorIsDark'

interface Props {
	bars: GanttBarModel[]
	totalWidth: number
	dateFromDate: Date
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
}

const props = defineProps<Props>()

const emit = defineEmits<{
	(e: 'barPointerDown', bar: GanttBarModel, event: PointerEvent): void
	(e: 'startResize', bar: GanttBarModel, edge: 'start' | 'end', event: PointerEvent): void
}>()

// Direct SVG bar rendering functions
function computeBarX(startDate: Date) {
	const x = (startDate.getTime() - props.dateFromDate.getTime()) / (1000*60*60*24) * props.dayWidthPixels
	return x
}

function computeBarWidth(bar: GanttBarModel) {
	const diff = (bar.end.getTime() - bar.start.getTime()) / (1000*60*60*24)
	const width = diff * props.dayWidthPixels
	return width
}

// Computed properties for dynamic bar positions during drag/resize
const getBarX = computed(() => (bar: GanttBarModel) => {
	if (props.isDragging && props.dragState?.barId === bar.id) {
		const originalX = computeBarX(props.dragState.originalStart)
		const offset = props.dragState.currentDays * props.dayWidthPixels
		return originalX + offset
	}
	if (props.isResizing && props.dragState?.barId === bar.id && props.dragState.edge === 'start') {
		const newStart = new Date(props.dragState.originalStart)
		newStart.setDate(newStart.getDate() + props.dragState.currentDays)
		return computeBarX(newStart)
	}
	return computeBarX(bar.start)
})

const getBarWidth = computed(() => (bar: GanttBarModel) => {
	if (props.isResizing && props.dragState?.barId === bar.id) {
		if (props.dragState.edge === 'start') {
			const newStart = new Date(props.dragState.originalStart)
			newStart.setDate(newStart.getDate() + props.dragState.currentDays)
			const originalEndX = computeBarX(props.dragState.originalEnd)
			const newStartX = computeBarX(newStart)
			return Math.max(0, originalEndX - newStartX)
		} else {
			const newEnd = new Date(props.dragState.originalEnd)
			newEnd.setDate(newEnd.getDate() + props.dragState.currentDays)
			const originalStartX = computeBarX(props.dragState.originalStart)
			const newEndX = computeBarX(newEnd)
			return Math.max(0, newEndX - originalStartX)
		}
	}
	return computeBarWidth(bar)
})

const getBarTextX = computed(() => (bar: GanttBarModel) => {
	return getBarX.value(bar) + 8
})

function getBarFill(bar: GanttBarModel) {
	// For tasks with actual dates
	if (bar.meta?.hasActualDates) {
		// Use task color if available
		if (bar.meta?.color) {
			return bar.meta.color
		}
		// Default to primary color if no task color
		return 'var(--primary)'
	}
	
	// For tasks without dates, use gray
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
	const black = 'var(--grey-800)'
	
	// For tasks without actual dates, use dark text on gray background
	if (!bar.meta?.hasActualDates) {
		return black
	}
	
	// For tasks with actual dates
	if (bar.meta?.color) {
		// Use colorIsDark to determine text color based on background
		return colorIsDark(bar.meta.color) ? black : 'white'
	}
	
	// Default for primary color background (white text)
	return 'white'
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
	top: 0;
	left: 0;
	pointer-events: none;
	z-index: 4;
	
	:deep(.gantt-bar) {
		pointer-events: all;
		cursor: grab;
		
		&:hover {
			opacity: 0.8;
		}
	}

	:deep(text) {
		pointer-events: none;
		user-select: none;
	}
}

// SVG bar styling
.gantt-bar {
	cursor: grab;
	
	&:hover {
		opacity: 0.8;
	}
	
	&:active {
		cursor: grabbing;
	}
}

.gantt-bar-text {
	font-size: .85rem;
	pointer-events: none;
	user-select: none;
}

// Resize handles
:deep(.gantt-resize-handle) {
	cursor: col-resize !important;
	opacity: 0;
	transition: opacity 0.2s ease;
	pointer-events: all; // Ensure they receive pointer events
	
	&:hover {
		opacity: 1;
	}
}

// Show resize handles on bar hover
:deep(g:hover) .gantt-resize-handle {
	opacity: 0.8;
	
	&:hover {
		opacity: 1;
		cursor: inherit; // Use the specific cursor defined above
	}
}
</style>