const LOADING_TIMEOUT = 100

export function setModuleLoading(loadFunc: (isLoading: boolean) => void) {
	const timeout = setTimeout(() => loadFunc(true), LOADING_TIMEOUT)
	return () => {
		clearTimeout(timeout)
		loadFunc(false)
	}
}
