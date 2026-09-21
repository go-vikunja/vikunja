import {queryOptions, useMutation} from '@tanstack/vue-query'
import {caldavTokensList, caldavTokensCreate, caldavTokensDelete, type Token} from '@/client/generated'
import {fetchAllPages} from './fetchAllPages'
import {contextMutationOptions} from './contextMutation'
import {i18n} from '@/i18n'

export const caldavTokenKeys = {all: ['caldav-tokens'] as const}

export function caldavTokensQuery() {
	return queryOptions({
		queryKey: caldavTokenKeys.all,
		queryFn: ({signal}) => fetchAllPages(async page => (await caldavTokensList({query: {page}, signal})).data),
	})
}

export function createCaldavTokenMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async () => (await caldavTokensCreate()).data,
			onSettled: (_input, client) => client.invalidateQueries({queryKey: caldavTokenKeys.all}),
		}),
		// Result holds the plaintext token.
		gcTime: 0,
	}
}

export function useCreateCaldavTokenMutation() {
	return useMutation(createCaldavTokenMutationOptions())
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
