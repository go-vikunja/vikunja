import AbstractService from './abstractService'
import UserModel from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

export default class UserService extends AbstractService<IUser> {
	constructor() {
		super({
			getAll: '/api/v1/users',
		})
	}

	modelFactory(data) {
		return new UserModel(data)
	}
}
