import {useMutation} from '@tanstack/vue-query'
import {userDeletionRequest, userDeletionConfirm, userDeletionCancel, userShow} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {accountKeys, reconcileAccount} from './account'
import {i18n} from '@/i18n'

export function requestDeletionMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (password: string) => (await userDeletionRequest({body: {password}})).data,
			successMessage: () => i18n.global.t('user.deletion.requestSuccess'),
		}),
		// Input holds the plaintext password.
		gcTime: 0,
	}
}

export function useRequestDeletionMutation() {
	return useMutation(requestDeletionMutationOptions())
}

export function confirmDeletionMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (token: string) => {
				await userDeletionConfirm({body: {token}})
				// The write already landed; a failing re-read must not fail the mutation.
				return userShow().then(({data}) => data).catch(() => undefined)
			},
			onSuccess: (account, _input, client) => {
				if (account) reconcileAccount(account, client)
			},
			onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
			successMessage: () => i18n.global.t('user.deletion.confirmSuccess'),
		}),
		gcTime: 0,
	}
}

export function useConfirmDeletionMutation() {
	return useMutation(confirmDeletionMutationOptions())
}

export function cancelDeletionMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (password: string) => {
				await userDeletionCancel({body: {password}})
				return userShow().then(({data}) => data).catch(() => undefined)
			},
			onSuccess: (account, _input, client) => {
				if (account) reconcileAccount(account, client)
			},
			onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
			successMessage: () => i18n.global.t('user.deletion.scheduledCancelSuccess'),
		}),
		gcTime: 0,
	}
}

export function useCancelDeletionMutation() {
	return useMutation(cancelDeletionMutationOptions())
}
