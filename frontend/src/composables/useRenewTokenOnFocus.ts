import {computed} from 'vue'
import {useRouter} from 'vue-router'
import {useEventListener} from '@vueuse/core'

import {useAuthStore} from '@/stores/auth'
import {MILLISECONDS_A_SECOND, SECONDS_A_HOUR} from '@/constants/date'

const SECONDS_TOKEN_VALID = 60 * SECONDS_A_HOUR

export function useRenewTokenOnFocus() {
	const router = useRouter()
	const authStore = useAuthStore()

	const userInfo = computed(() => authStore.info)
	const authenticated = computed(() => authStore.authenticated)

	// Try renewing the token every time vikunja is loaded initially
	// (When opening the browser the focus event is not fired)
	authStore.renewToken()

	// Check if the token is still valid if the window gets focus again to maybe renew it
	useEventListener('focus', async () => {
		if (!authenticated.value) {
			return
		}

		const nowInSeconds = new Date().getTime() / MILLISECONDS_A_SECOND
		const expiresIn = userInfo.value
			? userInfo.value.exp - nowInSeconds
			: 0

		// If the token expiry is negative, it is already expired and we have no choice but to redirect
		// the user to the login page
		if (expiresIn <= 0) {
			await authStore.checkAuth()
			await router.push({name: 'user.login'})
			return
		}

		// Check if the token is valid for less than 60 hours and renew if thats the case
		if (expiresIn < SECONDS_TOKEN_VALID) {
			authStore.renewToken()
			console.debug('renewed token')
		}
	})
}
