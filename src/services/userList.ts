import {formatISO} from 'date-fns'

import AbstractService from './abstractService'
import UserListModel from '@/models/userList'
import type {IUserList} from '@/modelTypes/IUserList'
import UserModel from '@/models/user'

export default class UserListService extends AbstractService<IUserList> {
	constructor() {
		super({
			create: '/lists/{listId}/users',
			getAll: '/lists/{listId}/users',
			update: '/lists/{listId}/users/{userId}',
			delete: '/lists/{listId}/users/{userId}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new UserListModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}