import { objectToSnakeCase } from '@/helpers/case'
import AbstractModel from './abstractModel'
import UserModel from '@/models/user'

import type {ISavedFilter} from '@/modelTypes/ISavedFilter'
import type {IUser} from '@/modelTypes/IUser'

export default class SavedFilterModel extends AbstractModel<ISavedFilter> implements ISavedFilter {
	id = 0
	title = ''
	description = ''
	filters: ISavedFilter['filters'] = {
		sort_by: ['done', 'id'],
		order_by: ['asc', 'desc'],
		filter: 'done = false',
		filter_include_nulls: true,
		s: '',
	}

	owner: IUser = new UserModel()
	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ISavedFilter> = {}) {
		super()
		this.assignData(data)

		this.owner = new UserModel(this.owner)

		// Filters are in snake_case for the API - this makes it consistent with the way filter params are used with one-off filters.
		// Should probably be camelCase everywhere, but that's a task for another day.
		this.filters = objectToSnakeCase(this.filters as unknown as Record<string, unknown>) as unknown as ISavedFilter['filters']

		this.created = new Date(this.created || Date.now())
		this.updated = new Date(this.updated || Date.now())
	}
}
