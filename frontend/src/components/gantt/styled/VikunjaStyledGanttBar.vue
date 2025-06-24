<template>
	<GanttBarPrimitive
		:model="model"
		:timeline-start="timelineStart"
		:timeline-end="timelineEnd"
		:on-move="onMove"
		as="rect"
	>
		<template #default="{ dragging, selected, focused }">
			<rect
				:x="computeX(model.start)"
				:width="computeWidth(model)"
				y="4"
				height="16"
				rx="3"
				:fill="dragging ? 'var(--bar-bg-drag)' : (selected ? 'var(--bar-bg-active)' : 'var(--bar-bg)')"
				:stroke="focused ? 'var(--bar-stroke-focus)' : 'none'"
				stroke-width="2"
			/>
			<text
				:x="computeX(model.start) + 4"
				y="16"
				class="small-label"
			>
				{{ model.meta?.label || model.id }}
			</text>
		</template>
	</GanttBarPrimitive>
</template>

<script setup lang="ts">
import GanttBarPrimitive from '../primitives/GanttBarPrimitive.vue'
import type {GanttBarModel} from '@/composables/useGanttBar'
const props = defineProps<{ model:GanttBarModel; timelineStart:Date; timelineEnd:Date; onMove:(id:string,start:Date,end:Date)=>void }>()
function computeX(date: Date) {
  return (date.getTime() - props.timelineStart.getTime()) / (1000*60*60*24) * 24
}
function computeWidth(bar: GanttBarModel) {
  const diff = (bar.end.getTime() - bar.start.getTime()) / (1000*60*60*24)
  return diff * 24
}
</script>

<style scoped>
.small-label { font-family:sans-serif; font-size:10px; font-weight:500; fill:var(--text-on-bar); pointer-events:none; }
</style>
