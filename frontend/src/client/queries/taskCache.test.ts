import {describe, expect, it} from 'vitest'
import {QueryClient, type InfiniteData} from '@tanstack/vue-query'
import type {PaginatedTask, Task} from '@/client/generated'
import {taskKeys} from './tasks'
import {kanbanKeys, type BoardData} from './kanban'
import {replaceTaskEverywhere, removeTaskEverywhere} from './taskCache'

describe('task cache reconciliation', () => {
	it.each([1, -1, -2])('updates populated lists and boards for project %s', project => {
		const client = new QueryClient()
		const task = {id: 3, title: 'old', project_id: 1, position: 20, bucket_id: 5}
		const list = taskKeys.list({project})
		const board = kanbanKeys.board(project, 8)
		client.setQueryData(taskKeys.detail(3), task)
		client.setQueryData(list, {pages: [{items: [task], total: 1}], pageParams: [1]})
		client.setQueryData(board, {buckets: [{id: 5, count: 1, tasks: [task]}], pages: {5: 2}, hasMore: {5: false}})
		replaceTaskEverywhere(client, {id: 3, title: 'new', position: 0})
		expect(client.getQueryData<Task>(taskKeys.detail(3))?.title).toBe('new')
		expect(client.getQueryData<InfiniteData<PaginatedTask>>(list)?.pages[0].items?.[0]).toMatchObject({title: 'new', position: 20})
		expect(client.getQueryData<BoardData>(board)).toMatchObject({buckets: [{tasks: [{title: 'new', bucket_id: 5}]}], pages: {5: 2}})
		removeTaskEverywhere(client, 3)
		expect(client.getQueryData(taskKeys.detail(3))).toBeUndefined()
		expect(client.getQueryData<InfiniteData<PaginatedTask>>(list)?.pages[0].items).toEqual([])
		expect(client.getQueryData<BoardData>(board)?.buckets[0]).toMatchObject({tasks: [], count: 0})
	})
	it('does not create queries for unmounted resources', () => {
		const client = new QueryClient()
		replaceTaskEverywhere(client, {id: 3, title: 'new'})
		expect(client.getQueryCache().getAll()).toHaveLength(0)
	})
})
