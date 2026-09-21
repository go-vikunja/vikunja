import {queryOptions, useMutation} from '@tanstack/vue-query'
import {notificationsList, notificationsMarkRead, notificationsMarkAllRead, notificationsDeleteAll, type DatabaseNotification} from '@/client/generated'
import {fetchAllPages} from './fetchAllPages'
import {contextMutationOptions} from './contextMutation'
import {i18n} from '@/i18n'

export const notificationKeys = {all: ['notifications'] as const}

export function notificationsQuery() {
	return queryOptions({
		queryKey: notificationKeys.all,
		queryFn: ({signal}) => fetchAllPages(async page => (await notificationsList({query: {page}, signal})).data),
	})
}

export function markNotificationReadMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await notificationsMarkRead({path: {notificationid: id}, body: {read: true}})).data,
		onSuccess: (updated, id, client) => client.setQueryData<DatabaseNotification[]>(notificationKeys.all, current => current?.map(row => row.id === id ? {...row, ...updated} : row)),
		onSettled: (_id, client) => client.invalidateQueries({queryKey: notificationKeys.all}),
	})
}
export function markAllNotificationsReadMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await notificationsMarkAllRead()).data,
		onSettled: (_input, client) => client.invalidateQueries({queryKey: notificationKeys.all}),
		successMessage: () => i18n.global.t('notification.markAllReadSuccess'),
	})
}
export function clearNotificationsMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await notificationsDeleteAll()).data,
		onSuccess: (_data, _input, client) => client.setQueryData<DatabaseNotification[]>(notificationKeys.all, current => current && []),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: notificationKeys.all}),
		successMessage: () => i18n.global.t('notification.clearAllSuccess'),
	})
}
export function useMarkNotificationReadMutation() { return useMutation(markNotificationReadMutationOptions()) }
export function useMarkAllNotificationsReadMutation() { return useMutation(markAllNotificationsReadMutationOptions()) }
export function useClearNotificationsMutation() { return useMutation(clearNotificationsMutationOptions()) }
