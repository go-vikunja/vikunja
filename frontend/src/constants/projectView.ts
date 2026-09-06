import type {ProjectView} from '@/client/generated'

export type ProjectViewKind = NonNullable<ProjectView['view_kind']>

export const PROJECT_VIEW_KINDS = {
	LIST: 'list',
	GANTT: 'gantt',
	TABLE: 'table',
	KANBAN: 'kanban',
} as const satisfies Record<string, ProjectViewKind>

export const DEFAULT_PROJECT_VIEW_SETTINGS = {
	FIRST: 'first',
	...PROJECT_VIEW_KINDS,
} as const
export type DefaultProjectViewKind = typeof DEFAULT_PROJECT_VIEW_SETTINGS[keyof typeof DEFAULT_PROJECT_VIEW_SETTINGS]

export const PROJECT_VIEW_BUCKET_CONFIGURATION_MODES = ['none', 'manual', 'filter'] as const satisfies readonly NonNullable<ProjectView['bucket_configuration_mode']>[]
