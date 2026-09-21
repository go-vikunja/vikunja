import {useMutation} from '@tanstack/vue-query'
import {authPasswordReset, authPasswordToken, userChangePassword, type PasswordResetWritable, type PasswordTokenRequestWritable, type UserChangePasswordRequestWritable} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {i18n} from '@/i18n'

export function usePasswordResetMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async (body: PasswordResetWritable) => (await authPasswordReset({body})).data,
		toastError: () => false,
	}))
}

export function useRequestPasswordResetMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async (body: PasswordTokenRequestWritable) => (await authPasswordToken({body})).data,
		toastError: () => false,
	}))
}

export function useChangePasswordMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async (body: UserChangePasswordRequestWritable) => (await userChangePassword({body})).data,
		successMessage: () => i18n.global.t('user.settings.passwordUpdateSuccess'),
	}))
}
