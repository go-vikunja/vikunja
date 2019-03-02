import AbstractService from './abstractService'
import UserNamespaceModel from '../models/userNamespace'
import UserModel from '../models/user'

export default class UserNamespaceService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceID}/users',
			getAll: '/namespaces/{namespaceID}/users',
			update: '/namespaces/{namespaceID}/users/{userID}',
			delete: '/namespaces/{namespaceID}/users/{userID}',
		})
	}

	modelFactory(data) {
		return new UserNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}