import {beforeEach, describe, expect, it, vi} from 'vitest'

import {checkAndSetApiUrl} from './checkAndSetApiUrl'

const mocks = vi.hoisted(() => ({
	clear: vi.fn(),
	configure: vi.fn(),
	update: vi.fn(),
}))

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({update: mocks.update}),
}))

vi.mock('@/client/http', () => ({
	configureApiClient: mocks.configure,
}))

vi.mock('@/client/queryClient', () => ({
	queryClient: {clear: mocks.clear},
}))

describe('checkAndSetApiUrl query lifecycle', () => {
	beforeEach(() => {
		window.API_URL = 'https://old.example.com/api/v2'
		localStorage.clear()
		mocks.clear.mockReset()
		mocks.configure.mockReset()
		mocks.update.mockReset()
	})

	it('reconfigures the client and clears cache after accepting a different server', async () => {
		mocks.update.mockResolvedValue(true)

		await expect(checkAndSetApiUrl('https://new.example.com/api/v2')).resolves.toBe('https://new.example.com/api/v2')

		expect(mocks.configure).toHaveBeenCalledOnce()
		expect(mocks.clear).toHaveBeenCalledOnce()
	})

	it('upgrades a saved v1 server URL before probing it', async () => {
		mocks.update.mockResolvedValue(true)
		await expect(checkAndSetApiUrl('https://new.example.com/root/api/v1')).resolves.toBe('https://new.example.com/root/api/v2')
		expect(localStorage.getItem('API_URL')).toBe('https://new.example.com/root/api/v2')
	})

	it('keeps pending queries when a v1 URL resolves to the same v2 server', async () => {
		window.API_URL = 'https://old.example.com/api/v1'
		mocks.update.mockResolvedValue(true)
		await checkAndSetApiUrl(window.API_URL)
		expect(window.API_URL).toBe('https://old.example.com/api/v2')
		expect(mocks.clear).not.toHaveBeenCalled()
		expect(mocks.configure).not.toHaveBeenCalled()
	})

	it('keeps the current client and cache when the server does not change', async () => {
		mocks.update.mockResolvedValue(true)

		await checkAndSetApiUrl('https://old.example.com/api/v2')

		expect(mocks.configure).not.toHaveBeenCalled()
		expect(mocks.clear).not.toHaveBeenCalled()
	})

	it('keeps the current client and cache when every candidate is rejected', async () => {
		mocks.update.mockRejectedValue(new Error('unreachable'))

		await expect(checkAndSetApiUrl('https://new.example.com')).rejects.toThrow('unreachable')

		expect(window.API_URL).toBe('https://old.example.com/api/v2')
		expect(mocks.configure).not.toHaveBeenCalled()
		expect(mocks.clear).not.toHaveBeenCalled()
	})
})
