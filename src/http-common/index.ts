import axios from 'axios'
import {getToken} from '@/helpers/auth'

export function HTTPFactory() {
	return axios.create({
		baseURL: window.API_URL,
	})
}

export function AuthenticatedHTTPFactory(token = getToken()) {	
	return axios.create({
		baseURL: window.API_URL,
		headers: {
			Authorization: `Bearer ${token}`,
			'Content-Type': 'application/json',
		},
	})
}
