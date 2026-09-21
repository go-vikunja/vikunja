vi.mock('@/client/generated', () => ({
	authLinkShare: sdk.authLinkShare,
	tokenRenew: sdk.tokenRenew,
}))
import {createPinia, setActivePinia} from 'pinia'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'
import {labelKeys} from '@/client/queries/labels'
import {AUTH_TYPES, type AuthType} from '@/constants/auth'

const auth = vi.hoisted(() => ({
	token: null as string | null,
}))

const sdk = vi.hoisted(() => ({
	authLinkShare: vi.fn(),
	tokenRenew: vi.fn(),
}))

vi.mock('@/helpers/auth', () => ({
	getToken: () => auth.token,
	refreshToken: vi.fn(),
	removeToken: vi.fn(() => {
		auth.token = null
	}),
	saveToken: vi.fn((token: string) => {
		auth.token = token
	}),
}))

vi.mock('@/router', () => ({
	default: {push: vi.fn()},
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	getRedirectUrlFromCurrentFrontendPath: vi.fn(),
	redirectToProvider: vi.fn(),
	redirectToProviderOnLogout: vi.fn(),
}))

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => ({
		disconnect: vi.fn(),
		closeStaleConnection: vi.fn(),
	}),
}))

import {useAuthStore} from './auth'

function jwt(type: AuthType, id: number): string {
	const payload = btoa(JSON.stringify({
		id,
		type,
		exp: Math.floor(Date.now() / 1000) + 3600,
	}))
	return `header.${payload}.signature`
}

describe('link share auth query lifecycle', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		queryClient.clear()
		auth.token = jwt(AUTH_TYPES.USER, 1)
		sdk.authLinkShare.mockReset()
		sdk.tokenRenew.mockReset()
	})

	it('removes the previous user query cache when entering a link share', async () => {
		const store = useAuthStore()
		store.setAuthenticated(true)
		store.setSession({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: 0,
		})
		store.setUser({id: 1}, false)
		queryClient.setQueryData(labelKeys.all, [{id: 1, title: 'private'}])
		queryClient.setQueryData(['projects'], [{id: 1, title: 'private'}])
		const linkToken = jwt(AUTH_TYPES.LINK_SHARE, 2)
		sdk.authLinkShare.mockResolvedValue({data: {token: linkToken, project_id: 42}})

		await store.linkShareAuth({hash: 'share', password: 'secret'})

		expect(queryClient.getQueryData(labelKeys.all)).toBeUndefined()
		expect(queryClient.getQueryData(['projects'])).toBeUndefined()
	})

	it('keeps the link share label cache during token renewal', async () => {
		const store = useAuthStore()
		const linkToken = jwt(AUTH_TYPES.LINK_SHARE, 2)
		sdk.authLinkShare.mockResolvedValueOnce({data: {token: linkToken, project_id: 42}})
		await store.linkShareAuth({hash: 'share', password: 'secret'})
		queryClient.setQueryData(labelKeys.all, [{id: 2, title: 'shared'}])
		sdk.tokenRenew.mockResolvedValueOnce({data: {token: jwt(AUTH_TYPES.LINK_SHARE, 2)}})

		await store.renewToken()

		expect(queryClient.getQueryData(labelKeys.all)).toEqual([{id: 2, title: 'shared'}])
	})
})
