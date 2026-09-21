import {useMutation, type QueryClient} from '@tanstack/vue-query'
import {
	userUpdateEmail,
	userCancelEmailUpdate,
	userResendEmailConfirmation,
	userShow,
} from '@/client/generated'
import type {
	UserInfoBody,
	UserUpdateEmailRequestWritable,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {accountKeys} from './account'
import {i18n} from '@/i18n'

function reconcileAccount(account: UserInfoBody, client: QueryClient) {
	client.setQueriesData<UserInfoBody>({queryKey: accountKeys.current}, current => current ? account : current)
}

export function updateEmailMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (body: UserUpdateEmailRequestWritable) => {
				await userUpdateEmail({body})
				return (await userShow()).data
			},
			onSuccess: (account, _input, client) => reconcileAccount(account, client),
			onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
			successMessage: account => {
				const key = account.pending_email
					? 'user.settings.updateEmailPendingSuccess'
					: 'user.settings.updateEmailSuccess'
				return i18n.global.t(key)
			},
		}),
		// Input holds the plaintext password.
		gcTime: 0,
	}
}

export function useUpdateEmailMutation() {
	return useMutation(updateEmailMutationOptions())
}

export function cancelEmailUpdateMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => {
			await userCancelEmailUpdate()
			return (await userShow()).data
		},
		onSuccess: (account, _input, client) => reconcileAccount(account, client),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
		successMessage: () => i18n.global.t('user.settings.updateEmailCancelSuccess'),
	})
}

export function useCancelEmailUpdateMutation() {
	return useMutation(cancelEmailUpdateMutationOptions())
}

export function resendEmailConfirmationMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await userResendEmailConfirmation()).data,
		successMessage: () => i18n.global.t('user.settings.updateEmailResendSuccess'),
	})
}

export function useResendEmailConfirmationMutation() {
	return useMutation(resendEmailConfirmationMutationOptions())
}
