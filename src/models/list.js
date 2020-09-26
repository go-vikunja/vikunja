import AbstractModel from './abstractModel'
import TaskModel from './task'
import UserModel from './user'
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'

export default class ListModel extends AbstractModel {

	constructor(data) {
		super(data)

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		// Make all tasks to task models
		this.tasks = this.tasks.map(t => {
			return new TaskModel(t)
		})

		this.owner = new UserModel(this.owner)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	// Default attributes that define the "empty" state.
	defaults() {
		return {
			id: 0,
			title: '',
			description: '',
			owner: UserModel,
			tasks: [],
			namespaceId: 0,
			isArchived: false,
			hexColor: '',
			identifier: '',
			backgroundInformation: null,
			isFavorite: false,

			created: null,
			updated: null,
		}
	}

	isSavedFilter() {
		return this.getSavedFilterId() > 0
	}

	getSavedFilterId() {
		return getSavedFilterIdFromListId(this.id)
	}
}