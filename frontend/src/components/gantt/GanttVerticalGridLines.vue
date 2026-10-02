<template>
	<div class="gantt-grid-lines">
		<svg
			class="gantt-vertical-lines"
			:width="totalWidth"
			:height="height"
			xmlns="http://www.w3.org/2000/svg"
		>
			<!-- Weekend shading -->
			<rect
				v-for="band in weekendBands"
				:key="`weekend-${band.x}`"
				:x="band.x"
				:y="0"
				:width="band.width"
				:height="height"
				class="gantt-weekend"
			/>
			<line
				v-for="x in lineXs"
				:key="x"
				:x1="x"
				:y1="0"
				:x2="x"
				:y2="height"
				stroke="var(--grey-400)"
				stroke-width="0.5"
				opacity="0.6"
			/>
			<!-- Today -->
			<line
				v-if="todayX !== null"
				:x1="todayX"
				:y1="0"
				:x2="todayX"
				:y2="height"
				class="gantt-today-line"
			/>
		</svg>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {gridLineXs, type GanttZoom, type WeekendBand} from '@/helpers/ganttZoom'

const props = withDefaults(defineProps<{
	timelineData: Date[]
	totalWidth: number
	height: number
	dayWidthPixels: number
	zoom?: GanttZoom
	weekendBands?: WeekendBand[]
	// x of the current moment, null when it is not in the shown range
	todayX?: number | null
}>(), {
	zoom: 'day',
	weekendBands: () => [],
	todayX: null,
})

const lineXs = computed(() => gridLineXs(props.timelineData, props.dayWidthPixels, props.zoom))
</script>

<style scoped lang="scss">
.gantt-grid-lines {
	position: absolute;
	inset-inline-start: 0;
	z-index: 1;
	pointer-events: none;
}

.gantt-vertical-lines {
	position: absolute;
	inset: 0;
}

.gantt-weekend {
	fill: var(--grey-200);
	opacity: 0.45;
}

.gantt-today-line {
	stroke: var(--danger);
	stroke-width: 1.5;
	stroke-dasharray: 4 3;
	opacity: 0.9;
}
</style>
