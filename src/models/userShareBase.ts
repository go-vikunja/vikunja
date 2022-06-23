import AbstractModel from './abstractModel'
import type UserModel from './user'
import type {Right} from '@/models/constants/rights'

export default class UserShareBaseModel extends AbstractModel {
	userId: UserModel['id']
	right: Right

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
			right: 0,

			created: null,
			updated: null,
		}
	}
}