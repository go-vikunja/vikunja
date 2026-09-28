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
	totalPagesFor,
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
			return {
				items: data.items ?? [],
				page: data.page ?? page,
				per_page: data.per_page ?? 0,
				total: data.total ?? 0,
				total_pages: data.total_pages ?? 0,
			}
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
		onSuccess: (_data, id, client) => client.setQueriesData<SessionPage>({queryKey: sessionKeys.all}, current => {
			if (!current) return current
			const total = Math.max(0, current.total - 1)
			return {
				...current,
				items: current.items.filter(session => session.id !== id),
				total,
				total_pages: totalPagesFor(current, total),
			}
		}),
		onSettled: (_id, client) => client.invalidateQueries({queryKey: sessionKeys.all}),
		successMessage: () => i18n.global.t('user.settings.sessions.deleteSuccess'),
	})
}

export function useDeleteSessionMutation() {
	return useMutation(deleteSessionMutationOptions())
}
