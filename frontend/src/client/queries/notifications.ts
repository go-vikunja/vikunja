import {
	infiniteQueryOptions,
	useMutation,
	type InfiniteData,
	type QueryClient,
} from '@tanstack/vue-query'
import {
	notificationsList,
	notificationsMarkRead,
	notificationsMarkAllRead,
	notificationsDeleteAll,
	type DatabaseNotification,
} from '@/client/generated'
import type {Paginated} from './pagination'
import {contextMutationOptions} from './contextMutation'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {i18n} from '@/i18n'

export const notificationKeys = {all: ['notifications'] as const}

export const NOTIFICATIONS_PER_PAGE = 20

export type NotificationPage = Paginated<DatabaseNotification>
export type NotificationPages = InfiniteData<NotificationPage, number>

export function notificationsQuery() {
	return infiniteQueryOptions({
		queryKey: notificationKeys.all,
		queryFn: async ({pageParam, signal}): Promise<NotificationPage> => {
			const {data} = await notificationsList({
				query: {
					page: pageParam,
					per_page: NOTIFICATIONS_PER_PAGE,
				},
				signal,
			})
			return {
				items: data.items ?? [],
				page: data.page ?? pageParam,
				per_page: data.per_page ?? NOTIFICATIONS_PER_PAGE,
				total: data.total ?? 0,
				total_pages: data.total_pages ?? 0,
			}
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage, _pages, lastPageParam) =>
			lastPageParam < lastPage.total_pages ? lastPageParam + 1 : undefined,
	})
}

function patchNotifications(client: QueryClient, patch: (row: DatabaseNotification) => DatabaseNotification) {
	client.setQueryData<NotificationPages>(notificationKeys.all, current => current && {
		...current,
		pages: current.pages.map(page => ({
			...page,
			items: page.items.map(patch),
		})),
	})
}

export function markNotificationReadMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await notificationsMarkRead({
			path: {notificationid: id},
			body: {read: true},
		})).data,
		onSuccess: (updated, id, client) => patchNotifications(
			client,
			row => row.id === id ? {...row, ...updated} : row,
		),
		onSettled: (_id, client) => client.invalidateQueries({queryKey: notificationKeys.all}),
	})
}
export function markAllNotificationsReadMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await notificationsMarkAllRead()).data,
		onSuccess: (_data, _input, client) => {
			const readAt = new Date().toISOString()
			patchNotifications(
				client,
				row => parseDateOrNull(row.read_at) === null ? {...row, read_at: readAt} : row,
			)
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: notificationKeys.all}),
		successMessage: () => i18n.global.t('notification.markAllReadSuccess'),
	})
}
export function clearNotificationsMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await notificationsDeleteAll()).data,
		onSuccess: (_data, _input, client) => client.setQueryData<NotificationPages>(
			notificationKeys.all,
			current => current && {
				pages: current.pages.slice(0, 1).map(page => ({
					...page,
					items: [],
					total: 0,
					total_pages: 0,
				})),
				pageParams: current.pageParams.slice(0, 1),
			},
		),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: notificationKeys.all}),
		successMessage: () => i18n.global.t('notification.clearAllSuccess'),
	})
}
export function useMarkNotificationReadMutation() { return useMutation(markNotificationReadMutationOptions()) }
export function useMarkAllNotificationsReadMutation() { return useMutation(markAllNotificationsReadMutationOptions()) }
export function useClearNotificationsMutation() { return useMutation(clearNotificationsMutationOptions()) }
