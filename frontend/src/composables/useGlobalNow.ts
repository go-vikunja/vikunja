import { ref } from 'vue'
import { createGlobalState, useIntervalFn } from '@vueuse/core'
import { onBeforeRouteUpdate } from 'vue-router'

import { MILLISECONDS_A_SECOND } from '@/constants/date'

const GLOBAL_NOW_INTERVAL = 60 * MILLISECONDS_A_SECOND

/**
 * A global shared state that provides the current time, updated at a regular interval.
 * 
 * Sharing this state globally ensures that all components accessing this hook use the same time reference, avoiding redundant intervals and ensuring consistency across the application.
 */
export const useGlobalNow = createGlobalState(() => {
	const now = ref(new Date())

	const update = () => now.value = new Date()

	useIntervalFn(update, GLOBAL_NOW_INTERVAL, { immediate: true })

	// ensure the now value is refreshed when the route changes
	onBeforeRouteUpdate(() => {
		update()
	})

	return {
		now,
		update,
	}
})
