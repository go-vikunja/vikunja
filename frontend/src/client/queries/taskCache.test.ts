import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient, QueryObserver} from '@tanstack/vue-query'
import {
	normalizeTask,
	taskKeys,
	taskQuery,
	tasksQuery,
	type PaginatedTaskResponse,
	type TaskResponse,
} from './tasks'
import {kanbanKeys, kanbanQuery, normalizeBucket, type BoardData} from './kanban'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
	replaceTaskEverywhere,
	removeTaskEverywhere,
	removeTaskFromBoard,
} from './taskCache'

const sdk = vi.hoisted(() => ({
	tasksRead: vi.fn(),
	tasksList: vi.fn(),
	projectTasksList: vi.fn(),
	projectViewTasksList: vi.fn(),
	projectViewBucketsTasksList: vi.fn(),
	bucketsList: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)

describe('task cache reconciliation', () => {
	beforeEach(() => vi.resetAllMocks())
	it('decrements the list total and the bucket count of a deleted task', () => {
		const client = new QueryClient()
		const task = normalizeTask({id: 3, title: 'task', project_id: 1, bucket_id: 5})
		const list = taskKeys.list({project: 1})
		const board = kanbanKeys.board(1, 8)
		client.setQueryData(list, {items: [task], total: 4, page: 1, per_page: 25, total_pages: 1})
		client.setQueryData(board, {buckets: [normalizeBucket({id: 5, count: 7, tasks: [task]})], pages: {5: 2}})
		removeTaskEverywhere(client, 3)
		expect(client.getQueryData<PaginatedTaskResponse>(list)?.total).toBe(3)
		expect(client.getQueryData<BoardData>(board)?.buckets[0].count).toBe(6)
	})
	it('leaves cache entries without the task untouched', async () => {
		const client = new QueryClient()
		const other = normalizeTask({id: 9, title: 'other', related_tasks: {subtask: [{id: 10, title: 'nested'}]}})
		const all = taskKeys.allList({project: 2})
		const list = taskKeys.list({project: 2})
		const board = kanbanKeys.board(2, 4)
		const detail = taskKeys.detail(9)
		const keys = [all, list, board, detail]
		client.setQueryData(all, [other])
		client.setQueryData(list, {items: [other], total: 1})
		client.setQueryData(board, {buckets: [normalizeBucket({id: 3, count: 1, tasks: [other]})], pages: {3: 1}})
		client.setQueryData(detail, other)
		await client.invalidateQueries({queryKey: list, refetchType: 'none'})
		const before = keys.map(key => client.getQueryState(key)?.dataUpdateCount)
		mapTaskEverywhere(client, 2, task => ({...task, done: true}))
		expect(keys.map(key => client.getQueryState(key)?.dataUpdateCount)).toEqual(before)
		expect(client.getQueryState(list)?.isInvalidated).toBe(true)
		// task 10 exists only as a relation of task 9, which still makes every entry hold it
		mapTaskEverywhere(client, 10, task => ({...task, done: true}))
		expect(keys.map(key => client.getQueryState(key)?.dataUpdateCount))
			.toEqual(before.map(count => (count ?? 0) + 1))
	})
	it('rewrites the totals of every cached page of a scope, not just the page holding the task', () => {
		const client = new QueryClient()
		const page = (items: TaskResponse[], number: number): PaginatedTaskResponse => ({
			items,
			total: 26,
			page: number,
			per_page: 25,
			total_pages: 2,
		})
		const first = taskKeys.list({project: 1}, 1)
		const second = taskKeys.list({project: 1}, 2)
		const otherScope = taskKeys.list({project: 2}, 1)
		client.setQueryData(first, page([normalizeTask({id: 3, project_id: 1, title: 'gone'})], 1))
		client.setQueryData(second, page([normalizeTask({id: 4, project_id: 1, title: 'stays'})], 2))
		client.setQueryData(otherScope, page([normalizeTask({id: 5, project_id: 2, title: 'elsewhere'})], 1))
		const untouched = client.getQueryState(otherScope)?.dataUpdateCount
		removeTaskEverywhere(client, 3)
		expect(client.getQueryData<PaginatedTaskResponse>(second))
			.toMatchObject({total: 25, total_pages: 1})
		expect(client.getQueryData<PaginatedTaskResponse>(first))
			.toMatchObject({items: [], total: 25, total_pages: 1})
		expect(client.getQueryState(otherScope)?.dataUpdateCount).toBe(untouched)
	})
	it('keeps a moved task in the relations of the tasks staying behind', () => {
		const client = new QueryClient()
		const child = normalizeTask({id: 2, project_id: 1, title: 'child'})
		const parent = normalizeTask({id: 1, project_id: 1, title: 'parent', related_tasks: {subtask: [child]}})
		const list = taskKeys.list({project: 1})
		const board = kanbanKeys.board(1, 4)
		client.setQueryData(list, {items: [parent, child], total: 2})
		client.setQueryData(board, {
			buckets: [normalizeBucket({id: 3, count: 2, tasks: [parent, child]})],
			pages: {3: 1},
		})
		replaceTaskEverywhere(client, {id: 2, project_id: 2})
		const moved = normalizeTask({id: 2, project_id: 2, title: 'child'})
		const kept = {...parent, related_tasks: {subtask: [moved]}}
		expect(client.getQueryData<PaginatedTaskResponse>(list)).toEqual({items: [kept], total: 1})
		expect(client.getQueryData<BoardData>(board)).toEqual({
			buckets: [normalizeBucket({id: 3, count: 1, tasks: [kept]})],
			pages: {3: 1},
		})
	})
	it('updates a parent and its subtask one after the other without recursing forever', () => {
		const client = new QueryClient()
		const list = taskKeys.allList({project: 1})
		client.setQueryData(list, [
			normalizeTask({id: 1, project_id: 1, related_tasks: {subtask: [{id: 2, project_id: 1}]}}),
			normalizeTask({id: 2, project_id: 1, related_tasks: {parenttask: [{id: 1, project_id: 1}]}}),
		])
		const cached = (id: number) => client.getQueryData<TaskResponse[]>(list)!.find(task => task.id === id)!
		replaceTaskEverywhere(client, {...cached(2), title: 'moved child'})
		replaceTaskEverywhere(client, {...cached(1), title: 'moved parent'})
		const [parent, child] = client.getQueryData<TaskResponse[]>(list)!
		expect(parent.related_tasks.subtask).toMatchObject([{id: 2, title: 'moved child', related_tasks: {}}])
		expect(child.related_tasks.parenttask).toMatchObject([{id: 1, title: 'moved parent', related_tasks: {}}])
	})
	it('drops a card from the board it was dragged out of, but not from other boards', () => {
		const client = new QueryClient()
		const dragged = normalizeTask({id: 2, project_id: 1, title: 'dragged'})
		const parent = normalizeTask({id: 1, project_id: 1, title: 'parent', related_tasks: {subtask: [dragged]}})
		const source = kanbanKeys.board(1, 4)
		const otherBoard = kanbanKeys.board(1, 5)
		const seed = (): BoardData => ({
			buckets: [normalizeBucket({id: 3, count: 9, tasks: [parent, dragged]})],
			pages: {3: 1},
		})
		client.setQueryData(source, seed())
		client.setQueryData(otherBoard, seed())
		const untouched = client.getQueryState(otherBoard)?.dataUpdateCount
		removeTaskFromBoard(client, source, 2)
		const bucket = client.getQueryData<BoardData>(source)?.buckets[0]
		expect(bucket?.tasks).toEqual([parent])
		expect(bucket?.count).toBe(8)
		expect(client.getQueryState(otherBoard)?.dataUpdateCount).toBe(untouched)
	})
	it('invalidates an open task detail without refetching it unless asked', async () => {
		sdk.tasksRead.mockResolvedValue({data: {id: 3, title: 'server'}})
		const client = new QueryClient()
		const {queryKey, queryFn} = taskQuery(3)
		const unsubscribe = new QueryObserver(client, {queryKey, queryFn}).subscribe(() => {})
		await vi.waitFor(() => expect(client.getQueryData(taskKeys.detail(3))).toBeDefined())
		await invalidateTaskMembership(client, 3)
		expect(sdk.tasksRead).toHaveBeenCalledOnce()
		expect(client.getQueryState(taskKeys.detail(3))?.isInvalidated).toBe(true)
		await invalidateTaskMembership(client, 3, 'active')
		expect(sdk.tasksRead).toHaveBeenCalledTimes(2)
		unsubscribe()
	})
	it('marks an open board stale without refetching it while lists still refetch', async () => {
		sdk.projectViewBucketsTasksList.mockResolvedValue({data: {items: [{id: 5, tasks: []}]}})
		sdk.projectTasksList.mockResolvedValue({data: {items: [], total: 0}})
		const client = new QueryClient()
		const board = kanbanQuery(1, 8)
		const list = tasksQuery({project: 1})
		const unsubscribeBoard = new QueryObserver(client, {
			queryKey: board.queryKey,
			queryFn: board.queryFn,
		}).subscribe(() => {})
		const unsubscribeList = new QueryObserver(client, {
			queryKey: list.queryKey,
			queryFn: list.queryFn,
		}).subscribe(() => {})
		await vi.waitFor(() => {
			expect(client.getQueryData(board.queryKey)).toBeDefined()
			expect(client.getQueryData(list.queryKey)).toBeDefined()
		})
		await invalidateTaskMembership(client, undefined, 'active')
		expect(sdk.projectViewBucketsTasksList).toHaveBeenCalledOnce()
		expect(sdk.projectTasksList).toHaveBeenCalledTimes(2)
		expect(client.getQueryState(board.queryKey)?.isInvalidated).toBe(true)
		unsubscribeBoard()
		unsubscribeList()
	})
	it('does not create queries for unmounted resources', () => {
		const client = new QueryClient()
		const list = taskKeys.list({project: 1})
		client.setQueryData(list, {items: [normalizeTask({id: 3, project_id: 1, title: 'old'})], total: 1})
		replaceTaskEverywhere(client, {id: 3, project_id: 2, title: 'new'})
		expect(client.getQueryData<PaginatedTaskResponse>(list)).toEqual({items: [], total: 0})
		expect(client.getQueryCache().getAll().map(query => query.queryKey)).toEqual([list])
	})
})
