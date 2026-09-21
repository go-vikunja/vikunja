export const API_PATH_SUFFIX = '/api/v2'

// upgrades a stored v1 base to v2; whatever deployment prefix sits in front of it is kept
export function getApiBaseUrl(): string {
	const url = window.API_URL
	return (url?.endsWith('/') ? url : url + '/').replace(/\/api\/v1\/$/, `${API_PATH_SUFFIX}/`)
}

export function getLegacyApiBaseUrl(): string {
	return getApiBaseUrl().replace(/\/api\/v2\/$/, '/api/v1/')
}
