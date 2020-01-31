import {HTTP} from '../http-common'
import router from '../router'
// const API_URL = 'http://localhost:8082/api/v1/'
// const LOGIN_URL = 'http://localhost:8082/login'

export default {

	user: {
		authenticated: false,
		infos: {},
	},

	login(context, creds, redirect) {
		localStorage.removeItem('token') // Delete an eventually preexisting old token

		HTTP.post('login', {
			username: creds.username,
			password: creds.password
		})
			.then(response => {
				// Save the token to local storage for later use
				localStorage.setItem('token', response.data.token)

				// Tell others the user is autheticated
				this.user.authenticated = true
				this.user.isLinkShareAuth = false
				const inf = this.getUserInfos()
				// eslint-disable-next-line
				console.log(inf)

				// Hide the loader
				context.loading = false

				// Redirect if nessecary
				if (redirect) {
					router.push({name: redirect})
				}
			})
			.catch(e => {
				// Hide the loader
				context.loading = false
				if (e.response) {
					context.errorMsg = e.response.data.message
					if (e.response.status === 401) {
						context.errorMsg = 'Wrong username or password.'
					}
				}
			})
	},

	register(context, creds, redirect) {
		HTTP.post('register', {
			username: creds.username,
			email: creds.email,
			password: creds.password
		})
			.then(() => {
				this.login(context, creds, redirect)
			})
			.catch(e => {
				// Hide the loader
				context.loading = false
				if (e.response) {
					context.errorMsg = e.response.data.message
					if (e.response.status === 401) {
						context.errorMsg = 'Wrong username or password.'
					}
				}
			})
	},

	logout() {
		localStorage.removeItem('token')
		router.push({name: 'login'})
		this.user.authenticated = false
	},

	linkShareAuth(hash) {
		return HTTP.post('/shares/' + hash + '/auth')
			.then(r => {
				localStorage.setItem('token', r.data.token)
				this.getUserInfos()
				return Promise.resolve(r.data)
			}).catch(e => {
				return Promise.reject(e)
			})
	},

	renewToken() {
		HTTP.post('user/token', null, {
			headers: {
				Authorization: 'Bearer ' + localStorage.getItem('token'),
			}
		})
			.then(r => {
				localStorage.setItem('token', r.data.token)
			})
			.catch(e => {
				// eslint-disable-next-line
				console.log('Error renewing token: ', e)
			})
	},

	checkAuth() {
		let jwt = localStorage.getItem('token')
		this.getUserInfos()
		this.user.authenticated = false
		if (jwt) {
			let infos = this.user.infos
			let ts = Math.round((new Date()).getTime() / 1000)
			if (infos.exp >= ts) {
				this.user.authenticated = true
			}
		}
	},

	getUserInfos() {
		let jwt = localStorage.getItem('token')
		if (jwt) {
			this.user.infos = this.parseJwt(localStorage.getItem('token'))
			return this.parseJwt(localStorage.getItem('token'))
		} else {
			return {}
		}
	},

	parseJwt(token) {
		let base64Url = token.split('.')[1]
		let base64 = base64Url.replace('-', '+').replace('_', '/')
		return JSON.parse(window.atob(base64))
	},

	getAuthHeader() {
		return {
			'Authorization': 'Bearer ' + localStorage.getItem('token')
		}
	},

	getToken() {
		return localStorage.getItem('token')
	}
}
