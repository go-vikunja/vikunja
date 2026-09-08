import {getFullBaseUrl} from '@/helpers/getFullBaseUrl'
import {
	CODE_VERIFIER_STORAGE_KEY,
	NONCE_STORAGE_KEY,
	createCodeChallenge,
	createCodeVerifier,
	createNonce,
} from '@/helpers/pkce'
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

export const redirectToProvider = async (provider: IProvider) => {

	const redirectUrl = getRedirectUrlFromCurrentFrontendPath(provider)
	const state = createRandomID(24)
	localStorage.setItem('state', state)

	// The nonce binds the ID token to this browser session (OpenID Connect Core 1.0, §3.1.2.1).
	// Some providers reject an authorization request without it.
	const nonce = createNonce()
	localStorage.setItem(NONCE_STORAGE_KEY, nonce)

	let scope = 'openid email profile'
	if (provider.scope !== null){
		scope = provider.scope
	}

	const params = new URLSearchParams({
		client_id: provider.clientId,
		redirect_uri: redirectUrl,
		response_type: 'code',
		scope,
		state,
		nonce,
	})

	// PKCE (RFC 7636). Providers which do not implement it ignore both parameters (§5), so
	// sending them is safe. The challenge needs SHA-256 from crypto.subtle though, which
	// browsers only expose in secure contexts – over plain http we fall back to an
	// authorization request without PKCE instead of breaking the login.
	const codeVerifier = createCodeVerifier()
	const codeChallenge = await createCodeChallenge(codeVerifier)
	if (typeof codeChallenge === 'undefined') {
		localStorage.removeItem(CODE_VERIFIER_STORAGE_KEY)
	} else {
		localStorage.setItem(CODE_VERIFIER_STORAGE_KEY, codeVerifier)
		params.set('code_challenge', codeChallenge)
		params.set('code_challenge_method', 'S256')
	}

	window.location.href = `${provider.authUrl}?${params.toString()}`
}

export const redirectToProviderOnLogout = (provider: IProvider): boolean => {
	if (provider.logoutUrl.length > 0) {
		window.location.href = `${provider.logoutUrl}`
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
