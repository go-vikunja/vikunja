import {useConfigStore} from '@/stores/config'

const API_DEFAULT_PORT = '3456'

export const ERROR_NO_API_URL = 'noApiUrlProvided'


export const checkAndSetApiUrl = (url: string): Promise<string> => {
	if (url.startsWith('/')) {
		url = window.location.host + url
	}

	// Check if the url has a http prefix
	if (
		!url.startsWith('http://') &&
		!url.startsWith('https://')
	) {
		url = `${window.location.protocol}//${url}`
	}

	const urlToCheck: URL = new URL(url)
	const origUrlToCheck = urlToCheck

	const oldUrl = window.API_URL
	window.API_URL = urlToCheck.toString()

	const configStore = useConfigStore()
	const updateConfig = () => configStore.update()

	// Check if the api is reachable at the provided url
	return updateConfig()
		.catch(e => {
			// Check if it is reachable at /api/v1 and http
			if (
				!urlToCheck.pathname.endsWith('/api/v1') &&
				!urlToCheck.pathname.endsWith('/api/v1/')
			) {
				urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
				window.API_URL = urlToCheck.toString()
				return updateConfig()
			}
			throw e
		})
		.catch(e => {
			// Check if it is reachable at /api/v1 and https
			urlToCheck.pathname = origUrlToCheck.pathname
			if (
				!urlToCheck.pathname.endsWith('/api/v1') &&
				!urlToCheck.pathname.endsWith('/api/v1/')
			) {
				urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
				window.API_URL = urlToCheck.toString()
				return updateConfig()
			}
			throw e
		})
		.catch(e => {
			// Check if it is reachable at port API_DEFAULT_PORT and https
			if (urlToCheck.port !== API_DEFAULT_PORT) {
				urlToCheck.port = API_DEFAULT_PORT
				window.API_URL = urlToCheck.toString()
				return updateConfig()
			}
			throw e
		})
		.catch(e => {
			// Check if it is reachable at :API_DEFAULT_PORT and /api/v1
			urlToCheck.pathname = origUrlToCheck.pathname
			if (
				!urlToCheck.pathname.endsWith('/api/v1') &&
				!urlToCheck.pathname.endsWith('/api/v1/')
			) {
				urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
				window.API_URL = urlToCheck.toString()
				return updateConfig()
			}
			throw e
		})
		.catch(e => {
			window.API_URL = oldUrl
			throw e
		})
		.then(success => {
			if (success) {
				localStorage.setItem('API_URL', window.API_URL)
				return window.API_URL
			}

			throw new Error(ERROR_NO_API_URL)
		})
}
