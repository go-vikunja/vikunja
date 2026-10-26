import {beforeEach, describe, expect, it, vi} from 'vitest'
import {reactive} from 'vue'
import {createPinia, setActivePinia} from 'pinia'

import {accountKeys} from '@/client/queries/account'
import {queryClient} from '@/client/queryClient'
import {useAuthStore} from '@/stores/auth'
import {useProjectChromeBackground} from '@/composables/useProjectChromeBackground'

const baseStoreState = reactive({background: '' as string | null})

vi.mock('@/stores/base', () => ({
	useBaseStore: () => baseStoreState,
}))

function setExtendProjectBackgroundToChrome(enabled: boolean) {
	const authStore = useAuthStore()

	queryClient.setQueryData(accountKeys.user(0), {
		settings: {
			...authStore.settings,
			frontend_settings: {
				...authStore.settings.frontend_settings,
				extend_project_background_to_chrome: enabled,
			},
		},
	})
}

describe('useProjectChromeBackground', () => {
	beforeEach(() => {
		queryClient.clear()
		setActivePinia(createPinia())
		useAuthStore()
		baseStoreState.background = null
		document.body.classList.remove('has-project-chrome-background')
	})

	it('enables the chrome background only when the setting is on and a project background exists', () => {
		setExtendProjectBackgroundToChrome(true)
		baseStoreState.background = 'https://example.com/bg.jpg'

		const {enabled, chromeBackgroundStyle} = useProjectChromeBackground()

		expect(enabled.value).toBe(true)
		expect(chromeBackgroundStyle.value?.backgroundImage).toContain('https://example.com/bg.jpg')
		expect(document.body.classList.contains('has-project-chrome-background')).toBe(true)
	})

	it('stays disabled when the setting is off', () => {
		setExtendProjectBackgroundToChrome(false)
		baseStoreState.background = 'https://example.com/bg.jpg'

		const {enabled} = useProjectChromeBackground()

		expect(enabled.value).toBe(false)
		expect(document.body.classList.contains('has-project-chrome-background')).toBe(false)
	})
})
