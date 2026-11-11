import {describe, it, expect, vi, beforeEach} from 'vitest'

const {authStore, baseStore} = vi.hoisted(() => ({
	authStore: {
		currentAccount: null as Record<string, unknown> | null,
		refreshUserInfo: vi.fn(async () => {}),
	},
	baseStore: {appReady: Promise.resolve()},
}))

vi.mock('@/stores/auth', () => ({useAuthStore: () => authStore}))
vi.mock('@/stores/base', () => ({useBaseStore: () => baseStore}))

import {loadUserInfo} from './loadUserInfo'

const fullInfo = {id: 1, username: 'user1', is_admin: false, is_local_user: true}

describe('loadUserInfo', () => {
	beforeEach(() => {
		authStore.currentAccount = null
		authStore.refreshUserInfo.mockReset()
		baseStore.appReady = Promise.resolve()
	})

	it('fetches /user when only the JWT fallback is loaded', async () => {
		authStore.refreshUserInfo.mockImplementation(async () => {
			authStore.currentAccount = fullInfo
		})

		expect(await loadUserInfo()).toEqual(fullInfo)
		expect(authStore.refreshUserInfo).toHaveBeenCalledOnce()
	})

	it('returns a loaded user without refetching', async () => {
		authStore.currentAccount = fullInfo

		expect(await loadUserInfo()).toEqual(fullInfo)
		expect(authStore.refreshUserInfo).not.toHaveBeenCalled()
	})

	it('reads the user only once the app is ready', async () => {
		baseStore.appReady = new Promise(resolve => setTimeout(() => {
			authStore.currentAccount = fullInfo
			resolve()
		}, 0))

		expect(await loadUserInfo()).toEqual(fullInfo)
		expect(authStore.refreshUserInfo).not.toHaveBeenCalled()
	})
})
