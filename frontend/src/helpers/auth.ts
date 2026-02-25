import {HTTPFactory} from '@/helpers/fetcher'

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
}

/**
 * Refreshes an auth token while ensuring it is updated everywhere.
 * The refresh token is sent automatically as an HttpOnly cookie.
 * The server rotates the cookie on every call.
 *
 * Uses the Web Locks API to coordinate across browser tabs. Only one tab
 * performs the actual refresh; other tabs waiting for the lock detect that
 * the token in localStorage was already updated and adopt it directly.
 */
export async function refreshToken(persist: boolean): Promise<void> {
	// Capture the token before waiting for the lock so we can detect
	// if another tab refreshed while we were queued.
	const tokenBeforeLock = localStorage.getItem('token')

	const doRefresh = async () => {
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
		await navigator.locks.request('vikunja-token-refresh', doRefresh)
	} else {
		// Fallback for environments without Web Locks (e.g. insecure HTTP)
		await doRefresh()
	}
}

