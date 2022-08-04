import AbstractModel from './abstractModel'
import ListModel from './list'
import UserModel from './user'
import SubscriptionModel from '@/models/subscription'

import type {INamespace} from '@/modelTypes/INamespace'
import type {IUser} from '@/modelTypes/IUser'
import type {IList} from '@/modelTypes/IList'
import type {ISubscription} from '@/modelTypes/ISubscription'

export default class NamespaceModel extends AbstractModel implements INamespace {
	id = 0
	title = ''
	description = ''
	owner: IUser = UserModel
	lists: IList[] = []
	isArchived = false
	hexColor = ''
	subscription: ISubscription = null

	created: Date = null
	updated: Date = null

	constructor(data: Partial<INamespace>) {
		super()
		this.assignData(data)

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		this.lists = this.lists.map(l => {
			return new ListModel(l)
		})

		this.owner = new UserModel(this.owner)

		if(typeof this.subscription !== 'undefined' && this.subscription !== null) {
			this.subscription = new SubscriptionModel(this.subscription)
		}

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
