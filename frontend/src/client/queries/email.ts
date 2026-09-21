import {useMutation, type QueryClient} from '@tanstack/vue-query'
import {userUpdateEmail, userCancelEmailUpdate, userResendEmailConfirmation, userShow, type UserInfoBody, type UserUpdateEmailRequestWritable} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {accountKeys} from './account'
import {i18n} from '@/i18n'

function reconcileAccount(account: UserInfoBody, client: QueryClient) {
	client.setQueriesData<UserInfoBody>({queryKey: accountKeys.current}, current => current ? account : current)
}

export function updateEmailMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: UserUpdateEmailRequestWritable) => {
			await userUpdateEmail({body})
			return (await userShow()).data
		},
		onSuccess: (account, _input, client) => reconcileAccount(account, client),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
		successMessage: account => i18n.global.t(account.pending_email ? 'user.settings.updateEmailPendingSuccess' : 'user.settings.updateEmailSuccess'),
	})
}

export function useUpdateEmailMutation() {
	return useMutation(updateEmailMutationOptions())
}

export function useCancelEmailUpdateMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async () => {
			await userCancelEmailUpdate()
			return (await userShow()).data
		},
		onSuccess: (account, _input, client) => reconcileAccount(account, client),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
		successMessage: () => i18n.global.t('user.settings.updateEmailCancelSuccess'),
	}))
}

export function useResendEmailConfirmationMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async () => (await userResendEmailConfirmation()).data,
		successMessage: () => i18n.global.t('user.settings.updateEmailResendSuccess'),
	}))
}
