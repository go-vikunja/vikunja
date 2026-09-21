import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	accountKeys,
	normalizeUserSettings,
	updateFrontendSettingsMutationOptions,
	updateSettingsMutationOptions,
	type UserInfoResponse,
} from './account'
import {queryClient} from '@/client/queryClient'
import {invalidateAvatarQueries} from './avatars'

const sdk = vi.hoisted(() => ({
	userUpdateSettings: vi.fn(),
	userShow: vi.fn(),
	userGetAvatarProvider: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))
vi.mock('./avatars', () => ({invalidateAvatarQueries: vi.fn()}))

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
		let merged: unknown
		// The reconcile refetch runs in onSettled, after the write-through this pins.
		sdk.userShow.mockImplementation(async () => {
			merged = client.getQueryData(accountKeys.user(1))
			return {data: SERVER_ACCOUNT}
		})
		const settings = {
			name: 'New',
			frontend_settings: {sidebar_width: 280},
		}

		await client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				settings,
				showMessage: false,
			})

		expect(sdk.userUpdateSettings).toHaveBeenCalledWith({body: settings})
		expect(merged).toEqual({
			id: 1,
			is_admin: true,
			username: 'ada',
			name: 'New',
			settings: {
				name: 'New',
				frontend_settings: {sidebar_width: 280},
			},
		})
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
				settings: {name: 'New'},
				showMessage: false,
			})

		expect(sdk.userShow).toHaveBeenCalledOnce()
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
				settings: {name: 'Rejected'},
			})).rejects.toEqual({status: 403})

		expect(client.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings).toEqual({name: 'Original'})
	})

	it('writes one frontend setting back together with every stored setting', async () => {
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Keep',
			settings: {
				name: 'Keep',
				timezone: 'Europe/Berlin',
				default_project_id: 7,
				frontend_settings: {comment_sort_order: 'asc'},
			},
		})

		await queryClient.getMutationCache()
			.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {comment_sort_order: 'desc'},
			})

		const body = sdk.userUpdateSettings.mock.calls[0][0].body
		expect(body.name).toBe('Keep')
		expect(body.timezone).toBe('Europe/Berlin')
		expect(body.default_project_id).toBe(7)
		expect(body.frontend_settings.comment_sort_order).toBe('desc')
	})

	it('refuses to write a frontend setting before the account is loaded', async () => {
		await expect(queryClient.getMutationCache()
			.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {comment_sort_order: 'desc'},
			})).rejects.toThrow('Cannot store a frontend setting before the account is loaded')

		expect(sdk.userUpdateSettings).not.toHaveBeenCalled()
	})

	it('keeps a successful write when the reconcile re-read fails', async () => {
		const client = new QueryClient()
		client.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Old',
			settings: {name: 'Old'},
		})
		sdk.userShow.mockRejectedValue({status: 503})

		await expect(client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				settings: {name: 'New'},
				showMessage: false,
			})).resolves.toEqual({name: 'New'})

		expect(sdk.userShow).toHaveBeenCalledOnce()
		expect(client.getQueryState(accountKeys.user(1))?.isInvalidated).toBe(true)
		expect(client.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings).toEqual({name: 'New'})
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
