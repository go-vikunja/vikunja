import {beforeEach, describe, expect, it} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'

import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'
import {useProjectChromeBackground} from '@/composables/useProjectChromeBackground'

describe('useProjectChromeBackground', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		document.body.classList.remove('has-project-chrome-background')
	})

	it('enables the chrome background only when the setting is on and a project background exists', () => {
		const authStore = useAuthStore()
		const baseStore = useBaseStore()

		authStore.settings.frontend_settings.extend_project_background_to_chrome = true
		baseStore.background = 'https://example.com/bg.jpg'

		const {enabled, chromeBackgroundStyle} = useProjectChromeBackground()

		expect(enabled.value).toBe(true)
		expect(chromeBackgroundStyle.value?.backgroundImage).toContain('https://example.com/bg.jpg')
		expect(document.body.classList.contains('has-project-chrome-background')).toBe(true)
	})

	it('stays disabled when the setting is off', () => {
		const authStore = useAuthStore()
		const baseStore = useBaseStore()

		baseStore.background = 'https://example.com/bg.jpg'

		const {enabled} = useProjectChromeBackground()

		expect(enabled.value).toBe(false)
		expect(document.body.classList.contains('has-project-chrome-background')).toBe(false)
	})
})
