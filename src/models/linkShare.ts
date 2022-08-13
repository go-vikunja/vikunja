import AbstractModel from './abstractModel'
import UserModel, { type IUser } from './user'
import {RIGHTS, type Right} from '@/constants/rights'

export interface ILinkShare extends AbstractModel {
	id: number
	hash: string
	right: Right
	sharedBy: IUser
	sharingType: number // FIXME: use correct numbers
	listId: number
	name: string
	password: string
	created: Date
	updated: Date
}

export default class LinkShareModel extends AbstractModel implements ILinkShare {
	id!: number
	hash!: string
	right!: Right
	sharedBy: IUser
	sharingType!: number // FIXME: use correct numbers
	listId!: number
	name!: string
	password!: string
	created: Date
	updated: Date

	constructor(data) {
		// The constructor of AbstractModel handles all the default parsing.
		super(data)

		this.sharedBy = new UserModel(this.sharedBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	// Default attributes that define the "empty" state.
	defaults() {
		return {
			id: 0,
			hash: '',
			right: RIGHTS.READ,
			sharedBy: UserModel,
			sharingType: 0,
			listId: 0,
			name: '',
			password: '',

			created: null,
			updated: null,
		}
	}
}