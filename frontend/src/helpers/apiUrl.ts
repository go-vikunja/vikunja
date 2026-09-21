export function getApiBaseUrl(): string {
	const url = window.API_URL
	return (url?.endsWith('/') ? url : url + '/').replace(/\/api\/v1\/$/, '/api/v2/')
}
