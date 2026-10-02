import {MILLISECONDS_A_DAY} from '@/constants/date'
import type {TaskResponse} from '@/client/queries/tasks'
import {buildGanttTaskTree} from '@/helpers/ganttTaskTree'
import {buildWbs, criticalPath, criticalPathInput, durationInDays, taskSpan} from '@/helpers/ganttSchedule'
import {buildTimelineTiers, type GanttZoom} from '@/helpers/ganttZoom'
import {MAX_SHEET_COLUMNS, type XlsxCell, type XlsxSheet, type XlsxStyle} from './xlsx'

// A timeline sheet with more columns than this is cut: it would be unreadable anyway.
export const MAX_TIMELINE_COLUMNS = 400

export interface GanttExportLabels {
	tasksSheet: string
	timelineSheet: string
	wbs: string
	name: string
	duration: string
	start: string
	finish: string
	percent: string
	predecessors: string
	assignees: string
	critical: string
	truncated: string
}

export interface GanttExportInput {
	tasks: Map<number, TaskResponse>
	zoom: GanttZoom
	from: Date
	to: Date
	labels: GanttExportLabels
	// Formats the header of a timeline column.
	formatUnit: (date: Date, zoom: GanttZoom) => string
	showCritical?: boolean
}

interface Row {
	task: TaskResponse
	indent: number
	isParent: boolean
	start: Date | null
	end: Date | null
}

// What a timeline column stands for: a day at fit and day zoom, otherwise a week, month or quarter.
export function timelineUnits(from: Date, to: Date, zoom: GanttZoom): {start: Date, end: Date}[] {
	const days: Date[] = []
	for (const d = new Date(from.getFullYear(), from.getMonth(), from.getDate()); d <= to; d.setDate(d.getDate() + 1)) {
		days.push(new Date(d))
	}
	if (zoom === 'fit' || zoom === 'day') {
		return days.map(d => ({start: d, end: new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1)}))
	}
	const tiers = buildTimelineTiers(days, 1, zoom)
	return tiers.lower.map(cell => ({
		start: cell.start,
		end: new Date(cell.start.getFullYear(), cell.start.getMonth(), cell.start.getDate() + cell.width),
	}))
}

export function buildGanttSheets(input: GanttExportInput): {sheets: XlsxSheet[], truncated: boolean} {
	const {labels} = input
	const nodes = buildGanttTaskTree(input.tasks)
	const wbs = buildWbs(nodes)

	// The dates and the critical path come from the same helpers as the chart on screen, so the file
	// shows what the screen shows.
	const rows: Row[] = nodes.map(node => {
		const {start, end} = taskSpan(node)
		return {task: node.task, indent: node.indentLevel, isParent: node.isParent, start, end}
	})

	let critical: ReturnType<typeof criticalPath> | null = null
	if (input.showCritical) {
		const cpm = criticalPathInput(rows.map(r => ({task: r.task, span: {start: r.start, end: r.end}})), input.tasks.values())
		critical = criticalPath(cpm.tasks, cpm.edges)
	}

	const label = (r: Row) => `${'  '.repeat(r.indent)}${r.task.title}`
	const header = [
		labels.wbs, labels.name, labels.duration, labels.start, labels.finish, labels.percent,
		labels.predecessors, labels.assignees, ...(critical ? [labels.critical] : []),
	]

	const tasksSheet: XlsxSheet = {
		name: labels.tasksSheet,
		header,
		widths: [8, 44, 10, 12, 12, 10, 16, 24, 10],
		rows: rows.map(r => {
			const row: XlsxCell[] = [
				wbs.get(r.task.id) ?? '',
				label(r),
				durationInDays(r.start, r.end, Boolean(r.task.is_milestone)),
				{value: r.start, style: 'date'},
				{value: r.end, style: 'date'},
				{value: r.task.done ? 1 : r.task.percent_done, style: 'percent'},
				(r.task.related_tasks.follows ?? []).map(p => wbs.get(p.id) ?? `#${p.id}`).join(', '),
				r.task.assignees.map(a => a.name || a.username).join(', '),
			]
			if (critical) {
				row.push(critical.taskIds.has(r.task.id) ? 'x' : '')
			}
			return row
		}),
		freezeColumns: 2,
	}

	let units = timelineUnits(input.from, input.to, input.zoom)
	const limit = Math.min(MAX_TIMELINE_COLUMNS, MAX_SHEET_COLUMNS - 3)
	const truncated = units.length > limit
	if (truncated) {
		units = units.slice(0, limit)
	}

	const barStyle = (r: Row): XlsxStyle => {
		if (r.task.is_milestone) return 'barMilestone'
		if (critical?.taskIds.has(r.task.id)) return 'barCritical'
		return r.isParent ? 'barSummary' : 'barTask'
	}

	const timelineSheet: XlsxSheet = {
		name: labels.timelineSheet,
		header: [labels.wbs, labels.name, ...units.map(u => input.formatUnit(u.start, input.zoom))],
		widths: [8, 44, ...units.map(() => (input.zoom === 'fit' || input.zoom === 'day' ? 3.5 : 6))],
		rows: rows.map(r => {
			const style = barStyle(r)
			const cells: XlsxCell[] = [wbs.get(r.task.id) ?? '', label(r)]
			for (const unit of units) {
				cells.push(spans(r, unit.start, unit.end) ? {value: null, style} : null)
			}
			return cells
		}),
		freezeColumns: 2,
	}

	return {sheets: [tasksSheet, timelineSheet], truncated}
}

// Whether the task is on at any time within [from, to). A milestone is a point at its end.
function spans(row: Row, from: Date, to: Date): boolean {
	if (!row.end) {
		return false
	}
	const end = row.end.getTime()
	const start = (row.start ?? row.end).getTime()
	if (row.task.is_milestone) {
		return end >= from.getTime() && end < to.getTime()
	}
	// The end day counts as a whole day, like the chart draws it.
	const lastMoment = new Date(row.end.getFullYear(), row.end.getMonth(), row.end.getDate()).getTime() + MILLISECONDS_A_DAY
	return start < to.getTime() && lastMoment > from.getTime()
}
