import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {normalizePageNumber, normalizeTask, taskKeys, taskQuery, tasksQuery, allTasksQuery} from './tasks'
import {bucketHasMore, bucketKeys, bucketsQuery, kanbanKeys, kanbanQuery, TASKS_PER_BUCKET} from './kanban'
const sdk = vi.hoisted(() => ({
	tasksRead: vi.fn(),
	tasksList: vi.fn(),
	projectTasksList: vi.fn(),
	projectViewTasksList: vi.fn(),
	projectViewBucketsTasksList: vi.fn(),
	bucketsList: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)

describe('task queries', () => {
	beforeEach(() => vi.resetAllMocks())
	it('reads detail with requested expansions and leaves dates as strings', async () => {
		const task = {id: 1, due_date: '2026-09-17T12:00:00Z'}
		sdk.tasksRead.mockResolvedValue({data: task})
		const client = new QueryClient()
		expect(await client.fetchQuery(taskQuery(1, ['buckets']))).toEqual(normalizeTask(task))
		expect(sdk.tasksRead).toHaveBeenCalledWith(expect.objectContaining({
			path: {task: 1},
			query: {expand: ['buckets']},
		}))
	})
	it('keeps the permission the read-one body adds', async () => {
		sdk.tasksRead.mockResolvedValue({data: {id: 1, max_permission: 2}})
		expect(await new QueryClient().fetchQuery(taskQuery(1))).toMatchObject({id: 1, max_permission: 2})
	})
	it('fills every guaranteed field at the read boundary', () => {
		expect(normalizeTask({id: 1})).toEqual({
			id: 1, title: '', description: '', done: false, priority: 0, percent_done: 0, hex_color: '',
			is_favorite: false, identifier: '', index: 0, position: 0, project_id: 0, bucket_id: 0,
			repeat_after: 0, repeat_mode: 0, cover_image_attachment_id: 0,
			labels: [], assignees: [], reminders: [], attachments: [], related_tasks: {}, reactions: {},
		})
	})
	it('replaces null collections and normalizes related tasks', () => {
		const task = normalizeTask({
			id: 1, labels: null, assignees: null, reminders: null, attachments: null,
			reactions: {'👍': null}, related_tasks: {subtask: [{id: 2, title: 'Child'}], parenttask: null},
		})
		expect(task).toMatchObject({
			labels: [],
			assignees: [],
			reminders: [],
			attachments: [],
			reactions: {'👍': []},
			related_tasks: {parenttask: []},
		})
		expect(task.related_tasks.subtask[0]).toMatchObject({id: 2, title: 'Child', labels: [], related_tasks: {}})
	})
	it('keeps related tasks one level deep like the API', () => {
		const task = normalizeTask({id: 1, related_tasks: {subtask: [{id: 2, related_tasks: {parenttask: [{id: 1}]}}]}})
		expect(task.related_tasks.subtask[0].related_tasks).toEqual({})
	})
	it('leaves expand-only fields absent', () => {
		const task = normalizeTask({id: 1})
		for (const field of [
			'buckets',
			'comments',
			'comment_count',
			'time_entries_count',
			'is_unread',
			'subscription',
		]) {
			expect(task).not.toHaveProperty(field)
		}
	})
	it('rejects a task without an id', () => {
		expect(() => normalizeTask({})).toThrow('missing an id')
	})
	it.each([
		{scope: {project: null}, fn: 'tasksList', path: undefined},
		{scope: {project: 3}, fn: 'projectTasksList', path: {project: 3}},
		{scope: {project: -1}, fn: 'projectTasksList', path: {project: -1}},
		{scope: {project: -2}, fn: 'projectTasksList', path: {project: -2}},
		{scope: {project: 3, view: 4}, fn: 'projectViewTasksList', path: {project: 3, view: 4}},
	] as const)('routes $scope to $fn', async ({scope, fn, path}) => {
		sdk[fn].mockResolvedValue({data: {items: [{id: 1}], page: 1, total_pages: 2}})
		const client = new QueryClient()
		const data = await client.fetchQuery(tasksQuery({...scope, params: {q: 'hello'}}))
		expect(data.items).toEqual([normalizeTask({id: 1})])
		expect(sdk[fn]).toHaveBeenCalledExactlyOnceWith({
			...(path && {path}),
			query: {q: 'hello', page: 1},
			signal: expect.any(AbortSignal),
		})
		for (const [name, mock] of Object.entries(sdk)) {
			if (name !== fn) expect(mock).not.toHaveBeenCalled()
		}
		await client.fetchQuery(tasksQuery({...scope, params: {q: 'hello'}}, 2))
		expect(sdk[fn]).toHaveBeenLastCalledWith({
			...(path && {path}),
			query: {q: 'hello', page: 2},
			signal: expect.any(AbortSignal),
		})
	})
	it('fills the paging envelope when the response omits it', async () => {
		sdk.tasksList.mockResolvedValue({data: {items: null}})
		expect(await new QueryClient().fetchQuery(tasksQuery())).toEqual({
			items: [],
			total: 0,
			page: 1,
			per_page: 0,
			total_pages: 0,
		})
	})
	it.each([NaN, 0])('falls back to the first page for page %s', async page => {
		sdk.tasksList.mockResolvedValue({data: {items: [], total_pages: 1}})
		const options = tasksQuery({}, page)
		expect(options.queryKey).toEqual(taskKeys.list({}, 1))
		await new QueryClient().fetchQuery(options)
		expect(sdk.tasksList).toHaveBeenCalledWith(expect.objectContaining({query: {page: 1}}))
	})
	it.each([
		['2', 2],
		['abc', 1],
		['0', 1],
		['-3', 1],
		['1.5', 1],
		[undefined, 1],
		[7, 7],
		[null, 1],
		['', 1],
		[['1', '2'], 1],
	])('normalizes page %s to %i', (raw, expected) => {
		expect(normalizePageNumber(raw)).toBe(expected)
	})
	it('retains paging metadata separately for each bucket on pseudo projects', async () => {
		sdk.projectViewBucketsTasksList.mockResolvedValue({
			data: {
				items: [
					{id: 4, count: 30, tasks: [{id: 1}]},
					{id: 5, count: 0, tasks: []},
				],
			},
		})
		const board = await new QueryClient().fetchQuery(kanbanQuery(-1, 2))
		expect(board.pages).toEqual({4: 1, 5: 1})
		expect(board.buckets.map(bucketHasMore)).toEqual([true, false])
		expect(board.buckets[0].tasks).toEqual([normalizeTask({id: 1})])
		expect(sdk.projectViewBucketsTasksList).toHaveBeenCalledExactlyOnceWith({
			path: {project: -1, view: 2},
			query: {per_page: TASKS_PER_BUCKET, expand: ['comment_count', 'is_unread']},
			signal: expect.any(AbortSignal),
		})
	})
	it('defaults the fields board consumers rely on', async () => {
		sdk.projectViewBucketsTasksList.mockResolvedValue({data: {items: [{id: 4}]}})
		const board = await new QueryClient().fetchQuery(kanbanQuery(1, 2))
		expect(board.buckets[0]).toEqual({id: 4, title: '', count: 0, limit: 0, position: 0, tasks: []})
	})
	it('lists buckets without their tasks under a key of its own', async () => {
		sdk.bucketsList.mockResolvedValue({data: {items: [{id: 4, title: 'Backlog'}]}})
		const buckets = await new QueryClient().fetchQuery(bucketsQuery(1, 2))
		expect(buckets).toEqual([{id: 4, title: 'Backlog'}])
		expect(sdk.bucketsList).toHaveBeenCalledExactlyOnceWith({
			path: {project: 1, view: 2},
			signal: expect.any(AbortSignal),
		})
		expect(sdk.projectViewBucketsTasksList).not.toHaveBeenCalled()
	})
	it('leaves bucket lists alone when the board key root is written to', () => {
		const client = new QueryClient()
		client.setQueryData(kanbanKeys.board(1, 2), 'board')
		client.setQueryData(bucketKeys.list(1, 2), 'buckets')
		client.setQueriesData({queryKey: kanbanKeys.all}, () => 'touched')
		expect(client.getQueryData(bucketKeys.list(1, 2))).toBe('buckets')
		expect(client.getQueryData(kanbanKeys.board(1, 2))).toBe('touched')
	})
	it('does not list buckets without a project or view', () => {
		expect(bucketsQuery(0, 2).enabled).toBe(false)
		expect(bucketsQuery(1, 0).enabled).toBe(false)
		expect(bucketsQuery(-1, 2).enabled).toBe(true)
	})
	it('does not load a board without a project or view', () => {
		expect(kanbanQuery(0, 2).enabled).toBe(false)
		expect(kanbanQuery(1, 0).enabled).toBe(false)
		expect(kanbanQuery(-1, 2).enabled).toBe(true)
	})
	it('reads the project out of list keys only', () => {
		expect(taskKeys.projectOf(taskKeys.list({project: 3, view: 4}, 2))).toBe(3)
		expect(taskKeys.projectOf(taskKeys.allList({project: -1}))).toBe(-1)
		expect(taskKeys.projectOf(taskKeys.list({}))).toBeNull()
		expect(taskKeys.projectOf(taskKeys.detail(7))).toBeUndefined()
		expect(kanbanKeys.projectOf(kanbanKeys.board(5, 2))).toBe(5)
		expect(kanbanKeys.projectOf(kanbanKeys.view(-2, 2))).toBe(-2)
	})
	it('reads the view out of list and board keys', () => {
		expect(taskKeys.viewOf(taskKeys.list({project: 3, view: 4}, 2))).toBe(4)
		expect(taskKeys.viewOf(taskKeys.allList({project: 3, view: 4}))).toBe(4)
		expect(taskKeys.viewOf(taskKeys.list({project: 3}))).toBe(0)
		expect(taskKeys.viewOf(taskKeys.detail(7))).toBeUndefined()
		expect(kanbanKeys.viewOf(kanbanKeys.board(5, 2))).toBe(2)
		expect(kanbanKeys.viewOf(kanbanKeys.view(-2, 6))).toBe(6)
	})
	it('reads the filter params out of list keys only', () => {
		expect(
			taskKeys.paramsOf(taskKeys.list({project: 3, view: 4, params: {filter: 'done = false'}}, 2)),
		).toEqual({filter: 'done = false'})
		expect(
			taskKeys.paramsOf(taskKeys.allList({project: 3, params: {sort_by: ['position']}})),
		).toEqual({sort_by: ['position']})
		expect(taskKeys.paramsOf(taskKeys.list({}))).toEqual({})
		expect(taskKeys.paramsOf(taskKeys.detail(7))).toBeUndefined()
	})
	it('exhausts all pages for a Gantt scope', async () => {
		sdk.projectViewTasksList.mockImplementation(({query}) => ({
			data: {items: [{id: query.page}], total_pages: 3},
		}))
		const tasks = await new QueryClient().fetchQuery(allTasksQuery({project: 1, view: 2}))
		expect(tasks.map(task => task.id)).toEqual([1, 2, 3])
		expect(tasks[0]).toEqual(normalizeTask({id: 1}))
		expect(sdk.projectViewTasksList).toHaveBeenLastCalledWith(expect.objectContaining({
			path: {project: 1, view: 2},
			query: {page: 3, per_page: 1000},
		}))
	})
})
