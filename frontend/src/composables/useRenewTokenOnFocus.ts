import {computed, ref, watch} from 'vue'
import {useRouter} from 'vue-router'
import {useEventListener} from '@vueuse/core'

import {useAuthStore} from '@/stores/auth'
import {MILLISECONDS_A_SECOND} from '@/constants/date'

// Refresh the token 60 seconds before it expires to avoid API calls hitting 401.
const REFRESH_BUFFER_SECONDS = 60

export function useRenewTokenOnFocus() {
	const router = useRouter()
	const authStore = useAuthStore()

	const userInfo = computed(() => authStore.info)
	const authenticated = computed(() => authStore.authenticated)
	const refreshTimer = ref<ReturnType<typeof setTimeout> | null>(null)

	function clearRefreshTimer() {
		if (refreshTimer.value !== null) {
			clearTimeout(refreshTimer.value)
			refreshTimer.value = null
		}
	}

	// Schedule a proactive refresh based on the JWT's exp claim.
	// Called after every successful auth check or token refresh.
	function scheduleProactiveRefresh() {
		clearRefreshTimer()

		if (!authenticated.value || !userInfo.value?.exp) {
			return
		}

		const nowInSeconds = Date.now() / MILLISECONDS_A_SECOND
		const expiresIn = userInfo.value.exp - nowInSeconds
		const refreshIn = Math.max(expiresIn - REFRESH_BUFFER_SECONDS, 0)

		refreshTimer.value = setTimeout(() => {
			authStore.renewToken()
		}, refreshIn * MILLISECONDS_A_SECOND)
	}

	// Re-schedule whenever the user info (and thus exp) changes.
	watch(
		() => userInfo.value?.exp,
		() => scheduleProactiveRefresh(),
	)

	// Also re-schedule when authentication state changes (e.g. logout clears it).
	watch(authenticated, (isAuth) => {
		if (!isAuth) {
			clearRefreshTimer()
		}
	})

	// Try renewing the token every time vikunja is loaded initially
	// (When opening the browser the focus event is not fired)
	authStore.renewToken()

	// Check if the token is still valid if the window gets focus again to maybe renew it.
	// This handles the case where the laptop was suspended and the timer didn't fire.
	useEventListener('focus', async () => {
		if (!authenticated.value) {
			return
		}

		const nowInSeconds = Date.now() / MILLISECONDS_A_SECOND
		const expiresIn = userInfo.value
			? userInfo.value.exp - nowInSeconds
			: 0

		// If the token is already expired, try to refresh immediately.
		// The 401 interceptor would also handle this, but refreshing here
		// avoids a brief error flash on the first API call after focus.
		if (expiresIn <= 0) {
			try {
				await authStore.renewToken()
			} catch {
				await authStore.checkAuth()
				await router.push({name: 'user.login'})
			}
			return
		}

		// If the token expires within the buffer window, refresh now.
		if (expiresIn < REFRESH_BUFFER_SECONDS) {
			authStore.renewToken()
		}
	})
}
