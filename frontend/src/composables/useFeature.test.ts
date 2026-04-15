import {describe, it, expect, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useFeature} from './useFeature'
import {useConfigStore} from '@/stores/config'

describe('useFeature', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('returns true when the feature is in the enabledProFeatures list', () => {
		const store = useConfigStore()
		store.enabledProFeatures = ['admin_panel']
		expect(useFeature('admin_panel').value).toBe(true)
	})

	it('returns false for features not present in the list', () => {
		const store = useConfigStore()
		store.enabledProFeatures = ['admin_panel']
		expect(useFeature('time_tracking').value).toBe(false)
	})

	it('returns false when the list is empty (free mode)', () => {
		const store = useConfigStore()
		store.enabledProFeatures = []
		expect(useFeature('admin_panel').value).toBe(false)
	})

	it('reacts to store updates', () => {
		const store = useConfigStore()
		store.enabledProFeatures = []
		const enabled = useFeature('admin_panel')
		expect(enabled.value).toBe(false)
		store.enabledProFeatures = ['admin_panel']
		expect(enabled.value).toBe(true)
	})
})
