import {
	keepPreviousData,
	queryOptions,
	useMutation,
} from '@tanstack/vue-query'
import {
	caldavTokensList,
	caldavTokensCreate,
	caldavTokensDelete,
	type Token,
} from '@/client/generated'
import {
	totalPagesFor,
	type Paginated,
} from './pagination'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {i18n} from '@/i18n'

const PER_PAGE = 25

export type CaldavTokenPage = Paginated<Token>

export const caldavTokenKeys = {
	all: ['caldav-tokens'] as const,
	list: (page: number) => ['caldav-tokens', page] as const,
}

export function caldavTokensQuery(page: number) {
	return queryOptions({
		queryKey: caldavTokenKeys.list(page),
		queryFn: async ({signal}): Promise<CaldavTokenPage> => {
			const {data} = await caldavTokensList({
				query: {
					page,
					per_page: PER_PAGE,
				},
				signal,
			})
			return {
				items: data.items ?? [],
				page: data.page ?? page,
				per_page: data.per_page ?? PER_PAGE,
				total: data.total ?? 0,
				total_pages: data.total_pages ?? 0,
			}
		},
		placeholderData: keepPreviousData,
	})
}

export function createCaldavTokenMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await caldavTokensCreate()).data,
		onSettled: (_input, client) => client.invalidateQueries({queryKey: caldavTokenKeys.all}),
	})
}

export function useCreateCaldavTokenMutation() {
	return useSecretMutation(createCaldavTokenMutationOptions())
}

export function deleteCaldavTokenMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await caldavTokensDelete({path: {id}})).data,
		onSuccess: (_data, id, client) => client.setQueriesData<CaldavTokenPage>({queryKey: caldavTokenKeys.all}, current => {
			if (!current) return current
			const total = Math.max(0, current.total - 1)
			return {
				...current,
				items: current.items.filter(token => token.id !== id),
				total,
				total_pages: totalPagesFor(current, total),
			}
		}),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: caldavTokenKeys.all}),
		successMessage: () => i18n.global.t('user.settings.caldav.deleteSuccess'),
	})
}

export function useDeleteCaldavTokenMutation() {
	return useMutation(deleteCaldavTokenMutationOptions())
}
