import type {RouteDetail} from '@/client/generated'

export type ApiTokenRoutes = Record<string, Record<string, RouteDetail>>
export type ApiTokenPresetGroups = Record<string, string[] | '*'>

export interface ApiTokenPreset {
	id: string
	groups: ApiTokenPresetGroups
	label?: string
}
