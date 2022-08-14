import AbstractModel, { type IAbstract } from './abstractModel'
import UserModel, { type IUser } from './user'
import {RIGHTS, type Right} from '@/constants/rights'

export interface ILinkShare extends IAbstract {
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
	id = 0
	hash = ''
	right: Right = RIGHTS.READ
	sharedBy: IUser = UserModel
	sharingType = 0 // FIXME: use correct numbers
	listId = 0
	name: ''
	password: ''
	created: Date = null
	updated: Date = null

	constructor(data: Partial<ILinkShare>) {
		super()
		this.assignData(data)

		this.sharedBy = new UserModel(this.sharedBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}