import {useMutation, type QueryClient} from '@tanstack/vue-query'
import {
	userUpdateEmail,
	userCancelEmailUpdate,
	userResendEmailConfirmation,
	userShow,
} from '@/client/generated'
import type {UserUpdateEmailRequestWritable} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {
	accountKeys,
	normalizeUserInfo,
	reconcileAccount,
	type AccountIdentity,
	type UserInfoResponse,
} from './account'
import {i18n} from '@/i18n'

function writeAccount(account: UserInfoResponse, {id, type}: AccountIdentity, client: QueryClient) {
	client.setQueryData<UserInfoResponse>(accountKeys.user(id, type), current => current ? account : current)
}

export function updateEmailMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async ({body}: AccountIdentity & {body: UserUpdateEmailRequestWritable}) => {
				await userUpdateEmail({body})
				return normalizeUserInfo((await userShow()).data)
			},
			onSuccess: writeAccount,
			onSettled: reconcileAccount,
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
	return contextMutationOptions<UserInfoResponse, AccountIdentity>({
		mutationFn: async () => {
			await userCancelEmailUpdate()
			return normalizeUserInfo((await userShow()).data)
		},
		onSuccess: writeAccount,
		onSettled: reconcileAccount,
		successMessage: () => i18n.global.t('user.settings.updateEmailCancelSuccess'),
	})
}

export function useCancelEmailUpdateMutation() {
	return useMutation(cancelEmailUpdateMutationOptions())
}

export function resendEmailConfirmationMutationOptions() {
	return contextMutationOptions<void, AccountIdentity>({
		mutationFn: async () => {
			await userResendEmailConfirmation()
		},
		// A cooldown rejection means another device already moved the pending state on.
		onSettled: reconcileAccount,
		successMessage: () => i18n.global.t('user.settings.updateEmailResendSuccess'),
	})
}

export function useResendEmailConfirmationMutation() {
	return useMutation(resendEmailConfirmationMutationOptions())
}
