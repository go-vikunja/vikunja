import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import type {AxiosResponse} from 'axios'

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
 */
export async function refreshToken(persist: boolean): Promise<AxiosResponse> {
	const HTTP = AuthenticatedHTTPFactory()
	try {
		const response = await HTTP.post('user/token')
		saveToken(response.data.token, persist)
		return response

	} catch(e) {
		throw new Error('Error renewing token: ', { cause: e })
	}
}

