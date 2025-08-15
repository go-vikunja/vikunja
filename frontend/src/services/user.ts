import AbstractService from './abstractService'
import UserModel from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

export default class UserService extends AbstractService<IUser> {
	constructor() {
		super({
			getAll: 'users',
			get: 'users/{id}',
			update: 'users/{id}',
		})
	}

	modelFactory(data) {
		return new UserModel(data)
	}
}
