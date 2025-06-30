import AbstractService from './abstractService'
import UserModel from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

export default class UserService extends AbstractService<IUser> {
	constructor() {
		super({
			getAll: '/users',
		})
	}

	modelFactory(data: any): IUser {
		return new UserModel(data) as IUser
	}
}
