import AbstractModel from '@/models/abstractModel'
import UserModel from '@/models/user'

export default class SubscriptionModel extends AbstractModel {
	id: number
	entity: string // FIXME: correct type?
	entityId: number // FIXME: correct type?
	user: UserModel

	created: Date

	constructor(data) {
		super(data)

		this.created = new Date(this.created)
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
