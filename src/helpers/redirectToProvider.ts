interface Provider {
	name: string
	key: string
	authUrl: string
	clientId: string
}

export const redirectToProvider = (provider: Provider, redirectUrl: string) => {
	const state = Math.random().toString(36).substring(2, 24)
	localStorage.setItem('state', state)

	window.location.href = `${provider.authUrl}?client_id=${provider.clientId}&redirect_uri=${redirectUrl}${provider.key}&response_type=code&scope=openid email profile&state=${state}`
}
