import {describe, it, expect, vi, beforeEach} from 'vitest'
import type {RouteLocation} from 'vue-router'

const {success, error} = vi.hoisted(() => ({
	success: vi.fn(),
	error: vi.fn(),
}))

vi.mock('@/message', () => ({success, error}))
vi.mock('@/i18n', () => ({
	i18n: {global: {t: (key: string) => key}},
}))

import router, {getAuthForRoute} from './index'

describe('settings route matching', () => {
	it('routes negative project ids to saved-filter settings', () => {
		expect(router.resolve('/projects/-2/settings/edit').name).toBe('filter.settings.edit')
		expect(router.resolve('/projects/-2/settings/delete').name).toBe('filter.settings.delete')
		expect(router.resolve('/projects/-10/settings/edit').name).toBe('filter.settings.edit')
	})

	it('routes non-negative project ids to project settings', () => {
		expect(router.resolve('/projects/2/settings/edit').name).toBe('project.settings.edit')
		expect(router.resolve('/projects/2/settings/delete').name).toBe('project.settings.delete')
	})

	it('does not route the Favorites pseudo-project to saved-filter settings', () => {
		expect(router.resolve('/projects/-1/settings/edit').name).toBe('bad-not-found')
		expect(router.resolve('/projects/-1/settings/delete').name).toBe('bad-not-found')
	})

	it('builds settings urls from named routes with projectId params', () => {
		expect(router.resolve({name: 'project.settings.edit', params: {projectId: 2}}).path)
			.toBe('/projects/2/settings/edit')
		expect(router.resolve({name: 'project.settings.delete', params: {projectId: 2}}).path)
			.toBe('/projects/2/settings/delete')
		expect(router.resolve({name: 'filter.settings.edit', params: {projectId: -10}}).path)
			.toBe('/projects/-10/settings/edit')
		expect(router.resolve({name: 'filter.settings.delete', params: {projectId: -10}}).path)
			.toBe('/projects/-10/settings/delete')
	})
})

function route(query: RouteLocation['query'] = {}) {
	return {
		name: 'home',
		hash: '',
		query,
		params: {},
		fullPath: '/',
	} as unknown as RouteLocation
}

function authStoreStub({pendingEmail = 'new@example.com', ...overrides}: {pendingEmail?: string} & Record<string, unknown> = {}) {
	const server = {pendingEmail}
	const store = {
		authUser: true,
		authLinkShare: false,
		// stale on purpose: this session was opened before the change was requested elsewhere
		info: {pendingEmail: ''},
		verifyEmail: vi.fn(async () => {
			server.pendingEmail = ''
			return true
		}),
		refreshUserInfo: vi.fn(async () => {
			store.info = {...server}
		}),
		...overrides,
	}
	return store
}

describe('getAuthForRoute email confirmation', () => {
	beforeEach(() => {
		success.mockReset()
		error.mockReset()
		localStorage.clear()
	})

	it('redeems the token from the query and lands on the email settings page', async () => {
		const authStore = authStoreStub()

		const result = await getAuthForRoute(route({userEmailConfirm: 'token-123'}), authStore)

		expect(authStore.verifyEmail).toHaveBeenCalledWith('token-123')
		expect(localStorage.getItem('emailConfirmToken')).toBeNull()
		expect(result).toEqual({name: 'user.settings.email-update'})
		expect(success).toHaveBeenCalledOnce()
	})

	it('refreshes before redeeming so a stale session still sees the pending change', async () => {
		const authStore = authStoreStub()

		await getAuthForRoute(route({userEmailConfirm: 'token-123'}), authStore)

		expect(authStore.refreshUserInfo).toHaveBeenCalledTimes(2)
		expect(success).toHaveBeenCalledOnce()
	})

	it('uses the first value when the token is repeated in the query', async () => {
		const authStore = authStoreStub()

		await getAuthForRoute(route({userEmailConfirm: ['token-123', 'token-456']}), authStore)

		expect(authStore.verifyEmail).toHaveBeenCalledWith('token-123')
	})

	it('does not claim success when nothing was pending', async () => {
		const authStore = authStoreStub({pendingEmail: '', info: {pendingEmail: 'stale@example.com'}})

		const result = await getAuthForRoute(route({userEmailConfirm: 'token-123'}), authStore)

		expect(result).toEqual({name: 'home'})
		expect(success).not.toHaveBeenCalled()
	})

	it('redirects home and reports the original error when verification fails', async () => {
		const cause = {response: {data: {code: 4021}}}
		const authStore = authStoreStub({
			verifyEmail: vi.fn().mockRejectedValue({message: 'nope', cause} as unknown as Error),
		})

		const result = await getAuthForRoute(route({userEmailConfirm: 'token-123'}), authStore)

		expect(result).toEqual({name: 'home'})
		expect(error).toHaveBeenCalledWith(cause)
	})

	it('ignores an empty token', async () => {
		const authStore = authStoreStub()

		await getAuthForRoute(route({userEmailConfirm: ''}), authStore)

		expect(authStore.verifyEmail).not.toHaveBeenCalled()
	})
})
