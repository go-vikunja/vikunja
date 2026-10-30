import {computed, watch} from 'vue'
import {createSharedComposable, tryOnBeforeUnmount} from '@vueuse/core'

import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

const CHROME_BACKGROUND_CLASS = 'has-project-chrome-background'

export const useProjectChromeBackground = createSharedComposable(() => {
	const authStore = useAuthStore()
	const baseStore = useBaseStore()

	const enabled = computed(() =>
		authStore.settings.frontend_settings.extend_project_background_to_chrome === true &&
		Boolean(baseStore.background),
	)

	const chromeBackgroundStyle = computed(() => {
		if (!enabled.value || !baseStore.background) {
			return undefined
		}

		const brightness = authStore.settings.frontend_settings.background_brightness
		const filters = ['blur(24px)', 'saturate(0.85)']
		if (typeof brightness === 'number') {
			filters.push(`brightness(${brightness}%)`)
		}

		return {
			backgroundImage: `url(${baseStore.background})`,
			filter: filters.join(' '),
		}
	})

	function syncBodyClass(active: boolean) {
		document.body.classList.toggle(CHROME_BACKGROUND_CLASS, active)
	}

	watch(enabled, syncBodyClass, {immediate: true, flush: 'post'})
	tryOnBeforeUnmount(() => syncBodyClass(false))

	return {
		enabled,
		chromeBackgroundStyle,
	}
})
