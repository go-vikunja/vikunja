import type {TokenRoutesResponse} from '@/client/generated'

export type ApiTokenRoutes = TokenRoutesResponse
export type ApiTokenPresetGroups = Record<string, string[] | '*'>

export interface ApiTokenPreset {
	id: string
	groups: ApiTokenPresetGroups
	label?: string
}
