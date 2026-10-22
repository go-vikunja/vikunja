import {
	keepPreviousData,
	queryOptions,
	useMutation,
} from '@tanstack/vue-query'
import {
	sessionsList,
	sessionsDelete,
	type Session,
} from '@/client/generated'
import {
	removeFromPages,
	toPaginated,
	type Paginated,
} from './pagination'
import {contextMutationOptions} from './contextMutation'
import {i18n} from '@/i18n'

export type SessionPage = Paginated<Session>

export const sessionKeys = {
	all: ['sessions'] as const,
	list: (page: number) => ['sessions', 'list', page] as const,
}

export function sessionsQuery(page: number) {
	return queryOptions({
		queryKey: sessionKeys.list(page),
		queryFn: async ({signal}): Promise<SessionPage> => {
			const {data} = await sessionsList({
				query: {page},
				signal,
			})
			return toPaginated(data, page)
		},
		placeholderData: keepPreviousData,
		staleTime: 0,
	})
}

export function deleteSessionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: string) => {
			await sessionsDelete({path: {session: id}})
		},
		onSuccess: (_data, id, client) => removeFromPages<Session>(client, sessionKeys.all, session => session.id === id),
		onSettled: (_id, client) => client.invalidateQueries({queryKey: sessionKeys.all}),
		successMessage: () => i18n.global.t('user.settings.sessions.deleteSuccess'),
	})
}

export function useDeleteSessionMutation() {
	return useMutation(deleteSessionMutationOptions())
}
