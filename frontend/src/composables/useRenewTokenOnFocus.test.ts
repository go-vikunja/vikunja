import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {effectScope, nextTick, type EffectScope} from 'vue'

import {AUTH_TYPES} from '@/constants/auth'
import type {SessionClaims} from '@/stores/auth'
import {useRenewTokenOnFocus} from './useRenewTokenOnFocus'

const {authStore, routerPushMock} = await vi.hoisted(async () => {
	const {reactive} = await import('vue')
	return {
		authStore: reactive({
			authenticated: true,
			session: null as SessionClaims | null,
			renewToken: vi.fn(),
			checkAuth: vi.fn(),
		}),
		routerPushMock: vi.fn(),
	}
})

vi.mock('vue-router', () => ({
	useRouter: () => ({push: routerPushMock}),
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => authStore,
}))

async function flushMicrotasks() {
	for (let i = 0; i < 10; i++) {
		await Promise.resolve()
	}
}

const NOW_SECONDS = 1_800_000_000

function sessionExpiringIn(seconds: number): SessionClaims {
	return {
		id: 1,
		type: AUTH_TYPES.USER,
		exp: NOW_SECONDS + seconds,
	}
}

describe('useRenewTokenOnFocus', () => {
	let scope: EffectScope

	beforeEach(async () => {
		vi.useFakeTimers()
		vi.setSystemTime(NOW_SECONDS * 1000)
		authStore.authenticated = true
		authStore.session = null
		authStore.renewToken.mockReset().mockResolvedValue(undefined)
		authStore.checkAuth.mockReset().mockResolvedValue(undefined)
		routerPushMock.mockReset()
		scope = effectScope()
		scope.run(() => useRenewTokenOnFocus())
		await nextTick()
	})

	afterEach(() => {
		scope.stop()
		vi.useRealTimers()
	})

	it('renews once on load', () => {
		expect(authStore.renewToken).toHaveBeenCalledTimes(1)
	})

	it('schedules the proactive refresh 60s before the session expires', async () => {
		authStore.session = sessionExpiringIn(600)
		await nextTick()

		vi.advanceTimersByTime(539_000)
		expect(authStore.renewToken).toHaveBeenCalledTimes(1)

		vi.advanceTimersByTime(1_000)
		expect(authStore.renewToken).toHaveBeenCalledTimes(2)
	})

	it('reschedules when a refresh moves the session expiry', async () => {
		authStore.session = sessionExpiringIn(600)
		await nextTick()
		authStore.session = sessionExpiringIn(1200)
		await nextTick()

		vi.advanceTimersByTime(540_000)
		expect(authStore.renewToken).toHaveBeenCalledTimes(1)

		vi.advanceTimersByTime(600_000)
		expect(authStore.renewToken).toHaveBeenCalledTimes(2)
	})

	it('cancels the scheduled refresh on logout', async () => {
		authStore.session = sessionExpiringIn(600)
		await nextTick()
		authStore.authenticated = false
		await nextTick()

		vi.advanceTimersByTime(600_000)
		expect(authStore.renewToken).toHaveBeenCalledTimes(1)
	})

	it('renews on focus when the session expires within the buffer', async () => {
		authStore.session = sessionExpiringIn(30)

		window.dispatchEvent(new Event('focus'))
		await nextTick()

		expect(authStore.renewToken).toHaveBeenCalledTimes(2)
	})

	it('does not renew on focus when the session is far from expiry', async () => {
		authStore.session = sessionExpiringIn(3600)

		window.dispatchEvent(new Event('focus'))
		await nextTick()

		expect(authStore.renewToken).toHaveBeenCalledTimes(1)
	})

	it('sends the user to login when an expired session cannot be renewed on focus', async () => {
		authStore.session = sessionExpiringIn(-10)
		authStore.renewToken.mockRejectedValueOnce(new Error('refresh failed'))

		window.dispatchEvent(new Event('focus'))
		await flushMicrotasks()

		expect(authStore.renewToken).toHaveBeenCalledTimes(2)
		expect(authStore.checkAuth).toHaveBeenCalledTimes(1)
		expect(routerPushMock).toHaveBeenCalledWith({name: 'user.login'})
	})
})
