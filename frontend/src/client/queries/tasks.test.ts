import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {taskQuery, tasksQuery} from './tasks'
import {kanbanQuery} from './kanban'
const sdk = vi.hoisted(() => ({tasksRead: vi.fn(), tasksList: vi.fn(), projectTasksList: vi.fn(), projectViewTasksList: vi.fn(), projectViewBucketsTasksList: vi.fn()}))
vi.mock('@/client/generated', () => sdk)

describe('task queries', () => {
	beforeEach(() => vi.resetAllMocks())
	it('reads detail with requested expansions and leaves dates as strings', async () => {
		const task = {id: 1, due_date: '2026-09-17T12:00:00Z'}
		sdk.tasksRead.mockResolvedValue({data: task})
		const client = new QueryClient()
		expect(await client.fetchQuery(taskQuery(1, ['buckets']))).toEqual(task)
		expect(sdk.tasksRead).toHaveBeenCalledWith(expect.objectContaining({path: {task: 1}, query: {expand: ['buckets']}}))
	})
	it.each([null, 3, -1, -2])('routes project scope %s to the right endpoint', async project => {
		const response = {data: {items: [{id: 1}], page: 1, total_pages: 2}}
		sdk.tasksList.mockResolvedValue(response)
		sdk.projectTasksList.mockResolvedValue(response)
		const client = new QueryClient()
		const options = tasksQuery({project, params: {q: 'hello'}})
		const data = await client.fetchInfiniteQuery(options)
		expect(data.pages[0].items).toEqual([{id: 1}])
		const call = project === null ? sdk.tasksList : sdk.projectTasksList
		expect(call).toHaveBeenCalledWith(expect.objectContaining({query: {q: 'hello', page: 1}}))
		expect(options.getNextPageParam!(data.pages[0], data.pages, 1, [1])).toBe(2)
	})
	it('keeps separate cache data when filters or views change', async () => {
		sdk.projectViewTasksList.mockImplementation(({query}) => Promise.resolve({data: {items: [{title: query.filter}], total_pages: 1}}))
		const client = new QueryClient()
		const first = tasksQuery({project: 1, view: 2, params: {filter: 'done = false'}})
		const second = tasksQuery({project: 1, view: 2, params: {filter: 'done = true'}})
		await client.fetchInfiniteQuery(first)
		await client.fetchInfiniteQuery(second)
		expect(client.getQueryData(first.queryKey)?.pages[0].items?.[0].title).toBe('done = false')
	})
	it('retains paging metadata separately for each bucket on pseudo projects', async () => {
		sdk.projectViewBucketsTasksList.mockResolvedValue({data: {items: [{id: 4, count: 30, tasks: [{id: 1}]}, {id: 5, count: 0, tasks: []}]}})
		const board = await new QueryClient().fetchQuery(kanbanQuery(-1, 2))
		expect(board.pages).toEqual({4: 1, 5: 1})
		expect(board.hasMore).toEqual({4: true, 5: false})
		expect(sdk.projectViewBucketsTasksList).toHaveBeenCalledWith(expect.objectContaining({path: {project: -1, view: 2}}))
	})
})
