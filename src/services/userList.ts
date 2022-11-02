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

	modelFactory(data) {
		return new UserListModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}