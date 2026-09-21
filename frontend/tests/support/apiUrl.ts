// The frontend stores its API base without the /api/vN suffix the env URLs may carry.
export function apiRootUrl(url: string): string {
	return url.replace(/\/+$/, '').replace(/\/api\/v[12]$/, '')
}

export function apiV1Url(): string {
	return `${apiRootUrl(process.env.API_URL || 'http://localhost:3456')}/api/v1`
}
