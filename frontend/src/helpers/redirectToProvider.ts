import {getFullBaseUrl} from '@/helpers/getFullBaseUrl'
import {createRandomID} from '@/helpers/randomId'
import type {IProvider} from '@/types/IProvider'
import {parseURL} from 'ufo'

export function getRedirectUrlFromCurrentFrontendPath(provider: IProvider): string {
	// We're not using the redirect url provided by the server to allow redirects when using the electron app.
	// The implications are not quite clear yet hence the logic to pass in another redirect url still exists.
	const url = parseURL(window.location.href)
	const base = getFullBaseUrl()
	return `${url.protocol}//${url.host}${base}auth/openid/${provider.key}`
}

const LINKING_KEY = 'openidLinking'

function isNonEmptyString(state: unknown): state is string {
	return typeof state === 'string' && state !== ''
}

export function isLinkingIdentity(state: unknown): boolean {
	return isNonEmptyString(state) && localStorage.getItem(LINKING_KEY) === state
}

// Consumes only this flow's entries: Back/reload then fails the state check, other tabs' flows survive.
export function endLinkFlow(state: unknown) {
	if (!isNonEmptyString(state)) {
		return
	}
	for (const key of [LINKING_KEY, 'state']) {
		if (localStorage.getItem(key) === state) {
			localStorage.removeItem(key)
		}
	}
}

// Linking reuses the login redirect URI so admins don't have to register a second one at the provider.
export const redirectToProvider = (provider: IProvider, {linking = false}: {linking?: boolean} = {}) => {
	const redirectUrl = getRedirectUrlFromCurrentFrontendPath(provider)
	const state = createRandomID(24)
	localStorage.setItem('state', state)
	if (linking) {
		localStorage.setItem(LINKING_KEY, state)
	} else {
		localStorage.removeItem(LINKING_KEY)
	}

	window.location.href = `${provider.auth_url}?client_id=${provider.client_id}&redirect_uri=${redirectUrl}&response_type=code&scope=${provider.scope}&state=${state}${linking ? '&prompt=select_account' : ''}`
}

export const redirectToProviderOnLogout = (provider: IProvider): boolean => {
	if (provider.logout_url.length > 0) {
		window.location.href = `${provider.logout_url}`
		return true
	}
	return false
}

interface AutoRedirectContext {
	localAuthEnabled: boolean
	ldapAuthEnabled: boolean
	openIdEnabled: boolean
	providers: IProvider[]
	isDesktopApp: boolean
	justLoggedOut: boolean
	hasCopyableRedirect: boolean
}

/**
 * The provider the login page should redirect to without the user clicking anything,
 * or undefined when it must render the login form instead.
 */
export function getAutoRedirectProvider(ctx: AutoRedirectContext): IProvider | undefined {
	// The Electron window hands login off to the system browser via DesktopLogin – redirecting
	// to the provider in-window would strand the user there with no way back to the app.
	if (ctx.isDesktopApp) {
		return undefined
	}

	// Otherwise we'd immediately re-authenticate the user we just logged out.
	if (ctx.justLoggedOut) {
		return undefined
	}

	// A native client's authorize URL is parked in the login hash so it stays copyable into the
	// browser the user is actually signed in to (#2654). Redirecting to the provider replaces it
	// before it can be copied, and the provider URL itself is not transferable: the OIDC state
	// lives in this browser's localStorage, so finishing the flow elsewhere fails the state check.
	if (ctx.hasCopyableRedirect) {
		return undefined
	}

	if (ctx.localAuthEnabled || ctx.ldapAuthEnabled) {
		return undefined
	}

	if (!ctx.openIdEnabled || ctx.providers.length !== 1) {
		return undefined
	}

	return ctx.providers[0]
}
