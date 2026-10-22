import {hashKey, type QueryClient, type QueryKey} from '@tanstack/vue-query'
import type {Task} from '@/client/generated'
import {mapTasksDeep, mergeTask, removeTask} from '@/helpers/task'
import {normalizeTask, taskKeys, type PaginatedTaskResponse, type TaskResponse} from './tasks'
import {kanbanKeys, type BoardData} from './kanban'
import {totalPagesFor} from './pagination'

// Applies `update` to every cached copy of task `id` (detail, lists, boards, nested relations),
// so a mutation response patches all mounted views without refetching them.
export function mapTaskEverywhere(
	client: QueryClient,
	id: number,
	update: (task: TaskResponse) => TaskResponse,
) {
	const map = (tasks: readonly TaskResponse[]) =>
		mapTasksDeep(tasks, id, task => update(normalizeTask(task))).map(normalizeTask)
	const holds = (tasks: readonly TaskResponse[]) => containsTask(tasks, id)
	client.setQueriesData<TaskResponse[]>({queryKey: taskKeys.allLists}, current =>
		current && holds(current) ? map(current) : undefined,
	)
	client.setQueriesData<TaskResponse>({queryKey: taskKeys.details}, current =>
		current && holds([current]) ? map([current])[0] : undefined,
	)
	client.setQueriesData<PaginatedTaskResponse>({queryKey: taskKeys.lists}, current =>
		current && holds(current.items) ? {...current, items: map(current.items)} : undefined,
	)
	client.setQueriesData<BoardData>({queryKey: kanbanKeys.all}, current => {
		if (!current) return undefined
		const buckets = current.buckets.map(bucket =>
			holds(bucket.tasks) ? {...bucket, tasks: map(bucket.tasks)} : bucket,
		)
		return buckets.some((bucket, index) => bucket !== current.buckets[index])
			? {...current, buckets}
			: undefined
	})
}

// undefined: task absent, entry stays untouched
type TaskRemoval = (tasks: readonly TaskResponse[], id: number) => TaskResponse[] | undefined

// Relations cross projects, so a move only drops top-level membership.
const dropMembership: TaskRemoval = (tasks, id) => tasks.some(task => task.id === id)
	? tasks.filter(task => task.id !== id)
	: undefined

const dropEverywhere: TaskRemoval = (tasks, id) => containsTask(tasks, id)
	? removeTask(tasks, id).map(normalizeTask)
	: undefined

// Favorites and saved filters span projects.
const isRealProject = (project: number | null | undefined): project is number =>
	typeof project === 'number' && project > 0

function containsTask(tasks: readonly Task[], id: number): boolean {
	return tasks.some(task => task.id === id
		|| Object.values(task.related_tasks ?? {}).some(children => containsTask(children ?? [], id)))
}

export function taskQueryKeys(client: QueryClient, id: number): QueryKey[] {
	return [
		...client.getQueriesData<TaskResponse>({queryKey: taskKeys.details})
			.filter(([, task]) => task && containsTask([task], id)),
		...client.getQueriesData<TaskResponse[]>({queryKey: taskKeys.allLists})
			.filter(([, tasks]) => tasks && containsTask(tasks, id)),
		...client.getQueriesData<PaginatedTaskResponse>({queryKey: taskKeys.lists})
			.filter(([, list]) => list && containsTask(list.items, id)),
		...client.getQueriesData<BoardData>({queryKey: kanbanKeys.all})
			.filter(([, board]) => board?.buckets.some(bucket => containsTask(bucket.tasks, id))),
	].map(([key]) => key)
}

type CollectionRemoval = {kind: 'move', project: number} | {kind: 'delete'}

function removeTaskFromCollections(client: QueryClient, id: number, removal: CollectionRemoval) {
	const movedTo = removal.kind === 'move' ? removal.project : undefined
	const remove: TaskRemoval = movedTo === undefined ? dropEverywhere : dropMembership
	const inScope = (project: number | null | undefined) =>
		movedTo === undefined || (isRealProject(project) && project !== movedTo)
	for (const [key, tasks] of client.getQueriesData<TaskResponse[]>({queryKey: taskKeys.allLists})) {
		if (!tasks || !inScope(taskKeys.projectOf(key))) continue
		const items = remove(tasks, id)
		if (items) client.setQueryData(key, items)
	}
	// Every cached page of a scope shares total and total_pages, not only the page holding the task.
	const removedPerScope = new Map<string, number>()
	const scopeOf = (key: QueryKey) => hashKey(key.slice(0, -1))
	for (const [key, list] of client.getQueriesData<PaginatedTaskResponse>({queryKey: taskKeys.lists})) {
		if (!list || !inScope(taskKeys.projectOf(key))) continue
		const items = remove(list.items, id)
		if (!items) continue
		client.setQueryData(key, {...list, items})
		removedPerScope.set(scopeOf(key), (removedPerScope.get(scopeOf(key)) ?? 0) + list.items.length - items.length)
	}
	for (const [key, list] of client.getQueriesData<PaginatedTaskResponse>({queryKey: taskKeys.lists})) {
		const removed = removedPerScope.get(scopeOf(key))
		if (!list || !removed) continue
		const total = Math.max(0, list.total - removed)
		const total_pages = totalPagesFor(list, total)
		client.setQueryData(key, {...list, total, total_pages})
	}
	for (const [key, board] of client.getQueriesData<BoardData>({queryKey: kanbanKeys.all})) {
		if (!board || !inScope(kanbanKeys.projectOf(key))) continue
		const next = removeFromBoard(board, id, remove)
		if (next) client.setQueryData(key, next)
	}
}

function removeFromBoard(board: BoardData, id: number, remove: TaskRemoval): BoardData | undefined {
	const buckets = board.buckets.map(bucket => {
		const tasks = remove(bucket.tasks, id)
		if (!tasks) return bucket
		return {...bucket, tasks, count: Math.max(0, bucket.count - (bucket.tasks.length - tasks.length))}
	})
	return buckets.some((bucket, index) => bucket !== board.buckets[index]) ? {...board, buckets} : undefined
}

// A move keeps pseudo-project boards (membership is server-decided), so the drag source drops its own card.
export function removeTaskFromBoard(client: QueryClient, key: QueryKey, id: number) {
	client.setQueryData<BoardData>(key, board => board && removeFromBoard(board, id, dropMembership))
}

export function replaceTaskEverywhere(client: QueryClient, updated: Task) {
	if (updated.id === undefined) return
	if (updated.project_id) {
		removeTaskFromCollections(client, updated.id, {kind: 'move', project: updated.project_id})
	}
	mapTaskEverywhere(client, updated.id, current => normalizeTask(mergeTask(current, updated)))
}

export function removeTaskEverywhere(client: QueryClient, id: number) {
	removeTaskFromCollections(client, id, {kind: 'delete'})
	client.removeQueries({queryKey: [...taskKeys.details, id]})
	client.setQueriesData<TaskResponse>({queryKey: taskKeys.details}, current =>
		current ? removeTask([current], id).map(normalizeTask)[0] : current,
	)
}

export async function invalidateTaskMembership(
	client: QueryClient,
	id?: number,
	refetchType: 'none' | 'active' = 'none',
) {
	await Promise.all([
		client.invalidateQueries({queryKey: taskKeys.lists, refetchType}),
		client.invalidateQueries({queryKey: taskKeys.allLists, refetchType}),
		// Refetching a board returns only the first page per bucket, so mutations patch it instead of refetching.
		client.invalidateQueries({queryKey: kanbanKeys.all, refetchType: 'none'}),
		...(id === undefined ? [] : [client.invalidateQueries({queryKey: [...taskKeys.details, id], refetchType})]),
	])
}
