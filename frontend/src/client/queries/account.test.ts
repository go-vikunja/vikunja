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
import {error} from '@/message'

const sdk = vi.hoisted(() => ({
	patchUserSettingsRead: vi.fn(),
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

const PATCHED = {
	data: {message: 'The settings were updated successfully.'},
	response: {status: 200},
}

function patchFailure(status: number, problem: unknown = {status}) {
	return {
		error: problem,
		response: {status},
	}
}

describe('account settings mutations', () => {
	beforeEach(() => {
		vi.clearAllMocks()
		queryClient.clear()
		tokenIdentity.current = {id: 1, type: 1}
		sdk.patchUserSettingsRead.mockResolvedValue(PATCHED)
		sdk.userShow.mockResolvedValue({data: SERVER_ACCOUNT})
		sdk.userGetAvatarProvider.mockResolvedValue({data: {avatar_provider: 'initials'}})
	})

	it('sends only the edited settings as a merge patch', async () => {
		const client = new QueryClient()

		await client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {
					name: 'New',
					timezone: undefined,
					frontend_settings: {},
				},
				showMessage: false,
			})

		const [options] = sdk.patchUserSettingsRead.mock.calls[0]
		expect(options.body).toStrictEqual({name: 'New'})
		expect(options.headers).toEqual({'Content-Type': 'application/merge-patch+json'})
		expect(sdk.userShow).toHaveBeenCalledOnce()
	})

	it('applies the edits to the cached account while preserving profile facts', async () => {
		const client = new QueryClient()
		client.setQueryData(accountKeys.user(1), {
			id: 1,
			is_admin: true,
			username: 'ada',
			name: 'Old',
			settings: {
				name: 'Old',
				frontend_settings: {sidebar_width: 376},
			},
		})
		let merged: UserInfoResponse | undefined
		// The reconcile refetch runs in onSettled, after the write-through this pins.
		sdk.userShow.mockImplementation(async () => {
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
			settings: {
				name: 'New',
				frontend_settings: {sidebar_width: 376},
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
				edits: {name: 'New'},
				showMessage: false,
			})

		expect(client.getQueryData(accountKeys.user(1))).toEqual(SERVER_ACCOUNT)
	})

	it('treats a patch that changes nothing as saved', async () => {
		const client = new QueryClient()
		client.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Old',
			settings: {name: 'Old'},
		})
		sdk.patchUserSettingsRead.mockResolvedValue(patchFailure(304, {}))
		sdk.userShow.mockRejectedValue({status: 503})

		await client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {week_start: 1},
				showMessage: false,
			})

		expect(client.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings?.week_start).toBe(1)
		expect(error).not.toHaveBeenCalled()
	})

	it('leaves cached settings intact when the server rejects an update', async () => {
		const client = new QueryClient()
		const account = {
			id: 1,
			settings: {name: 'Original'},
		}
		client.setQueryData(accountKeys.user(1), account)
		sdk.patchUserSettingsRead.mockResolvedValue(patchFailure(403))
		sdk.userShow.mockResolvedValue({data: account})

		await expect(client.getMutationCache()
			.build(client, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'Rejected'},
			})).rejects.toEqual({status: 403})

		expect(client.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings).toEqual({name: 'Original'})
		expect(error).toHaveBeenCalledExactlyOnceWith({status: 403})
	})

	it.each([
		['id', {id: 2, type: 1}],
		['type', {id: 1, type: 2}],
	])('does not write when the input %s differs from the token identity', async (_label, identity) => {
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

		expect(sdk.patchUserSettingsRead).not.toHaveBeenCalled()
		expect(error).not.toHaveBeenCalled()
	})

	it('patches a single frontend setting and shows it while saving', async () => {
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Old',
			settings: {
				name: 'Old',
				frontend_settings: {
					sidebar_width: 280,
					comment_sort_order: 'desc',
				},
			},
		})
		let shownWhileSaving: unknown
		sdk.patchUserSettingsRead.mockImplementationOnce(async () => {
			shownWhileSaving = queryClient.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings?.frontend_settings
			return PATCHED
		})

		await queryClient.getMutationCache()
			.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {sidebar_width: 400},
			})

		expect(sdk.patchUserSettingsRead.mock.calls[0][0].body).toEqual({frontend_settings: {sidebar_width: 400}})
		expect(shownWhileSaving).toMatchObject({
			sidebar_width: 400,
			comment_sort_order: 'desc',
		})
	})

	it('does not write a frontend setting for another identity', async () => {
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

		expect(sdk.patchUserSettingsRead).not.toHaveBeenCalled()
		expect(queryClient.getQueryData(accountKeys.user(1))).toEqual(account)
	})

	it('rolls back a frontend setting the server rejects', async () => {
		const account = {
			id: 1,
			settings: {frontend_settings: {sidebar_width: 280}},
		}
		queryClient.setQueryData(accountKeys.user(1), account)
		sdk.patchUserSettingsRead.mockResolvedValue(patchFailure(503))
		sdk.userShow.mockRejectedValue({status: 503})

		await expect(queryClient.getMutationCache()
			.build(queryClient, updateFrontendSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				frontendSettings: {sidebar_width: 400},
			})).rejects.toEqual({status: 503})

		expect(queryClient.getQueryData(accountKeys.user(1))).toEqual(account)
		expect(error).toHaveBeenCalledExactlyOnceWith({status: 503})
	})

	it('keeps a successful write when the reconcile re-read fails', async () => {
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Old',
			settings: {name: 'Old'},
		})
		sdk.userShow.mockRejectedValue({status: 503})

		await queryClient.getMutationCache()
			.build(queryClient, updateSettingsMutationOptions())
			.execute({
				id: 1,
				type: 1,
				edits: {name: 'Newer'},
				showMessage: false,
			})

		expect(queryClient.getQueryState(accountKeys.user(1))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings?.name).toBe('Newer')
		expect(error).not.toHaveBeenCalled()
	})

	it('sends an overlapping frontend setting only after the pending one', async () => {
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			name: 'Keep',
			settings: {
				name: 'Keep',
				frontend_settings: {sidebar_width: 280},
			},
		})
		let releaseFirst = () => {}
		sdk.patchUserSettingsRead.mockImplementationOnce(() => new Promise(resolve => {
			releaseFirst = () => resolve(PATCHED)
		}))
		sdk.userShow.mockImplementation(async () => ({
			data: {
				id: 1,
				name: 'Keep',
				settings: queryClient.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings,
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

		await vi.waitFor(() => expect(sdk.patchUserSettingsRead).toHaveBeenCalledTimes(1))
		const pending = queryClient.getQueryData<UserInfoResponse>(accountKeys.user(1))?.settings?.frontend_settings
		expect(pending).toMatchObject({
			comment_sort_order: 'desc',
			sidebar_width: 400,
		})
		releaseFirst()
		await Promise.all([first, second])

		expect(sdk.patchUserSettingsRead.mock.calls.map(([options]) => options.body)).toEqual([
			{frontend_settings: {comment_sort_order: 'desc'}},
			{frontend_settings: {sidebar_width: 400}},
		])
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
