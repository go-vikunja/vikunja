import AbstractModel from '@/models/abstractModel'
import UserModel from '@/models/user'

export default class SubscriptionModel extends AbstractModel {
	constructor(data) {
		super(data)

		/** @type {number} */
		this.id

		/** @type {string} */
		this.entity

		/** @type {number} */
		this.entityId

		/** @type {Date} */
		this.created = new Date(this.created)

		/** @type {UserModel} */
		this.user = new UserModel(this.user)
	}

	defaults() {
		return {
			id: 0,
			entity: '',
			entityId: 0,
			created: null,
			user: {},
		}
	}
}
