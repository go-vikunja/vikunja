import AbstractService from './abstractService'
import UserModel from '../models/user'

export default class UserService extends AbstractService {
	constructor() {
		super({
			getAll: '/users'
		})
	}
	
	modelFactory(data) {
		return new UserModel(data)
	}
}