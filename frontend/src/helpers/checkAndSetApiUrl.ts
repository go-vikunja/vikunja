import {useConfigStore} from '@/stores/config'
import {configureApiClient} from '@/client/http'
import {queryClient} from '@/client/queryClient'
import {
	API_PATH_SUFFIX,
	getApiBaseUrl,
	InvalidApiUrlProvidedError,
	NoApiUrlProvidedError,
} from '@/helpers/apiUrl'

const API_DEFAULT_PORT = '3456'

export const ERROR_NO_API_URL = 'noApiUrlProvided'

/**
 * Join a base pathname with the API_DEFAULT_PATH, normalizing slashes between them.
 */
function joinPath(base: string, suffix: string): string {
	const normalizedBase = base.endsWith('/') ? base.slice(0, -1) : base
	return normalizedBase + suffix
}

/**
 * Check whether a pathname already ends with the API default path,
 * with or without a trailing slash.
 */
function hasApiPath(pathname: string): boolean {
	const clean = pathname.endsWith('/') ? pathname.slice(0, -1) : pathname
	return clean.endsWith(API_PATH_SUFFIX)
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

	urlToCheck.pathname = urlToCheck.pathname.replace(/\/api\/v1\/?$/, API_PATH_SUFFIX)
	const origPathname = urlToCheck.pathname

	const oldUrl = window.API_URL
	const oldApiBase = oldUrl ? getApiBaseUrl() : null
	window.API_URL = urlToCheck.toString()

	const configStore = useConfigStore()

	// Check if the api is reachable at the provided url
	return configStore.update()
		.catch(e => {
			console.warn(`Could not fetch 'info' from the provided endpoint ${pUrl} on ${window.API_URL}/info. Some automatic fallback will be tried.`)
			// Check if it is reachable at the base path + /api/v2 via http
			if (!hasApiPath(urlToCheck.pathname)) {
				urlToCheck.pathname = joinPath(urlToCheck.pathname, API_PATH_SUFFIX)
				window.API_URL = urlToCheck.toString()
				return configStore.update()
			}
			throw e
		})
		.catch(e => {
			// Check if it is reachable at the base path + /api/v2 via https
			urlToCheck.pathname = origPathname
			if (!hasApiPath(urlToCheck.pathname)) {
				urlToCheck.pathname = joinPath(urlToCheck.pathname, API_PATH_SUFFIX)
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
			// Check if it is reachable at :API_DEFAULT_PORT with base path + /api/v2
			urlToCheck.pathname = origPathname
			if (!hasApiPath(urlToCheck.pathname)) {
				urlToCheck.pathname = joinPath(urlToCheck.pathname, API_PATH_SUFFIX)
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
				if (getApiBaseUrl() !== oldApiBase) {
					configureApiClient()
					queryClient.clear()
				}
				localStorage.setItem('API_URL', window.API_URL)
				return window.API_URL
			}

			throw new InvalidApiUrlProvidedError()
		})
}
