import type {IAbstract} from './IAbstract'
import type {IProject} from './IProject'
import type {IUser} from './IUser'
import type {ISubscription} from './ISubscription'

export interface INamespace extends IAbstract {
	id: number
	title: string
	description: string
	owner: IUser
	projects: IProject[]
	isArchived: boolean
	hexColor: string
	subscription: ISubscription

	created: Date
	updated: Date
}