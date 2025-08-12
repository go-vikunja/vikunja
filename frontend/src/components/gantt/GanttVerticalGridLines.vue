<template>
	<div class="gantt-grid-lines">
		<svg
			class="gantt-vertical-lines"
			:width="totalWidth"
			:height="height"
			xmlns="http://www.w3.org/2000/svg"
		>
			<line
				v-for="(date, index) in timelineData"
				:key="date.toISOString()"
				:x1="index * dayWidthPixels"
				:y1="0"
				:x2="index * dayWidthPixels"
				:y2="height"
				stroke="var(--grey-400)"
				stroke-width="0.5"
				opacity="0.6"
			/>
		</svg>
	</div>
</template>

<script setup lang="ts">
defineProps<{
	timelineData: Date[]
	totalWidth: number
	height: number
	dayWidthPixels: number
}>()
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
</style>
