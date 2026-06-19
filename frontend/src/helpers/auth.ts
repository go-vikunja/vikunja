import {HTTPFactory} from '@/helpers/fetcher'
import {isDesktopApp, refreshDesktopToken} from '@/helpers/desktopAuth'

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

/**
 * Removes all tokens everywhere.
 */
export const removeToken = () => {
	savedToken = null
	localStorage.removeItem('token')
	localStorage.removeItem('desktopOAuthRefreshToken')
}

// Coalesces concurrent same-tab refreshes into one POST. Web Locks (below) is
// secure-context-only, so on insecure HTTP there's no cross-tab coordination —
// without this guard, refreshes firing close together each spend the single-use
// cookie and all but one get a 401.
let inFlightRefresh: Promise<void> | null = null

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
	inFlightRefresh = doRefresh(persist).finally(() => {
		inFlightRefresh = null
	})
	return inFlightRefresh
}

async function doRefresh(persist: boolean): Promise<void> {
	// In desktop mode, refresh via IPC to the Electron main process
	if (isDesktopApp()) {
		const storedRefreshToken = localStorage.getItem('desktopOAuthRefreshToken')
		if (!storedRefreshToken) {
			throw new Error('No desktop OAuth refresh token available')
		}
		try {
			const tokens = await refreshDesktopToken(window.API_URL, storedRefreshToken)
			saveToken(tokens.access_token, persist)
			localStorage.setItem('desktopOAuthRefreshToken', tokens.refresh_token)
		} catch (e) {
			throw new Error('Error renewing token: ', {cause: e})
		}
		return
	}

	// Capture the token before waiting for the lock so we can detect
	// if another tab refreshed while we were queued.
	const tokenBeforeLock = localStorage.getItem('token')

	const refreshUnderLock = async () => {
		// If the token in localStorage changed while waiting for the lock,
		// another tab already refreshed. Just adopt the new token.
		const currentToken = localStorage.getItem('token')
		if (currentToken && currentToken !== tokenBeforeLock) {
			savedToken = currentToken
			return
		}

		// We hold the lock and no one else refreshed — make the API call.
		const HTTP = HTTPFactory()
		try {
			const response = await HTTP.post('user/token/refresh')
			saveToken(response.data.token, persist)
		} catch (e) {
			throw new Error('Error renewing token: ', {cause: e})
		}
	}

	if (navigator.locks) {
		await navigator.locks.request('vikunja-token-refresh', refreshUnderLock)
	} else {
		// Fallback for environments without Web Locks (e.g. insecure HTTP)
		await refreshUnderLock()
	}
}

