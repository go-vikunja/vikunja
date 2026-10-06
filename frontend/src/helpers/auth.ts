import {getApiBaseUrl} from '@/helpers/apiUrl'
import {canonicalApiBaseUrl} from '@/client/requestContext'
import {authRefreshToken} from '@/client/generated'
import {publicClient} from '@/client/publicClient'
import {ApiError} from '@/client/problem'
import {isDesktopApp, refreshDesktopToken} from '@/helpers/desktopAuth'
import {clearServerClock, recordServerClock, serverNowSeconds} from '@/helpers/serverClock'
import {MILLISECONDS_A_MINUTE, MILLISECONDS_A_SECOND} from '@/constants/date'

let savedToken: string | null = null

/**
 * Saves a token while optionally saving it to lacal storage. This is used when viewing a link share:
 * It enables viewing multiple link shares indipendently from each in multiple tabs other without overriding any other open ones.
 */
export const saveToken = (token: string, persist: boolean) => {
	savedToken = token
	if (persist) {
		localStorage.setItem('token', token)
	}
	recordServerClock(getTokenPayload(token)?.iat, persist)
}

/**
 * Returns a saved token. If there is one saved in memory it will use that before anything else.
 */
export const getToken = (): string | null => {
	if (savedToken !== null) {
		return savedToken
	}

	savedToken = localStorage.getItem('token')
	return savedToken
}

function getTokenPayload(token: string | null): Record<string, unknown> | null {
	if (!token) return null
	try {
		const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
		return JSON.parse(atob(base64))
	} catch {
		return null
	}
}

export function getTokenType(token: string | null): number | null {
	const payload = getTokenPayload(token)
	return typeof payload?.type === 'number' ? payload.type : null
}

export function isTokenExpired(token: string | null, marginSeconds = 0): boolean {
	const exp = getTokenPayload(token)?.exp
	return typeof exp !== 'number' || exp <= serverNowSeconds() + marginSeconds
}

export function getTokenIdentity(token: string | null): {id: number; type: number} | null {
	const payload = getTokenPayload(token)
	if (typeof payload?.id !== 'number' || typeof payload.type !== 'number') {
		return null
	}

	return {id: payload.id, type: payload.type}
}

/**
 * Removes all tokens everywhere.
 */
export const removeToken = () => {
	savedToken = null
	localStorage.removeItem('token')
	localStorage.removeItem('desktopOAuthRefreshToken')
	clearServerClock()

	// Bump the epoch and drop the in-flight refresh so a refresh that started
	// before this logout can't re-persist a token after we cleared it.
	authEpoch++
	inFlightRefresh = null
}

// Coalesces concurrent same-tab refreshes into one POST. Web Locks (below) is
// secure-context-only, so on insecure HTTP there's no cross-tab coordination —
// without this guard, refreshes firing close together each spend the single-use
// cookie and all but one get a 401.
let inFlightRefresh: Promise<void> | null = null

const refreshListeners: (() => void)[] = []

export function onTokenRefreshed(listener: () => void) {
	refreshListeners.push(listener)
	return () => {
		const index = refreshListeners.indexOf(listener)
		if (index !== -1) {
			refreshListeners.splice(index, 1)
		}
	}
}

// Incremented on every removeToken()/logout. A refresh captures the epoch when
// it starts and only persists its result if the epoch is unchanged, so a
// refresh that resolves after a logout can't undo it.
let authEpoch = 0

export function getAuthSessionEpoch(): number {
	return authEpoch
}

// The backend's refresh limit window is a minute; this caps a bogus Retry-After.
export const MAX_RETRY_AFTER_MS = 10 * MILLISECONDS_A_MINUTE

export type RefreshFailure =
	| {kind: 'network'}
	| {kind: 'server', status: number}
	| {kind: 'rate-limited', retryAt?: number}
	| {kind: 'rejected', status: number, code?: number}

// Desktop IPC failures and unusable 200 responses stay plain Errors: no HTTP status to classify.
export class RefreshTokenError extends Error {
	readonly failure: RefreshFailure

	constructor(failure: RefreshFailure, cause: unknown) {
		super('Error renewing token: ', {cause})
		this.failure = failure
	}
}

type TimedRateLimitError = RefreshTokenError & {readonly failure: {kind: 'rate-limited', retryAt: number}}

function isTimedRateLimit(e: RefreshTokenError): e is TimedRateLimitError {
	return e.failure.kind === 'rate-limited' && e.failure.retryAt !== undefined
}

// Not cleared on logout: the backend limits refreshes per IP, not per session.
let rateLimit: {error: TimedRateLimitError, apiBase: string} | null = null

function retryAtFromHeader(value: string | null): number | undefined {
	if (value === null || !/^\d+$/.test(value)) {
		return undefined
	}
	return Date.now() + Math.min(Number(value) * MILLISECONDS_A_SECOND, MAX_RETRY_AFTER_MS)
}

function httpRefreshFailure({status, code}: {status: number, code?: unknown}, headers: Headers): RefreshFailure {
	if (status >= 500) {
		return {kind: 'server', status}
	}
	if (status === 429) {
		return {kind: 'rate-limited', retryAt: retryAtFromHeader(headers.get('Retry-After'))}
	}
	return {
		kind: 'rejected',
		status,
		code: typeof code === 'number' ? code : undefined,
	}
}

/**
 * Refreshes an auth token while ensuring it is updated everywhere.
 * The refresh token is sent automatically as an HttpOnly cookie.
 * The server rotates the cookie on every call.
 *
 * Same-tab concurrent calls share one in-flight refresh (always-on dedup); the
 * Web Locks API inside adds cross-tab coordination only in secure contexts.
 */
export async function refreshToken(persist: boolean): Promise<void> {
	if (inFlightRefresh) {
		return inFlightRefresh
	}
	if (rateLimit?.apiBase === getApiBaseUrl() && Date.now() < rateLimit.error.failure.retryAt) {
		throw rateLimit.error
	}
	const p = doRefresh(persist)
	inFlightRefresh = p
	// Only clear if it still points to this promise — a logout (or a newer
	// refresh started after it) may have replaced inFlightRefresh meanwhile.
	// .catch: callers get the rejection through p; avoid a second unhandled one.
	p.finally(() => {
		if (inFlightRefresh === p) {
			inFlightRefresh = null
		}
	}).catch(() => {})
	p.then(() => {
		for (const listener of [...refreshListeners]) {
			try {
				listener()
			} catch (e) {
				console.error('Token refresh listener failed', e)
			}
		}
	}, () => {})
	return p
}

async function doRefresh(persist: boolean): Promise<void> {
	// Snapshot the epoch so we can tell if a logout happened while we awaited.
	const epochAtStart = authEpoch
	const serverAtStart = getApiBaseUrl()
	const loggedOutSinceStart = () => authEpoch !== epochAtStart || getApiBaseUrl() !== serverAtStart

	// Capture the tokens before waiting for the lock so we can detect
	// if another tab refreshed while we were queued.
	const tokenBeforeLock = localStorage.getItem('token')
	const desktopRefreshTokenBeforeLock = localStorage.getItem('desktopOAuthRefreshToken')

	const refreshUnderLock = async () => {
		// A logout may have happened while we waited for the lock — don't
		// re-adopt or re-fetch a token after the user signed out.
		if (loggedOutSinceStart()) {
			return
		}

		// In desktop mode, refresh via IPC to the Electron main process
		if (isDesktopApp()) {
			const storedRefreshToken = localStorage.getItem('desktopOAuthRefreshToken')

			if (storedRefreshToken !== desktopRefreshTokenBeforeLock) {
				const currentToken = localStorage.getItem('token')
				if (currentToken) {
					savedToken = currentToken
					return
				}
			}

			if (!storedRefreshToken) {
				throw new Error('No desktop OAuth refresh token available')
			}

			try {
				const tokens = await refreshDesktopToken(window.API_URL, storedRefreshToken)
				if (loggedOutSinceStart()) {
					return
				}
				saveToken(tokens.access_token, persist)
				localStorage.setItem('desktopOAuthRefreshToken', tokens.refresh_token)
			} catch (e) {
				throw new Error('Error renewing token: ', {cause: e})
			}
			return
		}

		// If the token in localStorage changed while waiting for the lock,
		// another tab already refreshed. Just adopt the new token.
		const currentToken = localStorage.getItem('token')
		if (currentToken && currentToken !== tokenBeforeLock) {
			savedToken = currentToken
			return
		}

		// We hold the lock and no one else refreshed — make the API call.
		try {
			const baseUrl = canonicalApiBaseUrl(getApiBaseUrl())
			const {data, error, response} = await authRefreshToken({client: publicClient, baseUrl, throwOnError: false})
			if (error) {
				const body: unknown = error
				// fetch and body reads reject with a TypeError when the connection drops.
				if (body instanceof TypeError) {
					throw new RefreshTokenError({kind: 'network'}, body)
				}
				// hey-api sets `response` before reading the body, so body read/parse errors arrive with a 200 status.
				if (!response || (body instanceof Error && !(body instanceof ApiError))) {
					throw body
				}
				// Proxy error pages aren't JSON, so the status comes from the response.
				const problem = typeof body === 'object'
					? {...body, status: response.status}
					: {status: response.status, detail: String(body)}
				const failure = httpRefreshFailure(problem, response.headers)
				const refreshError = new RefreshTokenError(failure, problem)
				if (isTimedRateLimit(refreshError)) {
					rateLimit = {error: refreshError, apiBase: serverAtStart}
				}
				throw refreshError
			}
			if (loggedOutSinceStart()) {
				return
			}
			if (!data?.token) throw new Error('Refresh response has no token')
			saveToken(data.token, persist)
		} catch (e) {
			// Another tab's refresh can land while our POST is in flight.
			const storedToken = localStorage.getItem('token')
			if (!loggedOutSinceStart() && storedToken && storedToken !== tokenBeforeLock) {
				savedToken = storedToken
				return
			}
			throw e instanceof RefreshTokenError ? e : new Error('Error renewing token: ', {cause: e})
		}
	}

	if (navigator.locks) {
		await navigator.locks.request('vikunja-token-refresh', refreshUnderLock)
	} else {
		// Fallback for environments without Web Locks (e.g. insecure HTTP)
		await refreshUnderLock()
	}
}
