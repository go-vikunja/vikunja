import {queryOptions, useMutation} from '@tanstack/vue-query'
import {sessionsList, sessionsDelete, type Session} from '@/client/generated'
import {fetchAllPages} from './fetchAllPages'
import {contextMutationOptions} from './contextMutation'
import {i18n} from '@/i18n'

export const sessionKeys = {all: ['sessions'] as const}

export function sessionsQuery() {
	return queryOptions({
		queryKey: sessionKeys.all,
		queryFn: ({signal}) => fetchAllPages(async page => (await sessionsList({query: {page}, signal})).data),
	})
}

export function deleteSessionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: string) => (await sessionsDelete({path: {session: id}})).data,
		onSuccess: (_data, id, client) => client.setQueryData<Session[]>(sessionKeys.all, current => current?.filter(session => session.id !== id)),
		onSettled: (_id, client) => client.invalidateQueries({queryKey: sessionKeys.all}),
		successMessage: () => i18n.global.t('user.settings.sessions.deleteSuccess'),
	})
}

export function useDeleteSessionMutation() {
	return useMutation(deleteSessionMutationOptions())
}
