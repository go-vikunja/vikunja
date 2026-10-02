import type {TaskResponse} from '@/client/queries/tasks'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {getTaskIdentifier} from '@/helpers/task'
import type {XlsxCell, XlsxSheet} from './xlsx'

// The most tasks one export holds. Beyond it the file is cut and the user is told.
export const MAX_EXPORT_TASKS = 5000

export type ExportColumnKey =
	| 'identifier' | 'title' | 'project' | 'bucket' | 'assignor' | 'assignees' | 'priority' | 'labels'
	| 'start' | 'due' | 'end' | 'percent' | 'done' | 'doneAt' | 'created' | 'updated'

export interface ExportColumn {
	key: ExportColumnKey
	width: number
}

export const ALL_EXPORT_COLUMNS: ExportColumn[] = [
	{key: 'identifier', width: 12},
	{key: 'title', width: 44},
	{key: 'project', width: 22},
	{key: 'bucket', width: 16},
	{key: 'assignor', width: 20},
	{key: 'assignees', width: 24},
	{key: 'priority', width: 10},
	{key: 'labels', width: 20},
	{key: 'start', width: 12},
	{key: 'due', width: 12},
	{key: 'end', width: 12},
	{key: 'percent', width: 10},
	{key: 'done', width: 8},
	{key: 'doneAt', width: 16},
	{key: 'created', width: 16},
	{key: 'updated', width: 16},
]

export interface ExportLookups {
	projectTitle: (projectId: number) => string
	// Only given by the kanban view. Without it the bucket column is left out.
	bucketTitle?: (bucketId: number) => string
}

export interface ExportLabels {
	columns: Record<ExportColumnKey, string>
	sheet: string
}

function userName(user: {name?: string, username?: string} | null | undefined): string {
	return user ? (user.name || user.username || '') : ''
}

function date(value: string | undefined | null) {
	return parseDateOrNull(value)
}

export function taskCell(task: TaskResponse, key: ExportColumnKey, lookups: ExportLookups): XlsxCell {
	switch (key) {
		case 'identifier': return getTaskIdentifier(task)
		case 'title': return task.title
		case 'project': return lookups.projectTitle(task.project_id)
		case 'bucket': return lookups.bucketTitle ? lookups.bucketTitle(task.bucket_id) : ''
		case 'assignor': return userName(task.created_by)
		case 'assignees': return task.assignees.map(userName).filter(Boolean).join(', ')
		case 'priority': return task.priority || null
		case 'labels': return task.labels.map(l => l.title).filter(Boolean).join(', ')
		case 'start': return {value: date(task.start_date), style: 'date'}
		case 'due': return {value: date(task.due_date), style: 'date'}
		case 'end': return {value: date(task.end_date), style: 'date'}
		case 'percent': return {value: task.done ? 1 : task.percent_done, style: 'percent'}
		case 'done': return task.done
		case 'doneAt': return {value: date(task.done_at), style: 'datetime'}
		case 'created': return {value: date(task.created), style: 'datetime'}
		case 'updated': return {value: date(task.updated), style: 'datetime'}
	}
}

export function columnsFor(lookups: ExportLookups, only?: ExportColumnKey[]): ExportColumn[] {
	return ALL_EXPORT_COLUMNS.filter(column => {
		if (column.key === 'bucket' && !lookups.bucketTitle) {
			return false
		}
		return !only || only.includes(column.key)
	})
}

// One flat sheet: what the table and list views export, and the base of the kanban sheet.
export function tasksSheet(
	tasks: TaskResponse[],
	lookups: ExportLookups,
	labels: ExportLabels,
	only?: ExportColumnKey[],
): XlsxSheet {
	const columns = columnsFor(lookups, only)
	return {
		name: labels.sheet,
		header: columns.map(c => labels.columns[c.key]),
		widths: columns.map(c => c.width),
		rows: tasks.map(task => columns.map(c => taskCell(task, c.key, lookups))),
	}
}

export interface BucketGroup {
	id: number
	title: string
}

// Kanban: the tasks in the order of the board, bucket by bucket, and the ones without a known
// bucket at the end.
export function sortByBoard(tasks: TaskResponse[], buckets: BucketGroup[]): TaskResponse[] {
	const order = new Map(buckets.map((b, i) => [b.id, i]))
	return tasks
		.map((task, index) => ({task, index, rank: order.get(task.bucket_id) ?? buckets.length}))
		.sort((a, b) => a.rank - b.rank || a.index - b.index)
		.map(item => item.task)
}

export function capTasks<T>(tasks: T[]): {tasks: T[], truncated: boolean} {
	return tasks.length > MAX_EXPORT_TASKS
		? {tasks: tasks.slice(0, MAX_EXPORT_TASKS), truncated: true}
		: {tasks, truncated: false}
}

const UNSAFE_FILENAME = /[\\/:*?"<>|\u0000-\u001F]/g

// <project>-<view>-<yyyy-mm-dd>.<ext>, safe on every file system.
export function exportFileName(project: string, view: string, extension: string, now = new Date()): string {
	const pad = (n: number) => String(n).padStart(2, '0')
	const day = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
	const base = `${project}-${view}`
		.replace(UNSAFE_FILENAME, '_')
		.replace(/\s+/g, ' ')
		.replace(/^[.\s-]+|[.\s-]+$/g, '')
		.slice(0, 120)
	return `${base || 'export'}-${day}.${extension}`
}
