import AbstractModel from './abstractModel'
import UserModel from './user'

import {PERMISSIONS, type Permission} from '@/constants/permissions'
import type {ILinkShare} from '@/modelTypes/ILinkShare'
import type {IUser} from '@/modelTypes/IUser'

export default class LinkShareModel extends AbstractModel<ILinkShare> implements ILinkShare {
	id = 0
	hash = ''
	permission: Permission = PERMISSIONS.READ
	sharedBy: IUser = UserModel
	sharingType = 0 // FIXME: use correct numbers
	projectId = 0
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
