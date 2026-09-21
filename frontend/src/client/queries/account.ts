import {queryOptions, useMutation} from '@tanstack/vue-query'
import {userShow, userUpdateSettings, userTimezones, userGetAvatarProvider, type UserGeneralSettingsWritable, type UserInfoBody} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {contextMutationOptions} from './contextMutation'
import {invalidateAvatarCache} from '@/helpers/user'
import {i18n, setLanguage, type SupportedLocale} from '@/i18n'

export const accountKeys = {
	current: ['account', 'current'] as const,
	user: (id: number, type = 1) => ['account', 'current', id, type] as const,
	timezones: ['account', 'timezones'] as const,
}

export function currentUserQuery(id = 0, type = 1) {
	return queryOptions({queryKey: accountKeys.user(id, type), queryFn: async ({signal}) => (await userShow({signal})).data})
}

export function refreshCurrentUser(id = 0) {
	return queryClient.fetchQuery({...currentUserQuery(id), staleTime: 0})
}

export function timezonesQuery() {
	return queryOptions({queryKey: accountKeys.timezones, queryFn: async ({signal}) => (await userTimezones({signal})).data ?? [], staleTime: Infinity})
}

export function updateSettingsMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({settings}: {settings: UserGeneralSettingsWritable; showMessage?: boolean}) => {
			await userUpdateSettings({body: settings})
			return settings
		},
		onSuccess: (settings, _input, client) => {
			const previous = client.getQueriesData<UserInfoBody>({queryKey: accountKeys.current}).map(([, data]) => data).find(Boolean)
			client.setQueriesData<UserInfoBody>({queryKey: accountKeys.current}, current => current ? {
				...current,
				name: settings.name ?? current.name,
				settings: {...current.settings, ...settings},
			} : current)
			if (settings.language) void setLanguage(settings.language as SupportedLocale)
			if (previous && previous.name !== settings.name) {
				void userGetAvatarProvider().then(({data}) => {
					if (data.avatar_provider === 'initials') invalidateAvatarCache(previous)
				}).catch(() => {})
			}
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: accountKeys.current}),
		successMessage: (_data, {showMessage}) => showMessage === false ? undefined : i18n.global.t('user.settings.general.savedSuccess'),
	})
}

export function useUpdateSettingsMutation() {
	return useMutation(updateSettingsMutationOptions())
}
