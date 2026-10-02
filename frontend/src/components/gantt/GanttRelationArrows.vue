<template>
	<svg
		class="gantt-relation-arrows"
		:width="width"
		:height="height"
		xmlns="http://www.w3.org/2000/svg"
		aria-hidden="true"
	>
		<defs>
			<marker
				id="arrowhead-danger"
				markerWidth="6"
				markerHeight="6"
				refX="5"
				refY="3"
				orient="auto"
			>
				<polygon
					points="0,0 6,3 0,6"
					fill="var(--danger)"
				/>
			</marker>
			<marker
				id="arrowhead-grey"
				markerWidth="6"
				markerHeight="6"
				refX="5"
				refY="3"
				orient="auto"
			>
				<polygon
					points="0,0 6,3 0,6"
					fill="var(--grey-500)"
				/>
			</marker>
		</defs>

		<template
			v-for="(arrow, index) in arrows"
			:key="`arrow-${index}`"
		>
			<path
				:d="computePath(arrow)"
				:stroke="arrow.color"
				:stroke-width="arrow.critical ? 2.5 : 1.5"
				fill="none"
				:stroke-dasharray="arrow.relationKind === 'precedes' && !arrow.critical ? '6,4' : 'none'"
				:marker-end="getMarkerEnd(arrow)"
				class="gantt-arrow"
				:class="{'is-critical': arrow.critical}"
			/>
			<!-- A wide invisible line to click: only dependencies can be removed from here -->
			<path
				v-if="deletable && arrow.relationKind === 'precedes'"
				:d="computePath(arrow)"
				stroke="transparent"
				stroke-width="12"
				fill="none"
				class="gantt-arrow-hit"
				role="button"
				tabindex="0"
				:aria-label="$t('project.gantt.removeDependency')"
				@click.stop="emit('deleteArrow', arrow.fromTaskId, arrow.toTaskId)"
				@keydown.enter.stop="emit('deleteArrow', arrow.fromTaskId, arrow.toTaskId)"
			>
				<title>{{ $t('project.gantt.removeDependency') }}</title>
			</path>
		</template>
	</svg>
</template>

<script setup lang="ts">
import type {GanttArrow} from '@/helpers/ganttRelationArrows'

withDefaults(defineProps<{
	arrows: GanttArrow[]
	width: number
	height: number
	rowHeight: number
	// Whether a dependency arrow can be clicked to remove it.
	deletable?: boolean
}>(), {
	deletable: false,
})

const emit = defineEmits<{
	// predecessor, successor
	(e: 'deleteArrow', fromTaskId: number, toTaskId: number): void
}>()

/**
 * Computes a bezier curve path for an arrow.
 * Uses horizontal bezier curves that curve around obstacles.
 */
function computePath(arrow: GanttArrow): string {
	const {startX, startY, endX, endY} = arrow

	// Horizontal distance
	const dx = endX - startX
	const dy = endY - startY

	// Control point offset (how much the curve bends)
	const cpOffset = Math.min(Math.abs(dx) * 0.4, 60)

	if (dx >= 0) {
		// Target is to the right - simple S-curve
		const cp1x = startX + cpOffset
		const cp1y = startY
		const cp2x = endX - cpOffset
		const cp2y = endY

		return `M ${startX} ${startY} C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${endX} ${endY}`
	} else {
		// Target is to the left - need to route around
		// Go right first, then down/up, then left to target
		const routeOffset = 30
		const midY = startY + (dy > 0 ? routeOffset : -routeOffset)

		const cp1x = startX + routeOffset
		const cp1y = startY
		const cp2x = startX + routeOffset
		const cp2y = midY

		const cp3x = endX - routeOffset
		const cp3y = midY
		const cp4x = endX - routeOffset
		const cp4y = endY

		return `M ${startX} ${startY} ` +
			`C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${startX + routeOffset} ${midY} ` +
			`L ${endX - routeOffset} ${midY} ` +
			`C ${cp3x} ${cp3y}, ${cp4x} ${cp4y}, ${endX} ${endY}`
	}
}

function getMarkerEnd(arrow: GanttArrow): string {
	return arrow.relationKind === 'blocking' || arrow.critical
		? 'url(#arrowhead-danger)'
		: 'url(#arrowhead-grey)'
}
</script>

<style scoped lang="scss">
.gantt-relation-arrows {
	position: absolute;
	inset-block-start: 0;
	inset-inline-start: 0;
	pointer-events: none;
	z-index: 3;
}

.gantt-arrow {
	opacity: 0.7;

	&.is-critical {
		opacity: 1;
	}
}

.gantt-arrow-hit {
	pointer-events: stroke;
	cursor: pointer;

	&:hover + .gantt-arrow,
	&:focus {
		outline: none;
	}
}
</style>

