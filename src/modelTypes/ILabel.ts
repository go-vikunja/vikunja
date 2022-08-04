import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

export interface ILabel extends IAbstract {
	id: number
	title: string
	hexColor: string
	description: string
	createdBy: IUser
	listId: number
	textColor: string

	created: Date
	updated: Date
}