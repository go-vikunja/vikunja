import type { StoreDefinition } from 'pinia'

export interface LoadingState {
	isLoading: boolean
}

const LOADING_TIMEOUT = 100

export const setModuleLoading = <LoadingStore extends StoreDefinition<string, LoadingState>>(store: LoadingStore, loadFunc : ((isLoading: boolean) => void) | null = null) => {
	const timeout = setTimeout(() => {
		if (loadFunc === null) {
			store.isLoading = true
		} else {
			loadFunc(true)
		}
	}, LOADING_TIMEOUT)
	return () => {
		clearTimeout(timeout)
		if (loadFunc === null) {
			store.isLoading = false
		} else {
			loadFunc(false)
		}
	}
}