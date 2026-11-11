import {queryOptions} from '@tanstack/vue-query'
import {
	userIdentitiesList,
	userIdentitiesLink,
	userIdentitiesPreview,
	userIdentitiesUnlink,
	type LinkIdentityRequestWritable,
	type LinkedIdentity,
	type UserIdentitiesPreviewRequestWritable,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {i18n} from '@/i18n'

export const identityKeys = {
	list: ['identities'] as const,
}

// The server never paginates identities: one per provider at most.
export function identitiesQuery() {
	return queryOptions({
		queryKey: identityKeys.list,
		queryFn: async ({signal}): Promise<LinkedIdentity[]> => {
			const {data} = await userIdentitiesList({signal})
			return data.items ?? []
		},
	})
}

export function usePreviewIdentityMutation() {
	return useSecretMutation(contextMutationOptions({
		mutationFn: async (body: UserIdentitiesPreviewRequestWritable) => (await userIdentitiesPreview({body})).data,
		toastError: () => false,
	}))
}

export function linkIdentityMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: LinkIdentityRequestWritable) => (await userIdentitiesLink({body})).data,
		onSettled: (_input, client) => client.invalidateQueries({queryKey: identityKeys.list}),
		successMessage: () => i18n.global.t('user.settings.linkedAccounts.linkSuccess'),
		// The callback page shows the error next to the form instead.
		toastError: () => false,
	})
}

export function useLinkIdentityMutation() {
	return useSecretMutation(linkIdentityMutationOptions())
}

export type UnlinkIdentityInput = {
	id: number,
	password: string,
	totp_passcode?: string,
}

export function unlinkIdentityMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, ...body}: UnlinkIdentityInput) => {
			await userIdentitiesUnlink({
				path: {identity: id},
				body,
			})
		},
		onSuccess: (_data, {id}, client) => {
			client.setQueryData<LinkedIdentity[]>(identityKeys.list, current => current?.filter(i => i.id !== id))
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: identityKeys.list}),
		successMessage: () => i18n.global.t('user.settings.linkedAccounts.unlinkSuccess'),
	})
}

export function useUnlinkIdentityMutation() {
	return useSecretMutation(unlinkIdentityMutationOptions())
}
