export const API_PATH_SUFFIX = '/api/v2'

// upgrades a v1 base to v2, keeping the deployment prefix in front of it
export function getApiBaseUrl(): string {
	const base = (window.API_URL ?? '').replace(/\/$/, '')
	if (base === '') {
		throw new InvalidApiUrlProvidedError()
	}

	return base.replace(/\/api\/v1$/, API_PATH_SUFFIX)
}

export function getLegacyApiBaseUrl(): string {
	return getApiBaseUrl().replace(/\/api\/v2$/, '/api/v1')
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
