import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {taskKeys} from './tasks'
import {kanbanKeys, type BoardData} from './kanban'
import {updateTaskMutationOptions, deleteTaskMutationOptions, addTaskAssigneeMutationOptions, bulkCreateTasksMutationOptions, moveTaskMutationOptions} from './taskMutations'
import type {Task} from '@/client/generated'
const sdk = vi.hoisted(() => ({tasksUpdate: vi.fn(), tasksDelete: vi.fn(), taskAssigneesCreate: vi.fn(), tasksBulkCreate: vi.fn(), taskBucketUpdate: vi.fn()}))
vi.mock('@/client/generated', async importOriginal => ({...await importOriginal<object>(), ...sdk}))
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn(), translatedError: (message: string) => new Error(message)}))
vi.mock('@/client/requestContext', () => ({captureClientRequestContext: () => 1, assertClientRequestContext: vi.fn(), isClientRequestContextCurrent: () => true}))

describe('task mutations', () => {
	beforeEach(() => vi.resetAllMocks())
	it('sends only writable fields and reconciles loaded detail', async () => {
		const client = new QueryClient()
		client.setQueryData(taskKeys.detail(1), {id: 1, title: 'old'})
		sdk.tasksUpdate.mockResolvedValue({data: {id: 1, title: 'new'}})
		await client.getMutationCache().build(client, updateTaskMutationOptions()).execute({id: 1, title: 'new', labels: [{id: 2}], created: '2026-01-01'})
		expect(sdk.tasksUpdate).toHaveBeenCalledWith({path: {task: 1}, body: {title: 'new'}})
		expect(client.getQueryData<Task>(taskKeys.detail(1))?.title).toBe('new')
	})
	it('removes a deleted task from loaded detail', async () => {
		const client = new QueryClient()
		client.setQueryData(taskKeys.detail(1), {id: 1})
		sdk.tasksDelete.mockResolvedValue({})
		await client.getMutationCache().build(client, deleteTaskMutationOptions()).execute(1)
		expect(client.getQueryData(taskKeys.detail(1))).toBeUndefined()
	})
	it('patches embedded assignees without creating an assignee cache', async () => {
		const client = new QueryClient()
		client.setQueryData(taskKeys.detail(1), {id: 1, assignees: []})
		sdk.taskAssigneesCreate.mockResolvedValue({data: {user_id: 2}})
		await client.getMutationCache().build(client, addTaskAssigneeMutationOptions()).execute({taskId: 1, user: {id: 2, username: 'alice'}})
		expect(client.getQueryData<Task>(taskKeys.detail(1))?.assignees).toEqual([{id: 2, username: 'alice'}])
		expect(client.getQueryCache().getAll()).toHaveLength(1)
	})
	it('rolls back an optimistic bucket move when the request fails', async () => {
		const client = new QueryClient()
		const key = kanbanKeys.board(1, 2)
		const original: BoardData = {buckets: [{id: 3, count: 1, tasks: [{id: 1, bucket_id: 3}]}, {id: 4, count: 0, tasks: []}], pages: {3: 1, 4: 1}, hasMore: {3: false, 4: false}}
		client.setQueryData(key, original)
		sdk.taskBucketUpdate.mockImplementation(() => {
			expect(client.getQueryData<BoardData>(key)?.buckets[1].tasks).toMatchObject([{id: 1, bucket_id: 4}])
			throw new Error('denied')
		})
		await expect(client.getMutationCache().build(client, moveTaskMutationOptions()).execute({project: 1, view: 2, bucket: 4, task: {id: 1}})).rejects.toThrow('denied')
		expect(client.getQueryData(key)).toEqual(original)
	})
	it('retains successful bulk batches and input alignment after partial failure', async () => {
		const client = new QueryClient()
		sdk.tasksBulkCreate.mockImplementation(({body}) => {
			if (body.tasks[0].title === '0') throw new Error('failed batch')
			return {data: {tasks: body.tasks.map((task: Task) => ({...task, id: 7}))}}
		})
		const input = Array.from({length: 101}, (_, index) => ({title: String(index), project_id: 1}))
		const result = await client.getMutationCache().build(client, bulkCreateTasksMutationOptions()).execute(input)
		expect(result.tasks[100]).toMatchObject({id: 7, title: '100'})
		expect(result.tasks.slice(0, 100)).toEqual(Array(100).fill(null))
		expect(result.error).toBeInstanceOf(Error)
	})
})
