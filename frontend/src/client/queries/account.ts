import {queryOptions, mutationOptions, useMutation, type QueryClient} from '@tanstack/vue-query'
import type {DeepReadonly} from 'vue'
import isEqual from 'fast-deep-equal'
import {
	userShow,
	patchUserSettingsRead,
	userTimezones,
	userGetAvatarProvider,
} from '@/client/generated'
import type {
	JsonPatchOp,
	UserGeneralSettings,
	UserInfoBody,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {contextMutationOptions} from './contextMutation'
import type {ClientRequestContext} from '@/client/requestContext'
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
			quick_add_default_reminders: Array.isArray(frontend.quick_add_default_reminders)
				? frontend.quick_add_default_reminders.map(reminder => ({...reminder}))
				: [],
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

export type UserSettingsEdits = Partial<Omit<UserSettings, 'frontend_settings'>> & {
	frontend_settings?: Partial<FrontendSettings>
}

function changedEntries<T extends object>(initial: T, edited: T): Partial<T> {
	return Object.fromEntries(Object.entries(edited)
		.filter(([key, value]) => !isEqual(value, initial[key as keyof T]))) as Partial<T>
}

export function diffUserSettings(initial: UserSettings, edited: UserSettings): UserSettingsEdits {
	const {frontend_settings: initialFrontend, ...initialRest} = initial
	const {frontend_settings: editedFrontend, ...editedRest} = edited
	return {
		...changedEntries(initialRest, editedRest),
		frontend_settings: changedEntries(initialFrontend, editedFrontend),
	}
}

function withSettingsEdits(settings: DeepReadonly<UserSettings>, edits: UserSettingsEdits): UserSettings {
	const draft = createUserSettingsDraft(settings)
	return {
		...draft,
		...edits,
		frontend_settings: {
			...draft.frontend_settings,
			...edits.frontend_settings,
		},
	}
}

export function currentUserQuery(id: number, type: AuthType = AUTH_TYPES.USER) {
	return queryOptions({
		queryKey: accountKeys.user(id, type),
		queryFn: async ({signal}) => normalizeUserInfo((await userShow({signal})).data),
	})
}

function freshCurrentUserQuery(id: number, type: AuthType) {
	return {
		...currentUserQuery(id, type),
		staleTime: 0,
		retry: false,
	}
}

// checkAuth has to see a fresh /user before the first navigation decides on it.
export function refreshCurrentUser(id: number, type: AuthType = AUTH_TYPES.USER) {
	return queryClient.fetchQuery(freshCurrentUserQuery(id, type))
}

export function timezonesQuery() {
	return queryOptions({
		queryKey: accountKeys.timezones,
		queryFn: async ({signal}) => (await userTimezones({signal})).data ?? [],
		staleTime: Infinity,
	})
}

function withAccountSettingsEdits(account: UserInfoResponse, edits: UserSettingsEdits): UserInfoResponse {
	return {
		...account,
		name: edits.name ?? account.name,
		settings: {
			...account.settings,
			...withSettingsEdits(normalizeUserSettings(account.settings), edits),
		},
	}
}

function applySettingsUpdate(
	edits: UserSettingsEdits,
	{id, type}: AccountIdentity,
	client: QueryClient,
) {
	const key = accountKeys.user(id, type)
	const previous = client.getQueryData<UserInfoResponse>(key)
	client.setQueryData<UserInfoResponse>(key, current => current ? withAccountSettingsEdits(current, edits) : current)
	if (edits.language) setLanguage(edits.language).catch(error)
	if (previous?.username && edits.name !== undefined && previous.name !== edits.name) {
		const {username} = previous
		void userGetAvatarProvider().then(({data}) => {
			if (data.avatar_provider === 'initials') invalidateAvatarQueries(username)
		}).catch(() => {})
	}
}

// The write already landed; a failed re-read must not fail the mutation.
// The account query is mounted for every user session, so the invalidation actually refetches.
export function reconcileAccount({id, type}: AccountIdentity, client: QueryClient) {
	return client.fetchQuery({
		...freshCurrentUserQuery(id, type),
		meta: {handlesError: true},
	}).catch(() => client.invalidateQueries({queryKey: accountKeys.user(id, type)}))
}

// The server applies a PATCH as read, merge, write, so overlapping ones could revert each other.
const ACCOUNT_SETTINGS_SCOPE = {id: 'account-settings'}

function definedEntries<T extends object>(value: T): Partial<T> {
	return Object.fromEntries(Object.entries(value).filter(([, entry]) => entry !== undefined)) as Partial<T>
}

function settingsMergePatch({frontend_settings: frontend, ...rest}: UserSettingsEdits): UserSettingsEdits {
	const patch: UserSettingsEdits = definedEntries(rest)
	const frontendPatch = definedEntries(frontend ?? {})
	if (Object.keys(frontendPatch).length > 0) patch.frontend_settings = frontendPatch
	return patch
}

async function saveSettingsEdits(
	{id, type, edits}: AccountIdentity & {edits: UserSettingsEdits},
	{request}: {request: ClientRequestContext},
) {
	if (request.identity?.id !== id || request.identity.type !== type) {
		throw new DOMException('Account changed', 'AbortError')
	}
	const patch = settingsMergePatch(edits)
	const {error: patchError, response} = await patchUserSettingsRead({
		// Typed as JSON Patch only, which cannot add a key below a null frontend_settings.
		body: patch as unknown as JsonPatchOp[],
		headers: {'Content-Type': 'application/merge-patch+json'},
		throwOnError: false,
	})
	// A patch that changes nothing gets an empty 304.
	if (patchError !== undefined && response?.status !== 304) throw patchError
	return patch
}

export function updateSettingsMutationOptions() {
	return mutationOptions({
		...contextMutationOptions({
			mutationFn: (input: AccountIdentity & {
				edits: UserSettingsEdits
				showMessage?: boolean
			}, context) => saveSettingsEdits(input, context),
			onSuccess: (patch, input, client) => applySettingsUpdate(patch, input, client),
			onSettled: reconcileAccount,
			successMessage: (_data, {showMessage}) => showMessage === false
				? undefined
				: i18n.global.t('user.settings.general.savedSuccess'),
		}),
		scope: ACCOUNT_SETTINGS_SCOPE,
	})
}

export function useUpdateSettingsMutation() {
	return useMutation(updateSettingsMutationOptions())
}

export function updateFrontendSettingsMutationOptions() {
	return mutationOptions({
		...contextMutationOptions({
			mutationFn: ({id, type, frontendSettings}: AccountIdentity & {
				frontendSettings: Partial<FrontendSettings>
			}, context) => saveSettingsEdits({
				id,
				type,
				edits: {frontend_settings: frontendSettings},
			}, context),
			optimistic: {
				queryKeys: ({id, type}) => [accountKeys.user(id, type)],
				update: ({id, type, frontendSettings}, client) => {
					client.setQueryData<UserInfoResponse>(accountKeys.user(id, type), current => current
						? withAccountSettingsEdits(current, {frontend_settings: frontendSettings})
						: current)
				},
			},
			onSuccess: (patch, input, client) => applySettingsUpdate(patch, input, client),
			onSettled: reconcileAccount,
		}),
		scope: ACCOUNT_SETTINGS_SCOPE,
	})
}

export function useUpdateFrontendSettingsMutation() {
	return useMutation(updateFrontendSettingsMutationOptions())
}
