import AbstractModel, { type IAbstract } from './abstractModel'
import {RIGHTS, type Right} from '@/constants/rights'
import type { IUser } from './user'

export interface IUserShareBase extends IAbstract {
	userId: IUser['id']
	right: Right

	created: Date
	updated: Date
}

export default class UserShareBaseModel extends AbstractModel implements IUserShareBase {
	userId: IUser['id'] = ''
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