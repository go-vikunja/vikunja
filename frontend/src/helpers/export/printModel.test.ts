import {describe, expect, it} from 'vitest'

import type {TaskResponse} from '@/client/queries/tasks'
import {buildPrintBoard, buildPrintGantt, buildPrintTable, printValue, type PrintFormatters, type PrintLabels} from './printModel'
import {ALL_EXPORT_COLUMNS} from './taskExport'

function task(id: number, o: Partial<TaskResponse> = {}): TaskResponse {
	return {
		id, title: `T${id}`, description: '', done: false, priority: 0, percent_done: 0, hex_color: '', is_favorite: false,
		identifier: `P-${id}`, index: id, position: 0, project_id: 1, bucket_id: 0, repeat_after: 0, repeat_mode: 0,
		cover_image_attachment_id: 0, labels: [], assignees: [], reminders: [], attachments: [], related_tasks: {}, reactions: {},
		...o,
	} as TaskResponse
}

const fmt: PrintFormatters = {
	date: d => `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()}`,
	dateTime: d => `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()} ${d.getHours()}h`,
}
const labels: PrintLabels = {
	yes: 'yes', no: 'no',
	columns: Object.fromEntries(ALL_EXPORT_COLUMNS.map(c => [c.key, `H:${c.key}`])) as PrintLabels['columns'],
}
const lookups = {projectTitle: (id: number) => `Project ${id}`}
const d = (day: number) => new Date(2026, 10, day)

describe('printValue', () => {
	it('formats the dates and flags for paper', () => {
		const t = task(1, {due_date: d(5).toISOString(), done: true, percent_done: 0.3})
		expect(printValue(t, 'due', lookups, fmt, labels)).toBe('2026-11-5')
		expect(printValue(t, 'done', lookups, fmt, labels)).toBe('yes')
		expect(printValue(t, 'percent', lookups, fmt, labels)).toBe('100%')
		expect(printValue(task(2), 'due', lookups, fmt, labels)).toBe('')
		expect(printValue(task(2), 'priority', lookups, fmt, labels)).toBe('')
	})

	it('writes text as it is, so print markup is the template\'s job', () => {
		expect(printValue(task(1, {title: '<b>x</b> & y'}), 'title', lookups, fmt, labels)).toBe('<b>x</b> & y')
	})
})

describe('buildPrintTable', () => {
	it('headers and rows follow the column order of the export, not the order asked for', () => {
		const table = buildPrintTable([task(1), task(2, {done: true})], ['title', 'identifier'], lookups, fmt, labels)
		expect(table.headers).toEqual(['H:identifier', 'H:title'])
		expect(table.rows).toEqual([['P-1', 'T1'], ['P-2', 'T2']])
		expect(table.doneRows).toEqual([false, true])
	})

	it('leaves the bucket column out without a bucket lookup', () => {
		expect(buildPrintTable([task(1)], ['bucket', 'title'], lookups, fmt, labels).headers).toEqual(['H:title'])
		expect(buildPrintTable([task(1)], ['bucket', 'title'], {...lookups, bucketTitle: () => 'Doing'}, fmt, labels).rows).toEqual([['Doing', 'T1']])
	})
})

describe('buildPrintBoard', () => {
	it('groups cards by bucket in board order and keeps tasks of unknown buckets', () => {
		const tasks = [task(1, {bucket_id: 20}), task(2, {bucket_id: 10}), task(3, {bucket_id: 99}), task(4, {bucket_id: 10, done: true})]

		const board = buildPrintBoard(tasks, [{id: 10, title: 'To do'}, {id: 20, title: 'Doing'}], {...lookups, bucketTitle: () => 'Other'}, fmt)

		expect(board.columns.map(c => c.title)).toEqual(['To do', 'Doing', 'Other'])
		expect(board.columns[0].tasks.map(t => t.title)).toEqual(['T2', 'T4'])
		expect(board.columns[0].tasks[1].done).toBe(true)
		expect(board.columns[2].tasks.map(t => t.title)).toEqual(['T3'])
	})

	it('a card lists id, assignees, due date and labels, skipping what is empty', () => {
		const t = task(1, {bucket_id: 10, due_date: d(5).toISOString(), assignees: [{id: 1, name: 'Jane', username: 'j'}]})
		const board = buildPrintBoard([t], [{id: 10, title: 'A'}], lookups, fmt)
		expect(board.columns[0].tasks[0].details).toEqual(['P-1', 'Jane', '2026-11-5'])
	})
})

describe('buildPrintGantt', () => {
	const formatUnit = (date: Date) => String(date.getDate())
	const map = (...ts: TaskResponse[]) => new Map(ts.map(t => [t.id, t]))

	it('places bars by date: x from the start, width over the end day', () => {
		const t = task(1, {start_date: d(3).toISOString(), end_date: d(5).toISOString()})

		const gantt = buildPrintGantt(map(t), d(1), d(10), 'day', false, formatUnit, d(2))

		expect(gantt.width).toBe(10 * 16)
		expect(gantt.rows[0].bar).toEqual({x: 2 * 16, width: 3 * 16})
		expect(gantt.rows[0].duration).toBe(3)
		expect(gantt.todayX).toBe(16 + 8)
	})

	it('a milestone is a point in the middle of its day', () => {
		const gantt = buildPrintGantt(map(task(1, {is_milestone: true, end_date: d(4).toISOString()})), d(1), d(10), 'day', false, formatUnit, d(20))

		expect(gantt.rows[0].kind).toBe('milestone')
		expect(gantt.rows[0].bar).toEqual({x: 3 * 16 + 8, width: 0})
		expect(gantt.todayX).toBeNull()
	})

	it('a task without dates has no bar, a parent spans its children', () => {
		const parent = task(1, {related_tasks: {subtask: [task(2)]} as TaskResponse['related_tasks']})
		const child = task(2, {start_date: d(2).toISOString(), end_date: d(3).toISOString(), related_tasks: {parenttask: [task(1)]} as TaskResponse['related_tasks']})

		const gantt = buildPrintGantt(map(parent, child, task(3)), d(1), d(10), 'day', false, formatUnit, d(20))

		expect(gantt.rows[0].kind).toBe('summary')
		expect(gantt.rows[0].bar).not.toBeNull()
		expect(gantt.rows[1].wbs).toBe('1.1')
		expect(gantt.rows[2].bar).toBeNull()
	})

	it('dependencies connect the rows, critical ones are marked', () => {
		const a = task(1, {start_date: d(1).toISOString(), end_date: d(3).toISOString(), related_tasks: {precedes: [task(2)]} as TaskResponse['related_tasks']})
		const b = task(2, {start_date: d(3).toISOString(), end_date: d(5).toISOString(), related_tasks: {follows: [task(1)]} as TaskResponse['related_tasks']})

		const plain = buildPrintGantt(map(a, b), d(1), d(10), 'day', false, formatUnit, d(20))
		const crit = buildPrintGantt(map(a, b), d(1), d(10), 'day', true, formatUnit, d(20))

		expect(plain.deps).toEqual([{fromRow: 0, toRow: 1, critical: false}])
		expect(crit.deps).toEqual([{fromRow: 0, toRow: 1, critical: true}])
		expect(crit.rows.map(r => r.critical)).toEqual([true, true])
		expect(plain.rows.map(r => r.critical)).toEqual([false, false])
	})

	it('a dependency to a task that is not loaded is left out', () => {
		const a = task(1, {start_date: d(1).toISOString(), end_date: d(3).toISOString(), related_tasks: {precedes: [task(99)]} as TaskResponse['related_tasks']})
		expect(buildPrintGantt(map(a), d(1), d(10), 'day', false, formatUnit, d(20)).deps).toEqual([])
	})

	it('the header follows the zoom', () => {
		const t = task(1, {start_date: d(2).toISOString(), end_date: d(3).toISOString()})
		const week = buildPrintGantt(map(t), d(2), d(22), 'week', false, formatUnit, d(20))
		expect(week.lower.map(c => c.label)).toEqual(['2', '9', '16'])
		const day = buildPrintGantt(map(t), d(2), d(4), 'day', false, formatUnit, d(20))
		expect(day.lower.map(c => c.label)).toEqual(['2', '3', '4'])
	})
})
