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
	removeFromPages,
	toPaginated,
	type Paginated,
} from './pagination'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {i18n} from '@/i18n'

export type CaldavTokenPage = Paginated<Token>

export const caldavTokenKeys = {
	all: ['caldav-tokens'] as const,
	list: (page: number) => ['caldav-tokens', 'list', page] as const,
}

export function caldavTokensQuery(page: number) {
	return queryOptions({
		queryKey: caldavTokenKeys.list(page),
		queryFn: async ({signal}): Promise<CaldavTokenPage> => {
			const {data} = await caldavTokensList({
				query: {page},
				signal,
			})
			return toPaginated(data, page)
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
		onSuccess: (_data, id, client) => removeFromPages<Token>(client, caldavTokenKeys.all, token => token.id === id),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: caldavTokenKeys.all}),
		successMessage: () => i18n.global.t('user.settings.caldav.deleteSuccess'),
	})
}

export function useDeleteCaldavTokenMutation() {
	return useMutation(deleteCaldavTokenMutationOptions())
}
