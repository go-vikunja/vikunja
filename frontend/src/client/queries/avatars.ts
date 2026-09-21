import {queryOptions, useMutation} from '@tanstack/vue-query'
import {avatarGet, userAvatarUpload, userGetAvatarProvider, userSetAvatarProvider} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {queryClient} from '@/client/queryClient'
import {i18n} from '@/i18n'

export const avatarKeys = {
	user: (username: string) => ['avatars', username] as const,
	image: (username: string, size: number) => ['avatars', username, size] as const,
	provider: ['avatar-provider'] as const,
}

export function invalidateAvatarQueries(username: string) {
	void queryClient.invalidateQueries({queryKey: avatarKeys.user(username)})
}

export function avatarQuery(username: string, size: number) {
	return queryOptions({
		queryKey: avatarKeys.image(username, size),
		queryFn: async ({signal}) => {
			const {data} = await avatarGet({path: {username}, query: {size}, parseAs: 'blob', signal})
			if (!(data instanceof Blob)) throw new Error('Avatar response was not an image')
			return data
		},
		staleTime: Infinity,
		retry: false,
	})
}

export function avatarProviderQuery() {
	return queryOptions({
		queryKey: avatarKeys.provider,
		queryFn: async ({signal}) => (await userGetAvatarProvider({signal})).data,
	})
}

export function updateAvatarProviderMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({provider}: {username: string; provider: string}) => (await userSetAvatarProvider({body: {avatar_provider: provider}})).data,
		onSettled: ({username}, client) => Promise.all([
			client.invalidateQueries({queryKey: avatarKeys.provider}),
			client.invalidateQueries({queryKey: avatarKeys.user(username)}),
		]),
		successMessage: () => i18n.global.t('user.settings.avatar.statusUpdateSuccess'),
	})
}

export function useUpdateAvatarProviderMutation() {
	return useMutation(updateAvatarProviderMutationOptions())
}

export function uploadAvatarMutationOptions() {
	return contextMutationOptions({
		// The upload fails without a file name on the multipart part.
		mutationFn: async ({blob}: {username: string; blob: Blob}) => (await userAvatarUpload({body: {avatar: new File([blob], 'avatar.png', {type: blob.type})}})).data,
		onSettled: ({username}, client) => Promise.all([
			client.invalidateQueries({queryKey: avatarKeys.provider}),
			client.invalidateQueries({queryKey: avatarKeys.user(username)}),
		]),
		successMessage: () => i18n.global.t('user.settings.avatar.setSuccess'),
	})
}

export function useUploadAvatarMutation() {
	return useMutation(uploadAvatarMutationOptions())
}
