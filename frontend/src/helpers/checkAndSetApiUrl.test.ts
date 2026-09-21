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
		window.API_URL = 'https://old.example.com'
		localStorage.clear()
		mocks.clear.mockReset()
		mocks.configure.mockReset()
		mocks.update.mockReset()
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
