// The frontend stores its API base without the /api/vN suffix the env URLs may carry.
export function apiRootUrl(url: string): string {
	return url.replace(/\/+$/, '').replace(/\/api\/v[12]$/, '')
}

// API_URL may list one API per worker, comma-separated. 127.0.0.1 rather than
// localhost to match the frontend's origin for CORS.
export function apiRootUrls(): string[] {
	const urls = (process.env.API_URL || 'http://127.0.0.1:3456')
		.split(',')
		.map(url => url.trim())
		.filter(Boolean)
		.map(apiRootUrl)
	if (urls.length === 0) {
		throw new Error('API_URL does not contain any URL')
	}
	return urls
}

// Parallel index, not worker index: a restarted worker keeps its parallel index.
export function workerApiRootUrl(): string {
	const urls = apiRootUrls()
	const index = Number(process.env.TEST_PARALLEL_INDEX ?? 0)
	const url = urls[index]
	if (!url) {
		throw new Error(`No API in API_URL for parallel index ${index} (${urls.length} configured)`)
	}
	return url
}

export function apiV1Url(): string {
	return `${workerApiRootUrl()}/api/v1`
}
