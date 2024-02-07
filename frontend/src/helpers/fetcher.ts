import axios from 'axios'
import {getToken} from '@/helpers/auth'

export function HTTPFactory() {
	const instance = axios.create({baseURL: window.API_URL})

	instance.interceptors.request.use((config) => {
		// by setting the baseURL fresh for every request
		// we make sure that it is never outdated in case it is updated
		config.baseURL = window.API_URL

		return config
	})

	return instance
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

	return instance
}
