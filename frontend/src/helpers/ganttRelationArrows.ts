import type {ITask} from '@/modelTypes/ITask'

export interface GanttBarPosition {
	x: number       // left edge x position
	y: number       // vertical center y position
	width: number   // bar width in pixels
	rowIndex: number
}

export interface GanttArrow {
	fromTaskId: number
	toTaskId: number
	startX: number
	startY: number
	endX: number
	endY: number
	color: string
	relationKind: 'blocking' | 'precedes'
}

const ARROW_COLORS: Record<string, string> = {
	blocking: 'var(--danger)',
	precedes: 'var(--grey-500)',
}

/**
 * Builds arrow data for dependency relations between visible Gantt tasks.
 * Only processes `blocking` and `precedes` directions to avoid duplicates.
 */
export function buildRelationArrows(
	tasks: Map<number, ITask>,
	positions: Map<number, GanttBarPosition>,
	hiddenToAncestor: Map<number, number>,
): GanttArrow[] {
	const arrows: GanttArrow[] = []
	const seen = new Set<string>()

	for (const [taskId, task] of tasks) {
		const sourceKinds = ['blocking', 'precedes'] as const

		for (const kind of sourceKinds) {
			const relatedTasks = task.relatedTasks?.[kind] ?? []

			for (const related of relatedTasks) {
				let fromId = taskId
				let toId = related.id

				// Re-route hidden tasks to their visible ancestor
				if (hiddenToAncestor.has(fromId)) {
					fromId = hiddenToAncestor.get(fromId)!
				}
				if (hiddenToAncestor.has(toId)) {
					toId = hiddenToAncestor.get(toId)!
				}

				// Skip if either end is not visible
				if (!positions.has(fromId) || !positions.has(toId)) continue

				// Skip self-arrows (can happen after re-routing)
				if (fromId === toId) continue

				// Deduplicate
				const key = `${Math.min(fromId, toId)}-${Math.max(fromId, toId)}-${kind}`
				if (seen.has(key)) continue
				seen.add(key)

				const fromPos = positions.get(fromId)!
				const toPos = positions.get(toId)!

				arrows.push({
					fromTaskId: fromId,
					toTaskId: toId,
					startX: fromPos.x + fromPos.width,
					startY: fromPos.y,
					endX: toPos.x,
					endY: toPos.y,
					color: ARROW_COLORS[kind],
					relationKind: kind,
				})
			}
		}
	}

	return spreadOverlappingArrows(arrows)
}

const PREFERRED_SPREAD_PX = 6
const MAX_TOTAL_SPREAD_PX = 24

/**
 * When multiple arrows share the same source or target task,
 * offset their Y positions so they don't overlap visually.
 * The spread is capped to stay within the row height.
 */
function spreadOverlappingArrows(arrows: GanttArrow[]): GanttArrow[] {
	spreadByKey(arrows, 'fromTaskId', 'startY')
	spreadByKey(arrows, 'toTaskId', 'endY')
	return arrows
}

function spreadByKey(arrows: GanttArrow[], groupKey: 'fromTaskId' | 'toTaskId', yKey: 'startY' | 'endY') {
	const groups = new Map<number, GanttArrow[]>()
	for (const arrow of arrows) {
		const id = arrow[groupKey]
		let group = groups.get(id)
		if (!group) {
			group = []
			groups.set(id, group)
		}
		group.push(arrow)
	}

	for (const group of groups.values()) {
		if (group.length < 2) continue
		const totalSpread = Math.min((group.length - 1) * PREFERRED_SPREAD_PX, MAX_TOTAL_SPREAD_PX)
		const step = totalSpread / (group.length - 1)
		for (let i = 0; i < group.length; i++) {
			group[i][yKey] += -totalSpread / 2 + i * step
		}
	}
}

