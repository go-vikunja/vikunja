<template>
	<GanttBarPrimitive
		:model="model"
		:timeline-start="timelineStart"
		:timeline-end="timelineEnd"
		:on-move="onMove"
		:on-double-click="onDoubleClick"
		as="rect"
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

const props = defineProps<{ model:GanttBarModel; timelineStart:Date; timelineEnd:Date; onMove:(id:string,start:Date,end:Date)=>void; onDoubleClick?:(model:GanttBarModel)=>void }>()

function computeX(date: Date) {
	return (date.getTime() - props.timelineStart.getTime()) / (1000*60*60*24) * 24
}

function computeWidth(bar: GanttBarModel) {
	const diff = (bar.end.getTime() - bar.start.getTime()) / (1000*60*60*24)
	return diff * 24
}

function getBarFill(dragging: boolean, selected: boolean) {
	if (dragging) return 'var(--bar-bg-drag)'
	if (selected) return 'var(--bar-bg-active)'
	
	// Use task color if available and has actual dates
	if (props.model.meta?.hasActualDates && props.model.meta?.color) {
		return props.model.meta.color
	}
	
	// Default colors
	if (props.model.meta?.hasActualDates) {
		return 'var(--primary)'
	}
	
	return 'var(--bar-bg)'
}

function getBarStroke(focused: boolean) {
	if (focused) return 'var(--bar-stroke-focus)'
	if (!props.model.meta?.hasActualDates) return 'var(--grey-300)'
	return 'none'
}

function getStrokeWidth(focused: boolean) {
	if (focused) return '2'
	if (!props.model.meta?.hasActualDates) return '2'
	return '0'
}

function getTextColor() {
	// For tasks without actual dates, use default text color
	if (!props.model.meta?.hasActualDates) {
		return 'var(--text-on-bar)'
	}
	
	// For tasks with color, determine text color based on background
	if (props.model.meta?.color) {
		return colorIsDark(props.model.meta.color) ? 'var(--grey-800)' : 'white'
	}
	
	// Default for primary color background
	return 'white'
}
</script>

<style scoped>
.small-label { font-family:sans-serif; font-size:10px; font-weight:500; fill:var(--text-on-bar); pointer-events:none; }
</style>
