<template>
	<svg
		v-if="cfg.enabled && arrows.length > 0"
		class="gantt-dependency-arrows"
		:width="totalWidth"
		:height="totalHeight"
	>
		<defs>
			<marker
				v-for="(arrow, i) in arrows"
				:id="`dep-arrow-${i}`"
				:key="`marker-${i}`"
				:markerWidth="cfg.arrowSize"
				:markerHeight="cfg.arrowSize * 0.75"
				:refX="cfg.arrowSize - 1"
				:refY="cfg.arrowSize * 0.375"
				orient="auto"
			>
				<polygon
					:points="`0 0, ${cfg.arrowSize} ${cfg.arrowSize * 0.375}, 0 ${cfg.arrowSize * 0.75}`"
					:fill="arrow.color"
				/>
			</marker>
		</defs>

		<!-- Shadow layer -->
		<template v-if="cfg.showShadow">
			<path
				v-for="(arrow, i) in arrows"
				:key="`shadow-${i}`"
				:d="arrow.path"
				fill="none"
				:stroke="`rgba(0,0,0,${cfg.shadowOpacity})`"
				:stroke-width="cfg.shadowWidth"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</template>

		<!-- Arrow lines -->
		<path
			v-for="(arrow, i) in arrows"
			:key="`line-${i}`"
			:d="arrow.path"
			fill="none"
			:stroke="arrow.color"
			:stroke-width="cfg.strokeWidth"
			:stroke-dasharray="cfg.dashArray === 'none' ? undefined : cfg.dashArray"
			stroke-linecap="round"
			stroke-linejoin="round"
			:marker-end="`url(#dep-arrow-${i})`"
			class="dependency-arrow"
		>
			<title>{{ arrow.label }}</title>
		</path>

		<!-- Source dots -->
		<template v-if="cfg.showDots">
			<circle
				v-for="(arrow, i) in arrows"
				:key="`dot-${i}`"
				:cx="arrow.sx"
				:cy="arrow.sy"
				:r="cfg.dotRadius"
				:fill="arrow.color"
			/>
		</template>
	</svg>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import type {GanttBarModel} from '@/composables/useGanttBar'
import type {ITask} from '@/modelTypes/ITask'
import {useGanttArrowConfig} from '@/composables/useGanttArrowConfig'

const ROW_HEIGHT = 40
const BAR_HEIGHT = 30

const props = defineProps<{
	barsByRow: GanttBarModel[][]
	tasks: Map<ITask['id'], ITask>
	dateFrom: Date
	dayWidthPixels: number
	totalWidth: number
}>()

const {config: cfg} = useGanttArrowConfig()

const MULTI_COLORS = [
	[93, 165, 218], [250, 164, 58], [96, 189, 104], [241, 88, 84],
	[178, 118, 178], [222, 207, 63], [77, 201, 246], [241, 124, 176],
	[178, 145, 47], [0, 193, 166],
]

function getColor(index: number): string {
	const a = cfg.opacity
	if (cfg.palette === 'mono') return `rgba(150,150,200,${a})`
	const c = MULTI_COLORS[index % MULTI_COLORS.length]
	return `rgba(${c[0]},${c[1]},${c[2]},${a})`
}

const totalHeight = computed(() => props.barsByRow.length * ROW_HEIGHT)

// Bar position map: id -> { x (left edge), rightX, cy }
const barPositions = computed(() => {
	const map = new Map<string, { x: number; rightX: number; cy: number; width: number }>()
	const fromTime = props.dateFrom.getTime()

	props.barsByRow.forEach((bars, rowIndex) => {
		for (const bar of bars) {
			const startDays = (bar.start.getTime() - fromTime) / (1000 * 60 * 60 * 24)
			const endDays = (bar.end.getTime() - fromTime) / (1000 * 60 * 60 * 24)
			const x = startDays * props.dayWidthPixels
			const width = Math.max((endDays - startDays) * props.dayWidthPixels, 1)

			map.set(bar.id, {
				x,
				rightX: x + width,
				cy: rowIndex * ROW_HEIGHT + ROW_HEIGHT / 2,
				width,
			})
		}
	})

	return map
})

function getExitPoint(rightX: number, cy: number, barWidth: number) {
	const halfH = BAR_HEIGHT / 2
	if (cfg.exitDir === 'bottom') {
		const left = rightX - barWidth
		return {
			x: left + barWidth * cfg.exitOffset,
			y: cy + halfH,
		}
	}
	// Right edge
	return {
		x: rightX,
		y: cy - halfH + BAR_HEIGHT * cfg.exitOffset,
	}
}

function getEntryPoint(leftX: number, cy: number, barWidth: number) {
	const halfH = BAR_HEIGHT / 2
	if (cfg.entryDir === 'top') {
		return {
			x: leftX + barWidth * cfg.entryOffset,
			y: cy - halfH,
		}
	}
	// Left edge
	return {
		x: leftX,
		y: cy - halfH + BAR_HEIGHT * cfg.entryOffset,
	}
}

function buildBezierPath(sx: number, sy: number, tx: number, ty: number): string {
	const dx = tx - sx
	const dy = ty - sy

	// Bottom → Left: hook shape (drop down, curve to left of target)
	if (cfg.exitDir === 'bottom' && cfg.entryDir === 'left') {
		const c1x = sx
		const c1y = sy + Math.abs(dy) * cfg.cp1X + cfg.cp1Y
		const c2x = tx + cfg.cp2Y  // cp2Y repurposed as horizontal approach distance
		const c2y = ty
		return `M ${sx} ${sy} C ${c1x} ${c1y}, ${c2x} ${c2y}, ${tx} ${ty}`
	}

	// Bottom → Top: S-curve dropping down then approaching from above
	if (cfg.exitDir === 'bottom' && cfg.entryDir === 'top') {
		const midY = (sy + ty) / 2
		const c1x = sx
		const c1y = midY + cfg.cp1Y
		const c2x = tx
		const c2y = midY + cfg.cp2Y
		return `M ${sx} ${sy} C ${c1x} ${c1y}, ${c2x} ${c2y}, ${tx} ${ty}`
	}

	// Right → Top: curve from right side into top of target
	if (cfg.exitDir === 'right' && cfg.entryDir === 'top') {
		if (dx > 10) {
			const c1x = sx + dx * cfg.cp1X
			const c1y = sy + cfg.cp1Y
			const c2x = tx
			const c2y = ty - Math.abs(cfg.cp2Y || 30)
			return `M ${sx} ${sy} C ${c1x} ${c1y}, ${c2x} ${c2y}, ${tx} ${ty}`
		}
		const detour = 15
		return `M ${sx} ${sy} L ${sx + detour} ${sy} Q ${sx + detour + 10} ${sy}, ${sx + detour + 10} ${sy - detour} L ${sx + detour + 10} ${ty - detour} Q ${sx + detour + 10} ${ty - detour}, ${tx} ${ty - detour} L ${tx} ${ty}`
	}

	// Right → Left (default): standard horizontal bezier
	if (dx > 10) {
		const c1x = sx + dx * cfg.cp1X
		const c1y = sy + cfg.cp1Y
		const c2x = sx + dx * cfg.cp2X
		const c2y = ty + cfg.cp2Y
		return `M ${sx} ${sy} C ${c1x} ${c1y}, ${c2x} ${c2y}, ${tx} ${ty}`
	}
	// Detour for overlapping (right → left when target is to the left)
	const detour = 15
	const dir = ty > sy ? 1 : -1
	return `M ${sx} ${sy} L ${sx + detour} ${sy} Q ${sx + detour + 10} ${sy}, ${sx + detour + 10} ${sy + dir * detour} L ${sx + detour + 10} ${ty - dir * detour} Q ${sx + detour + 10} ${ty}, ${tx - detour} ${ty} L ${tx} ${ty}`
}

function buildSteppedPath(start: {x: number; y: number}, end: {x: number; y: number}): string {
	const r = cfg.pathMode === 'stepRounded' ? cfg.cornerRadius : 0
	const ex = cfg.exitLength
	const en = cfg.entryLength

	let pts: number[][]

	if (cfg.exitDir === 'bottom' && cfg.entryDir === 'left') {
		const dropY = start.y + ex
		const approachX = end.x - en
		pts = [[start.x, start.y], [start.x, dropY], [approachX, dropY], [approachX, end.y], [end.x, end.y]]
	} else if (cfg.exitDir === 'bottom' && cfg.entryDir === 'top') {
		const dropY = start.y + ex
		pts = [[start.x, start.y], [start.x, dropY], [end.x, dropY], [end.x, end.y]]
	} else if (cfg.exitDir === 'right' && cfg.entryDir === 'left') {
		const turnX = start.x + ex
		const approachX = end.x - en
		const midY = (start.y + end.y) / 2
		if (approachX <= turnX + 5) {
			pts = [[start.x, start.y], [turnX, start.y], [turnX, end.y], [end.x, end.y]]
		} else {
			pts = [[start.x, start.y], [turnX, start.y], [turnX, midY], [approachX, midY], [approachX, end.y], [end.x, end.y]]
		}
	} else if (cfg.exitDir === 'right' && cfg.entryDir === 'top') {
		const turnX = start.x + ex
		pts = [[start.x, start.y], [turnX, start.y], [turnX, end.y - en], [end.x, end.y - en], [end.x, end.y]]
	} else {
		pts = [[start.x, start.y], [end.x, end.y]]
	}

	return buildRoundedPath(pts, r)
}

function buildRoundedPath(pts: number[][], r: number): string {
	if (pts.length < 2) return ''
	if (r === 0 || pts.length === 2) {
		return 'M ' + pts.map(p => `${p[0]} ${p[1]}`).join(' L ')
	}

	let d = `M ${pts[0][0]} ${pts[0][1]}`

	for (let i = 1; i < pts.length - 1; i++) {
		const prev = pts[i - 1]
		const curr = pts[i]
		const next = pts[i + 1]

		const dx1 = curr[0] - prev[0], dy1 = curr[1] - prev[1]
		const dx2 = next[0] - curr[0], dy2 = next[1] - curr[1]
		const len1 = Math.sqrt(dx1 * dx1 + dy1 * dy1)
		const len2 = Math.sqrt(dx2 * dx2 + dy2 * dy2)

		const cr = Math.min(r, len1 / 2, len2 / 2)
		if (cr < 1) {
			d += ` L ${curr[0]} ${curr[1]}`
			continue
		}

		const sRX = curr[0] - (dx1 / len1) * cr
		const sRY = curr[1] - (dy1 / len1) * cr
		const eRX = curr[0] + (dx2 / len2) * cr
		const eRY = curr[1] + (dy2 / len2) * cr

		d += ` L ${sRX} ${sRY} Q ${curr[0]} ${curr[1]}, ${eRX} ${eRY}`
	}

	d += ` L ${pts[pts.length - 1][0]} ${pts[pts.length - 1][1]}`
	return d
}

const arrows = computed(() => {
	const result: {
		path: string
		color: string
		label: string
		sx: number
		sy: number
	}[] = []

	let colorIndex = 0

	for (const [taskId, task] of props.tasks) {
		const precedesTasks = task.relatedTasks?.precedes || []
		for (const target of precedesTasks) {
			const sourcePos = barPositions.value.get(String(taskId))
			const targetPos = barPositions.value.get(String(target.id))

			if (!sourcePos || !targetPos) continue

			const color = getColor(colorIndex)
			colorIndex++

			const start = getExitPoint(sourcePos.rightX, sourcePos.cy, sourcePos.width)
			const end = getEntryPoint(targetPos.x, targetPos.cy, targetPos.width)

			let path: string
			if (cfg.pathMode === 'bezier') {
				path = buildBezierPath(start.x, start.y, end.x, end.y)
			} else {
				path = buildSteppedPath(start, end)
			}

			const sourceName = task.title || `Task #${taskId}`
			const targetName = target.title || `Task #${target.id}`

			result.push({
				path,
				color,
				label: `${sourceName} → ${targetName}`,
				sx: start.x,
				sy: start.y,
			})
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
	z-index: 5;
}

.dependency-arrow {
	transition: opacity 0.2s;
}
</style>
