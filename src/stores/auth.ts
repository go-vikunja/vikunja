import {defineStore, acceptHMRUpdate} from 'pinia'

import {HTTPFactory, AuthenticatedHTTPFactory} from '@/http-common'
import {i18n, getCurrentLanguage, saveLanguage} from '@/i18n'
import {objectToSnakeCase} from '@/helpers/case'
import UserModel, { getAvatarUrl } from '@/models/user'
import UserSettingsService from '@/services/userSettings'
import {getToken, refreshToken, removeToken, saveToken} from '@/helpers/auth'
import {setLoadingPinia} from '@/store/helper'
import {success} from '@/message'
import {redirectToProvider} from '@/helpers/redirectToProvider'
import {AUTH_TYPES, type IUser} from '@/modelTypes/IUser'
import type {AuthState} from '@/store/types'
import type {IUserSettings} from '@/modelTypes/IUserSettings'
import router from '@/router'
import {useConfigStore} from '@/stores/config'
import UserSettingsModel from '@/models/userSettings'
import {store} from '@/store'

export const useAuthStore = defineStore('auth', {
	state: () : AuthState => ({
		authenticated: false,
		isLinkShareAuth: false,
		needsTotpPasscode: false,
		
		info: null,
		avatarUrl: '',
		settings: new UserSettingsModel(),
		
		lastUserInfoRefresh: null,
		isLoading: false,
		isLoadingGeneralSettings: false,
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
	actions: {
		setIsLoading(isLoading: boolean) {
			this.isLoading = isLoading 
		},

		setIsLoadingGeneralSettings(isLoading: boolean) {
			this.isLoadingGeneralSettings = isLoading 
		},

		setUser(info: IUser | null) {
			this.info = info
			if (info !== null) {
				this.reloadAvatar()

				if (info.settings) {
					this.settings = new UserSettingsModel(info.settings)
				}

				this.isLinkShareAuth = info.id < 0
			}
		},
		setUserSettings(settings: IUserSettings) {
			this.settings = new UserSettingsModel(settings)
			this.info = new UserModel({
				...this.info !== null ? this.info : {},
				name: settings.name,
			})
		},
		setAuthenticated(authenticated: boolean) {
			this.authenticated = authenticated
		},
		setIsLinkShareAuth(isLinkShareAuth: boolean) {
			this.isLinkShareAuth = isLinkShareAuth
		},
		setNeedsTotpPasscode(needsTotpPasscode: boolean) {
			this.needsTotpPasscode = needsTotpPasscode
		},
		reloadAvatar() {
			if (!this.info) return
			this.avatarUrl = `${getAvatarUrl(this.info)}&=${+new Date()}`
		},
		updateLastUserRefresh() {
			this.lastUserInfoRefresh = new Date()
		},

		// Logs a user in with a set of credentials.
		async login(credentials) {
			const HTTP = HTTPFactory()
			store.commit('loading', true)
			this.setIsLoading(true)

			// Delete an eventually preexisting old token
			removeToken()

			try {
				const response = await HTTP.post('login', objectToSnakeCase(credentials))
				// Save the token to local storage for later use
				saveToken(response.data.token, true)

				// Tell others the user is autheticated
				this.checkAuth()
			} catch (e) {
				if (
					e.response &&
					e.response.data.code === 1017 &&
					!credentials.totpPasscode
				) {
					this.setNeedsTotpPasscode(true)
				}

				throw e
			} finally {
				store.commit('loading', false)
				this.setIsLoading(false)
			}
		},

		// Registers a new user and logs them in.
		// Not sure if this is the right place to put the logic in, maybe a seperate js component would be better suited.
		async register(credentials) {
			const HTTP = HTTPFactory()
			store.commit('loading', true)
			this.setIsLoading(true)
			try {
				await HTTP.post('register', credentials)
				return this.login(credentials)
			} catch (e) {
				if (e.response?.data?.message) {
					throw e.response.data
				}

				throw e
			} finally {
			store.commit('loading', false)
				this.setIsLoading(false)
			}
		},

		async openIdAuth({provider, code}) {
			const HTTP = HTTPFactory()
			store.commit('loading', true)
			this.setIsLoading(true)

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
				this.checkAuth()
			} finally {
				store.commit('loading', false)
				this.setIsLoading(false)
			}
		},

		async linkShareAuth({hash, password}) {
			const HTTP = HTTPFactory()
			const response = await HTTP.post('/shares/' + hash + '/auth', {
				password: password,
			})
			saveToken(response.data.token, false)
			this.checkAuth()
			return response.data
		},

		// Populates user information from jwt token saved in local storage in store
		checkAuth() {

			// This function can be called from multiple places at the same time and shortly after one another.
			// To prevent hitting the api too frequently or race conditions, we check at most once per minute.
			if (
				this.lastUserInfoRefresh !== null &&
				this.lastUserInfoRefresh > (new Date()).setMinutes((new Date()).getMinutes() + 1)
			) {
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
				this.setUser(info)

				if (authenticated) {
					this.refreshUserInfo()
				}
			}

			this.setAuthenticated(authenticated)
			if (!authenticated) {
				this.setUser(null)
				this.redirectToProviderIfNothingElseIsEnabled()
			}
		},

		redirectToProviderIfNothingElseIsEnabled() {
			const {auth} = useConfigStore()
			if (
				auth.local.enabled === false &&
				auth.openidConnect.enabled &&
				auth.openidConnect.providers?.length === 1 &&
				window.location.pathname.startsWith('/login') // Kinda hacky, but prevents an endless loop.
			) {
				redirectToProvider(auth.openidConnect.providers[0], auth.openidConnect.redirectUrl)
			}
		},

		async refreshUserInfo() {
			const jwt = getToken()
			if (!jwt) {
				return
			}

			const HTTP = AuthenticatedHTTPFactory()
			try {
				const response = await HTTP.get('user')
				const info = new UserModel({
					...response.data,
					...(this.info?.type && {type: this.info?.type}),
					...(this.info?.email && {email: this.info?.email}),
					...(this.info?.exp && {exp: this.info?.exp}),
				})

				this.setUser(info)
				this.updateLastUserRefresh()

				if (
						info.type === AUTH_TYPES.USER &&
						(
							typeof info.settings.language === 'undefined' ||
							info.settings.language === ''
						)
				) {
					// save current language
					await this.saveUserSettings({
						settings: {
							...this.settings,
							language: getCurrentLanguage(),
						},
						showMessage: false,
					})
				}

				return info
			} catch (e) {
				if(e?.response?.data?.message === 'invalid or expired jwt') {
					this.logout()
					return
				}
				throw new Error('Error while refreshing user info:', {cause: e})
			}
		},

		async saveUserSettings({
			settings,
			showMessage = true,
		}: {
			settings: IUserSettings
			showMessage : boolean
		}) {
			const userSettingsService = new UserSettingsService()

			// FIXME
			const cancel = setLoadingPinia(this, this.setIsLoadingGeneralSettings)
			try {
				saveLanguage(settings.language)
				await userSettingsService.update(settings)
				this.setUserSettings({...settings})
				if (showMessage) {
					success({message: i18n.global.t('user.settings.general.savedSuccess')})
				}
			} catch (e) {
				throw new Error('Error while saving user settings:', {cause: e})
			} finally {
				cancel()
			}
		},

		// Renews the api token and saves it to local storage
		renewToken() {
			// FIXME: Timeout to avoid race conditions when authenticated as a user (=auth token in localStorage) and as a
			// link share in another tab. Without the timeout both the token renew and link share auth are executed at
			// the same time and one might win over the other.
			setTimeout(async () => {
				if (!this.authenticated) {
					return
				}

				try {
					await refreshToken(!this.isLinkShareAuth)
					this.checkAuth()
				} catch (e) {
					// Don't logout on network errors as the user would then get logged out if they don't have
					// internet for a short period of time - such as when the laptop is still reconnecting
					if (e?.request?.status) {
						this.logout()
					}
				}
			}, 5000)
		},

		logout() {
			removeToken()
			window.localStorage.clear() // Clear all settings and history we might have saved in local storage.
			router.push({name: 'user.login'})
			this.checkAuth()
		},
	},
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useAuthStore, import.meta.hot))
}