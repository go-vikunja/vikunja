import {onScopeDispose, readonly, ref, toValue, watch, type MaybeRefOrGetter} from 'vue'

import {LOADING_TIMEOUT} from '@/stores/helper'

// Spinner state only: keep guards on the raw flag, this one lags by LOADING_TIMEOUT.
export function useDelayedLoading(source: MaybeRefOrGetter<boolean>, delay = LOADING_TIMEOUT) {
	const loading = ref(false)
	let timer: ReturnType<typeof setTimeout> | undefined

	watch(() => toValue(source), isLoading => {
		clearTimeout(timer)
		if (!isLoading) {
			loading.value = false
			return
		}
		timer = setTimeout(() => loading.value = true, delay)
	}, {immediate: true, flush: 'sync'})

	onScopeDispose(() => clearTimeout(timer))

	return readonly(loading)
}
