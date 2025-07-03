import AbstractModel from './abstractModel'
import UserModel from './user'

import {RIGHTS, type Right} from '@/constants/rights'
import type {ILinkShare} from '@/modelTypes/ILinkShare'
import type {IUser} from '@/modelTypes/IUser'

export default class LinkShareModel extends AbstractModel<ILinkShare> implements ILinkShare {
	id = 0
	hash = ''
	right: Right = RIGHTS.READ
	sharedBy: IUser = new UserModel({}) as IUser
	sharingType = 0 // FIXME: use correct numbers
	projectId = 0
	name = ''
	password = ''
	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ILinkShare>) {
		super()
		this.assignData(data)

		if (this.sharedBy) this.sharedBy = new UserModel(this.sharedBy) as IUser

		if (this.created) this.created = new Date(this.created)
		if (this.updated) this.updated = new Date(this.updated)
	}
}
