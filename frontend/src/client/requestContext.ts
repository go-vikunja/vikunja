import {getAuthSessionEpoch, getToken, getTokenIdentity} from '@/helpers/auth'
import {getApiBaseUrl} from '@/helpers/apiUrl'

export type ClientRequestContext = {
	identity: ReturnType<typeof getTokenIdentity>
	authSessionEpoch: number
	apiBaseUrl: string
}

export function captureClientRequestContext(): ClientRequestContext {
	return {
		identity: getTokenIdentity(getToken()),
		authSessionEpoch: getAuthSessionEpoch(),
		apiBaseUrl: getApiBaseUrl(),
	}
}

export function isClientRequestContextCurrent(context: ClientRequestContext): boolean {
	const current = captureClientRequestContext()
	const identityMatches = context.identity === null
		? current.identity === null
		: current.identity?.id === context.identity.id && current.identity.type === context.identity.type

	return identityMatches &&
		current.authSessionEpoch === context.authSessionEpoch &&
		current.apiBaseUrl === context.apiBaseUrl
}

export function assertClientRequestContext(context: ClientRequestContext): void {
	if (!isClientRequestContextCurrent(context)) {
		throw new DOMException('Client request context changed', 'AbortError')
	}
}

// Fence aborts are internal: readers drop them instead of reporting them to the user.
export function isRequestContextAbort(cause: unknown): boolean {
	return (cause as {name?: string} | null)?.name === 'AbortError'
}

function canonicalApiBaseUrl(apiBaseUrl: unknown): string {
	if (typeof apiBaseUrl !== 'string') {
		throw new DOMException('Invalid client API URL', 'AbortError')
	}
	const isRootRelative = apiBaseUrl.startsWith('/') && !apiBaseUrl.startsWith('//')
	const isHttpUrl = /^https?:\/\//i.test(apiBaseUrl)
	if (!isRootRelative && !isHttpUrl) {
		throw new DOMException('Invalid client API URL', 'AbortError')
	}

	try {
		const normalized = new URL(apiBaseUrl, window.location.origin)
		if (
			(isRootRelative && normalized.origin !== window.location.origin) ||
			(normalized.protocol !== 'http:' && normalized.protocol !== 'https:')
		) {
			throw new DOMException('Invalid client API URL', 'AbortError')
		}
		return normalized.toString()
	} catch (error) {
		if (isRequestContextAbort(error)) {
			throw error
		}
		throw new DOMException('Invalid client API URL', 'AbortError')
	}
}

export function assertClientRequestMatchesContext(
	request: Request,
	context: ClientRequestContext,
	configuredApiBaseUrl: unknown,
): void {
	assertClientRequestContext(context)

	const requestUrl = new URL(request.url, window.location.origin)
	const expectedConfiguredBaseUrl = context.apiBaseUrl.endsWith('/')
		? context.apiBaseUrl.slice(0, -1)
		: context.apiBaseUrl
	if (canonicalApiBaseUrl(configuredApiBaseUrl) !== canonicalApiBaseUrl(expectedConfiguredBaseUrl)) {
		throw new DOMException('Client request API changed', 'AbortError')
	}

	const apiBaseUrl = new URL(context.apiBaseUrl, window.location.origin)
	if (!apiBaseUrl.pathname.endsWith('/')) {
		apiBaseUrl.pathname += '/'
	}
	if (requestUrl.origin !== apiBaseUrl.origin || !requestUrl.pathname.startsWith(apiBaseUrl.pathname)) {
		throw new DOMException('Client request API changed', 'AbortError')
	}
}
