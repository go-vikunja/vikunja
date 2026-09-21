import {queryOptions, useMutation} from '@tanstack/vue-query'
import {caldavTokensList, caldavTokensCreate, caldavTokensDelete, type Token} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {i18n} from '@/i18n'

export const caldavTokenKeys = {all: ['caldav-tokens'] as const}

export function caldavTokensQuery() {
	return queryOptions({
		queryKey: caldavTokenKeys.all,
		// The handler returns every token on each page, so later pages would only repeat them.
		queryFn: async ({signal}) => (await caldavTokensList({
			query: {page: 1},
			signal,
		})).data.items ?? [],
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
		onSuccess: (_data, id, client) => client.setQueryData<Token[]>(caldavTokenKeys.all, current => current?.filter(token => token.id !== id)),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: caldavTokenKeys.all}),
		successMessage: () => i18n.global.t('user.settings.caldav.deleteSuccess'),
	})
}

export function useDeleteCaldavTokenMutation() {
	return useMutation(deleteCaldavTokenMutationOptions())
}
