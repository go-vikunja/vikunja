import {createRandomID} from '@/helpers/randomId'
import type {IProvider} from '@/types/IProvider'
import {parseURL} from 'ufo'

export function getRedirectUrlFromCurrentFrontendPath(provider: IProvider): string {
	// We're not using the redirect url provided by the server to allow redirects when using the electron app.
	// The implications are not quite clear yet hence the logic to pass in another redirect url still exists.
	const url = parseURL(window.location.href)
	return `${url.protocol}//${url.host}/auth/openid/${provider.key}`
}

export const redirectToProvider = (provider: IProvider) => {

	console.log({provider})

	const redirectUrl = getRedirectUrlFromCurrentFrontendPath(provider)
	const state = createRandomID(24)
	localStorage.setItem('state', state)

	window.location.href = `${provider.authUrl}?client_id=${provider.clientId}&redirect_uri=${redirectUrl}&response_type=code&scope=openid email profile&state=${state}`
}
export const redirectToProviderOnLogout = (provider: IProvider) => {
	if (provider.logoutUrl.length > 0) {
		window.location.href = `${provider.logoutUrl}`
	}
}
