import {HTTPFactory} from '@/http-common'

let savedToken = null
let persisted = false

/**
 * Saves a token while optionally saving it to lacal storage. This is used when viewing a link share:
 * It enables viewing multiple link shares indipendently from each in multiple tabs other without overriding any other open ones.
 * @param token
 * @param persist
 */
export const saveToken = (token, persist = true) => {
	savedToken = token
	if (persist) {
		persisted = true
		localStorage.setItem('token', token)
	}
}

/**
 * Returns a saved token. If there is one saved in memory it will use that before anything else.
 * @returns {string|null}
 */
export const getToken = () => {
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
 * @returns {Promise<AxiosResponse<any>>}
 */
export const refreshToken = () => {
	const HTTP = HTTPFactory()
	return HTTP.post('user/token', null, {
		headers: {
			Authorization: `Bearer ${getToken()}`,
		},
	})
		.then(r => {
			saveToken(r.data.token, persisted)
			return Promise.resolve(r)
		})
		.catch(e => {
			// eslint-disable-next-line
			console.log('Error renewing token: ', e)
			return Promise.reject(e)
		})
}

