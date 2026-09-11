import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IUser} from './IUser'
import type {ISubscription} from './ISubscription'
import type {ProjectView} from '@/client/generated'


export interface IProject extends IAbstract {
	id: number
	title: string
	description: string
	owner: IUser
	tasks: ITask[]
	isArchived: boolean
	hexColor: string
	identifier: string
	backgroundInformation: unknown | null // FIXME: improve type
	isFavorite: boolean
	subscription: ISubscription
	position: number
	backgroundBlurHash: string
	parentProjectId: number
	views: ProjectView[]
	
	created: Date
	updated: Date
}
