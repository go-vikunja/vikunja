import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {IFilter} from '@/types/IFilter'

export interface ISavedFilter extends IAbstract {
	id: number
	title: string
	description: string
	filters: IFilter

	owner: IUser
	created: Date
	updated: Date
}