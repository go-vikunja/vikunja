import {queryOptions, type QueryKey} from '@tanstack/vue-query'
import {bucketsList, projectViewBucketsTasksList, type Bucket} from '@/client/generated'
import {normalizeTask, type TaskFilterParams, type TaskResponse} from './tasks'

export const TASKS_PER_BUCKET = 25
export const BUCKET_TASK_PARAMS = {
	per_page: TASKS_PER_BUCKET,
	expand: ['comment_count', 'is_unread'],
} satisfies TaskFilterParams

export type BucketResponse = Omit<Bucket,
	'id' |
	'title' |
	'count' |
	'limit' |
	'position' |
	'tasks'
> & {
	id: number
	title: string
	count: number
	limit: number
	position: number
	tasks: TaskResponse[]
}

export type BoardData = {buckets: BucketResponse[], pages: Record<number, number>}

export function normalizeBucket(bucket: Bucket): BucketResponse {
	if (typeof bucket.id !== 'number') {
		throw new Error('Bucket response is missing an id')
	}

	return {
		...bucket,
		id: bucket.id,
		title: bucket.title ?? '',
		count: bucket.count ?? 0,
		limit: bucket.limit ?? 0,
		position: bucket.position ?? 0,
		tasks: (bucket.tasks ?? []).map(normalizeTask),
	}
}

export function bucketHasMore(bucket: BucketResponse): boolean {
	return bucket.tasks.length < bucket.count
}

export const kanbanKeys = {
	all: ['kanban'] as const,
	view: (project: number, view: number) => [...kanbanKeys.all, project, view] as const,
	board: (project: number, view: number, params: TaskFilterParams = {}) =>
		[...kanbanKeys.view(project, view), params] as const,
	projectOf: (key: QueryKey): number | undefined => key[1] as number | undefined,
	viewOf: (key: QueryKey): number | undefined => key[2] as number | undefined,
}

// Own key root: everything under kanbanKeys holds BoardData.
export const bucketKeys = {
	all: ['buckets'] as const,
	list: (project: number, view: number) => [...bucketKeys.all, project, view] as const,
}

export function bucketsQuery(project: number, view: number) {
	return queryOptions({
		queryKey: bucketKeys.list(project, view),
		queryFn: async ({signal}): Promise<Bucket[]> =>
			(await bucketsList({path: {project, view}, signal})).data.items ?? [],
		enabled: project !== 0 && view !== 0,
	})
}

export function kanbanQuery(project: number, view: number, params: TaskFilterParams = {}) {
	return queryOptions({
		queryKey: kanbanKeys.board(project, view, params),
		queryFn: async ({signal}): Promise<BoardData> => {
			const {data} = await projectViewBucketsTasksList({
				path: {project, view},
				query: {...params, ...BUCKET_TASK_PARAMS},
				signal,
			})
			const buckets = (data.items ?? []).map(normalizeBucket)
			return {
				buckets,
				pages: Object.fromEntries(buckets.map(bucket => [bucket.id, 1])),
			}
		},
		enabled: project !== 0 && view !== 0,
	})
}
