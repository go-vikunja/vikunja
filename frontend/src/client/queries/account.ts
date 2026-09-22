import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'
import type {DeepReadonly} from 'vue'
import {
	userShow,
	userUpdateSettings,
	userTimezones,
	userGetAvatarProvider,
} from '@/client/generated'
import type {
	UserGeneralSettings,
	UserGeneralSettingsWritable,
	UserInfoBody,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {contextMutationOptions} from './contextMutation'
import {AUTH_TYPES, type AuthType} from '@/constants/auth'
import {invalidateAvatarQueries} from './avatars'
import {
	defaultFrontendSettings,
	isNavigableUrl,
	type ExtraSettingsLink,
	type FrontendSettings,
	type UserSettings,
} from '@/helpers/userSettings'
import {getBrowserLanguage, i18n, setLanguage, type SupportedLocale} from '@/i18n'
import {error} from '@/message'

export const accountKeys = {
	current: ['account', 'current'] as const,
	user: (id: number, type: AuthType = AUTH_TYPES.USER) => ['account', 'current', id, type] as const,
	timezones: ['account', 'timezones'] as const,
}

export type AccountIdentity = {
	id: number
	type: AuthType
}

export type UserInfoResponse = UserInfoBody &
	Required<Pick<UserInfoBody, 'id' | 'username' | 'name' | 'is_admin' | 'is_local_user'>>

export function normalizeUserInfo(user: UserInfoBody): UserInfoResponse {
	return {
		...user,
		id: user.id ?? 0,
		username: user.username ?? '',
		name: user.name ?? '',
		is_admin: user.is_admin ?? false,
		is_local_user: user.is_local_user ?? false,
	}
}

export function normalizeUserSettings(data: UserGeneralSettings = {}): UserSettings {
	const frontend = typeof data.frontend_settings === 'object' && data.frontend_settings !== null
		? data.frontend_settings as Partial<FrontendSettings> : {}
	return {
		name: data.name ?? '',
		email_reminders_enabled: data.email_reminders_enabled ?? true,
		discoverable_by_name: data.discoverable_by_name ?? false,
		discoverable_by_email: data.discoverable_by_email ?? false,
		overdue_tasks_reminders_enabled: data.overdue_tasks_reminders_enabled ?? true,
		overdue_tasks_reminders_time: data.overdue_tasks_reminders_time ?? '09:00',
		default_project_id: data.default_project_id ?? 0,
		week_start: data.week_start ?? 0,
		timezone: data.timezone ?? '',
		language: (data.language || getBrowserLanguage()) as SupportedLocale,
		extra_settings_links: Object.fromEntries(
			Object.entries(data.extra_settings_links ?? {}).filter((entry): entry is [string, ExtraSettingsLink] => {
				const value = entry[1]
				return !!value
					&& typeof value === 'object'
					&& 'text' in value
					&& typeof value.text === 'string'
					&& 'url' in value
					&& typeof value.url === 'string'
					&& isNavigableUrl(value.url)
			}),
		),
		frontend_settings: {
			...defaultFrontendSettings(),
			...frontend,
			quick_add_default_reminders: (frontend.quick_add_default_reminders ?? []).map(reminder => ({...reminder})),
		},
	}
}

// The edit form mutates its copy, so it must not share nested objects with the read path.
export function createUserSettingsDraft(settings: DeepReadonly<UserSettings>): UserSettings {
	return {
		...settings,
		extra_settings_links: {...settings.extra_settings_links},
		frontend_settings: {
			...settings.frontend_settings,
			quick_add_default_reminders: settings.frontend_settings.quick_add_default_reminders
				.map(reminder => ({...reminder})),
		},
	}
}

export function currentUserQuery(id: number, type: AuthType = AUTH_TYPES.USER) {
	return queryOptions({
		queryKey: accountKeys.user(id, type),
		queryFn: async ({signal}) => normalizeUserInfo((await userShow({signal})).data),
	})
}

// checkAuth has to see a fresh /user before the first navigation decides on it.
export function refreshCurrentUser(id: number, type: AuthType = AUTH_TYPES.USER) {
	return queryClient.fetchQuery({
		...currentUserQuery(id, type),
		staleTime: 0,
		retry: false,
	})
}

export function timezonesQuery() {
	return queryOptions({
		queryKey: accountKeys.timezones,
		queryFn: async ({signal}) => (await userTimezones({signal})).data ?? [],
		staleTime: Infinity,
	})
}

function applySettingsUpdate(
	settings: UserGeneralSettingsWritable,
	{id, type}: AccountIdentity,
	client: QueryClient,
) {
	const key = accountKeys.user(id, type)
	const previous = client.getQueryData<UserInfoResponse>(key)
	client.setQueryData<UserInfoResponse>(key, current => current ? {
		...current,
		name: settings.name ?? current.name,
		settings: {
			...current.settings,
			...settings,
		},
	} : current)
	if (settings.language) setLanguage(settings.language as SupportedLocale).catch(error)
	if (previous?.username && previous.name !== settings.name) {
		const {username} = previous
		void userGetAvatarProvider().then(({data}) => {
			if (data.avatar_provider === 'initials') invalidateAvatarQueries(username)
		}).catch(() => {})
	}
}

// The account query is mounted for every user session, so an invalidation actually refetches.
export function reconcileAccount({id, type}: AccountIdentity, client: QueryClient) {
	return client.fetchQuery({
		...currentUserQuery(id, type),
		staleTime: 0,
		retry: false,
	})
}

export function updateSettingsMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({settings}: AccountIdentity & {
			settings: UserGeneralSettingsWritable
			showMessage?: boolean
		}) => {
			await userUpdateSettings({body: settings})
			return settings
		},
		onSuccess: (settings, input, client) => applySettingsUpdate(settings, input, client),
		onSettled: reconcileAccount,
		successMessage: (_data, {showMessage}) => showMessage === false
			? undefined
			: i18n.global.t('user.settings.general.savedSuccess'),
	})
}

export function useUpdateSettingsMutation() {
	return useMutation(updateSettingsMutationOptions())
}

// PUT /user/settings/general is a full replace, so one frontend flag can only be stored together
// with every other setting the server currently holds.
export function updateFrontendSettingsMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, type, frontendSettings}: AccountIdentity & {
			frontendSettings: Partial<FrontendSettings>
		}) => {
			const current = queryClient.getQueryData<UserInfoResponse>(accountKeys.user(id, type))
			if (!current) {
				throw new Error('Cannot store a frontend setting before the account is loaded')
			}
			const stored = normalizeUserSettings(current.settings)
			const settings: UserGeneralSettingsWritable = {
				...stored,
				frontend_settings: {
					...stored.frontend_settings,
					...frontendSettings,
				},
			}
			await userUpdateSettings({body: settings})
			return settings
		},
		onSuccess: (settings, input, client) => applySettingsUpdate(settings, input, client),
		onSettled: reconcileAccount,
	})
}

export function useUpdateFrontendSettingsMutation() {
	return useMutation(updateFrontendSettingsMutationOptions())
}
