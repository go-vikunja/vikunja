import AbstractModel, { type IAbstract } from './abstractModel'
import ListModel, { type IList } from './list'
import UserModel, { type IUser } from './user'
import SubscriptionModel, { type ISubscription } from '@/models/subscription'

export interface INamespace extends IAbstract {
	id: number
	title: string
	description: string
	owner: IUser
	lists: IList[]
	isArchived: boolean
	hexColor: string
	subscription: ISubscription

	created: Date
	updated: Date
}

export default class NamespaceModel extends AbstractModel implements INamespace {
	id!: number
	title!: string
	description!: string
	owner: IUser
	lists: IList[]
	isArchived!: boolean
	hexColor!: string
	subscription!: ISubscription

	created: Date
	updated: Date

	constructor(data) {
		super(data)

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
