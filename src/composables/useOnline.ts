import {ref} from 'vue'
import {useOnline as useNetworkOnline, ConfigurableWindow} from '@vueuse/core'


export function useOnline(options?: ConfigurableWindow) {
	const fakeOnlineState = !!import.meta.env.VITE_IS_ONLINE
	if (fakeOnlineState) {
		console.log('Setting fake online state', fakeOnlineState)
	}

	return fakeOnlineState
		? ref(true)
		: useNetworkOnline(options)
}