import {createPinia, setActivePinia} from 'pinia'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {AUTH_TYPES} from '@/modelTypes/IUser'
import {ENTITLEMENT} from '@/constants/entitlements'
import UserModel from '@/models/user'

const http = vi.hoisted(() => ({
	get: vi.fn(),
}))

vi.mock('@/helpers/auth', () => ({
	getToken: () => 'token',
	refreshToken: vi.fn(),
	removeToken: vi.fn(),
	saveToken: vi.fn(),
}))

vi.mock('@/helpers/fetcher', () => ({
	AuthenticatedHTTPFactory: () => ({get: http.get, post: vi.fn(), interceptors: {request: {use: vi.fn()}, response: {use: vi.fn()}}}),
	HTTPFactory: () => ({get: http.get, post: vi.fn(), interceptors: {request: {use: vi.fn()}, response: {use: vi.fn()}}}),
	apiV2Url: (path: string) => `/api/v2/${path}`,
}))

vi.mock('@/router', () => ({
	default: {push: vi.fn()},
}))

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => ({disconnect: vi.fn()}),
}))

import {useAuthStore} from './auth'

describe('auth store entitlements', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		http.get.mockReset()
	})

	it('treats a JWT-only user as not loaded and unentitled', () => {
		const store = useAuthStore()
		store.setUser(new UserModel({id: 1, type: AUTH_TYPES.USER}), false)

		expect(store.entitlementsLoaded).toBe(false)
		expect(store.hasEntitlement(ENTITLEMENT.TIME_TRACKING)).toBe(false)
		expect(store.limit(ENTITLEMENT.MAX_PROJECTS)).toBeNull()
		expect(store.isAtLimit(ENTITLEMENT.MAX_PROJECTS)).toBe(false)
	})

	it('resolves flags, limits and usage from the user', () => {
		const store = useAuthStore()
		store.setUser(new UserModel({
			id: 1,
			type: AUTH_TYPES.USER,
			entitlements: {admin_panel: 0, audit_logs: 0, time_tracking: 1, team_creation: 0, max_projects: 3},
			usage: {max_projects: 3, max_storage_bytes: 42},
		}), false)

		expect(store.entitlementsLoaded).toBe(true)
		expect(store.hasEntitlement(ENTITLEMENT.TIME_TRACKING)).toBe(true)
		expect(store.hasEntitlement(ENTITLEMENT.TEAM_CREATION)).toBe(false)
		expect(store.limit(ENTITLEMENT.MAX_PROJECTS)).toBe(3)
		expect(store.limit(ENTITLEMENT.MAX_STORAGE_BYTES)).toBeNull()
		expect(store.usage(ENTITLEMENT.MAX_STORAGE_BYTES)).toBe(42)
		expect(store.isAtLimit(ENTITLEMENT.MAX_PROJECTS)).toBe(true)

		store.adjustUsage(ENTITLEMENT.MAX_PROJECTS, -1)
		expect(store.usage(ENTITLEMENT.MAX_PROJECTS)).toBe(2)
		expect(store.isAtLimit(ENTITLEMENT.MAX_PROJECTS)).toBe(false)
	})

	it('reads entitlements from the v2 user endpoint', async () => {
		const store = useAuthStore()
		http.get.mockResolvedValue({
			data: {
				id: 1,
				username: 'user1',
				settings: {},
				entitlements: {admin_panel: 0, audit_logs: 0, time_tracking: 0, team_creation: 1},
				usage: {max_projects: 5, max_storage_bytes: 200},
			},
		})

		await store.refreshUserInfo()

		expect(http.get).toHaveBeenCalledWith('/api/v2/user')
		expect(store.entitlementsLoaded).toBe(true)
		expect(store.hasEntitlement(ENTITLEMENT.TEAM_CREATION)).toBe(true)
		expect(store.usage(ENTITLEMENT.MAX_PROJECTS)).toBe(5)
	})
})
