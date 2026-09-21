import {useMutation, type QueryClient} from '@tanstack/vue-query'
import {userDeletionRequest, userDeletionConfirm, userDeletionCancel, userShow, type UserInfoBody} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {accountKeys} from './account'
import {i18n} from '@/i18n'

function reconcileAccount(account: UserInfoBody, client: QueryClient) {
	client.setQueriesData<UserInfoBody>({queryKey: accountKeys.current}, current => current ? account : current)
}

export function useRequestDeletionMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async (password: string) => (await userDeletionRequest({body: {password}})).data,
		successMessage: () => i18n.global.t('user.deletion.requestSuccess'),
	}))
}

export function useConfirmDeletionMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async (token: string) => {
			await userDeletionConfirm({body: {token}})
			return (await userShow()).data
		},
		onSuccess: (account, _input, client) => reconcileAccount(account, client),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
		successMessage: () => i18n.global.t('user.deletion.confirmSuccess'),
	}))
}

export function cancelDeletionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (password: string) => {
			await userDeletionCancel({body: {password}})
			return (await userShow()).data
		},
		onSuccess: (account, _input, client) => reconcileAccount(account, client),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
		successMessage: () => i18n.global.t('user.deletion.scheduledCancelSuccess'),
	})
}

export function useCancelDeletionMutation() {
	return useMutation(cancelDeletionMutationOptions())
}
