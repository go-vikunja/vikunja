import type {ITask} from '@/modelTypes/ITask'

const MAX_INDENT_LEVEL = 4

export interface GanttTaskTreeNode {
	task: ITask
	indentLevel: number
	isParent: boolean
	childIds: number[]
	derivedStartDate: Date | null
	derivedEndDate: Date | null
	hasDerivedDates: boolean
}

/**
 * Builds a hierarchical task tree from a flat task map using relatedTasks data,
 * then flattens it in depth-first order for Gantt row rendering.
 */
export function buildGanttTaskTree(tasks: Map<number, ITask>): GanttTaskTreeNode[] {
	// Step 1: Build parent -> children mapping
	const childrenMap = new Map<number, number[]>()
	const hasParentInView = new Set<number>()

	for (const [taskId, task] of tasks) {
		const subtasks = task.relatedTasks?.subtask ?? []
		const childIds = subtasks
			.map(s => s.id)
			.filter(id => tasks.has(id))

		if (childIds.length > 0) {
			childrenMap.set(taskId, childIds)
		}

		const parents = task.relatedTasks?.parenttask ?? []
		for (const parent of parents) {
			if (tasks.has(parent.id)) {
				hasParentInView.add(taskId)
			}
		}
	}

	// Step 2: Find root tasks (no parent in the current view)
	const rootIds: number[] = []
	for (const [taskId] of tasks) {
		if (!hasParentInView.has(taskId)) {
			rootIds.push(taskId)
		}
	}

	// Step 3: Depth-first flatten
	const result: GanttTaskTreeNode[] = []
	const visited = new Set<number>()

	function visit(taskId: number, level: number) {
		if (visited.has(taskId)) return
		visited.add(taskId)

		const task = tasks.get(taskId)
		if (!task) return

		const childIds = childrenMap.get(taskId) ?? []
		const isParent = childIds.length > 0
		const clampedLevel = Math.min(level, MAX_INDENT_LEVEL)

		// Calculate derived dates for dateless parents
		let derivedStartDate: Date | null = null
		let derivedEndDate: Date | null = null
		let hasDerivedDates = false

		if (isParent && !task.startDate && !task.endDate && !task.dueDate) {
			const dates = collectChildDates(childIds, tasks, childrenMap)
			derivedStartDate = dates.minStart
			derivedEndDate = dates.maxEnd
			hasDerivedDates = derivedStartDate !== null || derivedEndDate !== null
		}

		result.push({
			task,
			indentLevel: clampedLevel,
			isParent,
			childIds,
			derivedStartDate,
			derivedEndDate,
			hasDerivedDates,
		})

		for (const childId of childIds) {
			visit(childId, level + 1)
		}
	}

	for (const rootId of rootIds) {
		visit(rootId, 0)
	}

	// Add any unvisited tasks (shouldn't happen normally, but safety net)
	for (const [taskId] of tasks) {
		if (!visited.has(taskId)) {
			const task = tasks.get(taskId)!
			result.push({
				task,
				indentLevel: 0,
				isParent: false,
				childIds: [],
				derivedStartDate: null,
				derivedEndDate: null,
				hasDerivedDates: false,
			})
		}
	}

	return result
}

function collectChildDates(
	childIds: number[],
	tasks: Map<number, ITask>,
	childrenMap: Map<number, number[]>,
): { minStart: Date | null; maxEnd: Date | null } {
	let minStart: Date | null = null
	let maxEnd: Date | null = null

	for (const childId of childIds) {
		const child = tasks.get(childId)
		if (!child) continue

		const start = child.startDate ? new Date(child.startDate) : null
		const end = child.endDate || child.dueDate
			? new Date((child.endDate || child.dueDate) as Date)
			: null

		if (start && (!minStart || start < minStart)) {
			minStart = start
		}
		if (end && (!maxEnd || end > maxEnd)) {
			maxEnd = end
		}

		// Recurse into grandchildren
		const grandchildIds = childrenMap.get(childId) ?? []
		if (grandchildIds.length > 0) {
			const grandDates = collectChildDates(grandchildIds, tasks, childrenMap)
			if (grandDates.minStart && (!minStart || grandDates.minStart < minStart)) {
				minStart = grandDates.minStart
			}
			if (grandDates.maxEnd && (!maxEnd || grandDates.maxEnd > maxEnd)) {
				maxEnd = grandDates.maxEnd
			}
		}
	}

	return {minStart, maxEnd}
}

