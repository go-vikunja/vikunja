/**
 * Route names for authentication pages that don't require (and shouldn't show)
 * the authenticated app shell. Used by App.vue to gate the layout switch and
 * by the router guard to identify routes that don't need authentication.
 */
export const AUTH_ROUTE_NAMES = new Set([
	'user.login',
	'user.register',
	'user.password-reset.request',
	'user.password-reset.reset',
	'link-share.auth',
	'openid.auth',
	// Shown to a signed-in user who has to choose a new password. Listing it here renders it in the
	// bare wrapper instead of the app shell, whose start-up requests the server refuses until the
	// password was changed.
	'user.change-password',
])

// Pages that render alone, without the app shell. Used by the print view.
export const BARE_ROUTE_NAMES = new Set<string>(['project.print'])
