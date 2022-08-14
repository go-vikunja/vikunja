import AbstractModel, { type IAbstract } from '@/models/abstractModel'
import UserModel, { type IUser } from '@/models/user'

export interface ISavedFilter extends IAbstract {
	id: number
	title: string
	description: string
	filters: {
		sortBy: ('done' | 'id')[]
		orderBy: ('asc' | 'desc')[]
		filterBy: 'done'[]
		filterValue: 'false'[]
		filterComparator: 'equals'[]
		filterConcat: 'and'
		filterIncludeNulls: boolean
	}

	owner: IUser
	created: Date
	updated: Date
}

export default class SavedFilterModel extends AbstractModel implements ISavedFilter {
	id = 0
	title = ''
	description = ''
	filters: {
		sortBy: ('done' | 'id')[]
		orderBy: ('asc' | 'desc')[]
		filterBy: 'done'[]
		filterValue: 'false'[]
		filterComparator: 'equals'[]
		filterConcat: 'and'
		filterIncludeNulls: boolean
	} = {
		sortBy: ['done', 'id'],
		orderBy: ['asc', 'desc'],
		filterBy: ['done'],
		filterValue: ['false'],
		filterComparator: ['equals'],
		filterConcat: 'and',
		filterIncludeNulls: true,
	}

	owner: IUser = {}
	created: Date = null
	updated: Date = null

	constructor(data: Partial<ISavedFilter>) {
		super()
		this.assignData(data)

		this.owner = new UserModel(this.owner)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	/**
	 * Calculates the corresponding list id to this saved filter.
	 * This function matches the one in the api.
	 * @returns {number}
	 */
	getListId() {
		let listId = this.id * -1 - 1
		if (listId > 0) {
			listId = 0
		}
		return listId
	}
}
