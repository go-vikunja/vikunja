import AbstractService from './abstractService'
import UserListModel from '../models/userList'
import UserModel from '../models/user'
import moment from 'moment'

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
		model.created = moment(model.created).toISOString()
		model.updated = moment(model.updated).toISOString()
		return model
	}

	modelFactory(data) {
		return new UserListModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}