import {createPinia, setActivePinia} from 'pinia'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'
import {labelKeys} from '@/client/queries/labels'
import {AUTH_TYPES, type AuthType} from '@/modelTypes/IUser'

const auth = vi.hoisted(() => ({
	token: null as string | null,
	post: vi.fn(),
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

vi.mock('@/helpers/fetcher', () => ({
	AuthenticatedHTTPFactory: () => fakeHttp(),
	HTTPFactory: () => fakeHttp(),
}))

function fakeHttp() {
	return {
		post: auth.post,
		get: vi.fn(),
		request: vi.fn(),
		interceptors: {
			request: {use: vi.fn()},
			response: {use: vi.fn()},
		},
	}
}

vi.mock('@/router', () => ({
	default: {push: vi.fn()},
}))

vi.mock('@/helpers/redirectToProvider', () => ({
	getRedirectUrlFromCurrentFrontendPath: vi.fn(),
	redirectToProvider: vi.fn(),
	redirectToProviderOnLogout: vi.fn(),
}))

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => ({disconnect: vi.fn()}),
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
		auth.post.mockReset()
	})

	it('removes the previous user query cache when entering a link share', async () => {
		const store = useAuthStore()
		store.setAuthenticated(true)
		store.setUser({id: 1, type: AUTH_TYPES.USER} as never, false)
		queryClient.setQueryData(labelKeys.all, [{id: 1, title: 'private'}])
		queryClient.setQueryData(['projects'], [{id: 1, title: 'private'}])
		const linkToken = jwt(AUTH_TYPES.LINK_SHARE, 2)
		auth.post.mockResolvedValue({data: {token: linkToken, project_id: 42}})

		await store.linkShareAuth({hash: 'share', password: 'secret'})

		expect(queryClient.getQueryData(labelKeys.all)).toBeUndefined()
		expect(queryClient.getQueryData(['projects'])).toBeUndefined()
	})

	it('keeps the link share label cache during token renewal', async () => {
		const store = useAuthStore()
		const linkToken = jwt(AUTH_TYPES.LINK_SHARE, 2)
		auth.post.mockResolvedValueOnce({data: {token: linkToken, project_id: 42}})
		await store.linkShareAuth({hash: 'share', password: 'secret'})
		queryClient.setQueryData(labelKeys.all, [{id: 2, title: 'shared'}])
		auth.post.mockResolvedValueOnce({data: {token: jwt(AUTH_TYPES.LINK_SHARE, 2)}})

		await store.renewToken()

		expect(queryClient.getQueryData(labelKeys.all)).toEqual([{id: 2, title: 'shared'}])
	})
})
