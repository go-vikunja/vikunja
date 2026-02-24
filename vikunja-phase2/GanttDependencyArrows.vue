<template>
	<svg
		v-if="arrows.length > 0"
		class="gantt-dependency-arrows"
		:width="totalWidth"
		:height="totalHeight"
	>
		<defs>
			<marker
				id="arrowhead"
				markerWidth="8"
				markerHeight="6"
				refX="7"
				refY="3"
				orient="auto"
			>
				<polygon
					points="0 0, 8 3, 0 6"
					:fill="arrowColor"
				/>
			</marker>
		</defs>
		<path
			v-for="(arrow, i) in arrows"
			:key="i"
			:d="arrow.path"
			fill="none"
			:stroke="arrowColor"
			stroke-width="1.5"
			stroke-dasharray="4,2"
			marker-end="url(#arrowhead)"
			class="dependency-arrow"
		/>
	</svg>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import type {GanttBarModel} from '@/composables/useGanttBar'
import type {ITask} from '@/modelTypes/ITask'

const ROW_HEIGHT = 40

const props = defineProps<{
	// All bars, grouped by row index
	barsByRow: GanttBarModel[][]
	// The task map for accessing relations
	tasks: Map<ITask['id'], ITask>
	// Timeline start date
	dateFrom: Date
	// Pixels per day
	dayWidthPixels: number
	// Total width of the chart
	totalWidth: number
}>()

const arrowColor = 'rgba(150, 150, 200, 0.6)'

const totalHeight = computed(() => props.barsByRow.length * ROW_HEIGHT)

// Build a flat map of barId -> { x, y, width } for arrow calculations
const barPositions = computed(() => {
	const map = new Map<string, { x: number, y: number, width: number }>()
	const fromTime = props.dateFrom.getTime()

	props.barsByRow.forEach((bars, rowIndex) => {
		for (const bar of bars) {
			const startDays = (bar.start.getTime() - fromTime) / (1000 * 60 * 60 * 24)
			const endDays = (bar.end.getTime() - fromTime) / (1000 * 60 * 60 * 24)
			const x = startDays * props.dayWidthPixels
			const width = (endDays - startDays) * props.dayWidthPixels

			map.set(bar.id, {
				x,
				y: rowIndex * ROW_HEIGHT + ROW_HEIGHT / 2,
				width: Math.max(width, 1),
			})
		}
	})

	return map
})

// Find all precedes/follows relations and generate arrow paths
const arrows = computed(() => {
	const result: { path: string }[] = []

	for (const [taskId, task] of props.tasks) {
		const precedesTasks = task.relatedTasks?.precedes || []
		for (const target of precedesTasks) {
			const sourcePos = barPositions.value.get(String(taskId))
			const targetPos = barPositions.value.get(String(target.id))

			if (!sourcePos || !targetPos) continue

			// Arrow from source right edge to target left edge
			const sx = sourcePos.x + sourcePos.width // right edge of source
			const sy = sourcePos.y
			const tx = targetPos.x // left edge of target
			const ty = targetPos.y

			// Generate a smooth cubic bezier path
			const dx = tx - sx
			const midX = sx + dx * 0.5

			if (dx > 10) {
				// Normal case: target is to the right
				result.push({
					path: `M ${sx} ${sy} C ${midX} ${sy}, ${midX} ${ty}, ${tx} ${ty}`,
				})
			} else {
				// Target is overlapping or to the left — route around
				const detour = 15
				result.push({
					path: `M ${sx} ${sy} L ${sx + detour} ${sy} Q ${sx + detour + 10} ${sy}, ${sx + detour + 10} ${sy + (ty > sy ? detour : -detour)} L ${sx + detour + 10} ${ty - (ty > sy ? detour : -detour)} Q ${sx + detour + 10} ${ty}, ${tx - detour} ${ty} L ${tx} ${ty}`,
				})
			}
		}
	}

	return result
})
</script>

<style scoped lang="scss">
.gantt-dependency-arrows {
	position: absolute;
	inset-block-start: 0;
	inset-inline-start: 0;
	pointer-events: none;
	z-index: 1;
}

.dependency-arrow {
	transition: opacity 0.2s;
}
</style>
