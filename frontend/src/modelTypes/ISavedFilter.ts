import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

// FIXME: what makes this different from TaskFilterParams?
export interface IFilters {
	sort_by: ('start_date' | 'done' | 'id' | 'position')[],
	order_by: ('asc' | 'desc')[],
	filter: string,
	filter_include_nulls: boolean,
	s: string,
}

export interface ISavedFilter extends IAbstract {
	id: number
	title: string
	description: string
	filters: IFilters

	owner: IUser
	created: Date
	updated: Date
}
