import {useConfigStore} from '@/stores/config'

const API_DEFAULT_PORT = '3456'

export const ERROR_NO_API_URL = 'noApiUrlProvided'

export class NoApiUrlProvidedError extends Error {
	constructor() {
		super()
		this.message = 'No API URL provided'
		this.name = 'NoApiUrlProvidedError'
	}
}

export class InvalidApiUrlProvidedError extends Error {
	constructor() {
		super()
		this.message = 'The provided API URL is invalid.'
		this.name = 'InvalidApiUrlProvidedError'
	}
}

export const checkAndSetApiUrl = (pUrl: string | undefined | null): Promise<string> => {
	let url = pUrl
	if (url === '' || url === null || typeof url === 'undefined') {
		throw new NoApiUrlProvidedError()
	}

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
	
	let urlToCheck: URL
	try {
		urlToCheck = new URL(url)
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
	} catch (e) {
		throw new InvalidApiUrlProvidedError()
	}

	const origUrlToCheck = urlToCheck

	const oldUrl = window.API_URL
	window.API_URL = urlToCheck.toString()

	const configStore = useConfigStore()

	// Check if the api is reachable at the provided url
	return configStore.update()
		.catch(e => {
			console.warn(`Could not fetch 'info' from the provided endpoint ${pUrl} on ${window.API_URL}/info. Some automatic fallback will be tried.`)
			// Check if it is reachable at /api/v1 and http
			if (
				!urlToCheck.pathname.endsWith('/api/v1') &&
				!urlToCheck.pathname.endsWith('/api/v1/')
			) {
				urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
				window.API_URL = urlToCheck.toString()
				return configStore.update()
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
				return configStore.update()
			}
			throw e
		})
		.catch(e => {
			// Check if it is reachable at port API_DEFAULT_PORT and https
			if (urlToCheck.port !== API_DEFAULT_PORT) {
				urlToCheck.port = API_DEFAULT_PORT
				window.API_URL = urlToCheck.toString()
				return configStore.update()
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
				return configStore.update()
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

			throw new InvalidApiUrlProvidedError()
		})
}
