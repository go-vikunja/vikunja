import {ref, watch, readonly} from 'vue'
import {useLocalStorage, useMediaQuery} from '@vueuse/core'
import {useRoute} from 'vue-router'

const BULMA_MOBILE_BREAKPOINT = 768

export function useMenuActive() {
	const isMobile = useMediaQuery(`(max-width: ${BULMA_MOBILE_BREAKPOINT}px)`)

	const desktopPreference = useLocalStorage(
		'menuActiveDesktopPreference',
		true,
		// If we have two tabs open we want to be able to have the menu open in one window
		// and closed in the other. The last changed value will be the new preference
		{listenToStorageChanges: false},
	)

	const menuActive = ref(false)
	const route = useRoute()

	// set to prefered value
	watch(isMobile, (current) => {
		menuActive.value = current
			// On mobile we don't show the menu in an expanded state
			// because that would hide the main content
			? false
			: desktopPreference.value
	}, {immediate: true})

	watch(menuActive, (current) => {
		if (!isMobile.value) {
			desktopPreference.value = current
		}
	})

	// Hide the menu on mobile when the route changes (e.g. when the user taps a menu item)
	watch(() => route.fullPath, () => isMobile.value && setMenuActive(false))

	function setMenuActive(newMenuActive: boolean) {
		menuActive.value = newMenuActive
	}

	function toggleMenu() {
		menuActive.value = menuActive.value = !menuActive.value
	}

	return {
		menuActive: readonly(menuActive),
		setMenuActive,
		toggleMenu,
	}
}
