import { useConfigStore } from '@/stores/config'
import { HTTPFactory } from '@/helpers/fetcher'
import { objectToCamelCase } from '@/helpers/case'

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

// Tries to fetch the /info endpoint from a given URL.
async function testUrl(url: string): Promise<any> {
	const http = HTTPFactory()
	const response = await http.get(`${url}/info`, {
		headers: { 'Accept': 'application/json' },
	})
	if (typeof response.data.version === 'undefined') {
		throw new Error('Invalid info response')
	}
	return response.data
}

export const checkAndSetApiUrl = async (pUrl: string | undefined | null): Promise<string> => {
	let url = pUrl
	if (!url) {
		throw new NoApiUrlProvidedError()
	}

	if (url.startsWith('/')) {
		url = window.location.host + url
	}

	if (!url.startsWith('http://') && !url.startsWith('https://')) {
		url = `${window.location.protocol}//${url}`
	}

	// Candidate URLs to test
	const candidates = [
		url,
		url.endsWith('/') ? `${url}api/v1` : `${url}/api/v1`,
	]
	// Also test without any path
	try {
		const parsed = new URL(url)
		candidates.push(`${parsed.protocol}//${parsed.host}`)
	} catch(e) {
		// ignore parsing errors, we'll catch the invalid url later
	}


	for (const candidate of [...new Set(candidates)]) {
		try {
			const info = await testUrl(candidate)
			
			// We found a working URL. Now, strip any path to get the base.
			const parsed = new URL(candidate)
			const baseUrl = `${parsed.protocol}//${parsed.host}`
			
			const configStore = useConfigStore()
			configStore.setApiUrl(baseUrl)
			configStore.setConfig(objectToCamelCase(info))
			localStorage.setItem('API_URL', baseUrl)
			return baseUrl
		} catch (e) {
			console.warn(`Attempted to connect to ${candidate}, but failed.`, e)
		}
	}

	// If we get here, all candidates failed.
	throw new InvalidApiUrlProvidedError()
}
