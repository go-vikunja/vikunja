import {HTTPFactory} from '@/http-common'
import {LOADING} from '../mutation-types'
import UserModel from '../../models/user'
import {getToken, refreshToken, removeToken, saveToken} from '@/helpers/auth'

const AUTH_TYPES = {
	'UNKNOWN': 0,
	'USER': 1,
	'LINK_SHARE': 2,
}

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
	getters: {
		authUser(state) {
			return state.authenticated && (
				state.info &&
				state.info.type === AUTH_TYPES.USER
			)
		},
		authLinkShare(state) {
			return state.authenticated && (
				state.info &&
				state.info.type === AUTH_TYPES.LINK_SHARE
			)
		},
	},
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
		async login(ctx, credentials) {
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

			try {
				const response = await HTTP.post('login', data)
				// Save the token to local storage for later use
				saveToken(response.data.token, true)
				
				// Tell others the user is autheticated
				ctx.dispatch('checkAuth')
			} catch(e) {
				if (
					e.response &&
					e.response.data.code === 1017 &&
					!credentials.totpPasscode
				) {
					ctx.commit('needsTotpPasscode', true)
				}

				throw e
			} finally {
				ctx.commit(LOADING, false, {root: true})
			}
		},

		// Registers a new user and logs them in.
		// Not sure if this is the right place to put the logic in, maybe a seperate js component would be better suited.
		async register(ctx, credentials) {
			const HTTP = HTTPFactory()
			try {
				await HTTP.post('register', {
					username: credentials.username,
					email: credentials.email,
					password: credentials.password,
				})
				return ctx.dispatch('login', credentials)
			} catch(e) {
				if (e.response?.data?.message) {
					throw e.response.data
				}

				throw e
			} finally {
				ctx.commit(LOADING, false, {root: true})
			}
		},

		async openIdAuth(ctx, {provider, code}) {
			const HTTP = HTTPFactory()
			ctx.commit(LOADING, true, {root: true})

			const data = {
				code: code,
			}

			// Delete an eventually preexisting old token
			removeToken()
			try {
				const response = await HTTP.post(`/auth/openid/${provider}/callback`, data)
				// Save the token to local storage for later use
				saveToken(response.data.token, true)
				
				// Tell others the user is autheticated
				ctx.dispatch('checkAuth')
			} finally {
				ctx.commit(LOADING, false, {root: true})
			}
		},

		async linkShareAuth(ctx, {hash, password}) {
			const HTTP = HTTPFactory()
			const response = await HTTP.post('/shares/' + hash + '/auth', {
				password: password,
			})
			saveToken(response.data.token, false)
			ctx.dispatch('checkAuth')
			return response.data
		},

		// Populates user information from jwt token saved in local storage in store
		checkAuth(ctx) {

			// This function can be called from multiple places at the same time and shortly after one another.
			// To prevent hitting the api too frequently or race conditions, we check at most once per minute.
			if (ctx.state.lastUserInfoRefresh !== null && ctx.state.lastUserInfoRefresh > (new Date()).setMinutes((new Date()).getMinutes() + 1)) {
				return
			}

			const jwt = getToken()
			let authenticated = false
			if (jwt) {
				const base64 = jwt
					.split('.')[1]
					.replace('-', '+')
					.replace('_', '/')
				const info = new UserModel(JSON.parse(atob(base64)))
				const ts = Math.round((new Date()).getTime() / 1000)
				authenticated = info.exp >= ts
				ctx.commit('info', info)

				if (authenticated) {
					ctx.dispatch('refreshUserInfo')
				}
			}

			ctx.commit('authenticated', authenticated)
			if (!authenticated) {
				ctx.commit('info', null)
				ctx.dispatch('config/redirectToProviderIfNothingElseIsEnabled', null, {root: true})
			}
		},

		async refreshUserInfo(ctx) {
			const jwt = getToken()
			if (!jwt) {
				return
			}

			const HTTP = HTTPFactory()
			try {

				const response = await HTTP.get('user', {
					headers: {
						Authorization: `Bearer ${jwt}`,
					},
				})
				const info = new UserModel(response.data)
				info.type = ctx.state.info.type
				info.email = ctx.state.info.email
				info.exp = ctx.state.info.exp
				
				ctx.commit('info', info)
				ctx.commit('lastUserRefresh')
				return info
			} catch(e) {
				throw new Error('Error while refreshing user info:', { cause: e })
			}
		},

		// Renews the api token and saves it to local storage
		renewToken(ctx) {
			// FIXME: Timeout to avoid race conditions when authenticated as a user (=auth token in localStorage) and as a
			// link share in another tab. Without the timeout both the token renew and link share auth are executed at
			// the same time and one might win over the other.
			setTimeout(async () => {
				if (!ctx.state.authenticated) {
					return
				}

				try {
					await refreshToken(!ctx.state.isLinkShareAuth)
					ctx.dispatch('checkAuth')
				} catch(e) {
					// Don't logout on network errors as the user would then get logged out if they don't have
					// internet for a short period of time - such as when the laptop is still reconnecting
					if (e?.request?.status) {
						ctx.dispatch('logout')
					}
				}
			}, 5000)
		},
		logout(ctx) {
			removeToken()
			ctx.dispatch('checkAuth')
		},
	},
}