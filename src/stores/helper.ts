export interface LoadingState {
	isLoading: boolean
}

const LOADING_TIMEOUT = 100

export const setModuleLoading = <Store extends LoadingState>(store: Store, loadFunc : ((isLoading: boolean) => void) | null = null) => {
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