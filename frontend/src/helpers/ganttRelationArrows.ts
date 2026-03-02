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

	return arrows
}

