export type ApiTokenRoutes = Record<string, Record<string, {path: string, method: string}>>
export type ApiTokenPresetGroups = Record<string, string[] | '*'>

export interface ApiTokenPreset {
	id: string
	groups: ApiTokenPresetGroups
	label?: string
}

export interface IMcpInfo {
	endpoint: string
	routes: ApiTokenRoutes
	presets: {
		read_only: ApiTokenPresetGroups
		typed: ApiTokenPresetGroups
		full: ApiTokenPresetGroups
	}
}
