import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	accountKeys,
	createUserSettingsDraft,
	diffUserSettings,
	normalizeUserSettings,
	updateFrontendSettingsMutationOptions,
	updateSettingsMutationOptions,
	withSettingsEdits,
	type UserInfoResponse,
} from './account'
import type {UserSettings} from '@/helpers/userSettings'
import {queryClient} from '@/client/queryClient'
import {invalidateAvatarQueries} from './avatars'
import {error} from '@/message'

const sdk = vi.hoisted(() => ({
	userUpdateSettings: vi.fn(),
	userShow: vi.fn(),
	userGetAvatarProvider: vi.fn(),
}))
const tokenIdentity = vi.hoisted(() => ({current: {id: 1, type: 1}}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))
vi.mock('./avatars', () => ({invalidateAvatarQueries: vi.fn()}))
vi.mock('@/helpers/auth', async importOriginal => ({
	...await importOriginal<typeof import('@/helpers/auth')>(),
	getTokenIdentity: () => tokenIdentity.current,
}))

const SERVER_ACCOUNT = {
	id: 1,
	username: 'user1',
	name: 'New',
	is_admin: true,
	is_local_user: true,
	settings: {
		name: 'New',
		frontend_settings: {sidebar_width: 280},
	},
}

describe('account settings mutations', () => {
	beforeEach(() => {
		vi.clearAllMocks()
		queryClient.clear()
		tokenIdentity.current = {id: 1, type: 1}
		sdk.userUpdateSettings.mockResolvedValue({data: {}})
		sdk.userShow.mockResolvedValue({data: SERVER_ACCOUNT})
		sdk.userGetAvatarProvider.mockResolvedValue({data: {avatar_provider: 'initials'}})
	})

	it('merges settings into the current account while preserving profile facts', async () => {
		const client = new QueryClient()
		const previous = {
			id: 1,
			is_admin: true,
			username: 'ada',
			name: 'Old',
			settings: {name: 'Old'},
		}
		client.setQueryData(accountKeys.user(1), previous)
		let merged: UserInfoResponse | undefined
		// The reconcile refetch runs in onSettled, after the write-through this pins.
		sdk.userShow
			.mockResolvedValueOnce({data: previous})
			.mockImplementation(async () => {
				merged = client.getQueryData(accountKeys.user(1))
				return {data: SERVER_ACCOUNT}
			})

		await client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'New'},
				showMessage: false,
			})

		expect(merged).toMatchObject({
			id: 1,
			is_admin: true,
			username: 'ada',
			name: 'New',
		})
		expect(merged?.settings?.name).toBe('New')
		await vi.waitFor(() => expect(invalidateAvatarQueries).toHaveBeenCalledTimes(1))
		expect(invalidateAvatarQueries).toHaveBeenCalledWith('ada')
	})

	it('reconciles the account with what the server stored', async () => {
		const client = new QueryClient()
		client.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Old',
			settings: {name: 'Old'},
		})

		await client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'New'},
				showMessage: false,
			})

		expect(sdk.userShow).toHaveBeenCalledTimes(2)
		expect(client.getQueryData(accountKeys.user(1))).toEqual(SERVER_ACCOUNT)
	})

	it('leaves cached settings intact when the server rejects an update', async () => {
		const client = new QueryClient()
		const account = {
			id: 1,
			settings: {name: 'Original'},
		}
		client.setQueryData(accountKeys.user(1), account)
		sdk.userUpdateSettings.mockRejectedValue({status: 403})
		sdk.userShow.mockResolvedValue({data: account})

		await expect(client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'Rejected'},
			})).rejects.toEqual({status: 403})

		expect(client.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings).toEqual({name: 'Original'})
	})

	it('applies the edits to the settings the server holds, not the cached ones', async () => {
		const client = new QueryClient()
		client.setQueryData(accountKeys.user(1), {
			id: 1,
			settings: {
				name: 'Old',
				frontend_settings: {sidebar_width: 376},
			},
		})
		sdk.userShow.mockResolvedValue({
			data: {
				id: 1,
				settings: {
					name: 'Old',
					frontend_settings: {sidebar_width: 422},
				},
			},
		})

		await client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'New'},
				showMessage: false,
			})

		const body = sdk.userUpdateSettings.mock.calls[0][0].body
		expect(body.name).toBe('New')
		expect(body.frontend_settings.sidebar_width).toBe(422)
	})

	it('leaves the language out of the write when asked to', async () => {
		const client = new QueryClient()
		await client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {language: 'de-DE'},
				omitLanguage: true,
				showMessage: false,
			})

		expect(sdk.userUpdateSettings.mock.calls[0][0].body).not.toHaveProperty('language')
	})

	it('does not write when the current settings cannot be read', async () => {
		sdk.userShow.mockRejectedValue({status: 503})
		const client = new QueryClient()

		await expect(client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'New'},
				showMessage: false,
			})).rejects.toEqual({status: 503})

		expect(sdk.userUpdateSettings).not.toHaveBeenCalled()
		expect(error).toHaveBeenCalledExactlyOnceWith({status: 503})
	})

	it('does not write when the signed-in identity changes during the read', async () => {
		sdk.userShow.mockImplementationOnce(async () => {
			tokenIdentity.current = {id: 2, type: 1}
			return {data: SERVER_ACCOUNT}
		})
		const client = new QueryClient()

		await expect(client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'New'},
				showMessage: false,
			})).rejects.toMatchObject({name: 'AbortError'})

		expect(sdk.userUpdateSettings).not.toHaveBeenCalled()
		expect(error).not.toHaveBeenCalled()
	})

	it.each([
		['id', {id: 2, type: 1}],
		['type', {id: 1, type: 2}],
	])('neither reads nor writes when the input %s differs from the token identity', async (_label, identity) => {
		tokenIdentity.current = identity
		const client = new QueryClient()

		await expect(client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'New'},
				showMessage: false,
			})).rejects.toMatchObject({name: 'AbortError'})

		expect(sdk.userShow).not.toHaveBeenCalled()
		expect(sdk.userUpdateSettings).not.toHaveBeenCalled()
		expect(error).not.toHaveBeenCalled()
	})

	it('writes a frontend setting on top of the settings the server holds, not the cached ones', async () => {
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Old',
			settings: {
				name: 'Old',
				frontend_settings: {sidebar_width: 280},
			},
		})
		sdk.userShow.mockResolvedValueOnce({
			data: {
				id: 1,
				name: 'Other tab',
				settings: {
					name: 'Other tab',
					timezone: 'Europe/Berlin',
					frontend_settings: {
						sidebar_width: 280,
						comment_sort_order: 'desc',
					},
				},
			},
		})
		let shownWhileSaving: unknown
		sdk.userUpdateSettings.mockImplementationOnce(async () => {
			shownWhileSaving = queryClient.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings?.frontend_settings
			return {data: {}}
		})

		await queryClient.getMutationCache()
			.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {sidebar_width: 400},
			})

		const body = sdk.userUpdateSettings.mock.calls[0][0].body
		expect(body).toMatchObject({
			name: 'Other tab',
			timezone: 'Europe/Berlin',
			frontend_settings: {
				sidebar_width: 400,
				comment_sort_order: 'desc',
			},
		})
		expect(shownWhileSaving).toMatchObject({sidebar_width: 400})
	})

	it('neither reads nor writes a frontend setting for another identity', async () => {
		tokenIdentity.current = {id: 2, type: 1}
		const account = {
			id: 1,
			settings: {frontend_settings: {sidebar_width: 280}},
		}
		queryClient.setQueryData(accountKeys.user(1), account)

		await expect(queryClient.getMutationCache()
			.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {sidebar_width: 400},
			})).rejects.toMatchObject({name: 'AbortError'})

		expect(sdk.userShow).not.toHaveBeenCalled()
		expect(sdk.userUpdateSettings).not.toHaveBeenCalled()
		expect(queryClient.getQueryData(accountKeys.user(1))).toEqual(account)
	})

	it('rolls back a frontend setting without writing when the stored settings cannot be read', async () => {
		const account = {
			id: 1,
			settings: {frontend_settings: {sidebar_width: 280}},
		}
		queryClient.setQueryData(accountKeys.user(1), account)
		sdk.userShow.mockRejectedValue({status: 503})

		await expect(queryClient.getMutationCache()
			.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {sidebar_width: 400},
			})).rejects.toEqual({status: 503})

		expect(sdk.userUpdateSettings).not.toHaveBeenCalled()
		expect(queryClient.getQueryData(accountKeys.user(1))).toEqual(account)
		expect(error).toHaveBeenCalledExactlyOnceWith({status: 503})
	})

	it('keeps a successful write when the reconcile re-read fails', async () => {
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Old',
			settings: {name: 'Old'},
		})
		sdk.userShow
			.mockResolvedValueOnce({data: SERVER_ACCOUNT})
			.mockRejectedValue({status: 503})

		const saved = await queryClient.getMutationCache()
			.build(queryClient, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'Newer'},
				showMessage: false,
			})

		expect(saved.name).toBe('Newer')
		expect(sdk.userShow).toHaveBeenCalledTimes(2)
		expect(queryClient.getQueryState(accountKeys.user(1))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings?.name).toBe('Newer')
		expect(error).not.toHaveBeenCalled()
	})

	it('builds an overlapping frontend setting write on top of the pending one', async () => {
		let stored: Record<string, unknown> = {
			name: 'Keep',
			frontend_settings: {sidebar_width: 280},
		}
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Keep',
			settings: stored,
		})
		let releaseFirst = () => {}
		sdk.userUpdateSettings
			.mockImplementationOnce(({body}) => new Promise(resolve => {
				releaseFirst = () => {
					stored = body
					resolve({data: {}})
				}
			}))
			.mockImplementationOnce(async ({body}) => {
				stored = body
				return {data: {}}
			})
		sdk.userShow.mockImplementation(async () => ({
			data: {
				id: 1,
				name: 'Keep',
				settings: stored,
			},
		}))
		const cache = queryClient.getMutationCache()

		const first = cache.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {comment_sort_order: 'desc'},
			})
		const second = cache.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {sidebar_width: 400},
			})

		await vi.waitFor(() => expect(sdk.userUpdateSettings).toHaveBeenCalledTimes(1))
		const pending = queryClient.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings?.frontend_settings
		expect(pending).toMatchObject({
			comment_sort_order: 'desc',
			sidebar_width: 400,
		})
		releaseFirst()
		await Promise.all([first, second])

		expect(sdk.userUpdateSettings).toHaveBeenCalledTimes(2)
		expect(sdk.userUpdateSettings.mock.calls[1][0].body.frontend_settings).toMatchObject({
			comment_sort_order: 'desc',
			sidebar_width: 400,
		})
		expect(stored.frontend_settings).toMatchObject({
			comment_sort_order: 'desc',
			sidebar_width: 400,
		})
	})
})

describe('normalizeUserSettings', () => {
	it.each([
		['a string', 'nope'],
		['an object', {relative_to: 'due_date'}],
		['null', null],
	])('falls back to no default reminders when the stored value is %s', (_label, value) => {
		const settings = normalizeUserSettings({
			frontend_settings: {quick_add_default_reminders: value},
		})

		expect(settings.frontend_settings.quick_add_default_reminders).toEqual([])
	})
})

describe('withSettingsEdits', () => {
	const initial = normalizeUserSettings({
		name: 'Old',
		week_start: 0,
		frontend_settings: {
			sidebar_width: 376,
			quick_add_default_reminders: [{relative_period: 60}],
		},
	})

	function save(edit: (settings: UserSettings) => void, remote: Parameters<typeof normalizeUserSettings>[0]) {
		const edited = createUserSettingsDraft(initial)
		edit(edited)
		return withSettingsEdits(normalizeUserSettings(remote), diffUserSettings(initial, edited))
	}

	it('lets the local edit win over a remote change of the same setting', () => {
		const saved = save(s => {
			s.frontend_settings.sidebar_width = 500
		}, {frontend_settings: {sidebar_width: 422}})

		expect(saved.frontend_settings.sidebar_width).toBe(500)
	})

	it('keeps the remote value of a setting edited and then reverted', () => {
		const saved = save(s => {
			s.frontend_settings.sidebar_width = 500
			s.frontend_settings.sidebar_width = 376
		}, {frontend_settings: {sidebar_width: 422}})

		expect(saved.frontend_settings.sidebar_width).toBe(422)
	})

	it('replaces the default reminders with the edited list', () => {
		const saved = save(s => {
			s.frontend_settings.quick_add_default_reminders = [{relative_period: -3600}]
		}, {frontend_settings: {quick_add_default_reminders: [{relative_period: 120}]}})

		expect(saved.frontend_settings.quick_add_default_reminders).toEqual([{relative_period: -3600}])
	})
})
