import AbstractModel from '@/models/abstractModel'
import UserModel from '@/models/user'

export default class SavedFilterModel extends AbstractModel {
	constructor(data) {
		super(data)

		this.owner = new UserModel(this.owner)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			title: '',
			description: '',
			filters: {
				sortBy: ['done', 'id'],
				orderBy: ['asc', 'desc'],
				filterBy: ['done'],
				filterValue: ['false'],
				filterComparator: ['equals'],
				filterConcat: 'and',
				filterIncludeNulls: true,
			},

			owner: {},
			created: null,
			updated: null,
		}
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
