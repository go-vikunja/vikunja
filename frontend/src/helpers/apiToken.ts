import type {ApiToken, TokenRoutesResponse} from '@/client/generated'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

export type ApiTokenRoutes = TokenRoutesResponse
export type ApiTokenPresetGroups = Record<string, string[] | '*'>

export interface ApiTokenPreset {
	id: string
	groups: ApiTokenPresetGroups
	label?: string
}

// An unparseable expiry means "no expiry", not "expired".
export function isApiTokenExpired(token: ApiToken, now: number = Date.now()): boolean {
	return (parseDateOrNull(token.expires_at)?.getTime() ?? Infinity) < now
}

// Metadata document clients are named by an https URL; only its host is verified, the name is self-asserted.
export function oauthClientHost(token: ApiToken): string | undefined {
	const id = token.oauth_client_id
	if (!id?.startsWith('https://')) return undefined
	try {
		return new URL(id).host
	} catch {
		return undefined
	}
}
