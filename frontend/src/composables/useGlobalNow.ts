import { ref } from 'vue'
import { createGlobalState, useIntervalFn } from '@vueuse/core'

import { MILLISECONDS_A_SECOND } from '@/constants/date'

export const GLOBAL_NOW_INTERVAL = 60 * MILLISECONDS_A_SECOND

// Shared by formatters that may run during rendering or outside components.
export const useGlobalNow = createGlobalState(() => {
	const now = ref(new Date())

	const update = () => now.value = new Date()

	useIntervalFn(update, GLOBAL_NOW_INTERVAL, { immediate: true })

	return {
		now,
		update,
	}
})
