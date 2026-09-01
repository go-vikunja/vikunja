import {getAuthSessionEpoch, getToken, getTokenIdentity} from '@/helpers/auth'
import {getApiV2BaseUrl} from '@/helpers/fetcher'

export type ClientRequestContext = {
	identity: ReturnType<typeof getTokenIdentity>
	authSessionEpoch: number
	apiV2BaseUrl: string
}

export function captureClientRequestContext(): ClientRequestContext {
	return {
		identity: getTokenIdentity(getToken()),
		authSessionEpoch: getAuthSessionEpoch(),
		apiV2BaseUrl: getApiV2BaseUrl(),
	}
}

export function isClientRequestContextCurrent(context: ClientRequestContext): boolean {
	const current = captureClientRequestContext()
	const identityMatches = context.identity === null
		? current.identity === null
		: current.identity?.id === context.identity.id && current.identity.type === context.identity.type

	return identityMatches &&
		current.authSessionEpoch === context.authSessionEpoch &&
		current.apiV2BaseUrl === context.apiV2BaseUrl
}

export function assertClientRequestContext(context: ClientRequestContext): void {
	if (!isClientRequestContextCurrent(context)) {
		throw new DOMException('Client request context changed', 'AbortError')
	}
}

function canonicalApiBaseUrl(apiBaseUrl: string | undefined): string {
	const isRootRelative = apiBaseUrl?.startsWith('/') && !apiBaseUrl.startsWith('//')
	const isHttpUrl = /^https?:\/\//i.test(apiBaseUrl ?? '')
	if (!isRootRelative && !isHttpUrl) {
		throw new DOMException('Invalid client API URL', 'AbortError')
	}

	try {
		const normalized = new URL(apiBaseUrl, window.location.origin)
		if (normalized.protocol !== 'http:' && normalized.protocol !== 'https:') {
			throw new DOMException('Invalid client API URL', 'AbortError')
		}
		return normalized.toString()
	} catch (error) {
		if (error instanceof DOMException && error.name === 'AbortError') {
			throw error
		}
		throw new DOMException('Invalid client API URL', 'AbortError')
	}
}

export function assertClientRequestMatchesContext(
	request: Request,
	context: ClientRequestContext,
	configuredApiV2BaseUrl: string | undefined,
): void {
	assertClientRequestContext(context)

	const requestUrl = new URL(request.url, window.location.origin)
	const expectedConfiguredBaseUrl = context.apiV2BaseUrl.endsWith('/')
		? context.apiV2BaseUrl.slice(0, -1)
		: context.apiV2BaseUrl
	if (canonicalApiBaseUrl(configuredApiV2BaseUrl) !== canonicalApiBaseUrl(expectedConfiguredBaseUrl)) {
		throw new DOMException('Client request API changed', 'AbortError')
	}

	const apiV2BaseUrl = new URL(context.apiV2BaseUrl, window.location.origin)
	if (!apiV2BaseUrl.pathname.endsWith('/')) {
		apiV2BaseUrl.pathname += '/'
	}
	if (requestUrl.origin !== apiV2BaseUrl.origin || !requestUrl.pathname.startsWith(apiV2BaseUrl.pathname)) {
		throw new DOMException('Client request API changed', 'AbortError')
	}
}
