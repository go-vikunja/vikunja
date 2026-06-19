import type {OAuthTokens} from '@/types/desktop'

export function isDesktopApp(): boolean {
	return !!window.vikunjaDesktop?.isDesktop
}

export function startDesktopOAuthLogin(apiUrl: string): Promise<void> {
	return window.vikunjaDesktop!.startOAuthLogin(apiUrl)
}

export function listenForDesktopOAuthTokens(callback: (tokens: OAuthTokens) => void): void {
	window.vikunjaDesktop!.onOAuthTokens(callback)
}

export function listenForDesktopOAuthError(callback: (error: string) => void): void {
	window.vikunjaDesktop!.onOAuthError(callback)
}

export function refreshDesktopToken(apiUrl: string, refreshToken: string): Promise<OAuthTokens> {
	return window.vikunjaDesktop!.refreshToken(apiUrl, refreshToken)
}
