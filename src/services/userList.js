import AbstractService from './abstractService'
import UserListModel from '../models/userList'
import UserModel from '../models/user'
import {formatISO} from 'date-fns'

export default class UserListService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listID}/users',
			getAll: '/lists/{listID}/users',
			update: '/lists/{listID}/users/{userID}',
			delete: '/lists/{listID}/users/{userID}',
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