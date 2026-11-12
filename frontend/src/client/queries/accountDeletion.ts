import {userDeletionRequest, userDeletionConfirm, userDeletionCancel} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {reconcileAccount, type AccountIdentity} from './account'
import {i18n} from '@/i18n'

export function requestDeletionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (password: string) => (await userDeletionRequest({body: {password}})).data,
		successMessage: () => i18n.global.t('user.deletion.requestSuccess'),
	})
}

export function useRequestDeletionMutation() {
	return useSecretMutation(requestDeletionMutationOptions())
}

export function confirmDeletionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({token}: AccountIdentity & {token: string}) => {
			await userDeletionConfirm({body: {token}})
		},
		onSettled: reconcileAccount,
		successMessage: () => i18n.global.t('user.deletion.confirmSuccess'),
	})
}

export function useConfirmDeletionMutation() {
	return useSecretMutation(confirmDeletionMutationOptions())
}

export function cancelDeletionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({password}: AccountIdentity & {password: string}) => {
			await userDeletionCancel({body: {password}})
		},
		onSettled: reconcileAccount,
		successMessage: () => i18n.global.t('user.deletion.scheduledCancelSuccess'),
	})
}

export function useCancelDeletionMutation() {
	return useSecretMutation(cancelDeletionMutationOptions())
}
