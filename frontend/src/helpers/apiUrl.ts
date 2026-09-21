export const API_PATH_SUFFIX = '/api/v2'

// upgrades a v1 base to v2, keeping the deployment prefix in front of it
export function getApiBaseUrl(): string {
	const url = window.API_URL
	return (url?.endsWith('/') ? url : url + '/').replace(/\/api\/v1\/$/, `${API_PATH_SUFFIX}/`)
}

export function getLegacyApiBaseUrl(): string {
	return getApiBaseUrl().replace(/\/api\/v2\/$/, '/api/v1/')
}
