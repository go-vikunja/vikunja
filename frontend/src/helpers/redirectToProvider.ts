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

export const redirectToProvider = (provider: IProvider) => {

	const redirectUrl = getRedirectUrlFromCurrentFrontendPath(provider)
	const state = createRandomID(24)
	localStorage.setItem('state', state)

	let scope = 'openid email profile'
	if (provider.scope !== null){
		scope = provider.scope
	}
	window.location.href = `${provider.authUrl}?client_id=${provider.clientId}&redirect_uri=${redirectUrl}&response_type=code&scope=${scope}&state=${state}`
}

// JUST_LOGGED_OUT_KEY is read by Login.vue to short-circuit any single-provider
// auto-redirect right after a logout. Without it, an immediate bounce back to the
// IdP would silently re-authenticate the user, defeating the logout entirely.
export const JUST_LOGGED_OUT_KEY = 'justLoggedOut'

export const redirectToProviderOnLogout = (provider: IProvider) => {
	if (!provider.logoutUrl || provider.logoutUrl.length === 0) {
		return
	}

	// Mark that we just logged out so Login.vue skips its auto-redirect when the
	// IdP sends the user back via post_logout_redirect_uri. sessionStorage
	// survives the round-trip to the IdP within the same tab.
	sessionStorage.setItem(JUST_LOGGED_OUT_KEY, '1')

	let target = provider.logoutUrl
	try {
		const url = new URL(provider.logoutUrl)

		// client_id lets the IdP identify the relying party when no id_token_hint
		// is available, so it can skip the "are you sure you want to log out?"
		// prompt for known clients (Authentik does this).
		if (provider.clientId) {
			url.searchParams.set('client_id', provider.clientId)
		}

		// post_logout_redirect_uri tells the IdP where to send the user after
		// signing them out. We send them back to the frontend root, which the
		// router resolves to /login. Combined with JUST_LOGGED_OUT_KEY above,
		// the login page will render normally instead of auto-redirecting.
		const current = parseURL(window.location.href)
		const base = getFullBaseUrl()
		url.searchParams.set('post_logout_redirect_uri', `${current.protocol}//${current.host}${base}`)

		target = url.toString()
	} catch {
		// Fall back to the raw URL if it's not parseable as an absolute URL.
		// We still want logout to navigate even if the admin misconfigured it.
	}

	window.location.href = target
}
