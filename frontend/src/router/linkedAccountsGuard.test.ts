import {describe, it, expect, vi, beforeEach} from 'vitest'

const {authStore, baseStore, configStore} = vi.hoisted(() => {
	const authStore = {
		authUser: true,
		authLinkShare: false,
		currentAccount: {id: 1, username: 'user1', is_local_user: true} as Record<string, unknown> | null,
		checkAuth: vi.fn(async () => {}),
		refreshUserInfo: vi.fn(async () => {}),
		verifyEmail: vi.fn(),
	}
	const baseStore = {appReady: Promise.resolve()}
	const configStore = {
		auth: {openid_connect: {enabled: true, providers: [{key: 'dex', name: 'Dex'}] as unknown[]}},
	}
	return {authStore, baseStore, configStore}
})

vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/i18n', () => ({i18n: {global: {t: (key: string) => key}}}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => authStore}))
vi.mock('@/stores/base', () => ({useBaseStore: () => baseStore}))
vi.mock('@/stores/config', () => ({useConfigStore: () => configStore}))
vi.mock('@/views/user/Settings.vue', () => ({default: {render: () => null}}))
vi.mock('@/views/user/settings/LinkedAccounts.vue', () => ({default: {render: () => null}}))
vi.mock('@/views/user/settings/General.vue', () => ({default: {render: () => null}}))

import router from './index'

async function navigateToLinkedAccounts() {
	await router.push('/')
	await router.push('/user/settings/linked-accounts')
	return router.currentRoute.value.name
}

describe('linked accounts route guard', () => {
	beforeEach(() => {
		authStore.currentAccount = {id: 1, username: 'user1', is_local_user: true}
		authStore.refreshUserInfo.mockReset()
		configStore.auth.openid_connect = {enabled: true, providers: [{key: 'dex', name: 'Dex'}]}
	})

	it('fetches /user when only the JWT fallback is loaded', async () => {
		authStore.currentAccount = null
		authStore.refreshUserInfo.mockImplementation(async () => {
			authStore.currentAccount = {id: 1, username: 'user1', is_local_user: true}
		})

		expect(await navigateToLinkedAccounts()).toBe('user.settings.linked-accounts')
		expect(authStore.refreshUserInfo).toHaveBeenCalledOnce()
	})

	it('sends non-local users to the general settings', async () => {
		authStore.currentAccount = {id: 1, username: 'user1', is_local_user: false}

		expect(await navigateToLinkedAccounts()).toBe('user.settings.general')
	})

	it('sends everyone to the general settings without a provider', async () => {
		configStore.auth.openid_connect = {enabled: true, providers: []}

		expect(await navigateToLinkedAccounts()).toBe('user.settings.general')
	})
})
