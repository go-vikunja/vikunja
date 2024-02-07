import AbstractModel from './abstractModel'
import UserModel from '@/models/user'

import type {ISubscription} from '@/modelTypes/ISubscription'
import type {IUser} from '@/modelTypes/IUser'

export default class SubscriptionModel extends AbstractModel<ISubscription> implements ISubscription {
	id = 0
	entity = ''
	entityId = 0
	user: IUser = {}

	created: Date = null

	constructor(data : Partial<ISubscription>) {
		super()
		this.assignData(data)

		this.created = new Date(this.created)
		this.user = new UserModel(this.user)
	}
}
