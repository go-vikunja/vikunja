import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

// FIXME: what makes this different from TaskFilterParams?
export type FilterSortField =
	| 'id'
	| 'index'
	| 'done'
	| 'title'
	| 'priority'
	| 'due_date'
	| 'start_date'
	| 'end_date'
	| 'percent_done'
	| 'created'
	| 'updated'
	| 'done_at'
	| 'position'

export interface IFilters {
	sort_by: FilterSortField[],
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
	isFavorite: boolean

	owner: IUser
	created: Date
	updated: Date
}
