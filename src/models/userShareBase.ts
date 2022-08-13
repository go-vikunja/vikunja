import AbstractModel from './abstractModel'
import {RIGHTS, type Right} from '@/constants/rights'
import type { IUser } from './user'

export interface IUserShareBase extends AbstractModel {
	userId: IUser['id']
	right: Right

	created: Date
	updated: Date
}

export default class UserShareBaseModel extends AbstractModel implements IUserShareBase {
	declare userId: IUser['id']
	declare right: Right

	created: Date
	updated: Date

	constructor(data) {
		super(data)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			userId: '',
			right: RIGHTS.READ,

			created: null,
			updated: null,
		}
	}
}