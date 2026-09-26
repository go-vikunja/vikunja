const sdk = vi.hoisted(() => ({info: vi.fn()}))
vi.mock('@/client/generated', () => sdk)

import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {computed} from 'vue'

import {useConfigStore} from './config'
import {InvalidApiUrlProvidedError} from '@/helpers/apiUrl'

describe('config store', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	describe('isProFeatureEnabled', () => {
		it('returns true when the feature is in the enabled_pro_features list', () => {
			const store = useConfigStore()
			store.enabled_pro_features = ['admin_panel']
			expect(store.isProFeatureEnabled('admin_panel')).toBe(true)
		})

		it('returns false for features not present in the list', () => {
			const store = useConfigStore()
			store.enabled_pro_features = ['admin_panel']
			expect(store.isProFeatureEnabled('time_tracking')).toBe(false)
		})

		it('returns false when the list is empty (free mode)', () => {
			const store = useConfigStore()
			store.enabled_pro_features = []
			expect(store.isProFeatureEnabled('admin_panel')).toBe(false)
		})

		it('reacts to store updates when wrapped in computed', () => {
			const store = useConfigStore()
			store.enabled_pro_features = []
			const enabled = computed(() => store.isProFeatureEnabled('admin_panel'))
			expect(enabled.value).toBe(false)
			store.enabled_pro_features = ['admin_panel']
			expect(enabled.value).toBe(true)
		})
	})

	describe('public configuration transport', () => {
		beforeEach(() => {
			window.API_URL = 'https://example.test/api/v1/'
		})

		it('keeps wire fields', async () => {
			sdk.info.mockResolvedValue({
				data: {
					version: 'v2',
					enabled_pro_features: ['admin_panel'],
				},
			})
			const store = useConfigStore()

			await store.update()

			expect(sdk.info).toHaveBeenCalledWith(expect.objectContaining({baseUrl: 'https://example.test/api/v2'}))
			expect(store.version).toBe('v2')
			expect(store.enabled_pro_features).toEqual(['admin_panel'])
		})

		it('normalizes null collections to empty arrays', async () => {
			sdk.info.mockResolvedValue({
				data: {
					version: 'v2',
					enabled_pro_features: null,
					available_migrators: null,
					enabled_background_providers: null,
					auth: {
						openid_connect: {
							enabled: true,
							providers: null,
						},
					},
				},
			})
			const store = useConfigStore()

			await store.update()

			expect(store.version).toBe('v2')
			expect(store.enabled_pro_features).toEqual([])
			expect(store.available_migrators).toEqual([])
			expect(store.enabled_background_providers).toEqual([])
			expect(store.auth.openid_connect.enabled).toBe(true)
			expect(store.auth.openid_connect.providers).toEqual([])
		})

		it('fills missing provider fields with defaults', async () => {
			sdk.info.mockResolvedValue({
				data: {
					version: 'v2',
					auth: {
						openid_connect: {
							enabled: true,
							providers: [{key: 'x'}],
						},
					},
				},
			})
			const store = useConfigStore()

			await store.update()

			expect(store.auth.openid_connect.providers).toEqual([
				{
					key: 'x',
					name: '',
					auth_url: '',
					client_id: '',
					logout_url: '',
					scope: 'openid email profile',
				},
			])
		})

		it('rejects when the response carries no version', async () => {
			sdk.info.mockResolvedValue({data: {}})
			const store = useConfigStore()

			await expect(store.update()).rejects.toBeInstanceOf(InvalidApiUrlProvidedError)
		})

		it('normalizes a non-Error rejection into InvalidApiUrlProvidedError', async () => {
			sdk.info.mockRejectedValue({
				code: 4001,
				message: 'not found',
			})
			const store = useConfigStore()

			await expect(store.update()).rejects.toBeInstanceOf(InvalidApiUrlProvidedError)
		})

		it('passes an Error rejection through unwrapped', async () => {
			const failure = new TypeError('Failed to fetch')
			sdk.info.mockRejectedValue(failure)
			const store = useConfigStore()

			await expect(store.update()).rejects.toBe(failure)
		})

		it('blanks legal and provider urls that are not http', async () => {
			sdk.info.mockResolvedValue({
				data: {
					version: 'v2',
					legal: {
						imprint_url: 'javascript:alert(1)',
						privacy_policy_url: 'https://example.test/privacy',
					},
					auth: {
						openid_connect: {
							enabled: true,
							providers: [{
								key: 'x',
								auth_url: 'javascript:alert(2)',
								logout_url: 'http://example.test/logout',
							}],
						},
					},
				},
			})
			const store = useConfigStore()

			await store.update()

			expect(store.legal.imprint_url).toBe('')
			expect(store.legal.privacy_policy_url).toBe('https://example.test/privacy')
			expect(store.auth.openid_connect.providers[0].auth_url).toBe('')
			expect(store.auth.openid_connect.providers[0].logout_url).toBe('http://example.test/logout')
		})

		it('does not apply a response for an API that is no longer configured', async () => {
			sdk.info.mockImplementation(async () => {
				window.API_URL = 'https://other.test/api/v1/'
				return {data: {version: 'v2'}}
			})
			const store = useConfigStore()

			await expect(store.update()).resolves.toBe(false)

			expect(store.version).toBe('')
		})
	})
})
