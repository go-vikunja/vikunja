import AbstractService from './abstractService'
import UserListModel from '../models/userList'
import UserModel from '../models/user'
import {formatISO} from 'date-fns'

export default class UserListService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listId}/users',
			getAll: '/lists/{listId}/users',
			update: '/lists/{listId}/users/{userId}',
			delete: '/lists/{listId}/users/{userId}',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new UserListModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}