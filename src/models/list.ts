import AbstractModel from './abstractModel'
import TaskModel from './task'
import UserModel from './user'
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'
import SubscriptionModel from '@/models/subscription'

export default class ListModel extends AbstractModel {

	constructor(data) {
		super(data)

		this.owner = new UserModel(this.owner)

		/** @type {number} */
		this.id

		/** @type {string} */
		this.title

		/** @type {string} */
		this.description
		
		/** @type {UserModel} */
		this.owner

		/** @type {TaskModel[]} */
		this.tasks

		// Make all tasks to task models
		this.tasks = this.tasks.map(t => {
			return new TaskModel(t)
		})
		
		/** @type {number} */
		this.namespaceId

		/** @type {boolean} */
		this.isArchived

		/** @type {string} */
		this.hexColor

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		/** @type {string} */
		this.identifier

		/** @type */
		this.backgroundInformation

		/** @type {boolean} */
		this.isFavorite

		/** @type */
		this.subscription

		if (typeof this.subscription !== 'undefined' && this.subscription !== null) {
			this.subscription = new SubscriptionModel(this.subscription)
		}

		/** @type {number} */
		this.position

		/** @type {Date} */
		this.created = new Date(this.created)

		/** @type {Date} */
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
			subscription: null,
			position: 0,
			backgroundBlurHash: '',

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