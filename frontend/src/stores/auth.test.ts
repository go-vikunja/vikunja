import {beforeEach, describe, expect, it, vi} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {useAuthStore} from './auth'
import {useConfigStore} from './config'
import * as fetcher from '@/helpers/fetcher'

vi.mock('@/helpers/fetcher', () => {
	const mockHttp = {
		post: vi.fn(() => Promise.resolve({ data: { token: 'fake-token' } })),
		get: vi.fn(() => Promise.resolve({ data: {} })),
		interceptors: {
			request: { use: vi.fn() },
			response: { use: vi.fn() },
		},
	}
	return {
		HTTPFactory: () => mockHttp,
		AuthenticatedHTTPFactory: () => mockHttp,
	}
})

describe('auth store', () => {
	let httpMock

	beforeEach(() => {
		setActivePinia(createPinia())
		httpMock = fetcher.HTTPFactory()
		vi.clearAllMocks()
	})

	it('login action should call post with the full url', async () => {
		const configStore = useConfigStore()
		configStore.setApiUrl('http://localhost:3456/api/v1')

		const authStore = useAuthStore()
		await authStore.login({ username: 'test', password: 'password' })

		expect(httpMock.post).toHaveBeenCalledWith(
			'http://localhost:3456/api/v1/login',
			expect.any(Object)
		)
	})

	it('register action should call post with the full url', async () => {
		const configStore = useConfigStore()
		configStore.setApiUrl('http://localhost:3456/api/v1')

		const authStore = useAuthStore()
		await authStore.register({ username: 'test', password: 'password' })

		expect(httpMock.post).toHaveBeenCalledWith(
			'http://localhost:3456/api/v1/register',
			expect.any(Object)
		)
	})
})
