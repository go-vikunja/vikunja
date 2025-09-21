import AbstractModel from './abstractModel'

import {PERMISSIONS, type Permission} from '@/constants/permissions'
import type {IUserShareBase} from '@/modelTypes/IUserShareBase'
import type {IUser} from '@/modelTypes/IUser'

export default class UserShareBaseModel extends AbstractModel<IUserShareBase> implements IUserShareBase {
	username: IUser['username'] = ''
	permission: Permission = PERMISSIONS.READ

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<IUserShareBase>) {
		super()
		this.assignData(data)
	
		this.created = this.created ? new Date(this.created) : new Date()
		this.updated = this.updated ? new Date(this.updated) : new Date()
	}
}
