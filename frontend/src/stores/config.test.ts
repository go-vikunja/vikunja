const sdk = vi.hoisted(() => ({info: vi.fn()}))
vi.mock('@/client/generated', () => sdk)

import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {computed} from 'vue'

import {useConfigStore} from './config'

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
})


describe('public configuration transport', () => {
 it('keeps wire fields and normalizes missing collections', async () => {
  setActivePinia(createPinia())
  window.API_URL = 'https://example.test/api/v1/'
  sdk.info.mockResolvedValue({data: {version: 'v2', enabled_pro_features: ['admin_panel']}})
  const store = useConfigStore()
  await store.update()
  expect(sdk.info).toHaveBeenCalledWith(expect.objectContaining({baseUrl: 'https://example.test/api/v2'}))
  expect(store.enabled_pro_features).toEqual(['admin_panel'])
  expect(store.available_migrators).toEqual([])
  expect(store.auth.openid_connect.providers).toEqual([])
 })
})
