import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'
import {taskCommentsCreate, taskCommentsDelete, taskCommentsList, taskCommentsUpdate} from '@/client/generated'
import type {PaginatedTaskComment, TaskComment, TaskCommentsListData} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {mapTaskEverywhere} from './taskCache'
import {taskKeys} from './tasks'
import {i18n} from '@/i18n'

export type CommentOrder = NonNullable<NonNullable<TaskCommentsListData['query']>['order_by']>
export type CommentResponse = Omit<TaskComment, 'id' | 'comment' | 'reactions'> & {
	id: number,
	comment: string,
	reactions: Record<string, NonNullable<NonNullable<TaskComment['reactions']>[string]>>,
}
export type CommentPage = Omit<Required<PaginatedTaskComment>, 'items'> & {items: CommentResponse[]}
export const commentKeys = {
	all: ['comments'] as const,
	task: (taskId: number) => ['comments', taskId] as const,
	page: (taskId: number, order: CommentOrder, page: number) => ['comments', taskId, order, page] as const,
}

export function normalizeComment(comment: TaskComment): CommentResponse {
	return {
		...comment,
		id: comment.id ?? 0,
		comment: comment.comment ?? '',
		reactions: Object.fromEntries(Object.entries(comment.reactions ?? {}).map(([value, users]) => [value, users ?? []])),
	}
}

export function commentsQuery(taskId: number, order: CommentOrder, page: number) {
	return queryOptions({
		queryKey: commentKeys.page(taskId, order, page),
		enabled: taskId > 0,
		queryFn: async ({signal}): Promise<CommentPage> => {
			const {data} = await taskCommentsList({path: {task: taskId}, query: {order_by: order, page}, signal})
			return {
				...data,
				items: (data.items ?? []).map(normalizeComment),
				page: data.page ?? page,
				per_page: data.per_page ?? 50,
				total: data.total ?? 0,
				total_pages: data.total_pages ?? 0,
			}
		},
	})
}

function settle(client: QueryClient, taskId: number) {
	return Promise.all([
		client.invalidateQueries({queryKey: commentKeys.task(taskId)}),
		client.invalidateQueries({queryKey: taskKeys.all, refetchType: 'none'}),
	])
}

export function createCommentMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, comment}: {taskId: number, comment: string}) =>
			(await taskCommentsCreate({path: {task: taskId}, body: {comment}})).data,
		onSuccess: (comment, {taskId}, client) => {
			for (const [key, current] of client.getQueriesData<CommentPage>({queryKey: commentKeys.task(taskId)})) {
				if (!current) continue
				const total = current.total + 1
				const total_pages = Math.ceil(total / current.per_page)
				let items = current.items
				if (key[2] === 'desc' && current.page === 1) {
					items = [normalizeComment(comment), ...items].slice(0, current.per_page)
				} else if (key[2] === 'asc' && current.page === total_pages) {
					items = [...items, normalizeComment(comment)]
				}
				client.setQueryData(key, {...current, total, total_pages, items})
			}
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				comment_count: task.comment_count === undefined ? undefined : task.comment_count + 1,
			}))
		},
		onSettled: ({taskId}, client) => settle(client, taskId),
		successMessage: () => i18n.global.t('task.comment.addedSuccess'),
	})
}

export function updateCommentMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, id, comment}: {taskId: number, id: number, comment: string}) =>
			(await taskCommentsUpdate({path: {task: taskId, commentid: id}, body: {comment}})).data,
		onSuccess: (updated, {taskId, id}, client) => {
			const merge = (comment: TaskComment) => normalizeComment({
				...comment,
				...updated,
				reactions: updated.reactions ?? comment.reactions,
			})
			client.setQueriesData<CommentPage>({queryKey: commentKeys.task(taskId)}, current => current && ({
				...current,
				items: current.items.map(comment => comment.id === id ? merge(comment) : comment),
			}))
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				comments: task.comments?.map(comment => comment.id === id ? merge(comment) : comment),
			}))
		},
		onSettled: ({taskId}, client) => settle(client, taskId),
	})
}

export function deleteCommentMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, id}: {taskId: number, id: number}) =>
			(await taskCommentsDelete({path: {task: taskId, commentid: id}})).data,
		onSuccess: (_data, {taskId, id}, client) => {
			client.setQueriesData<CommentPage>({queryKey: commentKeys.task(taskId)}, current => {
				if (!current) return current
				const total = Math.max(0, current.total - 1)
				return {...current, items: current.items.filter(c => c.id !== id), total, total_pages: Math.ceil(total / current.per_page)}
			})
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				comment_count: task.comment_count === undefined ? undefined : Math.max(0, task.comment_count - 1),
				comments: task.comments?.filter(comment => comment.id !== id),
			}))
		},
		onSettled: ({taskId}, client) => settle(client, taskId),
		successMessage: () => i18n.global.t('task.comment.deleteSuccess'),
	})
}

export function useCreateCommentMutation() { return useMutation(createCommentMutationOptions()) }
export function useUpdateCommentMutation() { return useMutation(updateCommentMutationOptions()) }
export function useDeleteCommentMutation() { return useMutation(deleteCommentMutationOptions()) }
