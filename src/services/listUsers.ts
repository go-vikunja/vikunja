import AbstractService from './abstractService'
import UserModel from '../models/user'

export default class ListUserService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listId}/listusers',
		})
	}

	modelFactory(data) {
		return new UserModel(data)
	}
}