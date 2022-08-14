import AbstractModel, { type IAbstract } from '@/models/abstractModel'
import UserModel, { type IUser } from '@/models/user'

export interface ISubscription extends IAbstract {
	id: number
	entity: string // FIXME: correct type?
	entityId: number // FIXME: correct type?
	user: IUser

	created: Date
}

export default class SubscriptionModel extends AbstractModel implements ISubscription {
	id!: number
	entity!: string // FIXME: correct type?
	entityId!: number // FIXME: correct type?
	user: IUser

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
