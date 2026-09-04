import {apiV2Url, AuthenticatedHTTPFactory} from '@/helpers/fetcher'

export interface ILinkPreview {
	url: string
	title?: string
	description?: string
	image?: string
	site_name?: string
	favicon?: string
}

const cache = new Map<string, Promise<ILinkPreview | null>>()

// Fetches an OpenGraph/meta preview for an external URL from the SSRF-protected
// /api/v2/link-preview endpoint. Failures resolve to null so callers can just
// fall back to the plain link. Results are cached per URL for the session.
export function getLinkPreview(url: string): Promise<ILinkPreview | null> {
	const cached = cache.get(url)
	if (cached !== undefined) {
		return cached
	}

	const request = AuthenticatedHTTPFactory()
		.get(apiV2Url('link-preview'), {params: {url}})
		.then(({data}) => data as ILinkPreview)
		.catch(() => null)

	cache.set(url, request)
	return request
}
