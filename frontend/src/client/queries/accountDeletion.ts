import {useMutation, type QueryClient} from '@tanstack/vue-query'
import {userDeletionRequest, userDeletionConfirm, userDeletionCancel} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {reconcileAccount, type AccountIdentity} from './account'
import {i18n} from '@/i18n'

// The write already landed; a failing re-read must not fail the mutation.
function reconcileDeletionState(identity: AccountIdentity, client: QueryClient) {
	return reconcileAccount(identity, client).catch(() => undefined)
}

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
			mutationFn: async ({token}: AccountIdentity & {token: string}) => {
				await userDeletionConfirm({body: {token}})
			},
			onSettled: reconcileDeletionState,
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
			mutationFn: async ({password}: AccountIdentity & {password: string}) => {
				await userDeletionCancel({body: {password}})
			},
			onSettled: reconcileDeletionState,
			successMessage: () => i18n.global.t('user.deletion.scheduledCancelSuccess'),
		}),
		gcTime: 0,
	}
}

export function useCancelDeletionMutation() {
	return useMutation(cancelDeletionMutationOptions())
}
