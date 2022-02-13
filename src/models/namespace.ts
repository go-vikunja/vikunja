import AbstractModel from './abstractModel'
import ListModel from './list'
import UserModel from './user'
import SubscriptionModel from '@/models/subscription'

export default class NamespaceModel extends AbstractModel {
	constructor(data) {
		super(data)

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		/** @type {ListModel[]} */
		this.lists = this.lists.map(l => {
			return new ListModel(l)
		})

		this.owner = new UserModel(this.owner)

		if(typeof this.subscription !== 'undefined' && this.subscription !== null) {
			this.subscription = new SubscriptionModel(this.subscription)
		}

		/** @type {number} */
		this.id

		/** @type {string} */
		this.title

		/** @type {boolean} */
		this.isArchived

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	// Default attributes that define the 'empty' state.
	defaults() {
		return {
			id: 0,
			title: '',
			description: '',
			owner: UserModel,
			lists: [],
			isArchived: false,
			hexColor: '',
			subscription: null,

			created: null,
			updated: null,
		}
	}
}
