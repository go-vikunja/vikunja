import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

interface Filters {
	sortBy: ('start_date' | 'done' | 'id' | 'position')[],
	orderBy: ('asc' | 'desc')[],
	filter: string,
	filterIncludeNulls: boolean,
	s: string,
}

export interface ISavedFilter extends IAbstract {
	id: number
	title: string
	description: string
	filters: Filters

	owner: IUser
	created: Date
	updated: Date
}