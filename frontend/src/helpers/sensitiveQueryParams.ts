import type {RouteLocationNormalizedLoaded} from 'vue-router'

// One-time tokens from email links: a leaked password reset token is an account takeover.
export const SENSITIVE_QUERY_PARAMS = ['userPasswordReset', 'userEmailConfirm', 'accountDeletionConfirm']

// Emails carry the token in the fragment so it never reaches access logs; older emails use the query.
export function getSensitiveParam(route: Pick<RouteLocationNormalizedLoaded, 'query' | 'hash'>, name: string): string | undefined {
	const raw = route.query[name]
	const fromQuery = Array.isArray(raw) ? raw[0] : raw
	if (typeof fromQuery === 'string' && fromQuery !== '') {
		return fromQuery
	}

	return new URLSearchParams(route.hash.slice(1)).get(name) || undefined
}
