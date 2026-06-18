import {watch} from 'vue'
import {createSharedComposable, tryOnMounted} from '@vueuse/core'
import {storeToRefs} from 'pinia'

import {useTimeTrackingStore} from '@/stores/timeTracking'

const TRACKING_FAVICON = '/images/icons/favicon-tracking-32x32.png'

function getFaviconLink(): HTMLLinkElement | null {
	return document.querySelector<HTMLLinkElement>('link[rel="icon"]')
}

// Swaps in a favicon with a small red dot in the lower left corner while a timer
// is running, so an active time tracking session is visible even when the tab
// isn't focused.
export const useTimeTrackingFavicon = createSharedComposable(() => {
	const {hasActiveTimer} = storeToRefs(useTimeTrackingStore())

	const originalHref = getFaviconLink()?.getAttribute('href') ?? '/favicon.ico'

	function update(active: boolean) {
		const link = getFaviconLink()
		if (link === null) {
			return
		}
		link.href = active ? TRACKING_FAVICON : originalHref
	}

	watch(hasActiveTimer, update, {flush: 'post'})
	tryOnMounted(() => update(hasActiveTimer.value))
})
