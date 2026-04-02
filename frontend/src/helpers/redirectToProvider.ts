import {createRandomID} from '@/helpers/randomId'
import {generateCodeVerifier, generateCodeChallenge} from '@/helpers/pkce'
import type {IProvider} from '@/types/IProvider'
import {parseURL} from 'ufo'

export function getRedirectUrlFromCurrentFrontendPath(provider: IProvider): string {
	// We're not using the redirect url provided by the server to allow redirects when using the electron app.
	// The implications are not quite clear yet hence the logic to pass in another redirect url still exists.
	const url = parseURL(window.location.href)
	return `${url.protocol}//${url.host}/auth/openid/${provider.key}`
}

export const redirectToProvider = async (provider: IProvider) => {

	const redirectUrl = getRedirectUrlFromCurrentFrontendPath(provider)
	const state = createRandomID(24)
	localStorage.setItem('state', state)

	const codeVerifier = generateCodeVerifier()
	const codeChallenge = await generateCodeChallenge(codeVerifier)
	sessionStorage.setItem('pkceCodeVerifier', codeVerifier)

	let scope = 'openid email profile'
	if (provider.scope !== null){
		scope = provider.scope
	}
	window.location.href = `${provider.authUrl}?client_id=${provider.clientId}&redirect_uri=${redirectUrl}&response_type=code&scope=${scope}&state=${state}&code_challenge=${codeChallenge}&code_challenge_method=S256`
}

export const redirectToProviderOnLogout = (provider: IProvider) => {
	if (provider.logoutUrl.length > 0) {
		window.location.href = `${provider.logoutUrl}`
	}
}
