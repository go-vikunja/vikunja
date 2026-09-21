import {queryOptions, type QueryKey} from '@tanstack/vue-query'
import {projectTasksList, projectViewTasksList, tasksList, tasksRead} from '@/client/generated'
import type {
	Label,
	PaginatedTask,
	Task,
	TaskAttachment,
	TaskReadOneBody,
	TaskReminder,
	TasksListData,
	TasksReadData,
	User,
} from '@/client/generated'
import {fetchAllPages} from './fetchAllPages'
import {API_MAX_PER_PAGE, type Paginated} from './pagination'
import {queryClient} from '@/client/queryClient'

export type TaskFilterParams = Omit<NonNullable<TasksListData['query']>, 'format' | 'page'>
export type TaskExpansion = NonNullable<NonNullable<TasksReadData['query']>['expand']>
export type TaskScope = {
	project?: number | null,
	view?: number,
	params?: TaskFilterParams,
}

// The read-one body is a task plus max_permission, which the other task endpoints leave out.
export type TaskResponse = Omit<TaskReadOneBody,
	'id' |
	'title' |
	'description' |
	'done' |
	'priority' |
	'percent_done' |
	'hex_color' |
	'is_favorite' |
	'identifier' |
	'index' |
	'position' |
	'project_id' |
	'bucket_id' |
	'repeat_after' |
	'repeat_mode' |
	'cover_image_attachment_id' |
	'labels' |
	'assignees' |
	'reminders' |
	'attachments' |
	'related_tasks' |
	'reactions'
> & {
	id: number
	title: string
	description: string
	done: boolean
	priority: number
	percent_done: number
	hex_color: string
	is_favorite: boolean
	identifier: string
	index: number
	position: number
	project_id: number
	bucket_id: number
	repeat_after: number
	repeat_mode: number
	cover_image_attachment_id: number
	labels: Label[]
	assignees: User[]
	reminders: TaskReminder[]
	attachments: TaskAttachment[]
	related_tasks: Record<string, TaskResponse[]>
	reactions: Record<string, User[]>
}

export type PaginatedTaskResponse = Omit<PaginatedTask, keyof Paginated<unknown>> & Paginated<TaskResponse>

export function getDefaultTaskFilterParams(): TaskFilterParams {
	return {
		sort_by: ['position', 'id'],
		order_by: ['asc', 'desc'],
		filter: '',
		filter_include_nulls: false,
		filter_timezone: '',
		q: '',
		expand: ['subtasks'],
	}
}

export function normalizeTask(task: Task): TaskResponse {
	if (typeof task.id !== 'number') {
		throw new Error('Task response is missing an id')
	}

	return {
		...task,
		id: task.id,
		title: task.title ?? '',
		description: task.description ?? '',
		done: task.done ?? false,
		priority: task.priority ?? 0,
		percent_done: task.percent_done ?? 0,
		hex_color: task.hex_color ?? '',
		is_favorite: task.is_favorite ?? false,
		identifier: task.identifier ?? '',
		index: task.index ?? 0,
		position: task.position ?? 0,
		project_id: task.project_id ?? 0,
		bucket_id: task.bucket_id ?? 0,
		repeat_after: task.repeat_after ?? 0,
		repeat_mode: task.repeat_mode ?? 0,
		cover_image_attachment_id: task.cover_image_attachment_id ?? 0,
		labels: task.labels ?? [],
		assignees: task.assignees ?? [],
		reminders: task.reminders ?? [],
		attachments: task.attachments ?? [],
		related_tasks: Object.fromEntries(
			Object.entries(task.related_tasks ?? {})
				// The API nests relations one level deep; deeper copies can cycle back to the task.
				.map(([kind, related]) => [kind, (related ?? []).map(child => normalizeTask({...child, related_tasks: {}}))]),
		),
		reactions: Object.fromEntries(
			Object.entries(task.reactions ?? {})
				.map(([reaction, users]) => [reaction, users ?? []]),
		),
	}
}

function normalizeTaskList(page: PaginatedTask): PaginatedTaskResponse {
	return {
		...page,
		items: (page.items ?? []).map(normalizeTask),
		total: page.total ?? 0,
		page: page.page ?? 1,
		per_page: page.per_page ?? 0,
		total_pages: page.total_pages ?? 0,
	}
}

export function normalizePageNumber(page: unknown): number {
	const parsed = Number(page)
	return Number.isInteger(parsed) && parsed >= 1 ? parsed : 1
}

const taskKeyRoot = ['tasks'] as const

export const taskKeys = {
	all: taskKeyRoot,
	details: [...taskKeyRoot, 'detail'] as const,
	detail: (id: number, expand: TaskExpansion = []) => [...taskKeys.details, id, expand] as const,
	lists: [...taskKeyRoot, 'list'] as const,
	allLists: [...taskKeyRoot, 'all'] as const,
	allList: ({project = null, view = 0, params = {}}: TaskScope) =>
		[...taskKeys.allLists, project, view, params] as const,
	list: ({project = null, view = 0, params = {}}: TaskScope, page = 1) =>
		[...taskKeys.lists, project, view, params, normalizePageNumber(page)] as const,
	projectOf: (key: QueryKey): number | null | undefined => key[1] === 'list' || key[1] === 'all'
		? key[2] as number | null
		: undefined,
	viewOf: (key: QueryKey): number | undefined => key[1] === 'list' || key[1] === 'all'
		? key[3] as number
		: undefined,
	paramsOf: (key: QueryKey): TaskFilterParams | undefined => key[1] === 'list' || key[1] === 'all'
		? key[4] as TaskFilterParams
		: undefined,
}

async function fetchTaskList(
	{project = null, view = 0, params = {}}: TaskScope,
	page: number,
	signal?: AbortSignal,
) {
	const query = {...params, page}
	if (project === null) return (await tasksList({query, signal})).data
	if (view) return (await projectViewTasksList({path: {project, view}, query, signal})).data
	return (await projectTasksList({path: {project}, query, signal})).data
}

export function taskQuery(id: number, expand: TaskExpansion = []) {
	return queryOptions({
		queryKey: taskKeys.detail(id, expand),
		queryFn: async ({signal}) => normalizeTask((await tasksRead({
			path: {task: id},
			query: {expand},
			signal,
		})).data),
		enabled: id > 0,
	})
}

export function tasksQuery(scope: TaskScope = {}, requestedPage = 1) {
	const page = normalizePageNumber(requestedPage)
	return queryOptions({
		queryKey: taskKeys.list(scope, page),
		queryFn: async ({signal}) => normalizeTaskList(await fetchTaskList(scope, page, signal)),
	})
}

export function ensureTask(id: number, expand: TaskExpansion = []) {
	return queryClient.ensureQueryData(taskQuery(id, expand))
}

export function allTasksQuery(scope: TaskScope) {
	const exhaustiveScope = {...scope, params: {...scope.params, per_page: API_MAX_PER_PAGE}}
	return queryOptions({
		queryKey: taskKeys.allList(scope),
		queryFn: ({signal}) => fetchAllPages(
			async page => normalizeTaskList(await fetchTaskList(exhaustiveScope, page, signal)),
		),
	})
}
