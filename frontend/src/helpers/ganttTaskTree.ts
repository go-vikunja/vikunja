import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import type {TaskResponse} from '@/client/queries/tasks'

const MAX_INDENT_LEVEL = 4

export interface GanttTaskTreeNode {
	task: TaskResponse
	indentLevel: number
	isParent: boolean
	childIds: number[]
	derivedStartDate: Date | null
	derivedEndDate: Date | null
	hasDerivedDates: boolean
}

export interface GanttTaskTreeOptions {
	// 'relation' keeps subtasks in the order of the parent's relations, 'position' sorts them by their view position.
	childOrder?: 'relation' | 'position'
}

/**
 * Builds a hierarchical task tree from a flat task map using relatedTasks data,
 * then flattens it in depth-first order for Gantt row rendering.
 */
export function buildGanttTaskTree(
	tasks: Map<number, TaskResponse>,
	{childOrder = 'relation'}: GanttTaskTreeOptions = {},
): GanttTaskTreeNode[] {
	// Step 1: Build parent -> children mapping
	const childrenMap = new Map<number, number[]>()
	const hasParentInView = new Set<number>()

	for (const [taskId, task] of tasks) {
		const subtasks = task.related_tasks.subtask ?? []
		const childIds = subtasks
			.map(s => s.id)
			.filter(id => tasks.has(id))
		if (childOrder === 'position') {
			childIds.sort((a, b) => tasks.get(a)!.position - tasks.get(b)!.position || a - b)
		}

		if (childIds.length > 0) {
			childrenMap.set(taskId, childIds)
		}

		const parents = task.related_tasks.parenttask ?? []
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
	const collectChildDates = createChildDatesCollector(tasks, childrenMap)

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

		if (isParent && !parseDateOrNull(task.start_date) && !parseDateOrNull(task.end_date) && !parseDateOrNull(task.due_date)) {
			const dates = collectChildDates(childIds)
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

interface DateRange {
	minStart: Date | null
	maxEnd: Date | null
}

function mergeRange(into: DateRange, from: DateRange) {
	if (from.minStart && (!into.minStart || from.minStart < into.minStart)) {
		into.minStart = from.minStart
	}
	if (from.maxEnd && (!into.maxEnd || from.maxEnd > into.maxEnd)) {
		into.maxEnd = from.maxEnd
	}
}

// Memoized per task: relation graphs can be DAGs (shared descendants) or cyclic.
function createChildDatesCollector(
	tasks: Map<number, TaskResponse>,
	childrenMap: Map<number, number[]>,
) {
	const subtreeRanges = new Map<number, DateRange>()
	const inProgress = new Set<number>()

	function subtreeRange(taskId: number): DateRange {
		const cached = subtreeRanges.get(taskId)
		if (cached) return cached

		const range: DateRange = {minStart: null, maxEnd: null}
		const task = tasks.get(taskId)
		if (!task || inProgress.has(taskId)) return range

		inProgress.add(taskId)
		mergeRange(range, {
			minStart: parseDateOrNull(task.start_date),
			maxEnd: parseDateOrNull(task.end_date) ?? parseDateOrNull(task.due_date),
		})
		for (const childId of childrenMap.get(taskId) ?? []) {
			mergeRange(range, subtreeRange(childId))
		}
		inProgress.delete(taskId)

		subtreeRanges.set(taskId, range)
		return range
	}

	return (childIds: number[]): DateRange => {
		const range: DateRange = {minStart: null, maxEnd: null}
		for (const childId of childIds) {
			mergeRange(range, subtreeRange(childId))
		}
		return range
	}
}
