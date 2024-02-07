export interface IFilter {
	sortBy: ('done' | 'id')[]
	orderBy: ('asc' | 'desc')[]
	filterBy: 'done'[]
	filterValue: 'false'[]
	filterComparator: 'equals'[]
	filterConcat: 'and'
	filterIncludeNulls: boolean
}