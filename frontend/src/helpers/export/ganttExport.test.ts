import {describe, expect, it} from 'vitest'

import type {TaskResponse} from '@/client/queries/tasks'
import {buildGanttSheets, MAX_TIMELINE_COLUMNS, timelineUnits, type GanttExportLabels} from './ganttExport'

function task(id: number, overrides: Partial<TaskResponse> = {}): TaskResponse {
	return {
		id,
		title: `T${id}`,
		description: '',
		done: false,
		priority: 0,
		percent_done: 0,
		hex_color: '',
		is_favorite: false,
		identifier: `P-${id}`,
		index: id,
		position: 0,
		project_id: 1,
		bucket_id: 0,
		repeat_after: 0,
		repeat_mode: 0,
		cover_image_attachment_id: 0,
		labels: [],
		assignees: [],
		reminders: [],
		attachments: [],
		related_tasks: {},
		reactions: {},
		...overrides,
	} as TaskResponse
}

const labels: GanttExportLabels = {
	tasksSheet: 'Tasks', timelineSheet: 'Timeline', wbs: 'WBS', name: 'Name', duration: 'Duration',
	start: 'Start', finish: 'Finish', percent: '%', predecessors: 'Pred', assignees: 'Who', critical: 'Critical',
	truncated: 'cut',
}

const d = (day: number) => new Date(2026, 10, day) // Nov 2026

function map(...tasks: TaskResponse[]) {
	return new Map(tasks.map(t => [t.id, t]))
}

const base = {labels, formatUnit: (date: Date) => String(date.getDate())}

describe('timelineUnits', () => {
	it('one unit per day at fit and day zoom', () => {
		const units = timelineUnits(d(2), d(4), 'day')
		expect(units.map(u => u.start.getDate())).toEqual([2, 3, 4])
		expect(units[0].end.getDate()).toBe(3)
	})

	it('weeks at week zoom, ending where the next starts', () => {
		const units = timelineUnits(d(2), d(22), 'week') // Mon 2nd .. Sun 22nd
		expect(units.map(u => u.start.getDate())).toEqual([2, 9, 16])
		expect(units[0].end.getDate()).toBe(9)
	})
})

describe('buildGanttSheets', () => {
	it('writes the pane columns with WBS, indented names, duration and predecessors', () => {
		const parent = task(1, {related_tasks: {subtask: [task(2)]} as TaskResponse['related_tasks']})
		const child = task(2, {
			start_date: d(3).toISOString(), end_date: d(5).toISOString(), percent_done: 0.5,
			related_tasks: {parenttask: [task(1)], follows: [task(3)]} as TaskResponse['related_tasks'],
			assignees: [{id: 1, name: 'Jane', username: 'jane'}],
		})
		const other = task(3, {start_date: d(1).toISOString(), end_date: d(2).toISOString()})

		const {sheets} = buildGanttSheets({...base, tasks: map(parent, child, other), zoom: 'day', from: d(1), to: d(8)})
		const rows = sheets[0].rows

		const childRow = rows.find(r => String(r[1]).trim() === 'T2') as unknown[]
		expect(sheets[0].header.slice(0, 3)).toEqual(['WBS', 'Name', 'Duration'])
		expect(childRow[0]).toBe('1.1')
		expect(String(childRow[1])).toBe('  T2')
		expect(childRow[2]).toBe(3)
		expect(childRow[6]).toBe('2')
		expect(childRow[7]).toBe('Jane')
	})

	it('a parent without dates spans its children', () => {
		const parent = task(1, {related_tasks: {subtask: [task(2)]} as TaskResponse['related_tasks']})
		const child = task(2, {start_date: d(3).toISOString(), end_date: d(5).toISOString(), related_tasks: {parenttask: [task(1)]} as TaskResponse['related_tasks']})

		const {sheets} = buildGanttSheets({...base, tasks: map(parent, child), zoom: 'day', from: d(1), to: d(8)})

		const parentRow = sheets[0].rows[0] as unknown[]
		expect(parentRow[2]).toBe(3)
	})

	it('the timeline paints the days a task is on, the end day included', () => {
		const t = task(1, {start_date: d(3).toISOString(), end_date: d(5).toISOString()})

		const {sheets} = buildGanttSheets({...base, tasks: map(t), zoom: 'day', from: d(1), to: d(8)})
		const timeline = sheets[1]
		const cells = timeline.rows[0].slice(2) as Array<{style?: string} | null>

		expect(timeline.header.slice(2)).toEqual(['1', '2', '3', '4', '5', '6', '7', '8'])
		expect(cells.map(c => (c ? c.style : null))).toEqual([null, null, 'barTask', 'barTask', 'barTask', null, null, null])
	})

	it('milestones are one cell in their own colour, summary bars in theirs', () => {
		const parent = task(1, {related_tasks: {subtask: [task(2)]} as TaskResponse['related_tasks']})
		const child = task(2, {start_date: d(2).toISOString(), end_date: d(3).toISOString(), related_tasks: {parenttask: [task(1)]} as TaskResponse['related_tasks']})
		const milestone = task(3, {is_milestone: true, end_date: d(6).toISOString()})

		const {sheets} = buildGanttSheets({...base, tasks: map(parent, child, milestone), zoom: 'day', from: d(1), to: d(8)})
		const style = (row: number) => (sheets[1].rows[row].slice(2) as Array<{style?: string} | null>).map(c => c?.style ?? null)

		expect(style(0)).toEqual([null, 'barSummary', 'barSummary', null, null, null, null, null])
		expect(style(2)).toEqual([null, null, null, null, null, 'barMilestone', null, null])
	})

	it('critical tasks get the critical colour and a marker in the table', () => {
		const a = task(1, {start_date: d(1).toISOString(), end_date: d(3).toISOString(), related_tasks: {precedes: [task(2)]} as TaskResponse['related_tasks']})
		const b = task(2, {start_date: d(3).toISOString(), end_date: d(5).toISOString(), related_tasks: {follows: [task(1)]} as TaskResponse['related_tasks']})

		const {sheets} = buildGanttSheets({...base, tasks: map(a, b), zoom: 'day', from: d(1), to: d(6), showCritical: true})

		expect(sheets[0].header[8]).toBe('Critical')
		expect(sheets[0].rows.map(r => r[8])).toEqual(['x', 'x'])
		expect((sheets[1].rows[0][2] as {style: string}).style).toBe('barCritical')
	})

	it('no critical column without the option', () => {
		const t = task(1, {start_date: d(1).toISOString(), end_date: d(2).toISOString()})
		const {sheets} = buildGanttSheets({...base, tasks: map(t), zoom: 'day', from: d(1), to: d(3)})
		expect(sheets[0].header).toHaveLength(8)
	})

	it('cuts an overlong timeline and says so', () => {
		const t = task(1, {start_date: d(1).toISOString(), end_date: d(2).toISOString()})
		const result = buildGanttSheets({
			...base, tasks: map(t), zoom: 'day', from: new Date(2026, 0, 1), to: new Date(2027, 11, 31),
		})

		expect(result.truncated).toBe(true)
		expect(result.sheets[1].header).toHaveLength(2 + MAX_TIMELINE_COLUMNS)
	})

	it('week zoom uses one column per week', () => {
		const t = task(1, {start_date: d(3).toISOString(), end_date: d(12).toISOString()})

		const {sheets} = buildGanttSheets({...base, tasks: map(t), zoom: 'week', from: d(2), to: d(22)})
		const cells = (sheets[1].rows[0].slice(2) as Array<{style?: string} | null>).map(c => c?.style ?? null)

		expect(sheets[1].header.slice(2)).toEqual(['2', '9', '16'])
		expect(cells).toEqual(['barTask', 'barTask', null])
	})

	it('a task without dates is listed but not drawn', () => {
		const {sheets} = buildGanttSheets({...base, tasks: map(task(1)), zoom: 'day', from: d(1), to: d(3)})
		expect(sheets[0].rows).toHaveLength(1)
		expect((sheets[1].rows[0].slice(2) as unknown[]).every(c => c === null)).toBe(true)
	})
})
