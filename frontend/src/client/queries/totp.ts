import {queryOptions, useMutation} from '@tanstack/vue-query'
import {totpGet, totpEnroll, totpEnable, totpDisable, totpQrcode, type Totp} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {i18n} from '@/i18n'

export const totpKeys = {current: ['totp'] as const, qr: ['totp-qr'] as const}

export function totpQuery() {
	return queryOptions({
		queryKey: totpKeys.current,
		queryFn: async ({signal}): Promise<Totp> => {
			try {
				return (await totpGet({signal})).data
			} catch (cause) {
				if ((cause as {code?: number})?.code === 1016) return {enabled: false}
				throw cause
			}
		},
	})
}

export function totpQrQuery() {
	return queryOptions({
		queryKey: totpKeys.qr,
		queryFn: async ({signal}) => (await totpQrcode({signal, parseAs: 'blob'})).data,
		gcTime: 0,
	})
}

export function useEnrollTotpMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async () => (await totpEnroll()).data,
		onSuccess: (data, _input, client) => {
			client.removeQueries({queryKey: totpKeys.qr})
			client.setQueryData<Totp>(totpKeys.current, current => current ? data : current)
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: totpKeys.current}),
	}))
}

export function useEnableTotpMutation() {
	return useMutation(contextMutationOptions({
		mutationFn: async (passcode: string) => (await totpEnable({body: {passcode}})).data,
		onSuccess: (_data, _input, client) => {
			client.setQueryData<Totp>(totpKeys.current, current => current ? {enabled: true} : current)
			client.removeQueries({queryKey: totpKeys.qr})
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: totpKeys.current}),
		successMessage: () => i18n.global.t('user.settings.totp.confirmSuccess'),
	}))
}

export function disableTotpMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (password: string) => (await totpDisable({body: {password}})).data,
		onSuccess: (_data, _input, client) => {
			client.setQueryData<Totp>(totpKeys.current, current => current ? {enabled: false} : current)
			client.removeQueries({queryKey: totpKeys.qr})
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: totpKeys.current}),
		successMessage: () => i18n.global.t('user.settings.totp.disableSuccess'),
	})
}

export function useDisableTotpMutation() {
	return useMutation(disableTotpMutationOptions())
}
