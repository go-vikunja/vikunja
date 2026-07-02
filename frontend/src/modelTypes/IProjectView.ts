import type {IAbstract} from './IAbstract'
import type {IProject} from '@/modelTypes/IProject'
import type {IFilters} from '@/modelTypes/ISavedFilter'

export const PROJECT_VIEW_KINDS = {
	LIST: 'list',
	GANTT: 'gantt',
	TABLE: 'table',
	KANBAN: 'kanban',
} as const
export type ProjectViewKind = typeof PROJECT_VIEW_KINDS[keyof typeof PROJECT_VIEW_KINDS]

export const DEFAULT_PROJECT_VIEW_SETTINGS = {
	FIRST: 'first',
	...PROJECT_VIEW_KINDS,
} as const
export type DefaultProjectViewKind = typeof DEFAULT_PROJECT_VIEW_SETTINGS[keyof typeof DEFAULT_PROJECT_VIEW_SETTINGS]

export const PROJECT_VIEW_BUCKET_CONFIGURATION_MODES = ['none', 'manual', 'filter']
export type ProjectViewBucketConfigurationMode = typeof PROJECT_VIEW_BUCKET_CONFIGURATION_MODES[number]

export interface IProjectViewBucketConfiguration {
	title: string
	filter: IFilters
}

export interface IProjectView extends IAbstract {
	id: number
	title: string
	projectId: IProject['id']
	viewKind: ProjectViewKind

	filter: IFilters|undefined
	position: number

	bucketConfigurationMode: ProjectViewBucketConfigurationMode
	bucketConfiguration: IProjectViewBucketConfiguration[]
	defaultBucketId: number
	doneBucketId: number

	created: Date
	updated: Date
}
