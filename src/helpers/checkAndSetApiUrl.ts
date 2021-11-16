import {store} from '@/store'

const API_DEFAULT_PORT = '3456'

export const ERROR_NO_API_URL = 'noApiUrlProvided'

const updateConfig = () => store.dispatch('config/update') 

export const checkAndSetApiUrl = (url: string): Promise<string> => {
	if(url.startsWith('/')) {
		url = window.location.host + url
	}
	
	// Check if the url has an http prefix
	if (
		!url.startsWith('http://') &&
		!url.startsWith('https://')
	) {
		url = `http://${url}`
	}

	const urlToCheck: URL = new URL(url)
	const origUrlToCheck = urlToCheck

	const oldUrl = window.API_URL
	window.API_URL = urlToCheck.toString()

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
			// Check if it has a port and if not check if it is reachable at https
			if (urlToCheck.protocol === 'http:') {
				urlToCheck.protocol = 'https:'
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
				urlToCheck.protocol = 'https:'
				urlToCheck.port = API_DEFAULT_PORT
				window.API_URL = urlToCheck.toString()
				return updateConfig()
			}
			throw e
		})
		.catch(e => {
			// Check if it is reachable at :API_DEFAULT_PORT and /api/v1 and https
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
			// Check if it is reachable at port API_DEFAULT_PORT and http
			if (urlToCheck.port !== API_DEFAULT_PORT) {
				urlToCheck.protocol = 'http:'
				urlToCheck.port = API_DEFAULT_PORT
				window.API_URL = urlToCheck.toString()
				return updateConfig()
			}
			throw e
		})
		.catch(e => {
			// Check if it is reachable at :API_DEFAULT_PORT and /api/v1 and http
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
		.then(r => {
			if (typeof r !== 'undefined') {
				localStorage.setItem('API_URL', window.API_URL)
				return window.API_URL
			}
			
			throw new Error(ERROR_NO_API_URL)
		})
}
