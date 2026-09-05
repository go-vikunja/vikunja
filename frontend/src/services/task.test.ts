import {describe, it, expect, vi, beforeEach} from 'vitest'

import TaskService from './task'
import TaskModel from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'

interface BulkPayload {
	tasks: {title: string}[],
}

interface BulkResponse {
	data: {tasks: {id: number, title: string}[]},
}

const post = vi.hoisted(() => vi.fn<(url: string, payload: BulkPayload) => Promise<BulkResponse>>())

vi.mock('@/helpers/fetcher', () => ({
	getApiBaseUrl: () => '/api/v1/',
	apiV2Url: (path: string) => `/api/v2/${path}`,
	HTTPFactory: () => ({post, interceptors: {request: {use: vi.fn()}, response: {use: vi.fn()}}}),
	AuthenticatedHTTPFactory: () => ({post, interceptors: {request: {use: vi.fn()}, response: {use: vi.fn()}}}),
}))

let nextId = 1

function echoResponse(payload: BulkPayload): BulkResponse {
	return {data: {tasks: payload.tasks.map(t => ({id: nextId++, title: t.title}))}}
}

function buildTasks(titles: string[], projectId: number): ITask[] {
	return titles.map(title => new TaskModel({title, projectId}))
}

function titlesOfCall(call: [string, BulkPayload]): string[] {
	return call[1].tasks.map(t => t.title)
}

describe('TaskService.bulkCreate', () => {
	beforeEach(() => {
		nextId = 1
		post.mockReset()
		post.mockImplementation(async (_url, payload) => echoResponse(payload))
	})

	it('creates all tasks of one project in a single request', async () => {
		const titles = ['first', 'second', 'third']
		const {tasks, error} = await new TaskService().bulkCreate(buildTasks(titles, 42))

		expect(post).toHaveBeenCalledTimes(1)
		expect(post.mock.calls[0][0]).toBe('/api/v2/projects/42/tasks/bulk')
		expect(error).toBeNull()
		expect(tasks.map(t => t?.title)).toEqual(titles)
	})

	it('sends only the fields the v2 task schema allows', async () => {
		await new TaskService().bulkCreate([new TaskModel({title: 'first', projectId: 42})])

		const payloadTask = post.mock.calls[0][1].tasks[0]
		expect(Object.keys(payloadTask).sort()).toEqual([
			'assignees',
			'bucket_id',
			'description',
			'done',
			'due_date',
			'end_date',
			'hex_color',
			'is_favorite',
			'percent_done',
			'priority',
			'reminders',
			'repeat_after',
			'repeat_mode',
			'start_date',
			'title',
		])
	})

	it('sends one request per project and keeps the input order', async () => {
		const input = [
			new TaskModel({title: 'a', projectId: 1}),
			new TaskModel({title: 'b', projectId: 2}),
			new TaskModel({title: 'c', projectId: 1}),
			new TaskModel({title: 'd', projectId: 2}),
		]

		const {tasks, error} = await new TaskService().bulkCreate(input)

		expect(post).toHaveBeenCalledTimes(2)
		expect(post.mock.calls.map(c => c[0])).toEqual([
			'/api/v2/projects/1/tasks/bulk',
			'/api/v2/projects/2/tasks/bulk',
		])
		expect(titlesOfCall(post.mock.calls[0])).toEqual(['a', 'c'])
		expect(titlesOfCall(post.mock.calls[1])).toEqual(['b', 'd'])
		expect(error).toBeNull()
		expect(tasks.map(t => t?.title)).toEqual(['a', 'b', 'c', 'd'])
	})

	it('chunks at 100 tasks and submits the last chunk first', async () => {
		const titles = Array.from({length: 205}, (_, i) => `task ${i}`)

		const {tasks, error} = await new TaskService().bulkCreate(buildTasks(titles, 7))

		expect(post).toHaveBeenCalledTimes(3)
		expect(post.mock.calls.map(c => c[1].tasks.length)).toEqual([5, 100, 100])
		expect(titlesOfCall(post.mock.calls[0])).toEqual(titles.slice(200))
		expect(titlesOfCall(post.mock.calls[1])).toEqual(titles.slice(100, 200))
		expect(titlesOfCall(post.mock.calls[2])).toEqual(titles.slice(0, 100))
		expect(error).toBeNull()
		expect(tasks.map(t => t?.title)).toEqual(titles)
	})

	it('does not send an empty trailing batch', async () => {
		const titles = Array.from({length: 200}, (_, i) => `task ${i}`)

		const {tasks} = await new TaskService().bulkCreate(buildTasks(titles, 7))

		expect(post.mock.calls.map(c => c[1].tasks.length)).toEqual([100, 100])
		expect(tasks.map(t => t?.title)).toEqual(titles)
	})

	it('stops after a failed batch and keeps the batches sent before it', async () => {
		const failure = new Error('boom')
		let calls = 0
		post.mockImplementation(async (_url, payload) => {
			calls++
			if (calls === 2) {
				throw failure
			}
			return echoResponse(payload)
		})

		const titles = Array.from({length: 205}, (_, i) => `task ${i}`)
		const {tasks, error} = await new TaskService().bulkCreate(buildTasks(titles, 7))

		expect(post).toHaveBeenCalledTimes(2)
		expect(error).toBe(failure)
		expect(tasks.slice(200).map(t => t?.title)).toEqual(titles.slice(200))
		expect(tasks.slice(0, 200).every(t => t === null)).toBe(true)
	})

	it('fails the batch when the response does not match the payload', async () => {
		post.mockImplementation(async () => ({data: {tasks: []}}))

		const {tasks, error} = await new TaskService().bulkCreate(buildTasks(['a', 'b'], 7))

		expect(error).toBeInstanceOf(Error)
		expect(tasks).toEqual([null, null])
	})

	it('keeps tasks created by other projects when a batch fails', async () => {
		const failure = new Error('boom')
		post.mockImplementation(async (url, payload) => {
			if (url === '/api/v2/projects/1/tasks/bulk') {
				throw failure
			}
			return echoResponse(payload)
		})

		const input = [
			new TaskModel({title: 'a', projectId: 1}),
			new TaskModel({title: 'b', projectId: 2}),
			new TaskModel({title: 'c', projectId: 1}),
		]

		const {tasks, error} = await new TaskService().bulkCreate(input)

		expect(error).toBe(failure)
		expect(tasks[0]).toBeNull()
		expect(tasks[2]).toBeNull()
		expect(tasks[1]?.title).toBe('b')
	})
})

interface ProcessedTask {
	repeat_after: number,
	hex_color: string,
	reminders: unknown[],
	related_tasks: Record<string, ProcessedTask[]>,
}

// A task as it reaches processModel: a plain object, not necessarily built by TaskModel.
function rawTask(overrides: Record<string, unknown> = {}) {
	return {
		id: 1,
		title: 'a task',
		projectId: 1,
		hexColor: '',
		reminders: [],
		repeatAfter: {type: 'days', amount: 0},
		relatedTasks: {},
		attachments: [],
		...overrides,
	} as unknown as ITask
}

function process(task: ITask): ProcessedTask {
	return new TaskService().processModel(task) as unknown as ProcessedTask
}

describe('TaskService.processModel', () => {
	it('converts a repeatAfter object to seconds', () => {
		expect(process(rawTask({repeatAfter: {type: 'hours', amount: 3}})).repeat_after).toBe(3 * 3600)
		expect(process(rawTask({repeatAfter: {type: 'days', amount: 2}})).repeat_after).toBe(2 * 86400)
		expect(process(rawTask({repeatAfter: {type: 'weeks', amount: 1}})).repeat_after).toBe(7 * 86400)
	})

	it('sends no repeat for an amount of 0', () => {
		expect(process(rawTask({repeatAfter: {type: 'days', amount: 0}})).repeat_after).toBe(0)
	})

	it('keeps a repeatAfter that is already in seconds', () => {
		expect(process(rawTask({repeatAfter: 3600})).repeat_after).toBe(3600)
	})

	it('sends no repeat when repeatAfter is missing or null', () => {
		expect(process(rawTask({repeatAfter: undefined})).repeat_after).toBe(0)
		expect(process(rawTask({repeatAfter: null})).repeat_after).toBe(0)
	})

	it('processes a nested related task that has no repeatAfter', () => {
		const task = rawTask({relatedTasks: {subtask: [{id: 2, title: 'sub', projectId: 1}]}})

		expect(process(task).related_tasks.subtask[0].repeat_after).toBe(0)
	})

	it('processes a nested related task that has no reminders', () => {
		const task = rawTask({relatedTasks: {subtask: [{id: 2, title: 'sub', repeatAfter: 0}]}})

		expect(process(task).related_tasks.subtask[0].reminders).toEqual([])
	})

	it('processes a nested related task that has no attachments', () => {
		const task = rawTask({relatedTasks: {subtask: [{id: 2, title: 'sub', reminders: []}]}})

		expect(() => process(task)).not.toThrow()
	})

	it('processes a task that has no relatedTasks', () => {
		expect(() => process(rawTask({relatedTasks: undefined}))).not.toThrow()
	})

	it('processes a task that has no hexColor', () => {
		expect(process(rawTask({hexColor: undefined})).hex_color).toBe('')
	})

	it('leaves the related tasks of the passed model untouched', () => {
		const subtask = rawTask({id: 2, title: 'sub', repeatAfter: {type: 'days', amount: 1}})
		const relatedTasks = {subtask: [subtask]}

		process(rawTask({relatedTasks}))

		expect(relatedTasks.subtask[0]).toBe(subtask)
		expect(subtask.repeatAfter).toEqual({type: 'days', amount: 1})
	})

	it('can process the same task twice', () => {
		const task = rawTask({relatedTasks: {subtask: [rawTask({id: 2, repeatAfter: {type: 'days', amount: 1}})]}})

		expect(process(task).related_tasks.subtask[0].repeat_after).toBe(86400)
		expect(process(task).related_tasks.subtask[0].repeat_after).toBe(86400)
	})

	it('normalises a missing relatedTasks to an empty object', () => {
		expect(process(rawTask({relatedTasks: undefined})).related_tasks).toEqual({})
	})
})
