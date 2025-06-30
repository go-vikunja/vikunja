import AbstractModel from './abstractModel'

import {RIGHTS, type Right} from '@/constants/rights'
import type {IUserShareBase} from '@/modelTypes/IUserShareBase'
import type {IUser} from '@/modelTypes/IUser'

export default class UserShareBaseModel extends AbstractModel<IUserShareBase> implements IUserShareBase {
	userId: IUser['id'] = 0
	right: Right = RIGHTS.READ

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<IUserShareBase>) {
		super()
		this.assignData(data)
	
		this.created = new Date(this.created || Date.now())
		this.updated = new Date(this.updated || Date.now())
	}
}
