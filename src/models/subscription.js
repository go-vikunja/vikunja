import AbstractModel from '@/models/abstractModel'
import UserModel from '@/models/user'

export default class SubscriptionModel extends AbstractModel {
	id = 0
	entity = ''
	entityId = 0
	created = null
	user = UserModel

	constructor(data) {
		super(data)
		this.user = new UserModel(this.user)
		this.created = new Date(this.created)
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
