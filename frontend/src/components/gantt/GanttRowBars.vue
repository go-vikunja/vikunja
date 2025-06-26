<template>
	<svg
		class="gantt-row-bars"
		:width="totalWidth"
		height="40"
		xmlns="http://www.w3.org/2000/svg"
	>
		<GanttBarPrimitive
			v-for="bar in bars"
			:key="bar.id"
			:model="bar"
			:timeline-start="dateFromDate"
			:timeline-end="dateToDate"
			:on-move="handleBarMove"
			:on-double-click="handleBarDoubleClick"
			as="g"
		>
			<template #default="{ dragging, selected, focused }">
				<!-- Main bar -->
				<rect
					:x="getBarX(bar)"
					:y="4"
					:width="getBarWidth(bar)"
					:height="32"
					:rx="4"
					:fill="getBarFill(bar)"
					:stroke="getBarStroke(bar, focused)"
					:stroke-width="getBarStrokeWidth(bar, focused)"
					:stroke-dasharray="!bar.meta?.hasActualDates ? '3,3' : 'none'"
					:style="{ textDecoration: bar.meta?.isDone ? 'line-through' : 'none' }"
					class="gantt-bar"
				/>
				
				<!-- Resize handles (only show when focused/selected) -->
				<g
					v-if="focused || selected"
					class="resize-handles"
				>
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
				</g>
				
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
					:x="getBarX(bar) + 8"
					:y="24"
					class="gantt-bar-text"
					:fill="getBarTextColor(bar)"
					:clip-path="`url(#clip-${bar.id})`"
					:style="{ textDecoration: bar.meta?.isDone ? 'line-through' : 'none' }"
				>
					{{ bar.meta?.label || bar.id }}
				</text>
			</template>
		</GanttBarPrimitive>
	</svg>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import type {GanttBarModel} from '@/composables/useGanttBar'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import GanttBarPrimitive from './primitives/GanttBarPrimitive.vue'

interface Props {
	bars: GanttBarModel[]
	totalWidth: number
	dateFromDate: Date
	dateToDate: Date
	dayWidthPixels: number
}

const props = defineProps<Props>()

const emit = defineEmits<{
	(e: 'updateTask', id: string, start: Date, end: Date): void
	(e: 'openTask', bar: GanttBarModel): void
	(e: 'startResize', bar: GanttBarModel, edge: 'start' | 'end', event: PointerEvent): void
}>()

// Simplified positioning - primitive handles drag state
function computeBarX(startDate: Date) {
	const x = (startDate.getTime() - props.dateFromDate.getTime()) / (1000*60*60*24) * props.dayWidthPixels
	return x
}

function computeBarWidth(bar: GanttBarModel) {
	const diff = (bar.end.getTime() - bar.start.getTime()) / (1000*60*60*24)
	return diff * props.dayWidthPixels
}

const getBarX = computed(() => (bar: GanttBarModel) => {
	return computeBarX(bar.start)
})

const getBarWidth = computed(() => (bar: GanttBarModel) => {
	return computeBarWidth(bar)
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

function getBarStroke(bar: GanttBarModel, focused: boolean) {
	if (focused) return 'var(--bar-stroke-focus)'
	if (!bar.meta?.hasActualDates) return 'var(--grey-400)'
	return 'none'
}

function getBarStrokeWidth(bar: GanttBarModel, focused: boolean) {
	if (focused) return '2'
	if (!bar.meta?.hasActualDates) return '2'
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

function handleBarMove(id: string, newStart: Date, newEnd: Date) {
	// Emit immediately for visual feedback - parent should handle debouncing
	emit('updateTask', id, newStart, newEnd)
}

function handleBarDoubleClick(bar: GanttBarModel) {
	emit('openTask', bar)
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

// Handle dragging and selected states via CSS using data-state from primitive
:deep([data-state*="dragging"]) .gantt-bar {
	opacity: 0.7;
}

:deep([data-state*="selected"]) .gantt-bar {
	opacity: 0.9;
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