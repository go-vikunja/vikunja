import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IUser} from './IUser'
import type {ISubscription} from './ISubscription'
import type {INamespace} from './INamespace'


export interface IProject extends IAbstract {
	id: number
	title: string
	description: string
	owner: IUser
	tasks: ITask[]
	namespaceId: INamespace['id']
	isArchived: boolean
	hexColor: string
	identifier: string
	backgroundInformation: unknown | null // FIXME: improve type
	isFavorite: boolean
	subscription: ISubscription
	position: number
	backgroundBlurHash: string
	
	created: Date
	updated: Date
}