import {HTTP} from '../../http-common'
import {ERROR_MESSAGE, LOADING} from "../mutation-types";
import UserModel from "../../models/user";

export default {
	namespaced: true,
	state: () => ({
		authenticated: false,
		isLinkShareAuth: false,
		info: {},
		needsTotpPasscode: false,
	}),
	mutations: {
		info(state, info) {
			state.info = info
		},
		authenticated(state, authenticated) {
			state.authenticated = authenticated
		},
		isLinkShareAuth(state, is) {
			state.isLinkShareAuth = is
		},
		needsTotpPasscode(state, needs) {
			state.needsTotpPasscode = needs
		},
	},
	actions: {
		// Logs a user in with a set of credentials.
		login(ctx, credentials) {
			ctx.commit(LOADING, true, {root: true})

			// Delete an eventually preexisting old token
			localStorage.removeItem('token')

			const data =  {
				username: credentials.username,
				password: credentials.password
			}

			if(credentials.totpPasscode) {
				data.totp_passcode = credentials.totpPasscode
			}

			return HTTP.post('login', data)
				.then(response => {
					// Save the token to local storage for later use
					localStorage.setItem('token', response.data.token)

					// Tell others the user is autheticated
					ctx.commit('isLinkShareAuth', false)
					ctx.dispatch('checkAuth')
					return Promise.resolve()
				})
				.catch(e => {
					if (e.response) {
						if (e.response.data.code === 1017 && !credentials.totpPasscode) {
							ctx.commit('needsTotpPasscode', true)
							return Promise.reject()
						}

						let errorMsg = e.response.data.message
						if (e.response.status === 401) {
							errorMsg = 'Wrong username or password.'
						}
						ctx.commit(ERROR_MESSAGE, errorMsg, {root: true})
					}
					return Promise.reject()
				})
				.finally(() => {
					ctx.commit(LOADING, false, {root: true})
				})
		},
		// Registers a new user and logs them in.
		// Not sure if this is the right place to put the logic in, maybe a seperate js component would be better suited.
		register(ctx, credentials) {
			return HTTP.post('register', {
				username: credentials.username,
				email: credentials.email,
				password: credentials.password
			})
				.then(() => {
					return ctx.dispatch('login', credentials)
				})
				.catch(e => {
					if (e.response) {
						ctx.commit(ERROR_MESSAGE, e.response.data.message, {root: true})
					}
					return Promise.reject()
				})
				.finally(() => {
					ctx.commit(LOADING, false, {root: true})
				})
		},

		linkShareAuth(ctx, hash) {
			return HTTP.post('/shares/' + hash + '/auth')
				.then(r => {
					localStorage.setItem('token', r.data.token)
					ctx.dispatch('checkAuth')
					return Promise.resolve(r.data)
				}).catch(e => {
					return Promise.reject(e)
				})
		},
		// Populates user information from jwt token saved in local storage in store
		checkAuth(ctx) {
			const jwt = localStorage.getItem('token')
			let authenticated = false
			if (jwt) {
				const base64 = jwt
					.split('.')[1]
					.replace('-', '+')
					.replace('_', '/')
				const info = new UserModel(JSON.parse(window.atob(base64)))
				const ts = Math.round((new Date()).getTime() / 1000)
				if (info.exp >= ts) {
					authenticated = true
				}
				ctx.commit('info', info)
			}
			ctx.commit('authenticated', authenticated)
			return Promise.resolve()
		},
		// Renews the api token and saves it to local storage
		renewToken(ctx) {
			if (!ctx.state.authenticated) {
				return
			}

			HTTP.post('user/token', null, {
				headers: {
					Authorization: 'Bearer ' + localStorage.getItem('token'),
				}
			})
				.then(r => {
					localStorage.setItem('token', r.data.token)
					ctx.dispatch('checkAuth')
				})
				.catch(e => {
					// eslint-disable-next-line
					console.log('Error renewing token: ', e)
				})
		},
		logout(ctx) {
			localStorage.removeItem('token')
			ctx.dispatch('checkAuth')
		}
	},
}