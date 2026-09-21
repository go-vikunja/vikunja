import type {ApiToken, TokenRoutesResponse} from '@/client/generated'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

export type ApiTokenRoutes = TokenRoutesResponse
export type ApiTokenPresetGroups = Record<string, string[] | '*'>

export interface ApiTokenPreset {
	id: string
	groups: ApiTokenPresetGroups
	label?: string
}

// An expiry the boundary parser rejects means "no expiry", not "expired".
export function isApiTokenExpired(token: ApiToken, now: number = Date.now()): boolean {
	return (parseDateOrNull(token.expires_at)?.getTime() ?? Infinity) < now
}
