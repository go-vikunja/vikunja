import {HTTPFactory} from '@/http-common'
import {ERROR_MESSAGE, LOADING} from '../mutation-types'
import UserModel from '../../models/user'
import {getToken, refreshToken, removeToken, saveToken} from '@/helpers/auth'

const defaultSettings = settings => {
	if (typeof settings.weekStart === 'undefined' || settings.weekStart === '') {
		settings.weekStart = 0
	}
	return settings
}

export default {
	namespaced: true,
	state: () => ({
		authenticated: false,
		isLinkShareAuth: false,
		info: null,
		needsTotpPasscode: false,
		avatarUrl: '',
		lastUserInfoRefresh: null,
		settings: {},
	}),
	mutations: {
		info(state, info) {
			state.info = info
			if (info !== null) {
				state.avatarUrl = info.getAvatarUrl()

				if (info.settings) {
					state.settings = defaultSettings(info.settings)
				}

				state.isLinkShareAuth = info.id < 0
			}
		},
		setUserSettings(state, settings) {
			state.settings = defaultSettings(settings)
			const info = state.info !== null ? state.info : {}
			info.name = settings.name
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
		reloadAvatar(state) {
			state.avatarUrl = `${state.info.getAvatarUrl()}&=${+new Date()}`
		},
		lastUserRefresh(state) {
			state.lastUserInfoRefresh = new Date()
		},
	},
	actions: {
		// Logs a user in with a set of credentials.
		login(ctx, credentials) {
			const HTTP = HTTPFactory()
			ctx.commit(LOADING, true, {root: true})

			// Delete an eventually preexisting old token
			removeToken()

			const data = {
				username: credentials.username,
				password: credentials.password,
			}

			if (credentials.totpPasscode) {
				data.totp_passcode = credentials.totpPasscode
			}

			return HTTP.post('login', data)
				.then(response => {
					// Save the token to local storage for later use
					saveToken(response.data.token, true)

					// Tell others the user is autheticated
					ctx.dispatch('checkAuth')
					return Promise.resolve()
				})
				.catch(e => {
					if (e.response) {
						if (e.response.data.code === 1017 && !credentials.totpPasscode) {
							ctx.commit('needsTotpPasscode', true)
							return Promise.reject(e)
						}
					}

					return Promise.reject(e)
				})
				.finally(() => {
					ctx.commit(LOADING, false, {root: true})
				})
		},
		// Registers a new user and logs them in.
		// Not sure if this is the right place to put the logic in, maybe a seperate js component would be better suited.
		register(ctx, credentials) {
			const HTTP = HTTPFactory()
			return HTTP.post('register', {
				username: credentials.username,
				email: credentials.email,
				password: credentials.password,
			})
				.then(() => {
					return ctx.dispatch('login', credentials)
				})
				.catch(e => {
					if (e.response && e.response.data && e.response.data.message) {
						ctx.commit(ERROR_MESSAGE, e.response.data.message, {root: true})
					}

					return Promise.reject(e)
				})
				.finally(() => {
					ctx.commit(LOADING, false, {root: true})
				})
		},
		openIdAuth(ctx, {provider, code}) {
			const HTTP = HTTPFactory()
			ctx.commit(LOADING, true, {root: true})

			const data = {
				code: code,
			}

			// Delete an eventually preexisting old token
			removeToken()
			return HTTP.post(`/auth/openid/${provider}/callback`, data)
				.then(response => {
					// Save the token to local storage for later use
					saveToken(response.data.token, true)

					// Tell others the user is autheticated
					ctx.dispatch('checkAuth')
					return Promise.resolve()
				})
				.catch(e => {
					return Promise.reject(e)
				})
				.finally(() => {
					ctx.commit(LOADING, false, {root: true})
				})
		},
		linkShareAuth(ctx, {hash, password}) {
			const HTTP = HTTPFactory()
			return HTTP.post('/shares/' + hash + '/auth', {
				password: password,
			})
				.then(r => {
					saveToken(r.data.token, false)
					ctx.dispatch('checkAuth')
					return Promise.resolve(r.data)
				}).catch(e => {
					return Promise.reject(e)
				})
		},
		// Populates user information from jwt token saved in local storage in store
		checkAuth(ctx) {

			// This function can be called from multiple places at the same time and shortly after one another.
			// To prevent hitting the api too frequently or race conditions, we check at most once per minute.
			if (ctx.state.lastUserInfoRefresh !== null && ctx.state.lastUserInfoRefresh > (new Date()).setMinutes((new Date()).getMinutes() + 1)) {
				return Promise.resolve()
			}

			const jwt = getToken()
			let authenticated = false
			if (jwt) {
				const base64 = jwt
					.split('.')[1]
					.replace('-', '+')
					.replace('_', '/')
				const info = new UserModel(JSON.parse(window.atob(base64)))
				const ts = Math.round((new Date()).getTime() / 1000)
				authenticated = info.exp >= ts
				ctx.commit('info', info)

				if (authenticated) {
					ctx.dispatch('refreshUserInfo')
					ctx.commit('authenticated', authenticated)
				}
			}

			ctx.commit('authenticated', authenticated)
			if (!authenticated) {
				ctx.commit('info', null)
				ctx.dispatch('config/redirectToProviderIfNothingElseIsEnabled', null, {root: true})
			}

			return Promise.resolve()
		},
		refreshUserInfo(ctx) {
			const jwt = getToken()
			if (!jwt) {
				return
			}

			const HTTP = HTTPFactory()
			// We're not returning the promise here to prevent blocking the initial ui render if the user is
			// accessing the site with a token in local storage
			HTTP.get('user', {
				headers: {
					Authorization: `Bearer ${jwt}`,
				},
			})
				.then(r => {
					const info = new UserModel(r.data)
					info.type = ctx.state.info.type
					info.email = ctx.state.info.email
					info.exp = ctx.state.info.exp

					ctx.commit('info', info)
					ctx.commit('lastUserRefresh')
				})
				.catch(e => {
					console.error('Error while refreshing user info:', e)
				})
		},
		// Renews the api token and saves it to local storage
		renewToken(ctx) {
			// Timeout to avoid race conditions when authenticated as a user (=auth token in localStorage) and as a
			// link share in another tab. Without the timeout both the token renew and link share auth are executed at
			// the same time and one might win over the other.
			setTimeout(() => {
				if (!ctx.state.authenticated) {
					return
				}

				refreshToken(!ctx.state.isLinkShareAuth)
					.then(() => {
						ctx.dispatch('checkAuth')
					})
					.catch(e => {
						// Don't logout on network errors as the user would then get logged out if they don't have
						// internet for a short period of time - such as when the laptop is still reconnecting
						if (e.request.status) {
							ctx.dispatch('logout')
						}
					})
			}, 5000)
		},
		logout(ctx) {
			removeToken()
			ctx.dispatch('checkAuth')
		},
	},
}