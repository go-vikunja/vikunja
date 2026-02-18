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
])
