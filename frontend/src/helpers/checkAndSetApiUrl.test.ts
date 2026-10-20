import {beforeEach, describe, expect, it, vi} from 'vitest'

import {checkAndSetApiUrl} from './checkAndSetApiUrl'

const mocks = vi.hoisted(() => ({
	clear: vi.fn(),
	configure: vi.fn(),
	update: vi.fn(),
	setAuthenticated: vi.fn(),
	setSession: vi.fn(),
}))

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({update: mocks.update}),
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({setAuthenticated: mocks.setAuthenticated, setSession: mocks.setSession}),
}))

vi.mock('@/client/http', () => ({
	configureApiClient: mocks.configure,
}))

vi.mock('@/client/queryClient', () => ({
	queryClient: {clear: mocks.clear},
}))

describe('checkAndSetApiUrl query lifecycle', () => {
	beforeEach(() => {
		window.API_URL = 'https://old.example.com'
		localStorage.clear()
		mocks.clear.mockReset()
		mocks.configure.mockReset()
		mocks.update.mockReset()
		mocks.setAuthenticated.mockReset()
		mocks.setSession.mockReset()
	})

	it('reconfigures the client and clears cache after accepting a different server', async () => {
		mocks.update.mockResolvedValue(true)

		await expect(checkAndSetApiUrl('https://new.example.com')).resolves.toBe('https://new.example.com')

		expect(localStorage.getItem('API_URL')).toBe('https://new.example.com')
		expect(mocks.configure).toHaveBeenCalledOnce()
		expect(mocks.clear).toHaveBeenCalledOnce()
	})

	it.each([
		['https://new.example.com/root/api/v1', 'https://new.example.com/root'],
		['https://new.example.com/root/api/v2/', 'https://new.example.com/root'],
		['https://new.example.com/root/', 'https://new.example.com/root'],
	])('stores %s without its version suffix', async (input, stored) => {
		mocks.update.mockResolvedValue(true)

		await expect(checkAndSetApiUrl(input)).resolves.toBe(stored)

		expect(window.API_URL).toBe(stored)
		expect(localStorage.getItem('API_URL')).toBe(stored)
	})

	it('resolves the same-origin default to the frontend origin', async () => {
		window.API_URL = ''
		mocks.update.mockResolvedValue(true)

		await expect(checkAndSetApiUrl('')).resolves.toBe('http://localhost:3000')

		expect(localStorage.getItem('API_URL')).toBe('http://localhost:3000')
	})

	it('keeps pending queries when a legacy URL resolves to the same server', async () => {
		window.API_URL = 'https://old.example.com/api/v1'
		mocks.update.mockResolvedValue(true)

		await checkAndSetApiUrl(window.API_URL)

		expect(window.API_URL).toBe('https://old.example.com')
		expect(mocks.clear).not.toHaveBeenCalled()
		expect(mocks.configure).not.toHaveBeenCalled()
	})

	it('keeps the current client and cache when the server does not change', async () => {
		mocks.update.mockResolvedValue(true)

		await checkAndSetApiUrl('https://old.example.com/')

		expect(mocks.configure).not.toHaveBeenCalled()
		expect(mocks.clear).not.toHaveBeenCalled()
	})

	it('falls back to the default API port', async () => {
		const probed: string[] = []
		mocks.update.mockImplementation(async () => {
			probed.push(window.API_URL)
			if (window.API_URL !== 'https://new.example.com:3456/root') throw new Error('unreachable')
			return true
		})

		await expect(checkAndSetApiUrl('https://new.example.com/root/api/v1')).resolves.toBe('https://new.example.com:3456/root')

		expect(probed).toEqual([
			'https://new.example.com/root',
			'https://new.example.com:3456/root',
		])
	})

	it('keeps the current client and cache when every candidate is rejected', async () => {
		const probed: string[] = []
		mocks.update.mockImplementation(async () => {
			probed.push(window.API_URL)
			throw new Error('unreachable')
		})

		await expect(checkAndSetApiUrl('https://new.example.com')).rejects.toThrow('unreachable')

		expect(probed).toEqual([
			'https://new.example.com',
			'https://new.example.com:3456',
		])
		expect(window.API_URL).toBe('https://old.example.com')
		expect(localStorage.getItem('API_URL')).toBeNull()
		expect(mocks.configure).not.toHaveBeenCalled()
		expect(mocks.clear).not.toHaveBeenCalled()
	})
})

describe('checkAndSetApiUrl stored credentials', () => {
	beforeEach(() => {
		window.API_URL = ''
		localStorage.clear()
		localStorage.setItem('token', 'jwt')
		localStorage.setItem('desktopOAuthRefreshToken', 'refresh')
		mocks.update.mockReset()
		mocks.update.mockResolvedValue(true)
		mocks.setAuthenticated.mockReset()
		mocks.setSession.mockReset()
	})

	it('drops tokens and auth state when switching to another server', async () => {
		localStorage.setItem('API_URL', 'https://old.example.com')

		await checkAndSetApiUrl('https://attacker.example.com')

		expect(localStorage.getItem('token')).toBeNull()
		expect(localStorage.getItem('desktopOAuthRefreshToken')).toBeNull()
		expect(mocks.setAuthenticated).toHaveBeenCalledWith(false)
		expect(mocks.setSession).toHaveBeenCalledWith(null)
	})

	it.each([
		['the same server', 'https://old.example.com', 'https://old.example.com/'],
		['a legacy stored url', 'https://old.example.com/api/v1', 'https://old.example.com'],
		['a same-origin stored url', '', ''],
	])('keeps tokens when booting against %s', async (_, stored, input) => {
		localStorage.setItem('API_URL', stored)

		await checkAndSetApiUrl(input)

		expect(localStorage.getItem('token')).toBe('jwt')
		expect(localStorage.getItem('desktopOAuthRefreshToken')).toBe('refresh')
		expect(mocks.setAuthenticated).not.toHaveBeenCalled()
	})

	it('keeps tokens when no server was stored yet', async () => {
		await checkAndSetApiUrl('https://new.example.com')

		expect(localStorage.getItem('token')).toBe('jwt')
	})

	it('keeps tokens when the new server is rejected', async () => {
		localStorage.setItem('API_URL', 'https://old.example.com')
		mocks.update.mockRejectedValue(new Error('unreachable'))

		await expect(checkAndSetApiUrl('https://attacker.example.com')).rejects.toThrow('unreachable')

		expect(localStorage.getItem('token')).toBe('jwt')
	})
})
