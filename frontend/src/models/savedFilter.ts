import AbstractModel from './abstractModel'
import UserModel from '@/models/user'

import type {ISavedFilter} from '@/modelTypes/ISavedFilter'
import type {IUser} from '@/modelTypes/IUser'

export default class SavedFilterModel extends AbstractModel<ISavedFilter> implements ISavedFilter {
	id = 0
	title = ''
	description = ''
	filters: ISavedFilter['filters'] = {
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

	constructor(data: Partial<ISavedFilter> = {}) {
		super()
		this.assignData(data)

		this.owner = new UserModel(this.owner)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
