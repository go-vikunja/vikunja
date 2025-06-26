<template>
	<GanttBarPrimitive
		:model="model"
		:timeline-start="timelineStart"
		:timeline-end="timelineEnd"
		:on-move="onMove"
		:on-double-click="onDoubleClick"
		as="svg"
	>
		<template #default="{ dragging, selected, focused }">
			<rect
				:x="computeX(model.start)"
				:width="computeWidth(model)"
				y="4"
				height="16"
				rx="3"
				:fill="getBarFill(dragging, selected)"
				:stroke="getBarStroke(focused)"
				:stroke-width="getStrokeWidth(focused)"
				:stroke-dasharray="!model.meta?.hasActualDates ? '3,3' : 'none'"
				:style="{ textDecoration: model.meta?.isDone ? 'line-through' : 'none' }"
			/>
			<text
				:x="computeX(model.start) + 4"
				y="16"
				class="small-label"
				:fill="getTextColor()"
				:style="{ textDecoration: model.meta?.isDone ? 'line-through' : 'none' }"
			>
				{{ model.meta?.label || model.id }}
			</text>
		</template>
	</GanttBarPrimitive>
</template>

<script setup lang="ts">
import GanttBarPrimitive from '../primitives/GanttBarPrimitive.vue'
import type {GanttBarModel} from '@/composables/useGanttBar'
import {colorIsDark} from '@/helpers/color/colorIsDark'

const props = defineProps<{ 
	model:GanttBarModel; 
	timelineStart:Date; 
	timelineEnd:Date; 
	onMove:(id:string,start:Date,end:Date)=>void; 
	onDoubleClick?:(model:GanttBarModel)=>void 
}>()
const PIXELS_PER_DAY = 30
const MILLISECONDS_PER_DAY = 1000 * 60 * 60 * 24

function computeX(date: Date) {
	return (date.getTime() - props.timelineStart.getTime()) / MILLISECONDS_PER_DAY * PIXELS_PER_DAY
}

function computeWidth(bar: GanttBarModel) {
	const diff = (bar.end.getTime() - bar.start.getTime()) / MILLISECONDS_PER_DAY
	return diff * PIXELS_PER_DAY
}

function getBarFill(dragging: boolean, selected: boolean) {
	if (dragging) return 'var(--bar-bg-drag)' // Blue for dragging
	if (selected) return 'var(--bar-bg-active)' // Darker blue for selected
	
	// Use task color if available and has actual dates
	if (props.model.meta?.hasActualDates && props.model.meta?.color) {
		return props.model.meta.color
	}
	
	// Default colors - use CSS variables for theming support
	if (props.model.meta?.hasActualDates) {
		return 'var(--bar-bg)' // Primary bar color
	}
	
	return 'var(--grey-300)' // Light gray for tasks without dates
}

function getBarStroke(focused: boolean) {
	if (focused) return 'var(--bar-stroke-focus)' // Primary color for focus
	if (!props.model.meta?.hasActualDates) return 'var(--grey-400)' // Gray for dashed border
	return 'none'
}

function getStrokeWidth(focused: boolean) {
	if (focused) return '2'
	if (!props.model.meta?.hasActualDates) return '2'
	return '0'
}

function getTextColor() {
	// For tasks without actual dates, use dark text
	if (!props.model.meta?.hasActualDates) {
		return '#2c3e50'
	}
	
	// For tasks with color, determine text color based on background
	if (props.model.meta?.color) {
		return colorIsDark(props.model.meta.color) ? '#2c3e50' : 'white'
	}
	
	// Default for primary color background (white text on green)
	return 'white'
}
</script>

<style scoped>
.small-label { font-family:sans-serif; font-size:10px; font-weight:500; fill:var(--text-on-bar); pointer-events:none; }
</style>
