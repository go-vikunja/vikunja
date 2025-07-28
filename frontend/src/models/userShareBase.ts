import AbstractModel from './abstractModel'

import {RIGHTS, type Right} from '@/constants/rights'
import type {IUserShareBase} from '@/modelTypes/IUserShareBase'
import type {IUser} from '@/modelTypes/IUser'

export default class UserShareBaseModel extends AbstractModel<IUserShareBase> implements IUserShareBase {
	username: IUser['username'] = ''
	right: Right = RIGHTS.READ

	created: Date = null
	updated: Date = null

	constructor(data: Partial<IUserShareBase>) {
		super()
		this.assignData(data)
	
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
