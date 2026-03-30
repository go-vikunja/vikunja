export interface OAuthTokens {
	access_token: string
	refresh_token: string
	expires_in: number
}

export interface VikunjaDesktop {
	isDesktop: boolean
	startOAuthLogin: (apiUrl: string) => Promise<void>
	onOAuthTokens: (callback: (tokens: OAuthTokens) => void) => void
	onOAuthError: (callback: (error: string) => void) => void
	refreshToken: (apiUrl: string, refreshToken: string) => Promise<OAuthTokens>
}

declare global {
	interface Window {
		vikunjaDesktop?: VikunjaDesktop
	}
}
