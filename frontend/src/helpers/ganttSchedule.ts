import {MILLISECONDS_A_DAY} from '@/constants/date'
import type {TaskResponse} from '@/client/queries/tasks'
import type {GanttTaskTreeNode} from '@/helpers/ganttTaskTree'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

// ---- Dates of a task as the Gantt chart sees them -------------------------------------------------

export interface TaskSpan {
	start: Date | null
	end: Date | null
}

/**
 * The start and end a task has in the Gantt chart, in one place so that the chart, its task table, the
 * PDF and the Excel export agree. The end is the end date, else the due date, else (for a parent
 * without dates of its own) the latest end of its children. The start is the start date, else the
 * earliest start of the children. A milestone is a point: its start is its end, and a milestone that
 * only has a start date sits at that date.
 */
export function taskSpan(node: Pick<GanttTaskTreeNode, 'task' | 'derivedStartDate' | 'derivedEndDate' | 'hasDerivedDates'>): TaskSpan {
	const t = node.task
	const ownEnd = parseDateOrNull(t.end_date) ?? parseDateOrNull(t.due_date)

	if (t.is_milestone) {
		const at = ownEnd ?? parseDateOrNull(t.start_date) ?? (node.hasDerivedDates ? node.derivedEndDate : null)
		return {start: at, end: at}
	}

	return {
		start: parseDateOrNull(t.start_date) ?? (node.hasDerivedDates ? node.derivedStartDate : null),
		end: ownEnd ?? (node.hasDerivedDates ? node.derivedEndDate : null),
	}
}

/**
 * The input of criticalPath for the shown rows: every dated task that is not done, and every
 * finish-to-start dependency between them.
 */
export function criticalPathInput(
	rows: {task: TaskResponse, span: TaskSpan}[],
	tasks: Iterable<TaskResponse>,
): {tasks: CpmTask[], edges: CpmEdge[]} {
	const cpmTasks: CpmTask[] = []
	for (const row of rows) {
		if (row.span.start && row.span.end && !row.task.done) {
			cpmTasks.push({id: row.task.id, start: row.span.start.getTime(), end: row.span.end.getTime()})
		}
	}
	const edges: CpmEdge[] = []
	for (const t of tasks) {
		for (const successor of t.related_tasks.precedes ?? []) {
			edges.push({from: t.id, to: successor.id})
		}
	}
	return {tasks: cpmTasks, edges}
}

// ---- WBS --------------------------------------------------------------------------------------

export interface WbsNode {
	task: {id: number}
	indentLevel: number
}

// "1", "1.1", "1.2", "2" ... from the depth-first task rows of the Gantt chart.
export function buildWbs(nodes: WbsNode[]): Map<number, string> {
	const result = new Map<number, string>()
	const counters: number[] = []
	for (const node of nodes) {
		const level = node.indentLevel
		while (counters.length <= level) {
			counters.push(0)
		}
		counters.length = level + 1
		counters[level]++
		result.set(node.task.id, counters.join('.'))
	}
	return result
}

// ---- Duration ---------------------------------------------------------------------------------

function calendarDay(date: Date): number {
	return Date.UTC(date.getFullYear(), date.getMonth(), date.getDate())
}

// Calendar days from start to end, both included (a task on a single day lasts 1 day). A milestone
// has no duration. Null when a date is missing.
export function durationInDays(start: Date | null, end: Date | null, isMilestone = false): number | null {
	if (isMilestone) {
		return 0
	}
	if (!start || !end) {
		return null
	}
	return Math.max(1, Math.round((calendarDay(end) - calendarDay(start)) / MILLISECONDS_A_DAY) + 1)
}

// ---- Critical path ----------------------------------------------------------------------------

export interface CpmTask {
	id: number
	// Milliseconds since the epoch.
	start: number
	end: number
}

export interface CpmEdge {
	// The predecessor: `from` has to end before `to` starts.
	from: number
	to: number
}

export interface CriticalPathResult {
	taskIds: Set<number>
	// "from-to"
	edges: Set<string>
}

export function edgeKey(from: number, to: number): string {
	return `${from}-${to}`
}

const ONE_MINUTE = 60_000

/**
 * Critical path over finish-to-start dependencies, with the dates the tasks have: a task is critical
 * when there is no slack between its end and the start of what depends on it, all the way to the
 * finish of the linked tasks. A dependency that is already violated (the successor starts before the
 * predecessor ends) counts as critical too. Tasks without a dependency are never highlighted, and
 * tasks in a dependency cycle are left out.
 */
export function criticalPath(tasks: CpmTask[], edges: CpmEdge[], toleranceMs = ONE_MINUTE): CriticalPathResult {
	const byId = new Map(tasks.map(t => [t.id, t]))
	const successors = new Map<number, number[]>()
	const inDegree = new Map<number, number>()
	const linked = new Set<number>()
	const seenEdges = new Set<string>()

	for (const edge of edges) {
		if (edge.from === edge.to || !byId.has(edge.from) || !byId.has(edge.to)) {
			continue
		}
		const key = edgeKey(edge.from, edge.to)
		if (seenEdges.has(key)) {
			continue
		}
		seenEdges.add(key)
		successors.set(edge.from, [...(successors.get(edge.from) ?? []), edge.to])
		inDegree.set(edge.to, (inDegree.get(edge.to) ?? 0) + 1)
		linked.add(edge.from)
		linked.add(edge.to)
	}

	// Topological order (Kahn). Whatever is left over sits on a cycle.
	const order: number[] = []
	const queue = [...linked].filter(id => !inDegree.get(id))
	const remaining = new Map(inDegree)
	while (queue.length > 0) {
		const id = queue.shift() as number
		order.push(id)
		for (const next of successors.get(id) ?? []) {
			const left = (remaining.get(next) ?? 0) - 1
			remaining.set(next, left)
			if (left === 0) {
				queue.push(next)
			}
		}
	}
	const sortable = new Set(order)
	if (sortable.size === 0) {
		return {taskIds: new Set(), edges: new Set()}
	}

	let finish = -Infinity
	for (const id of order) {
		finish = Math.max(finish, (byId.get(id) as CpmTask).end)
	}

	// Backward pass: the latest start each task can have without delaying the finish.
	const latestStart = new Map<number, number>()
	const slack = new Map<number, number>()
	for (const id of [...order].reverse()) {
		const task = byId.get(id) as CpmTask
		const next = (successors.get(id) ?? []).filter(s => sortable.has(s))
		const latestFinish = next.length > 0
			? Math.min(...next.map(s => latestStart.get(s) as number))
			: finish
		latestStart.set(id, latestFinish - (task.end - task.start))
		slack.set(id, latestFinish - task.end)
	}

	const taskIds = new Set<number>()
	for (const id of order) {
		if ((slack.get(id) as number) <= toleranceMs) {
			taskIds.add(id)
		}
	}

	const criticalEdges = new Set<string>()
	for (const id of order) {
		if (!taskIds.has(id)) {
			continue
		}
		for (const next of successors.get(id) ?? []) {
			if (!taskIds.has(next)) {
				continue
			}
			const gap = (byId.get(next) as CpmTask).start - (byId.get(id) as CpmTask).end
			if (gap <= toleranceMs) {
				criticalEdges.add(edgeKey(id, next))
			}
		}
	}

	return {taskIds, edges: criticalEdges}
}
