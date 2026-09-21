import {useConfigStore} from '@/stores/config'
import {configureApiClient} from '@/client/http'
import {queryClient} from '@/client/queryClient'
import {
	getApiBaseUrl,
	InvalidApiUrlProvidedError,
	NoApiUrlProvidedError,
	normalizeApiUrl,
} from '@/helpers/apiUrl'

const API_DEFAULT_PORT = '3456'

export const ERROR_NO_API_URL = 'noApiUrlProvided'

function candidateUrls(pUrl: string): string[] {
	let url = normalizeApiUrl(pUrl)
	if (url === '' || url.startsWith('/')) {
		url = window.location.host + url
	}

	if (!url.startsWith('http://') && !url.startsWith('https://')) {
		url = `${window.location.protocol}//${url}`
	}

	let urlToCheck: URL
	try {
		urlToCheck = new URL(url)
	} catch {
		throw new InvalidApiUrlProvidedError()
	}

	const candidates = [normalizeApiUrl(urlToCheck.toString())]
	if (urlToCheck.port !== API_DEFAULT_PORT) {
		urlToCheck.port = API_DEFAULT_PORT
		candidates.push(normalizeApiUrl(urlToCheck.toString()))
	}
	return candidates
}

export const checkAndSetApiUrl = (pUrl: string | undefined | null): Promise<string> => {
	if (pUrl === null || typeof pUrl === 'undefined') {
		throw new NoApiUrlProvidedError()
	}

	const candidates = candidateUrls(pUrl)
	const oldUrl = window.API_URL
	const oldApiBase = typeof oldUrl === 'string' ? getApiBaseUrl() : null
	const configStore = useConfigStore()

	let probe: Promise<boolean> = Promise.reject(new InvalidApiUrlProvidedError())
	for (const candidate of candidates) {
		probe = probe.catch(() => {
			window.API_URL = candidate
			return configStore.update()
		})
	}

	return probe
		.catch(e => {
			console.warn(`Could not fetch 'info' from the provided endpoint ${pUrl}.`)
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
