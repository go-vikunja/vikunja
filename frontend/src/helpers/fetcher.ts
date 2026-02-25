import axios from 'axios'
import type {AxiosRequestConfig} from 'axios'
import {getToken, refreshToken} from '@/helpers/auth'
import {AUTH_TYPES} from '@/modelTypes/IUser'

export function HTTPFactory() {
	const instance = axios.create({
		baseURL: window.API_URL,
		// Ensure the browser sends and accepts cookies (e.g. the HttpOnly
		// refresh token) even when the API is on a different origin.
		withCredentials: true,
	})

	instance.interceptors.request.use((config) => {
		// by setting the baseURL fresh for every request
		// we make sure that it is never outdated in case it is updated
		config.baseURL = window.API_URL

		return config
	})

	return instance
}

// Shared state for the 401 interceptor so that concurrent requests that all
// fail with 401 only trigger a single refresh, then all retry with the new token.
let refreshPromise: Promise<string | null> | null = null

async function doRefresh(): Promise<string | null> {
	try {
		await refreshToken(true)
		return getToken()
	} catch {
		// Refresh failed. Don't remove the token here — in a multi-tab scenario,
		// another tab may have successfully rotated the refresh token, and clearing
		// localStorage would log out that tab too. Let the caller decide.
		return null
	}
}

/**
 * Returns the `type` claim from a JWT without verifying the signature.
 * Returns null if the token is missing or malformed.
 */
function getTokenType(token: string | null): number | null {
	if (!token) return null
	try {
		const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
		const payload = JSON.parse(atob(base64))
		return typeof payload.type === 'number' ? payload.type : null
	} catch {
		return null
	}
}

export function AuthenticatedHTTPFactory() {
	const instance = HTTPFactory()

	instance.interceptors.request.use((config) => {
		config.headers = {
			...config.headers,
			'Content-Type': 'application/json',
		}

		// Set the default auth header if we have a token
		const token = getToken()
		if (token !== null) {
			config.headers['Authorization'] = `Bearer ${token}`
		}
		return config
	})

	// Response interceptor: on expired JWT 401, attempt a refresh and retry once.
	instance.interceptors.response.use(undefined, async (error) => {
		const originalRequest: AxiosRequestConfig & { _retried?: boolean } = error.config

		// Only intercept 401s, and don't retry a request that already retried.
		if (error.response?.status !== 401 || originalRequest._retried) {
			return Promise.reject(error)
		}

		// Only retry when the 401 is from an expired/invalid JWT. The backend
		// returns error code 11 for this case. Other 401s (disabled account,
		// wrong API token, etc.) are genuine auth failures — retrying would loop.
		const ERROR_CODE_INVALID_TOKEN = 11
		if (error.response?.data?.code !== ERROR_CODE_INVALID_TOKEN) {
			return Promise.reject(error)
		}

		// Don't try to refresh if we don't have a token at all (not logged in),
		// or if the token is a link share JWT (they don't use cookie-based refresh).
		const currentToken = getToken()
		if (!currentToken || getTokenType(currentToken) !== AUTH_TYPES.USER) {
			return Promise.reject(error)
		}

		originalRequest._retried = true

		// Coalesce concurrent refresh attempts into a single request.
		if (!refreshPromise) {
			refreshPromise = doRefresh().finally(() => {
				refreshPromise = null
			})
		}

		const newToken = await refreshPromise
		if (!newToken) {
			// Refresh failed — reject so the UI can redirect to login.
			return Promise.reject(error)
		}

		// Retry the original request with the new token.
		originalRequest.headers = {
			...originalRequest.headers,
			Authorization: `Bearer ${newToken}`,
		}
		return instance.request(originalRequest)
	})

	return instance
}
