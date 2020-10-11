import axios from 'axios'

export const HTTPFactory = () => {
	return axios.create({
		baseURL: window.API_URL,
	})
}
