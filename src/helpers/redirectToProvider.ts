import {createRandomID} from '@/helpers/randomId'
import type {IProvider} from '@/types/IProvider'
import {parseURL} from 'ufo'

export const redirectToProvider = (provider: IProvider) => {

	// We're not using the redirect url provided by the server to allow redirects when using the electron app.
	// The implications are not quite clear yet hence the logic to pass in another redirect url still exists.
	const url = parseURL(window.location.href)
	const redirectUrl = `${url.protocol}//${url.host}/auth/openid/`

	const state = createRandomID(24)
	localStorage.setItem('state', state)

	window.location.href = `${provider.authUrl}?client_id=${provider.clientId}&redirect_uri=${redirectUrl}${provider.key}&response_type=code&scope=openid email profile&state=${state}`
}
export const redirectToProviderOnLogout = (provider: IProvider) => {
	if (provider.logoutUrl.length > 0) {
		window.location.href = `${provider.logoutUrl}`
	}
}
