import type { StoreDefinition } from 'pinia'

export const setLoadingPinia = (store: StoreDefinition, loadFunc : ((isLoading: boolean) => void) | null = null) => {
	const timeout = setTimeout(() => {
		if (loadFunc === null) {
			store.isLoading = true
		} else {
			loadFunc(true)
		}
	}, 100)
	return () => {
		clearTimeout(timeout)
		if (loadFunc === null) {
			store.isLoading = false
		} else {
			loadFunc(false)
		}
	}
}