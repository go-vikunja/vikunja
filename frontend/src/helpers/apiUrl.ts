const API_VERSION_SUFFIX = /\/api\/v[12]$/

// Stored without /api/vN ('' = same origin); values saved before that still carry one.
export function normalizeApiUrl(url: string): string {
	return url.replace(/\/+$/, '').replace(API_VERSION_SUFFIX, '')
}

export function getApiRootUrl(): string {
	if (typeof window.API_URL !== 'string') {
		throw new InvalidApiUrlProvidedError()
	}

	return normalizeApiUrl(window.API_URL)
}

export function getApiBaseUrl(): string {
	return `${getApiRootUrl()}/api/v2`
}

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
