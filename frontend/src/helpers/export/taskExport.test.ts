import {describe, expect, it} from 'vitest'

import type {TaskResponse} from '@/client/queries/tasks'
import {
	ALL_EXPORT_COLUMNS,
	capTasks,
	columnsFor,
	exportFileName,
	MAX_EXPORT_TASKS,
	sortByBoard,
	taskCell,
	tasksSheet,
	type ExportLabels,
} from './taskExport'

function task(overrides: Partial<TaskResponse> = {}): TaskResponse {
	return {
		id: 1,
		title: 'Approve invoice',
		description: '',
		done: false,
		priority: 3,
		percent_done: 0.25,
		hex_color: '',
		is_favorite: false,
		identifier: 'PROJ-1',
		index: 1,
		position: 0,
		project_id: 7,
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

const lookups = {projectTitle: (id: number) => `Project ${id}`}
const labels: ExportLabels = {
	sheet: 'Tasks',
	columns: Object.fromEntries(ALL_EXPORT_COLUMNS.map(c => [c.key, c.key.toUpperCase()])) as ExportLabels['columns'],
}

describe('taskCell', () => {
	it('identifier, title and project', () => {
		expect(taskCell(task(), 'identifier', lookups)).toBe('PROJ-1')
		expect(taskCell(task(), 'title', lookups)).toBe('Approve invoice')
		expect(taskCell(task(), 'project', lookups)).toBe('Project 7')
	})

	it('assignor and assignees by name, falling back to the username', () => {
		const t = task({
			created_by: {id: 1, name: 'Jane Doe', username: 'jane'},
			assignees: [{id: 2, name: '', username: 'joe'}, {id: 3, name: 'Ann', username: 'ann'}],
		})
		expect(taskCell(t, 'assignor', lookups)).toBe('Jane Doe')
		expect(taskCell(t, 'assignees', lookups)).toBe('joe, Ann')
		expect(taskCell(task(), 'assignor', lookups)).toBe('')
	})

	it('dates are real dates, the API zero date and missing dates are empty', () => {
		const t = task({start_date: '2026-11-02T00:00:00Z', due_date: '0001-01-01T00:00:00Z'})
		expect(taskCell(t, 'start', lookups)).toEqual({value: new Date('2026-11-02T00:00:00Z'), style: 'date'})
		expect(taskCell(t, 'due', lookups)).toEqual({value: null, style: 'date'})
		expect(taskCell(t, 'end', lookups)).toEqual({value: null, style: 'date'})
	})

	it('percent is complete for done tasks, priority 0 is empty', () => {
		expect(taskCell(task({percent_done: 0.25}), 'percent', lookups)).toEqual({value: 0.25, style: 'percent'})
		expect(taskCell(task({done: true, percent_done: 0.25}), 'percent', lookups)).toEqual({value: 1, style: 'percent'})
		expect(taskCell(task({priority: 0}), 'priority', lookups)).toBeNull()
		expect(taskCell(task({done: true}), 'done', lookups)).toBe(true)
	})

	it('labels are joined', () => {
		const t = task({labels: [{id: 1, title: 'urgent'}, {id: 2, title: 'finance'}] as TaskResponse['labels']})
		expect(taskCell(t, 'labels', lookups)).toBe('urgent, finance')
	})

	it('a title that looks like a formula stays text, unchanged', () => {
		expect(taskCell(task({title: '=HYPERLINK("http://x")'}), 'title', lookups)).toBe('=HYPERLINK("http://x")')
	})
})

describe('columnsFor / tasksSheet', () => {
	it('leaves the bucket column out unless a bucket lookup exists', () => {
		expect(columnsFor(lookups).map(c => c.key)).not.toContain('bucket')
		expect(columnsFor({...lookups, bucketTitle: () => 'x'}).map(c => c.key)).toContain('bucket')
	})

	it('keeps the order of the full set when only some columns are asked for', () => {
		expect(columnsFor(lookups, ['title', 'identifier', 'due']).map(c => c.key)).toEqual(['identifier', 'title', 'due'])
	})

	it('builds a header, widths and one row per task', () => {
		const sheet = tasksSheet([task({id: 1}), task({id: 2, title: 'B'})], lookups, labels, ['identifier', 'title'])

		expect(sheet.name).toBe('Tasks')
		expect(sheet.header).toEqual(['IDENTIFIER', 'TITLE'])
		expect(sheet.widths).toEqual([12, 44])
		expect(sheet.rows).toEqual([['PROJ-1', 'Approve invoice'], ['PROJ-1', 'B']])
	})
})

describe('sortByBoard', () => {
	it('orders by bucket, keeps the order inside a bucket, unknown buckets last', () => {
		const tasks = [
			task({id: 1, bucket_id: 20}),
			task({id: 2, bucket_id: 10}),
			task({id: 3, bucket_id: 99}),
			task({id: 4, bucket_id: 20}),
			task({id: 5, bucket_id: 10}),
		]
		const sorted = sortByBoard(tasks, [{id: 10, title: 'To do'}, {id: 20, title: 'Doing'}])
		expect(sorted.map(t => t.id)).toEqual([2, 5, 1, 4, 3])
	})
})

describe('capTasks', () => {
	it('keeps a normal list and cuts an oversized one', () => {
		expect(capTasks([1, 2, 3])).toEqual({tasks: [1, 2, 3], truncated: false})
		const big = Array.from({length: MAX_EXPORT_TASKS + 5}, (_, i) => i)
		const capped = capTasks(big)
		expect(capped.truncated).toBe(true)
		expect(capped.tasks).toHaveLength(MAX_EXPORT_TASKS)
	})

	it('exactly at the limit is not cut', () => {
		expect(capTasks(Array.from({length: MAX_EXPORT_TASKS}, (_, i) => i)).truncated).toBe(false)
	})
})

describe('exportFileName', () => {
	const day = new Date(2026, 10, 3)

	it('is project, view and day', () => {
		expect(exportFileName('Website', 'Table', 'xlsx', day)).toBe('Website-Table-2026-11-03.xlsx')
	})

	it('replaces characters file systems reject and trims dots and spaces', () => {
		expect(exportFileName('A/B: C*?', ' "x" ', 'xlsx', day)).toBe('A_B_ C__- _x_-2026-11-03.xlsx')
		expect(exportFileName('..', '..', 'xlsx', day)).toBe('export-2026-11-03.xlsx')
	})

	it('keeps non-latin names and limits the length', () => {
		expect(exportFileName('資訊科技', '列表', 'xlsx', day)).toBe('資訊科技-列表-2026-11-03.xlsx')
		expect(exportFileName('x'.repeat(300), 'v', 'xlsx', day).length).toBeLessThan(140)
	})
})
