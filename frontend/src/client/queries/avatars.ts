import {queryOptions, useMutation} from '@tanstack/vue-query'
import {avatarGet, userAvatarUpload, userGetAvatarProvider, userSetAvatarProvider} from '@/client/generated'
import {expectBlob} from './blobResponse'
import {contextMutationOptions} from './contextMutation'
import {queryClient} from '@/client/queryClient'
import {i18n} from '@/i18n'

// A blob: url for an svg inherits our origin and can script; a data: url cannot.
const SCRIPTABLE_MIME_TYPE = 'image/svg+xml'

function readAsDataUrl(blob: Blob) {
	return new Promise<string>((resolve, reject) => {
		const reader = new FileReader()
		reader.onload = () => {
			if (typeof reader.result === 'string') {
				resolve(reader.result)
				return
			}
			reject(new Error('Avatar could not be read as a data url'))
		}
		reader.onerror = () => reject(reader.error ?? new Error('Avatar could not be read as a data url'))
		reader.readAsDataURL(blob)
	})
}

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
			const blob = expectBlob(data, 'Avatar')
			const isScriptable = blob.type.split(';')[0].trim().toLowerCase() === SCRIPTABLE_MIME_TYPE
			// FileReader is absent in iOS Lockdown Mode and some webviews, fall back to a blob url there.
			if (!isScriptable || typeof FileReader === 'undefined') return blob
			return readAsDataUrl(blob)
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
		mutationFn: async ({provider}: {
			username: string,
			provider: string,
		}) => {
			const {data} = await userSetAvatarProvider({body: {avatar_provider: provider}})
			return data
		},
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
		mutationFn: async ({blob}: {
			username: string,
			blob: Blob,
		}) => {
			// The upload fails without a file name.
			const avatar = new File([blob], 'avatar.png', {type: blob.type})
			const {data} = await userAvatarUpload({body: {avatar}})
			return data
		},
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
