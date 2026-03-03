import {computed, readonly, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

import {AuthenticatedHTTPFactory, HTTPFactory} from '@/helpers/fetcher'
import {getBrowserLanguage, i18n, setLanguage} from '@/i18n'
import {objectToSnakeCase} from '@/helpers/case'
import UserModel, {getDisplayName, fetchAvatarBlobUrl, invalidateAvatarCache} from '@/models/user'
import AvatarService from '@/services/avatar'
import UserSettingsService from '@/services/userSettings'
import {getToken, refreshToken, removeToken, saveToken} from '@/helpers/auth'
import {setModuleLoading} from '@/stores/helper'
import {success, error} from '@/message'
import {
	getRedirectUrlFromCurrentFrontendPath,
	redirectToProvider,
	redirectToProviderOnLogout,
} from '@/helpers/redirectToProvider'
import {AUTH_TYPES, type IUser} from '@/modelTypes/IUser'
import type {IUserSettings} from '@/modelTypes/IUserSettings'
import router from '@/router'
import {useConfigStore} from '@/stores/config'
import UserSettingsModel from '@/models/userSettings'
import {MILLISECONDS_A_SECOND} from '@/constants/date'
import {PrefixMode} from '@/modules/parseTaskText'
import {DATE_DISPLAY} from '@/constants/dateDisplay'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {RELATION_KIND} from '@/types/IRelationKind'
import type {IProvider} from '@/types/IProvider'

function redirectToSpecifiedProvider() {

	const {auth} = useConfigStore()
	const searchParams = new URLSearchParams(window.location.search)
	if (searchParams.has('redirectToProvider')) {

		const redirectToProviderValue = searchParams.get('redirectToProvider')

		if (
			auth.openidConnect.providers?.length === 1
			&& (window.location.pathname.startsWith('/login') || window.location.pathname === '/') // Kinda hacky, but prevents an endless loop.
			&& (redirectToProviderValue === null
				|| redirectToProviderValue === 'true'
				|| redirectToProviderValue === '1')
 		) {
			redirectToProvider(auth.openidConnect.providers[0])
		}

		// let's try to find the provider to logon to !
		const wantedProvider = auth.openidConnect.providers?.find(p => p.key === redirectToProviderValue)
		if (wantedProvider) {
			redirectToProvider(wantedProvider)
		}
		console.warn(`Could not find provider to redirect to.\nWanted: ${wantedProvider}\nAvailable: ${auth.openidConnect.providers?.map(p => p.key)}`)
	}
}

function getLoggedInVia(): string | null {
	return localStorage.getItem('loggedInViaProvider')
}

function setLoggedInVia(provider: string | null): void {
	if (provider) {
		localStorage.setItem('loggedInViaProvider', provider)
	} else {
		localStorage.removeItem('loggedInViaProvider')
	}
}

export const useAuthStore = defineStore('auth', () => {
	const configStore = useConfigStore()
	
	const authenticated = ref(false)
	const needsTotpPasscode = ref(false)
	
	const info = ref<IUser | null>(null)
	const avatarUrl = ref('')
	const settings = ref<IUserSettings>(new UserSettingsModel())
	
	const currentSessionId = ref<string | null>(null)
	const lastUserInfoRefresh = ref<Date | null>(null)
	const isLoading = ref(false)
	const isLoadingGeneralSettings = ref(false)

	const authUser = computed(() => {
		return authenticated.value && (
			info.value &&
			info.value.type === AUTH_TYPES.USER
		)
	})

	const authLinkShare = computed(() => {
		return authenticated.value && (
			info.value &&
			info.value.type === AUTH_TYPES.LINK_SHARE
		)
	})

	const userDisplayName = computed(() => info.value ? getDisplayName(info.value) : undefined)
	
	const isLinkShareAuth = computed(() => info.value?.type === AUTH_TYPES.LINK_SHARE)

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading 
	}

	function setIsLoadingGeneralSettings(isLoading: boolean) {
		isLoadingGeneralSettings.value = isLoading 
	}

	function setUser(newUser: IUser | null, saveSettings = true) {
		info.value = newUser
		if (newUser !== null && !isLinkShareAuth.value) {
			reloadAvatar()

			if (saveSettings && newUser.settings) {
				loadSettings(newUser.settings)
			}
		}
	}

	function setUserSettings(newSettings: IUserSettings) {
		loadSettings(newSettings)
		info.value = new UserModel({
			...info.value !== null ? info.value : {},
			name: newSettings.name,
		})
	}
	
	function loadSettings(newSettings: IUserSettings) {
		settings.value = new UserSettingsModel({
			...newSettings,
			frontendSettings: {
				// Need to set default settings here in case the user does not have any saved in the api already
				playSoundWhenDone: true,
				quickAddMagicMode: PrefixMode.Default,
				colorSchema: 'auto',
				allowIconChanges: true,
				dateDisplay: DATE_DISPLAY.RELATIVE,
				timeFormat: TIME_FORMAT.HOURS_24,
				defaultTaskRelationType: RELATION_KIND.RELATED,
				backgroundBrightness: 100,
				showIncludeSubprojectsToggle: false,
				sidebarWidth: null,
				commentSortOrder: 'asc',
				...newSettings.frontendSettings,
			},
		})
		// console.log('settings from auth store', {...settings.value.frontendSettings})
	}

	function setAuthenticated(newAuthenticated: boolean) {
		authenticated.value = newAuthenticated
	}


	function setNeedsTotpPasscode(newNeedsTotpPasscode: boolean) {
		needsTotpPasscode.value = newNeedsTotpPasscode
	}

	async function reloadAvatar() {
		if (!info.value || !info.value.username) {
			return
		}
		invalidateAvatarCache(info.value)
		avatarUrl.value = await fetchAvatarBlobUrl(info.value, 40)
	}

	function updateLastUserRefresh() {
		lastUserInfoRefresh.value = new Date()
	}

	// Logs a user in with a set of credentials.
	async function login(credentials) {
		const HTTP = HTTPFactory()
		setIsLoading(true)

		// Delete an eventually preexisting old token
		removeToken()

		try {
			const response = await HTTP.post('login', objectToSnakeCase(credentials))
			// Save the token to local storage for later use
			saveToken(response.data.token, true)

			// Tell others the user is authenticated
			await checkAuth()
		} catch (e) {
			if (
				e.response &&
				e.response.data.code === 1017 &&
				!credentials.totpPasscode
			) {
				setNeedsTotpPasscode(true)
			}

			throw e
		} finally {
			setIsLoading(false)
		}
	}

	/**
	 * Registers a new user and logs them in.
	 * Not sure if this is the right place to put the logic in, maybe a separate js component would be better suited. 
	 */
	async function register(credentials, language: string|null = null) {
		const HTTP = HTTPFactory()
		setIsLoading(true)
		
		if (!language) {
			language = i18n.global.locale.value ?? getBrowserLanguage()
		}
		
		try {
			await HTTP.post('register', {
				...credentials,
				language,
			})
			return login(credentials)
		} catch (e) {
			if (e.response?.data?.code === 2002 && e.response?.data?.invalid_fields[0]?.startsWith('language:')) {
				return register(credentials, 'en')
			}
			
			if (e.response?.data?.message) {
				throw e.response.data
			}

			throw e
		} finally {
			setIsLoading(false)
		}
	}

	async function openIdAuth({provider, code}) {
		const HTTP = HTTPFactory()
		setIsLoading(true)
		setLoggedInVia(null)

		const fullProvider: IProvider = configStore.auth.openidConnect.providers.find((p: IProvider) => p.key === provider)

		const data = {
			code: code,
			redirect_url: getRedirectUrlFromCurrentFrontendPath(fullProvider),
		}

		// Delete an eventually preexisting old token
		removeToken()
		try {
			const response = await HTTP.post(`/auth/openid/${provider}/callback`, data)
			// Save the token to local storage for later use
			saveToken(response.data.token, true)
			setLoggedInVia(provider)

			// Tell others the user is authenticated
			await checkAuth()
		} finally {
			setIsLoading(false)
		}
	}

	async function linkShareAuth({hash, password}) {
		const HTTP = HTTPFactory()
		const response = await HTTP.post('/shares/' + hash + '/auth', {
			password: password,
		})
		saveToken(response.data.token, false)
		await checkAuth()
		return response.data
	}

	/**
	 * Populates user information from jwt token saved in local storage in store
	 */
	async function checkAuth() {
		const now = new Date()
		const oneMinuteAgo = new Date(new Date().setMinutes(now.getMinutes() - 1))
		// This function can be called from multiple places at the same time and shortly after one another.
		// To prevent hitting the api too frequently or race conditions, we check at most once per minute.
		if (
			lastUserInfoRefresh.value !== null &&
			lastUserInfoRefresh.value > oneMinuteAgo
		) {
			return
		}

		const jwt = getToken()
		let isAuthenticated = false
		if (jwt) {
			try {
				const base64 = jwt
					.split('.')[1]
					.replace(/-/g, '+')
					.replace(/_/g, '/')
				const payload = JSON.parse(atob(base64))
				const jwtUser = new UserModel(payload)
				const ts = Math.round((new Date()).getTime() / MILLISECONDS_A_SECOND)

				isAuthenticated = jwtUser.exp >= ts
				currentSessionId.value = payload.sid ?? null

				if (isAuthenticated) {
					// Only set user from JWT if we don't already have a fully loaded
					// user with the same ID. The JWT lacks fields like `name`, so
					// overwriting a complete user object causes a visible flash
					// where the display name briefly reverts to the username.
					if (info.value === null || info.value.id !== jwtUser.id) {
						setUser(jwtUser, false)
					} else {
						// Always keep exp in sync so token renewal checks stay accurate
						info.value.exp = jwtUser.exp
					}
				} else if (jwtUser.type === AUTH_TYPES.USER) {
					// JWT expired but this is a user session — attempt a cookie-based
					// refresh before giving up. This lets users who reopen the app
					// after the short JWT TTL seamlessly resume their session.
					try {
						await refreshToken(true)
						const freshJwt = getToken()
						if (freshJwt) {
							const b64 = freshJwt.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
							const p = JSON.parse(atob(b64))
							const freshUser = new UserModel(p)
							isAuthenticated = freshUser.exp >= ts
							currentSessionId.value = p.sid ?? null
							if (info.value === null || info.value.id !== freshUser.id) {
								setUser(freshUser, false)
							} else {
								info.value.exp = freshUser.exp
							}
						}
					} catch {
						// Refresh failed — stay unauthenticated
					}
				}
			} catch (_) {
				logout()
			}

			if (isAuthenticated) {
				await refreshUserInfo()
			}
		}

		setAuthenticated(isAuthenticated)
		if (!isAuthenticated) {
			setUser(null)
			redirectToSpecifiedProvider()
		}
		
		return Promise.resolve(authenticated)
	}

	async function refreshUserInfo() {
		const jwt = getToken()
		if (!jwt) {
			return
		}

		const HTTP = AuthenticatedHTTPFactory()
		try {
			const response = await HTTP.get('user')
			const newUser = new UserModel({
				...response.data,
				...(info.value?.type && {type: info.value?.type}),
				...(info.value?.exp && {exp: info.value?.exp}),
			})

			if (newUser.settings.language) {
				await setLanguage(newUser.settings.language)
			}

			setUser(newUser)
			updateLastUserRefresh()

			return newUser
		} catch (e) {
			if((e?.response?.status >= 400 && e?.response?.status < 500) ||
				e?.response?.data?.message === 'missing, malformed, expired or otherwise invalid token provided') {
				await logout()
				return
			}
			
			const cause = {e}
			
			if (typeof e?.response?.data?.message !== 'undefined') {
				cause.message = e.response.data.message
			}
			
			console.error('Error refreshing user info:', e)
			
			throw new Error('Error while refreshing user info:', {cause})
		}
	}

	/**
	 * Try to verify the email
	 */
	async function verifyEmail(): Promise<boolean> {
		const emailVerifyToken = localStorage.getItem('emailConfirmToken')
		if (emailVerifyToken) {
			const stopLoading = setModuleLoading(setIsLoading)
			try {
				await HTTPFactory().post('user/confirm', {token: emailVerifyToken})
				return true
			} catch(e) {
				throw new Error(e.response.data.message)
			} finally {
				localStorage.removeItem('emailConfirmToken')
				stopLoading()
			}
		}
		return false
	}

	async function saveUserSettings({
		settings,
		showMessage = true,
	}: {
		settings: IUserSettings,
		showMessage: boolean,
	}) {
		const userSettingsService = new UserSettingsService()

		const cancel = setModuleLoading(setIsLoadingGeneralSettings)
		try {
			const oldName = info.value?.name
			let settingsUpdate = {...settings}
			if (configStore.demoModeEnabled) {
				settingsUpdate = {
					...settingsUpdate,
					language: null,
				}
			}
			const updateSettingsPromise = userSettingsService.update(settingsUpdate)
			setUserSettings(settingsUpdate)
			await setLanguage(settings.language)
			await updateSettingsPromise
			if (oldName !== undefined && oldName !== settingsUpdate.name) {
				const {avatarProvider} = await (new AvatarService()).get({})
				if (avatarProvider === 'initials') {
					await reloadAvatar()
				}
			}
			if (showMessage) {
				success({message: i18n.global.t('user.settings.general.savedSuccess')})
			}
		} catch (e) {
			error(e)
		} finally {
			cancel()
		}
	}

	/**
	 * Renews the api token and saves it to local storage
	 */
	async function renewToken() {
		if (!authenticated.value) {
			return
		}

		try {
			if (isLinkShareAuth.value) {
				// Link shares renew via the dedicated link-share endpoint (JWT-based).
				const HTTP = AuthenticatedHTTPFactory()
				const response = await HTTP.post('user/token')
				saveToken(response.data.token, false)
			} else {
				// User sessions renew via the refresh-token cookie.
				await refreshToken(true)
			}
			await checkAuth()
		} catch (e) {
			// Only logout if the JWT has actually expired and we can't refresh.
			// If the JWT is still valid, the proactive refresh failure is harmless
			// — the 401 interceptor will handle it when the token really expires.
			const nowInSeconds = Date.now() / MILLISECONDS_A_SECOND
			const isExpired = !info.value?.exp || info.value.exp < nowInSeconds
			if (isExpired && (e?.cause?.request?.status || e?.cause?.response?.status)) {
				await logout()
			}
		}
	}

	async function logout() {
		// Revoke the server session so the refresh token can't be reused.
		// Best-effort: if the network call fails, still clean up locally.
		try {
			const HTTP = AuthenticatedHTTPFactory()
			await HTTP.post('user/logout')
		} catch (_e) {
			// Ignore — session will expire naturally
		}

		removeToken()
		const loggedInVia = getLoggedInVia()
		window.localStorage.clear() // Clear all settings and history we might have saved in local storage.
		lastUserInfoRefresh.value = null
		await router.push({name: 'user.login'})
		await checkAuth()

		// if configured, redirect to OIDC Provider on logout
		const fullProvider: IProvider|undefined = configStore.auth.openidConnect.providers?.find((p: IProvider) => p.key === loggedInVia)
		if (fullProvider) {
			redirectToProviderOnLogout(fullProvider)
		}
	}

	return {
		// state
		authenticated: readonly(authenticated),
		needsTotpPasscode: readonly(needsTotpPasscode),

		info: readonly(info),
		avatarUrl: readonly(avatarUrl),
		settings: readonly(settings),

		currentSessionId: readonly(currentSessionId),
		lastUserInfoRefresh: readonly(lastUserInfoRefresh),

		authUser,
		authLinkShare,
		userDisplayName,
		isLinkShareAuth,

		isLoading: readonly(isLoading),
		setIsLoading,

		isLoadingGeneralSettings: readonly(isLoadingGeneralSettings),
		setIsLoadingGeneralSettings,

		setUser,
		setUserSettings,
		setAuthenticated,
		setNeedsTotpPasscode,

		reloadAvatar,
		updateLastUserRefresh,

		login,
		register,
		openIdAuth,
		linkShareAuth,
		checkAuth,
		refreshUserInfo,
		verifyEmail,
		saveUserSettings,
		renewToken,
		logout,
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useAuthStore, import.meta.hot))
}
