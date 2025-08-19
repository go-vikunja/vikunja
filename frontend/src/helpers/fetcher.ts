import axios from 'axios'
import {getToken} from '@/helpers/auth'

export function HTTPFactory() {
	return axios.create()
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
