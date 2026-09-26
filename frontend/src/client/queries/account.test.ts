import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	accountKeys,
	updateFrontendSettingsMutationOptions,
	updateSettingsMutationOptions,
	type UserInfoResponse,
} from './account'
import {queryClient} from '@/client/queryClient'
import {invalidateAvatarCache} from '@/helpers/user'

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
vi.mock('@/helpers/user', () => ({invalidateAvatarCache: vi.fn()}))

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
			name: 'New',
			settings: {
				name: 'New',
				frontend_settings: {sidebar_width: 280},
			},
		})
		await vi.waitFor(() => expect(invalidateAvatarCache).toHaveBeenCalledTimes(1))
		expect(invalidateAvatarCache).toHaveBeenCalledWith(previous)
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
})
