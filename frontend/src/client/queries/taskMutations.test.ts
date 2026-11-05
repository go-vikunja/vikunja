import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient, QueryObserver} from '@tanstack/vue-query'
import {taskKeys, tasksQuery} from './tasks'
import {
	bucketKeys,
	kanbanKeys,
	kanbanQuery,
	normalizeBucket,
	type BoardData,
} from './kanban'
import {
	addTaskAssigneeMutationOptions,
	bulkCreateTasksMutationOptions,
	createTaskMutationOptions,
	deleteTaskMutationOptions,
	duplicateTaskMutationOptions,
	favoriteTaskMutationOptions,
	moveTaskMutationOptions,
	splitWriteOutcome,
	taskWriteBody,
	updateTaskMutationOptions,
	updateTaskPositionMutationOptions,
	writeOutcomeOf,
	type UpdateTaskInput,
} from './taskMutations'
import {trashKeys} from './trash'
import {timeEntryKeys} from './timeEntries'
import type {Task, PaginatedTask} from '@/client/generated'

const sdk = vi.hoisted(() => ({
	bucketsList: vi.fn(),
	patchTasksRead: vi.fn(),
	projectTasksList: vi.fn(),
	projectViewBucketsTasksList: vi.fn(),
	taskAssigneesCreate: vi.fn(),
	taskBucketUpdate: vi.fn(),
	tasksBulkCreate: vi.fn(),
	tasksCreate: vi.fn(),
	tasksDelete: vi.fn(),
	tasksDuplicate: vi.fn(),
	tasksPositionUpdate: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
	translatedError: (message: string) => new Error(message),
}))
vi.mock('@/client/requestContext', () => ({
	captureClientRequestContext: () => 1,
	assertClientRequestContext: vi.fn(),
	isClientRequestContextCurrent: vi.fn(() => true),
	isRequestContextAbort: () => false,
}))

async function seedScrolledBoard() {
	const {projectKeys} = await import('./projects')
	const client = new QueryClient()
	sdk.projectViewBucketsTasksList.mockResolvedValue({
		data: {
			items: [{
				id: 3,
				count: 30,
				tasks: Array.from({length: 30}, (_, index) => ({id: index + 1, project_id: 1})),
			}],
		},
	})
	sdk.projectTasksList.mockResolvedValue({data: {items: [], total: 0}})
	client.setQueryData(projectKeys.detail(1), {
		id: 1,
		views: [{id: 2, bucket_configuration_mode: 'manual', default_bucket_id: 3}],
	})
	const board = kanbanQuery(1, 2)
	const list = tasksQuery({project: 1})
	const observers = [
		new QueryObserver(client, {queryKey: board.queryKey, queryFn: board.queryFn}),
		new QueryObserver(client, {queryKey: list.queryKey, queryFn: list.queryFn}),
	].map(observer => observer.subscribe(() => {}))
	await vi.waitFor(() => {
		expect(client.getQueryData(board.queryKey)).toBeDefined()
		expect(client.getQueryData(list.queryKey)).toBeDefined()
	})
	client.setQueryData<BoardData>(board.queryKey, current => current && {...current, pages: {3: 2}})
	return {client, board, list, unsubscribe: () => observers.forEach(stop => stop())}
}

async function seedDefaultBucketBoards() {
	const {projectKeys} = await import('./projects')
	const client = new QueryClient()
	client.setQueryData(projectKeys.detail(1), {
		id: 1,
		views: [
			{id: 2, bucket_configuration_mode: 'manual', default_bucket_id: 4},
			{id: 5, bucket_configuration_mode: 'manual'},
		],
	})
	client.setQueryData(kanbanKeys.board(1, 2), {
		buckets: [
			normalizeBucket({id: 3, count: 0, tasks: []}),
			normalizeBucket({id: 4, count: 0, tasks: []}),
		],
		pages: {3: 1, 4: 1},
	})
	client.setQueryData(kanbanKeys.board(1, 5), {
		buckets: [normalizeBucket({id: 6, count: 0, tasks: []})],
		pages: {6: 1},
	})
	return client
}

describe('task mutations', () => {
	beforeEach(() => vi.resetAllMocks())

	it('sends only writable fields and reconciles loaded detail', async () => {
		const client = new QueryClient()
		client.setQueryData(taskKeys.detail(1), {id: 1, title: 'old'})
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, title: 'new'}})
		await client.getMutationCache()
			.build(client, updateTaskMutationOptions())
			.execute({id: 1, title: 'new', labels: [{id: 2}], created: '2026-01-01'})
		expect(sdk.patchTasksRead).toHaveBeenCalledWith({
			path: {task: 1},
			body: [{op: 'add', path: '/title', value: 'new'}],
			throwOnError: false,
		})
		expect(client.getQueryData<Task>(taskKeys.detail(1))?.title).toBe('new')
	})

	it('clears a rule with an explicit null and never sends the read-only flag', async () => {
		const client = new QueryClient()
		client.setQueryData(taskKeys.detail(1), {id: 1, repeat: {freq: 'daily', interval: 1}})
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, repeats_from_current_date: false}})
		const update = updateTaskMutationOptions()
		await client.getMutationCache()
			.build(client, update)
			.execute({id: 1, repeat: null, repeats_from_current_date: false})
		expect(sdk.patchTasksRead).toHaveBeenCalledWith({
			path: {task: 1},
			body: [
				{op: 'add', path: '/repeat', value: null},
				{op: 'add', path: '/repeats_from_current_date', value: false},
			],
			throwOnError: false,
		})
		expect(client.getQueryData<Task>(taskKeys.detail(1))).not.toHaveProperty('repeat.freq')

		sdk.patchTasksRead.mockResolvedValue({data: {id: 1}})
		await client.getMutationCache()
			.build(client, update)
			.execute({id: 1, repeat: {freq: 'yearly', interval: 1, read_only: true}})
		expect(sdk.patchTasksRead).toHaveBeenLastCalledWith({
			path: {task: 1},
			body: [{op: 'add', path: '/repeat', value: {freq: 'yearly', interval: 1}}],
			throwOnError: false,
		})
	})

	it('leaves the stored rule alone when the input has no repeat', async () => {
		const client = new QueryClient()
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, title: 'new'}})
		await client.getMutationCache()
			.build(client, updateTaskMutationOptions())
			.execute({id: 1, title: 'new'})
		expect(sdk.patchTasksRead.mock.calls[0][0].body).not.toContainEqual(expect.objectContaining({path: '/repeat'}))
	})

	it('creates a task without a rule when the draft cleared it', async () => {
		const client = new QueryClient()
		sdk.tasksCreate.mockResolvedValue({data: {id: 9, project_id: 1}})
		await client.getMutationCache()
			.build(client, createTaskMutationOptions())
			.execute({project_id: 1, title: 'new', repeat: null})
		expect(sdk.tasksCreate.mock.calls[0][0].body).toEqual({title: 'new', project_id: 1})
	})

	it('addresses the delete endpoint by task id', async () => {
		const client = new QueryClient()
		sdk.tasksDelete.mockResolvedValue({})
		await client.getMutationCache()
			.build(client, deleteTaskMutationOptions())
			.execute(1)
		expect(sdk.tasksDelete).toHaveBeenCalledWith({path: {task: 1}})
	})

	it('patches embedded assignees without creating an assignee cache', async () => {
		const client = new QueryClient()
		client.setQueryData(taskKeys.detail(1), {id: 1, assignees: []})
		sdk.taskAssigneesCreate.mockResolvedValue({data: {user_id: 2}})
		await client.getMutationCache()
			.build(client, addTaskAssigneeMutationOptions())
			.execute({taskId: 1, user: {id: 2, username: 'alice'}})
		expect(client.getQueryData<Task>(taskKeys.detail(1))?.assignees).toEqual([{id: 2, username: 'alice'}])
		expect(client.getQueryCache().getAll()).toHaveLength(1)
	})

	it('rolls back an optimistic bucket move when the request fails', async () => {
		const client = new QueryClient()
		const key = kanbanKeys.board(1, 2)
		const original: BoardData = {
			buckets: [
				normalizeBucket({id: 3, count: 1, tasks: [{id: 1, bucket_id: 3}]}),
				normalizeBucket({id: 4, count: 0, tasks: []}),
			],
			pages: {3: 1, 4: 1},
		}
		client.setQueryData(key, original)
		sdk.taskBucketUpdate.mockImplementation(() => {
			expect(client.getQueryData<BoardData>(key)?.buckets[1].tasks).toMatchObject([{id: 1, bucket_id: 4}])
			throw new Error('denied')
		})
		await expect(client.getMutationCache()
			.build(client, moveTaskMutationOptions())
			.execute({project: 1, view: 2, bucket: 4, task: {id: 1}})).rejects.toThrow('denied')
		expect(client.getQueryData(key)).toEqual(original)
	})

	it('rolls back only the caches holding the task', async () => {
		const client = new QueryClient()
		const key = taskKeys.list({project: 1})
		const other = taskKeys.list({project: 2})
		client.setQueryData(key, {items: [{id: 1, title: 'old'}], total: 1})
		client.setQueryData(other, {items: [{id: 9, title: 'other'}], total: 1})
		sdk.patchTasksRead.mockImplementation(() => {
			client.setQueryData(other, {items: [{id: 9, title: 'meanwhile'}], total: 1})
			throw new Error('denied')
		})
		await expect(client.getMutationCache()
			.build(client, updateTaskMutationOptions(true))
			.execute({id: 1, title: 'new'})).rejects.toThrow('denied')
		expect(client.getQueryData<PaginatedTask>(other)?.items).toEqual([{id: 9, title: 'meanwhile'}])
		expect(client.getQueryData<PaginatedTask>(key)?.items).toEqual([{id: 1, title: 'old'}])
	})

	it('leaves fetches of caches without the task running', async () => {
		const client = new QueryClient()
		const other = taskKeys.list({project: 2})
		client.setQueryData(taskKeys.list({project: 1}), {items: [{id: 1, title: 'old'}], total: 1})
		let release: (page: PaginatedTask) => void = () => {}
		const pending = client.fetchQuery({
			queryKey: other,
			queryFn: () => new Promise<PaginatedTask>(resolve => { release = resolve }),
		})
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, title: 'new'}})
		await client.getMutationCache()
			.build(client, updateTaskMutationOptions(true))
			.execute({id: 1, title: 'new'})
		release({items: [{id: 9}], total: 1})
		await expect(pending).resolves.toMatchObject({items: [{id: 9}]})
		expect(client.getQueryData<PaginatedTask>(other)?.items).toEqual([{id: 9}])
	})

	it('retains successful bulk batches and input alignment after partial failure', async () => {
		const client = new QueryClient()
		sdk.tasksBulkCreate.mockImplementation(({body}) => {
			if (body.tasks[0].title === '0') throw new Error('failed batch')
			return {data: {tasks: body.tasks.map((task: Task) => ({...task, id: 7}))}}
		})
		const input = Array.from({length: 101}, (_, index) => ({title: String(index), project_id: 1}))
		const result = await client.getMutationCache()
			.build(client, bulkCreateTasksMutationOptions())
			.execute(input)
		expect(result.tasks[100]).toMatchObject({id: 7, title: '100'})
		expect(result.tasks.slice(0, 100)).toEqual(Array(100).fill(null))
		expect(result.error).toBeInstanceOf(Error)
	})

	it('stops a bulk write before the next batch when the identity changes', async () => {
		const {assertClientRequestContext, isClientRequestContextCurrent} = await import('@/client/requestContext')
		const client = new QueryClient()
		sdk.tasksBulkCreate.mockImplementation(({body}) => {
			vi.mocked(isClientRequestContextCurrent).mockReturnValue(false)
			vi.mocked(assertClientRequestContext).mockImplementation(() => {
				throw new DOMException('Changed identity', 'AbortError')
			})
			return {data: {tasks: body.tasks.map((task: Task) => ({...task, id: 1}))}}
		})
		await expect(client.getMutationCache()
			.build(client, bulkCreateTasksMutationOptions())
			.execute(Array.from({length: 101}, (_, index) => ({title: String(index), project_id: 1}))))
			.rejects.toThrow('Changed identity')
		expect(sdk.tasksBulkCreate).toHaveBeenCalledTimes(1)
	})

	it('keeps the cached per-view position when the moved task comes back without one', async () => {
		const client = new QueryClient()
		const key = kanbanKeys.board(1, 2)
		client.setQueryData(key, {
			buckets: [
				{id: 3, count: 1, tasks: [{id: 1, position: 42, bucket_id: 3}]},
				{id: 4, count: 0, tasks: []},
			],
			pages: {3: 1, 4: 1},
		})
		sdk.taskBucketUpdate.mockResolvedValue({data: {bucket_id: 4, task: {id: 1, position: 0}}})
		await client.getMutationCache()
			.build(client, moveTaskMutationOptions())
			.execute({project: 1, view: 2, bucket: 4, task: {id: 1}})
		expect(client.getQueryData<BoardData>(key)?.buckets[1].tasks)
			.toMatchObject([{id: 1, position: 42, bucket_id: 4}])
	})

	it('renames the moved view bucket on the cached task and leaves other views alone', async () => {
		const client = new QueryClient()
		const detail = taskKeys.detail(1, ['buckets'])
		client.setQueryData(detail, {
			id: 1,
			buckets: [
				{id: 3, project_view_id: 2, title: 'To-Do'},
				{id: 8, project_view_id: 5, title: 'Backlog'},
			],
		})
		client.setQueryData(kanbanKeys.board(1, 2), {
			buckets: [
				{id: 3, project_view_id: 2, title: 'To-Do', count: 1, tasks: [{id: 1}]},
				{id: 4, project_view_id: 2, title: 'Done', count: 0, tasks: []},
			],
			pages: {3: 1, 4: 1},
		})
		sdk.taskBucketUpdate.mockResolvedValue({data: {bucket_id: 4, task: {id: 1}}})
		await client.getMutationCache()
			.build(client, moveTaskMutationOptions())
			.execute({project: 1, view: 2, bucket: 4, task: {id: 1}})
		expect(client.getQueryData<Task>(detail)?.buckets).toMatchObject([
			{id: 4, project_view_id: 2, title: 'Done'},
			{id: 8, project_view_id: 5, title: 'Backlog'},
		])
	})

	it('renames the moved view bucket from the cached bucket list when no board is cached', async () => {
		const client = new QueryClient()
		const detail = taskKeys.detail(1, ['buckets'])
		const queryFn = vi.fn(() => ({id: 1, buckets: [{id: 3, project_view_id: 2, title: 'To-Do'}]}))
		const unsubscribe = new QueryObserver(client, {queryKey: detail, queryFn}).subscribe(() => {})
		await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
		client.setQueryData(bucketKeys.list(1, 2), [
			{id: 3, project_view_id: 2, title: 'To-Do'},
			{id: 4, project_view_id: 2, title: 'Done'},
		])
		sdk.taskBucketUpdate.mockResolvedValue({data: {bucket_id: 4, task: {id: 1}}})
		await client.getMutationCache()
			.build(client, moveTaskMutationOptions())
			.execute({project: 1, view: 2, bucket: 4, task: {id: 1}})
		expect(client.getQueryData<Task>(detail)?.buckets).toMatchObject([
			{id: 4, project_view_id: 2, title: 'Done'},
		])
		expect(queryFn).toHaveBeenCalledTimes(1)
		unsubscribe()
	})

	it('refetches the watched task detail when neither the board nor the bucket list is cached', async () => {
		const client = new QueryClient()
		const queryFn = vi.fn(() => ({id: 1, buckets: [{id: 3, project_view_id: 2, title: 'To-Do'}]}))
		const observer = new QueryObserver(client, {queryKey: taskKeys.detail(1, ['buckets']), queryFn})
		const unsubscribe = observer.subscribe(() => {})
		await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
		sdk.taskBucketUpdate.mockResolvedValue({data: {bucket_id: 4, task: {id: 1}}})
		await client.getMutationCache()
			.build(client, moveTaskMutationOptions())
			.execute({project: 1, view: 2, bucket: 4, task: {id: 1}})
		await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(2))
		unsubscribe()
	})

	it('reorders only cached lists for the affected view', async () => {
		const client = new QueryClient()
		const key = taskKeys.list({project: 1, view: 2, params: {sort_by: ['position']}})
		const other = taskKeys.list({project: 1, view: 3, params: {sort_by: ['position']}})
		const data = {items: [{id: 1, position: 1}, {id: 2, position: 2}], total: 2}
		client.setQueryData(key, data)
		client.setQueryData(other, data)
		sdk.tasksPositionUpdate.mockResolvedValue({data: {position: 3}})
		await client.getMutationCache()
			.build(client, updateTaskPositionMutationOptions())
			.execute({taskId: 1, project_view_id: 2, position: 3})
		expect(client.getQueryData<PaginatedTask>(key)?.items?.map(task => task.id)).toEqual([2, 1])
		expect(client.getQueryData(other)).toEqual(data)
	})

	it('reorders descending lists in their own direction and leaves unsorted ones alone', async () => {
		const client = new QueryClient()
		const descending = taskKeys.list({
			project: 1,
			view: 2,
			params: {sort_by: ['position'], order_by: ['desc']},
		})
		const search = taskKeys.list({project: 1, view: 2, params: {q: 'term'}})
		const data = {items: [{id: 2, position: 2}, {id: 1, position: 1}], total: 2}
		client.setQueryData(descending, data)
		client.setQueryData(search, data)
		sdk.tasksPositionUpdate.mockResolvedValue({data: {position: 3}})
		await client.getMutationCache()
			.build(client, updateTaskPositionMutationOptions())
			.execute({taskId: 1, project_view_id: 2, position: 3})
		expect(client.getQueryData<PaginatedTask>(descending)?.items?.map(task => task.id)).toEqual([1, 2])
		expect(client.getQueryData<PaginatedTask>(search)?.items?.map(task => task.id)).toEqual([2, 1])
	})

	it('serializes clearing a date and relative reminders as valid API dates', () => {
		expect(taskWriteBody({
			due_date: '',
			reminders: [{relative_to: 'due_date', relative_period: -60, reminder: ''}],
		})).toEqual({
			due_date: '0001-01-01T00:00:00Z',
			reminders: [{
				relative_to: 'due_date',
				relative_period: -60,
				reminder: '0001-01-01T00:00:00Z',
			}],
		})
	})

	it('leaves boards without the completed task untouched and unfetched', async () => {
		const {projectKeys} = await import('./projects')
		const client = new QueryClient()
		const otherKey = kanbanKeys.board(1, 5)
		const otherBoard: BoardData = {
			buckets: [
				normalizeBucket({id: 6, count: 0, tasks: []}),
				normalizeBucket({id: 7, count: 0, tasks: []}),
			],
			pages: {6: 1, 7: 1},
		}
		client.setQueryData(projectKeys.detail(1), {
			id: 1,
			views: [{id: 2, done_bucket_id: 4}, {id: 5, done_bucket_id: 7}],
		})
		client.setQueryData(kanbanKeys.board(1, 2), {
			buckets: [
				{id: 3, count: 1, tasks: [{id: 1, done: false}]},
				{id: 4, count: 0, tasks: []},
			],
			pages: {3: 1, 4: 1},
		})
		client.setQueryData(otherKey, otherBoard)
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, project_id: 1, done: true}})
		await client.getMutationCache()
			.build(client, updateTaskMutationOptions())
			.execute({id: 1, done: true})
		// structural sharing hides no-op rewrites, so count writes instead: 1 = seed only
		expect(client.getQueryState(otherKey)?.dataUpdateCount).toBe(1)
		expect(sdk.projectViewBucketsTasksList).not.toHaveBeenCalled()
	})

	it('inserts a created card into a scrolled board without refetching it, while lists still refetch', async () => {
		const {client, board, list, unsubscribe} = await seedScrolledBoard()
		sdk.tasksCreate.mockResolvedValue({data: {id: 99, title: 'new', project_id: 1, bucket_id: 3}})
		await client.getMutationCache()
			.build(client, createTaskMutationOptions())
			.execute({title: 'new', project_id: 1, bucket_id: 3})
		const data = client.getQueryData<BoardData>(board.queryKey)!
		expect(data.buckets[0].tasks[0]).toMatchObject({id: 99, bucket_id: 3})
		expect(data.buckets[0].tasks).toHaveLength(31)
		expect(data.buckets[0].count).toBe(31)
		expect(data.pages).toEqual({3: 2})
		expect(sdk.projectViewBucketsTasksList).toHaveBeenCalledOnce()
		expect(sdk.projectTasksList).toHaveBeenCalledTimes(2)
		expect(client.getQueryState(board.queryKey)?.isInvalidated).toBe(true)
		expect(client.getQueryState(list.queryKey)?.isInvalidated).toBe(false)
		unsubscribe()
	})

	it('drops a deleted card from a scrolled board without refetching it', async () => {
		const {client, board, unsubscribe} = await seedScrolledBoard()
		sdk.tasksDelete.mockResolvedValue({})
		await client.getMutationCache()
			.build(client, deleteTaskMutationOptions())
			.execute(1)
		const data = client.getQueryData<BoardData>(board.queryKey)!
		expect(data.buckets[0].tasks.map(task => task.id)).not.toContain(1)
		expect(data.buckets[0].count).toBe(29)
		expect(data.pages).toEqual({3: 2})
		expect(sdk.projectViewBucketsTasksList).toHaveBeenCalledOnce()
		expect(sdk.projectTasksList).toHaveBeenCalledTimes(2)
		unsubscribe()
	})

	it('marks the trash and time entries stale after deleting a task', async () => {
		const client = new QueryClient()
		client.setQueryData(trashKeys.taskList(1), {
			items: [],
			page: 1,
			per_page: 50,
			total: 0,
			total_pages: 0,
		})
		client.setQueryData(timeEntryKeys.list('', 'UTC', 1), {
			items: [],
			page: 1,
			per_page: 50,
			total: 0,
			total_pages: 0,
		})
		sdk.tasksDelete.mockResolvedValue({})
		await client.getMutationCache()
			.build(client, deleteTaskMutationOptions())
			.execute(5)
		expect(sdk.tasksDelete).toHaveBeenCalledWith({path: {task: 5}})
		expect(client.getQueryState(trashKeys.taskList(1))?.isInvalidated).toBe(true)
		expect(client.getQueryState(timeEntryKeys.list('', 'UTC', 1))?.isInvalidated).toBe(true)
	})

	it('inserts a duplicate into the default bucket of every cached view', async () => {
		const client = await seedDefaultBucketBoards()
		sdk.tasksDuplicate.mockResolvedValue({data: {duplicated_task: {id: 9, project_id: 1, bucket_id: 0}}})
		await client.getMutationCache()
			.build(client, duplicateTaskMutationOptions())
			.execute(1)
		expect(client.getQueryData<BoardData>(kanbanKeys.board(1, 2))?.buckets
			.map(bucket => bucket.tasks.map(task => task.id))).toEqual([[], [9]])
		expect(client.getQueryData<BoardData>(kanbanKeys.board(1, 5))?.buckets[0].tasks
			.map(task => task.id)).toEqual([9])
	})

	it('inserts a done card into the done bucket even when another bucket was requested', async () => {
		const {projectKeys} = await import('./projects')
		const client = await seedDefaultBucketBoards()
		client.setQueryData(projectKeys.detail(1), {
			id: 1,
			views: [{
				id: 2,
				bucket_configuration_mode: 'manual',
				default_bucket_id: 4,
				done_bucket_id: 3,
			}],
		})
		sdk.tasksCreate.mockResolvedValue({
			data: {
				id: 9,
				project_id: 1,
				bucket_id: 4,
				done: true,
			},
		})
		await client.getMutationCache()
			.build(client, createTaskMutationOptions())
			.execute({
				title: 'new',
				project_id: 1,
				bucket_id: 4,
				done: true,
			})
		expect(client.getQueryData<BoardData>(kanbanKeys.board(1, 2))?.buckets
			.map(bucket => bucket.tasks.map(task => task.id))).toEqual([
			[9],
			[],
		])
	})

	it('caches a created card with its stored position in each view', async () => {
		const client = await seedDefaultBucketBoards()
		sdk.tasksCreate.mockResolvedValue({data: {id: 9, project_id: 1, position: 0, positions: [
			{task_id: 9, project_view_id: 2, position: 250},
			{task_id: 9, project_view_id: 5, position: 100},
		]}})
		await client.getMutationCache()
			.build(client, createTaskMutationOptions())
			.execute({title: 'new', project_id: 1})
		expect(client.getQueryData<BoardData>(kanbanKeys.board(1, 2))?.buckets[1].tasks[0].position).toBe(250)
		expect(client.getQueryData<BoardData>(kanbanKeys.board(1, 5))?.buckets[0].tasks[0].position).toBe(100)
	})

	it('keeps the input order of a bulk batch by inserting it back to front', async () => {
		const client = await seedDefaultBucketBoards()
		sdk.tasksBulkCreate.mockImplementation(({body}) => ({
			data: {
				tasks: body.tasks.map((task: Task, index: number) => ({
					...task,
					id: 20 + index,
					project_id: 1,
				})),
			},
		}))
		await client.getMutationCache()
			.build(client, bulkCreateTasksMutationOptions())
			.execute([
				{title: 'a', project_id: 1},
				{title: 'b', project_id: 1},
			])
		expect(client.getQueryData<BoardData>(kanbanKeys.board(1, 2))?.buckets[1]).toMatchObject({
			count: 2,
			tasks: [{id: 20}, {id: 21}],
		})
	})

	it('leaves a filter-configured board to its stale mark', async () => {
		const {projectKeys} = await import('./projects')
		const client = new QueryClient()
		const filtered = kanbanKeys.board(1, 2)
		const manual = kanbanKeys.board(1, 5)
		const board = {buckets: [normalizeBucket({id: 3, count: 0, tasks: []})], pages: {3: 1}}
		client.setQueryData(projectKeys.detail(1), {
			id: 1,
			views: [
				{id: 2, bucket_configuration_mode: 'filter'},
				{id: 5, bucket_configuration_mode: 'manual'},
			],
		})
		client.setQueryData(filtered, board)
		client.setQueryData(manual, {
			buckets: [normalizeBucket({id: 6, count: 0, tasks: []})],
			pages: {6: 1},
		})
		sdk.tasksCreate.mockResolvedValue({data: {id: 99, project_id: 1}})
		await client.getMutationCache()
			.build(client, createTaskMutationOptions())
			.execute({title: 'new', project_id: 1})
		expect(client.getQueryData<BoardData>(filtered)).toEqual(board)
		expect(client.getQueryState(filtered)?.isInvalidated).toBe(true)
		expect(client.getQueryData<BoardData>(manual)?.buckets[0].tasks.map(task => task.id)).toEqual([99])
	})

	it('refetches the open boards a created card cannot be patched into', async () => {
		const {projectKeys} = await import('./projects')
		const client = new QueryClient()
		client.setQueryData(projectKeys.detail(1), {
			id: 1,
			views: [
				{
					id: 2,
					bucket_configuration_mode: 'filter',
				},
				{
					id: 5,
					bucket_configuration_mode: 'manual',
					default_bucket_id: 6,
				},
			],
		})
		const emptyBoard = () => Promise.resolve({
			buckets: [normalizeBucket({
				id: 6,
				count: 0,
				tasks: [],
			})],
			pages: {6: 1},
		})
		const readFilterView = vi.fn(emptyBoard)
		const readSearched = vi.fn(emptyBoard)
		const readPlain = vi.fn(emptyBoard)
		const searched = kanbanKeys.board(1, 5, {q: 'milk'})
		const plain = kanbanKeys.board(1, 5)
		const observers = [
			new QueryObserver(client, {
				queryKey: kanbanKeys.board(1, 2),
				queryFn: readFilterView,
			}),
			new QueryObserver(client, {
				queryKey: searched,
				queryFn: readSearched,
			}),
			new QueryObserver(client, {
				queryKey: plain,
				queryFn: readPlain,
			}),
		].map(observer => observer.subscribe(() => {}))
		await vi.waitFor(() => {
			expect(client.getQueryData(kanbanKeys.board(1, 2))).toBeDefined()
			expect(client.getQueryData(searched)).toBeDefined()
			expect(client.getQueryData(plain)).toBeDefined()
		})
		sdk.tasksCreate.mockResolvedValue({data: {id: 99, project_id: 1}})
		sdk.projectTasksList.mockResolvedValue({data: {items: [], total: 0}})

		await client.getMutationCache()
			.build(client, createTaskMutationOptions())
			.execute({title: 'new', project_id: 1})
		observers.forEach(stop => stop())

		expect(readFilterView).toHaveBeenCalledTimes(2)
		expect(readSearched).toHaveBeenCalledTimes(2)
		expect(client.getQueryData<BoardData>(searched)?.buckets[0].tasks).toEqual([])
		expect(readPlain).toHaveBeenCalledTimes(1)
		expect(client.getQueryData<BoardData>(plain)?.buckets[0].tasks.map(task => task.id)).toEqual([99])
	})

	it('renames the cached task bucket when toggling done moves the card', async () => {
		const {projectKeys} = await import('./projects')
		const client = new QueryClient()
		const detail = taskKeys.detail(1, ['buckets'])
		client.setQueryData(projectKeys.detail(1), {
			id: 1,
			views: [{id: 2, done_bucket_id: 4, default_bucket_id: 3}],
		})
		client.setQueryData(kanbanKeys.board(1, 2), {
			buckets: [
				{id: 3, project_view_id: 2, title: 'To-Do', count: 1, tasks: [{id: 1, done: false}]},
				{id: 4, project_view_id: 2, title: 'Done', count: 0, tasks: []},
			],
			pages: {3: 1, 4: 1},
		})
		client.setQueryData(detail, {
			id: 1,
			done: false,
			buckets: [{id: 3, project_view_id: 2, title: 'To-Do'}],
		})
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, project_id: 1, done: true}})
		await client.getMutationCache()
			.build(client, updateTaskMutationOptions())
			.execute({id: 1, done: true})
		expect(client.getQueryData<Task>(detail)?.buckets).toMatchObject([
			{id: 4, project_view_id: 2, title: 'Done'},
		])
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, project_id: 1, done: false}})
		await client.getMutationCache()
			.build(client, updateTaskMutationOptions())
			.execute({id: 1, done: false})
		expect(client.getQueryData<Task>(detail)?.buckets).toMatchObject([
			{id: 3, project_view_id: 2, title: 'To-Do'},
		])
	})

	it('renames the done bucket on a detail opened without a board from the cached bucket list', async () => {
		const {projectKeys} = await import('./projects')
		const client = new QueryClient()
		const detail = taskKeys.detail(1, ['buckets'])
		const queryFn = vi.fn(() => ({
			id: 1,
			project_id: 1,
			done: false,
			buckets: [{id: 3, project_view_id: 2, title: 'To-Do'}],
		}))
		const unsubscribe = new QueryObserver(client, {queryKey: detail, queryFn}).subscribe(() => {})
		await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
		client.setQueryData(projectKeys.detail(1), {
			id: 1,
			views: [{id: 2, default_bucket_id: 3, done_bucket_id: 4}],
		})
		client.setQueryData(bucketKeys.list(1, 2), [
			{id: 3, project_view_id: 2, title: 'To-Do'},
			{id: 4, project_view_id: 2, title: 'Done'},
		])
		sdk.patchTasksRead.mockResolvedValue({data: {id: 1, project_id: 1, done: true}})
		await client.getMutationCache()
			.build(client, updateTaskMutationOptions())
			.execute({id: 1, done: true})
		expect(client.getQueryData<Task>(detail)?.buckets).toMatchObject([
			{id: 4, project_view_id: 2, title: 'Done'},
		])
		expect(queryFn).toHaveBeenCalledTimes(1)
		expect(sdk.projectViewBucketsTasksList).not.toHaveBeenCalled()
		expect(sdk.bucketsList).not.toHaveBeenCalled()
		unsubscribe()
	})

	describe('initial due time', () => {
		const rule = {freq: 'daily', interval: 1}
		const update = (input: UpdateTaskInput) => {
			const client = new QueryClient()
			return client.getMutationCache().build(client, updateTaskMutationOptions()).execute(input)
		}
		const sentBody = (call = 0) => sdk.patchTasksRead.mock.calls[call][0].body

		beforeEach(() => sdk.patchTasksRead.mockResolvedValue({data: {id: 1}}))

		it.each([
			['no due date', {}],
			['a cleared due date', {due_date: ''}],
			['the zero date', {due_date: '0001-01-01T00:00:00Z'}],
		])('asks the server for a due date on an undated repeating task with %s', async (_name, dueDate) => {
			await update({id: 1, title: 'water plants', done: false, repeat: rule, ...dueDate, initialDueTime: '09:00'})
			expect(sentBody()).toEqual([
				{op: 'add', path: '/title', value: 'water plants'},
				{op: 'add', path: '/done', value: false},
				{op: 'add', path: '/repeat', value: rule},
				{op: 'add', path: '/initial_due_time', value: '09:00'},
			])
		})

		it.each([
			['has a due date', {due_date: '2026-11-02T09:00:00Z', repeat: rule}],
			['is done', {due_date: '', done: true, repeat: rule}],
			['has no rule', {due_date: ''}],
			['clears its rule', {due_date: '', repeat: null}],
		])('sends the plain patch when the task %s', async (_name, task) => {
			await update({id: 1, ...task, initialDueTime: '09:00'})
			await update({id: 1, ...task})
			expect(sentBody(0)).toEqual(sentBody(1))
			expect(sentBody(0)).not.toContainEqual(expect.objectContaining({path: '/initial_due_time'}))
		})

		it('keeps every other write free of an initial due time', async () => {
			const client = new QueryClient()
			const draft = {title: 'water plants', project_id: 1, due_date: '', repeat: rule}
			sdk.tasksCreate.mockResolvedValue({data: {id: 9, project_id: 1}})
			sdk.tasksBulkCreate.mockImplementation(({body}) => ({
				data: {tasks: body.tasks.map((task: Task) => ({...task, id: 9}))},
			}))
			sdk.patchTasksRead.mockResolvedValue({data: {id: 1}})
			await client.getMutationCache().build(client, createTaskMutationOptions()).execute(draft)
			await client.getMutationCache().build(client, bulkCreateTasksMutationOptions()).execute([draft])
			await client.getMutationCache().build(client, favoriteTaskMutationOptions()).execute({id: 1, ...draft})
			await client.getMutationCache().build(client, updateTaskMutationOptions()).execute({id: 1, ...draft})
			const created = {title: 'water plants', project_id: 1, due_date: '0001-01-01T00:00:00Z', repeat: rule}
			expect(sdk.tasksCreate.mock.calls[0][0].body).toEqual(created)
			expect(sdk.tasksBulkCreate.mock.calls[0][0].body).toEqual({tasks: [created]})
			expect(sentBody(0)).toEqual([{op: 'add', path: '/is_favorite', value: true}])
			expect(sentBody(1)).toEqual([
				{op: 'add', path: '/title', value: 'water plants'},
				{op: 'add', path: '/due_date', value: '0001-01-01T00:00:00Z'},
				{op: 'add', path: '/repeat', value: rule},
				{op: 'add', path: '/project_id', value: 1},
			])
		})

		it('keeps the request option out of the optimistic cache', async () => {
			const client = new QueryClient()
			client.setQueryData(taskKeys.detail(1), {id: 1, title: 'old'})
			sdk.patchTasksRead.mockImplementation(() => {
				expect(client.getQueryData(taskKeys.detail(1))).not.toHaveProperty('initialDueTime')
				return {data: {id: 1}}
			})
			await client.getMutationCache()
				.build(client, updateTaskMutationOptions(true))
				.execute({id: 1, title: 'new', repeat: rule, initialDueTime: '09:00'})
			expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		})
	})

	describe('write outcome', () => {
		it('splits write-only fields off the returned task', () => {
			expect(splitWriteOutcome({id: 1, due_date: '2026-11-02T09:00:00Z', due_date_source: 'generated'})).toEqual({
				task: {id: 1, due_date: '2026-11-02T09:00:00Z'},
				outcome: {dueDateSource: 'generated'},
			})
			expect(splitWriteOutcome({id: 1, reason: 'until_passed'})).toEqual({
				task: {id: 1},
				outcome: {reason: 'until_passed'},
			})
		})

		it('caches the task without the outcome and hands the outcome to the caller', async () => {
			const client = new QueryClient()
			client.setQueryData(taskKeys.detail(1), {id: 1, title: 'water plants'})
			sdk.patchTasksRead.mockResolvedValue({data: {
				id: 1,
				due_date: '2026-11-02T09:00:00Z',
				due_date_source: 'generated',
				reason: 'until_passed',
			}})
			const saved = await client.getMutationCache()
				.build(client, updateTaskMutationOptions())
				.execute({id: 1, repeat: {freq: 'daily', interval: 1}, initialDueTime: '09:00'})
			const cached = client.getQueryData<Task>(taskKeys.detail(1))
			expect(cached).toMatchObject({title: 'water plants', due_date: '2026-11-02T09:00:00Z'})
			expect(cached).not.toHaveProperty('due_date_source')
			expect(cached).not.toHaveProperty('reason')
			expect(saved).not.toHaveProperty('due_date_source')
			expect(saved).not.toHaveProperty('reason')
			expect(writeOutcomeOf(saved)).toEqual({dueDateSource: 'generated', reason: 'until_passed'})
		})

		it('treats an empty 304 as unchanged and leaves every cache alone', async () => {
			const {error} = await import('@/message')
			const client = new QueryClient()
			client.setQueryData(taskKeys.detail(1), {id: 1, title: 'same'})
			sdk.patchTasksRead.mockResolvedValue({data: undefined, error: '', response: {status: 304}})
			const saved = await client.getMutationCache()
				.build(client, updateTaskMutationOptions())
				.execute({id: 1, title: 'same', initialDueTime: '09:00'})
			expect(writeOutcomeOf(saved)).toEqual({unchanged: true})
			expect(saved).toEqual({id: 1, title: 'same'})
			expect(client.getQueryState(taskKeys.detail(1))?.dataUpdateCount).toBe(1)
			expect(error).not.toHaveBeenCalled()
		})

		it('still fails on any other error status', async () => {
			const client = new QueryClient()
			sdk.patchTasksRead.mockResolvedValue({data: undefined, error: {message: 'denied'}, response: {status: 403}})
			await expect(client.getMutationCache()
				.build(client, updateTaskMutationOptions())
				.execute({id: 1, title: 'new'})).rejects.toEqual({message: 'denied'})
		})

		it('rolls an optimistic edit back when the server reports nothing changed', async () => {
			const client = new QueryClient()
			const list = taskKeys.list({project: 1})
			client.setQueryData(taskKeys.detail(1), {id: 1, title: 'same'})
			client.setQueryData(list, {items: [{id: 1, title: 'same'}], total: 1})
			sdk.patchTasksRead.mockImplementation(() => {
				expect(client.getQueryData<Task>(taskKeys.detail(1))?.title).toBe('same  ')
				return {data: undefined, error: '', response: {status: 304}}
			})
			await client.getMutationCache()
				.build(client, updateTaskMutationOptions(true))
				.execute({id: 1, title: 'same  '})
			expect(sdk.patchTasksRead.mock.calls[0][0].body).toEqual([{op: 'add', path: '/title', value: 'same'}])
			expect(client.getQueryData<Task>(taskKeys.detail(1))).toEqual({id: 1, title: 'same'})
			expect(client.getQueryData<PaginatedTask>(list)?.items).toEqual([{id: 1, title: 'same'}])
		})

		it('keeps write-only fields out of created, bulk-created and favorited tasks', async () => {
			const client = await seedDefaultBucketBoards()
			const written = {due_date: '2026-11-02T09:00:00Z', due_date_source: 'explicit', reason: 'until_passed'}
			client.setQueryData(taskKeys.detail(5), {id: 5, project_id: 1, is_favorite: false})
			sdk.tasksCreate.mockResolvedValue({data: {id: 9, project_id: 1, ...written}})
			sdk.tasksBulkCreate.mockResolvedValue({data: {tasks: [{id: 10, project_id: 1, ...written}]}})
			sdk.patchTasksRead.mockResolvedValue({data: {id: 5, project_id: 1, is_favorite: true, ...written}})
			const created = await client.getMutationCache()
				.build(client, createTaskMutationOptions())
				.execute({title: 'new', project_id: 1})
			const bulk = await client.getMutationCache()
				.build(client, bulkCreateTasksMutationOptions())
				.execute([{title: 'bulk', project_id: 1}])
			const favorite = await client.getMutationCache()
				.build(client, favoriteTaskMutationOptions())
				.execute({id: 5, project_id: 1, is_favorite: false})
			const cachedCards = client.getQueryData<BoardData>(kanbanKeys.board(1, 2))!.buckets[1].tasks
			expect(cachedCards.map(task => task.id)).toEqual([10, 9])
			for (const task of [created, bulk.tasks[0], favorite, ...cachedCards, client.getQueryData(taskKeys.detail(5))]) {
				expect(task).toMatchObject({due_date: '2026-11-02T09:00:00Z'})
				expect(task).not.toHaveProperty('due_date_source')
				expect(task).not.toHaveProperty('reason')
			}
		})

		it('keeps write-only fields out of duplicated and moved tasks', async () => {
			const client = await seedDefaultBucketBoards()
			const written = {due_date: '2026-11-02T09:00:00Z', due_date_source: 'explicit', reason: 'until_passed'}
			client.setQueryData(kanbanKeys.board(1, 2), {
				buckets: [
					normalizeBucket({id: 3, project_view_id: 2, count: 1, tasks: [{id: 1, project_id: 1}]}),
					normalizeBucket({id: 4, project_view_id: 2, count: 0, tasks: []}),
				],
				pages: {3: 1, 4: 1},
			})
			client.setQueryData(taskKeys.detail(1), {id: 1, project_id: 1})
			sdk.tasksDuplicate.mockResolvedValue({data: {duplicated_task: {id: 9, project_id: 1, bucket_id: 0, ...written}}})
			sdk.taskBucketUpdate.mockResolvedValue({data: {bucket_id: 4, task: {id: 1, project_id: 1, ...written}}})
			const duplicate = await client.getMutationCache()
				.build(client, duplicateTaskMutationOptions())
				.execute(1)
			const moved = await client.getMutationCache()
				.build(client, moveTaskMutationOptions())
				.execute({project: 1, view: 2, bucket: 4, task: {id: 1, project_id: 1}})
			const cards = [kanbanKeys.board(1, 2), kanbanKeys.board(1, 5)]
				.flatMap(key => client.getQueryData<BoardData>(key)!.buckets.flatMap(bucket => bucket.tasks))
			expect(cards.map(task => task.id).sort()).toEqual([1, 9, 9])
			for (const task of [duplicate, moved.task, ...cards, client.getQueryData(taskKeys.detail(1))]) {
				expect(task).toMatchObject({due_date: '2026-11-02T09:00:00Z'})
				expect(task).not.toHaveProperty('due_date_source')
				expect(task).not.toHaveProperty('reason')
			}
		})

		it('reports no outcome for a task it did not write', () => {
			expect(writeOutcomeOf({id: 1})).toEqual({})
		})
	})
})
