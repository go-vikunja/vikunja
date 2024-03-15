import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IUser} from './IUser'
import type {ISubscription} from './ISubscription'
import type {IProject} from '@/modelTypes/IProject'


export interface IProjectView extends IAbstract {
	id: number
	title: string
	projectId: IProject['id']
	viewKind: 'list' | 'gantt' | 'table' | 'kanban'
	
	fitler: string
	position: number
	
	bucketConfigurationMode: 'none' | 'manual' | 'filter'
	bucketConfiguration: object
	defaultBucketId: number
	doneBucketId: number
	
	created: Date
	updated: Date
}