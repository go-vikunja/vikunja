import {ref} from 'vue'
import {useOnline as useNetworkOnline} from '@vueuse/core'
import type {ConfigurableWindow} from '@vueuse/core'

export function useOnline(options?: ConfigurableWindow) {
	const isOnline = useNetworkOnline(options)
	const fakeOnlineState = Boolean(import.meta.env.VITE_IS_ONLINE)
	if (isOnline.value === false && fakeOnlineState) {
		console.log('Setting fake online state', fakeOnlineState)
		return ref(true)
	}
	return isOnline
}
