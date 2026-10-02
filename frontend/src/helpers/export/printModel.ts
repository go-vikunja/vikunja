import type {TaskResponse} from '@/client/queries/tasks'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import {buildGanttTaskTree} from '@/helpers/ganttTaskTree'
import {buildWbs, criticalPath, criticalPathInput, durationInDays, taskSpan} from '@/helpers/ganttSchedule'
import {buildTimelineTiers, type GanttZoom} from '@/helpers/ganttZoom'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {getTaskIdentifier} from '@/helpers/task'
import {ALL_EXPORT_COLUMNS, type ExportColumnKey, type ExportLookups} from './taskExport'

// ---- Tables (table and list views) --------------------------------------------------------------

export interface PrintTable {
	headers: string[]
	rows: string[][]
	// Rows of a printed table can be set apart, for example done tasks.
	doneRows: boolean[]
}

export interface PrintLabels {
	columns: Record<ExportColumnKey, string>
	yes: string
	no: string
}

// How a value is written on paper: readable dates in the browser's locale.
export type PrintFormatters = {
	date: (d: Date) => string
	dateTime: (d: Date) => string
}

function name(user: {name?: string, username?: string} | null | undefined): string {
	return user ? (user.name || user.username || '') : ''
}

export function printValue(
	task: TaskResponse,
	key: ExportColumnKey,
	lookups: ExportLookups,
	fmt: PrintFormatters,
	labels: PrintLabels,
): string {
	const dt = (v: string | undefined) => {
		const d = parseDateOrNull(v)
		return d ? fmt.dateTime(d) : ''
	}
	const day = (v: string | undefined) => {
		const d = parseDateOrNull(v)
		return d ? fmt.date(d) : ''
	}
	switch (key) {
		case 'identifier': return getTaskIdentifier(task)
		case 'title': return task.title
		case 'project': return lookups.projectTitle(task.project_id)
		case 'bucket': return lookups.bucketTitle ? lookups.bucketTitle(task.bucket_id) : ''
		case 'assignor': return name(task.created_by)
		case 'assignees': return task.assignees.map(name).filter(Boolean).join(', ')
		case 'priority': return task.priority ? String(task.priority) : ''
		case 'labels': return task.labels.map(l => l.title).filter(Boolean).join(', ')
		case 'start': return day(task.start_date)
		case 'due': return day(task.due_date)
		case 'end': return day(task.end_date)
		case 'percent': return `${Math.round((task.done ? 1 : task.percent_done) * 100)}%`
		case 'done': return task.done ? labels.yes : labels.no
		case 'doneAt': return dt(task.done_at)
		case 'created': return dt(task.created)
		case 'updated': return dt(task.updated)
	}
}

export function buildPrintTable(
	tasks: TaskResponse[],
	columns: ExportColumnKey[],
	lookups: ExportLookups,
	fmt: PrintFormatters,
	labels: PrintLabels,
): PrintTable {
	const keys = ALL_EXPORT_COLUMNS
		.map(c => c.key)
		.filter(key => columns.includes(key) && (key !== 'bucket' || lookups.bucketTitle))
	return {
		headers: keys.map(key => labels.columns[key]),
		rows: tasks.map(task => keys.map(key => printValue(task, key, lookups, fmt, labels))),
		doneRows: tasks.map(task => task.done),
	}
}

// ---- Kanban ---------------------------------------------------------------------------------------

export interface PrintBoard {
	columns: {title: string, tasks: {title: string, details: string[], done: boolean}[]}[]
}

export function buildPrintBoard(
	tasks: TaskResponse[],
	buckets: {id: number, title: string}[],
	lookups: ExportLookups,
	fmt: PrintFormatters,
): PrintBoard {
	const byBucket = new Map<number, TaskResponse[]>()
	for (const task of tasks) {
		const list = byBucket.get(task.bucket_id) ?? []
		list.push(task)
		byBucket.set(task.bucket_id, list)
	}
	const card = (task: TaskResponse) => {
		const due = parseDateOrNull(task.due_date)
		const details = [
			getTaskIdentifier(task),
			task.assignees.map(name).filter(Boolean).join(', '),
			due ? fmt.date(due) : '',
			task.labels.map(l => l.title).filter(Boolean).join(', '),
		].filter(Boolean)
		return {title: task.title, details, done: task.done}
	}
	const known = new Set(buckets.map(b => b.id))
	const columns = buckets.map(bucket => ({title: bucket.title, tasks: (byBucket.get(bucket.id) ?? []).map(card)}))
	// Tasks whose bucket is not on the board (a bucket from another view) are kept, not dropped.
	const stray = tasks.filter(t => !known.has(t.bucket_id)).map(card)
	if (stray.length > 0) {
		columns.push({title: lookups.bucketTitle ? lookups.bucketTitle(0) : '', tasks: stray})
	}
	return {columns}
}

// ---- Gantt -----------------------------------------------------------------------------------------

export interface PrintGanttRow {
	id: number
	wbs: string
	title: string
	indent: number
	duration: number | null
	start: Date | null
	end: Date | null
	percent: number
	predecessors: string
	assignees: string
	kind: 'task' | 'summary' | 'milestone'
	critical: boolean
	// Bar in chart units (pixels), null without dates.
	bar: {x: number, width: number} | null
}

export interface PrintGantt {
	rows: PrintGanttRow[]
	width: number
	upper: {label: string, x: number, width: number}[]
	lower: {label: string, x: number, width: number}[]
	lines: number[]
	todayX: number | null
	deps: {fromRow: number, toRow: number, critical: boolean}[]
}

const PRINT_DAY_WIDTH: Record<GanttZoom, number> = {fit: 16, day: 16, week: 8, month: 3, quarter: 1.5}

export function buildPrintGantt(
	tasks: Map<number, TaskResponse>,
	from: Date,
	to: Date,
	zoom: GanttZoom,
	showCritical: boolean,
	formatUnit: (date: Date, unit: 'year' | 'quarter' | 'month' | 'week') => string,
	now: Date,
): PrintGantt {
	const days: Date[] = []
	for (const d = new Date(from.getFullYear(), from.getMonth(), from.getDate()); d <= to; d.setDate(d.getDate() + 1)) {
		days.push(new Date(d))
	}
	const dayWidth = PRINT_DAY_WIDTH[zoom]
	const width = days.length * dayWidth
	const firstDay = days[0]

	const xOf = (date: Date) => firstDay
		? ((date.getTime() - firstDay.getTime()) / MILLISECONDS_A_DAY) * dayWidth
		: 0

	const nodes = buildGanttTaskTree(tasks)
	const wbs = buildWbs(nodes)

	// Same dates and same critical path as the chart on screen and the Excel export.
	const dated = nodes.map(node => ({node, ...taskSpan(node)}))

	let critical: ReturnType<typeof criticalPath> | null = null
	if (showCritical) {
		const cpm = criticalPathInput(dated.map(d => ({task: d.node.task, span: {start: d.start, end: d.end}})), tasks.values())
		critical = criticalPath(cpm.tasks, cpm.edges)
	}

	const rowIndex = new Map<number, number>()
	const rows: PrintGanttRow[] = dated.map(({node, start, end}, index) => {
		const t = node.task
		rowIndex.set(t.id, index)
		let bar: PrintGanttRow['bar'] = null
		if (end) {
			const barStart = start ?? end
			// The end day counts as a whole day, like the chart draws it.
			const barEnd = new Date(end.getFullYear(), end.getMonth(), end.getDate() + 1)
			const x = xOf(new Date(barStart.getFullYear(), barStart.getMonth(), barStart.getDate()))
			bar = t.is_milestone
				? {x: xOf(new Date(end.getFullYear(), end.getMonth(), end.getDate())) + dayWidth / 2, width: 0}
				: {x, width: Math.max(xOf(barEnd) - x, dayWidth)}
		}
		return {
			id: t.id,
			wbs: wbs.get(t.id) ?? '',
			title: t.title,
			indent: node.indentLevel,
			duration: durationInDays(start, end, Boolean(t.is_milestone)),
			start,
			end,
			percent: Math.round((t.done ? 1 : t.percent_done) * 100),
			predecessors: (t.related_tasks.follows ?? []).map(p => wbs.get(p.id) ?? `#${p.id}`).join(', '),
			assignees: t.assignees.map(name).filter(Boolean).join(', '),
			kind: t.is_milestone ? 'milestone' : node.isParent ? 'summary' : 'task',
			critical: Boolean(critical?.taskIds.has(t.id)),
			bar,
		}
	})

	const deps: PrintGantt['deps'] = []
	for (const t of tasks.values()) {
		for (const succ of t.related_tasks.precedes ?? []) {
			const fromRow = rowIndex.get(t.id)
			const toRow = rowIndex.get(succ.id)
			if (fromRow === undefined || toRow === undefined) continue
			deps.push({fromRow, toRow, critical: Boolean(critical?.edges.has(`${t.id}-${succ.id}`))})
		}
	}

	const tiers = buildTimelineTiers(days, dayWidth, zoom, formatUnit)
	const lower = tiers.lower.length > 0 ? tiers.lower : days.map((d, i) => ({label: String(d.getDate()), x: i * dayWidth, width: dayWidth}))

	const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
	const todayX = firstDay && today >= firstDay && today <= days[days.length - 1] ? xOf(today) + dayWidth / 2 : null

	return {
		rows,
		width,
		upper: tiers.upper.map(c => ({label: c.label, x: c.x, width: c.width})),
		lower: lower.map(c => ({label: c.label, x: c.x, width: c.width})),
		lines: lower.map(c => c.x),
		todayX,
		deps,
	}
}
